package natsclient

import (
	"sync"
	"time"

	"github.com/blendle/go-streamprocessor/stream"
	"github.com/blendle/go-streamprocessor/streamconfig"
	"github.com/blendle/go-streamprocessor/streammsg"
	"github.com/blendle/go-streamprocessor/streamutils"
	nats "github.com/nats-io/go-nats"
	"go.uber.org/zap"
)

// Producer implements the stream.Producer interface for the Kafka client.
type Producer struct {
	// c represents the configuration passed into the producer on
	// initialization.
	c streamconfig.Producer

	logger   *zap.Logger
	nats     *nats.Conn
	wg       sync.WaitGroup
	messages chan<- streammsg.Message
	quit     chan bool
	once     *sync.Once
}

var _ stream.Producer = (*Producer)(nil)

// NewProducer returns a new Kafka producer.
func NewProducer(options ...func(*streamconfig.Producer)) (stream.Producer, error) {
	ch := make(chan streammsg.Message)

	producer, err := newProducer(ch, options)
	if err != nil {
		return nil, err
	}

	// add one to the WaitGroup. We reduce this count once Close() is called, and
	// the messages channel is closed.
	producer.wg.Add(1)

	// We listen to the produce channel in a goroutine. Every message delivered to
	// this producer gets prepared for a Kafka deliver, and then added to a queue
	// of messages ready to be sent to Kafka. This queue is handled asynchronously
	// for us. If the producer is closed, the close is blocked until the queue is
	// emptied. If the queue can't be emptied, the close call returns an error.
	go producer.produce(ch)

	// Finally, we monitor for any interrupt signals. Ideally, the user handles
	// these cases gracefully, but just in case, we try to close the producer if
	// any such interrupt signal is intercepted. If closing the producer fails, we
	// exit 1, and log a fatal message explaining what happened.
	//
	// This functionality is enabled by default, but can be disabled through a
	// configuration flag.
	if producer.c.HandleInterrupt {
		go streamutils.HandleInterrupts(producer.Close, producer.logger)
	}

	return producer, nil
}

// Messages returns the write channel for messages to be produced.
func (p *Producer) Messages() chan<- streammsg.Message {
	return p.messages
}

// Close closes the producer connection. This function blocks until all messages
// still in the channel have been processed, and the channel is properly closed.
func (p *Producer) Close() (err error) {
	p.once.Do(func() {
		// Trigger the quit channel, which terminates our internal goroutine to
		// process messages, and closes the messages channel. We do this first, to
		// prevent sending any left-over messages to a closed NATS producer
		// channel.
		p.quit <- true

		// Wait until the WaitGroup counter is zero. This makes sure we block the
		// close call until the writer has been closed, to prevent reading errors.
		p.wg.Wait()

		// After we are guaranteed to no longer deliver any messages to the NATS
		// producer, we make sure we flush any messages still waiting to be
		// delivered to NATS. We allow for a reasonable amount of time to pass
		// before we abort the flush operation. If any messages are still not
		// delivered after the timeout expires, we return an error, indicating that
		// something went wrong.
		err = p.nats.FlushTimeout(5 * time.Second)
		if err != nil {
			return
		}

		// This synchronous call closes the NATS producer. There are no errors to
		// handle from this close call.
		p.nats.Close()

		// Let's flush all logs still in the buffer, since this producer is no
		// longer useful after this point.
		_ = p.logger.Sync() // nolint: gas
	})

	return err
}

// Config returns a read-only representation of the producer configuration.
func (p *Producer) Config() streamconfig.Producer {
	return p.c
}

func (p *Producer) produce(ch <-chan streammsg.Message) {
	defer func() {
		close(p.messages)
		p.wg.Done()
	}()

	for {
		select {
		case <-p.quit:
			p.logger.Info("Received quit signal. Exiting producer.")
			return

		case m := <-ch:
			msg := p.newMessage(m)

			// We're using the synchronous `Produce` function instead of the channel-
			// based `ProduceChannel` function, since we want to make sure the message
			// was delivered to the queue. Please note that this does _not_ mean that
			// we wait for the message to actually be delivered to Kafka. rdkafka uses
			// an internal queue and thread to periodically deliver messages to the
			// Kafka broker, and reports back delivery messages on a separate channel.
			// These delivery reports are handled by us in `p.checkReports()`.
			err := p.nats.PublishMsg(msg)
			if err != nil {
				p.logger.Fatal(
					"Error while delivering message to NATS.",
					zap.Error(err),
				)
			}
		}
	}
}

func newProducer(ch chan streammsg.Message, options []func(*streamconfig.Producer)) (*Producer, error) {
	// Construct a full configuration object, based on the provided configuration,
	// the default configurations, and the static configurations.
	config, err := streamconfig.NewProducer(options...)
	if err != nil {
		return nil, err
	}

	config.Logger.Info(
		"Finished parsing NATS client configurations.",
		zap.Any("config", config),
	)

	no := nats.Options{Servers: config.NATS.Servers}

	// Instantiate a new NATS producer.
	natsconsumer, err := no.Connect()
	if err != nil {
		return nil, err
	}

	producer := &Producer{
		c:        config,
		logger:   &config.Logger,
		nats:     natsconsumer,
		messages: ch,
		quit:     make(chan bool),
		once:     &sync.Once{},
	}

	return producer, nil
}

func (p *Producer) newMessage(m streammsg.Message) *nats.Msg {

	msg := &nats.Msg{
		Data:    m.Value,
		Subject: m.Topic,
		Sub:     &nats.Subscription{Subject: p.c.NATS.Topic},
	}

	return msg
}
