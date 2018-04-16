package natsclient

import (
	"sync"

	"github.com/blendle/go-streamprocessor/stream"
	"github.com/blendle/go-streamprocessor/streamconfig"
	"github.com/blendle/go-streamprocessor/streammsg"
	"github.com/blendle/go-streamprocessor/streamutils"
	"github.com/nats-io/go-nats"
	"go.uber.org/zap"
)

// Consumer implements the stream.Consumer interface for the NATS client.
type Consumer struct {
	// c represents the configuration passed into the consumer on initialization.
	c streamconfig.Consumer

	logger   *zap.Logger
	messages chan streammsg.Message
	nats     *nats.Conn
	once     *sync.Once
	quit     chan bool
	recvCh   chan *nats.Msg
	sub      *nats.Subscription
	wg       sync.WaitGroup
}

var _ stream.Consumer = (*Consumer)(nil)

// NewConsumer returns a new NATS consumer.
func NewConsumer(options ...func(*streamconfig.Consumer)) (stream.Consumer, error) {
	consumer, err := newConsumer(options)
	if err != nil {
		return nil, err
	}

	// add one to the WaitGroup. We reduce this count once Close() is called, and
	// the messages channel is closed.
	consumer.wg.Add(1)

	// We start a goroutine to consume any messages being delivered to us from
	// NATS. We deliver these messages on a blocking channel, so as long as no one
	// is listening on the other end of the channel, there's no significant
	// overhead to starting the goroutine this early.
	go consumer.consume()

	// Finally, we monitor for any interrupt signals. Ideally, the user handles
	// these cases gracefully, but just in case, we try to close the consumer if
	// any such interrupt signal is intercepted. If closing the consumer fails, we
	// exit 1, and log a fatal message explaining what happened.
	//
	// This functionality is enabled by default, but can be disabled through a
	// configuration flag.
	if consumer.c.HandleInterrupt {
		go streamutils.HandleInterrupts(consumer.Close, consumer.logger)
	}

	return consumer, nil
}

// Messages returns the read channel for the messages that are returned by the
// stream.
func (c *Consumer) Messages() <-chan streammsg.Message {
	return c.messages
}

// Ack is a no-op implementation to satisfy the `stream.Consumer` interface.
// NATS has no support for message acknowledgments (I believe?).
func (c *Consumer) Ack(m streammsg.Message) error {
	return nil
}

// Nack is a no-op implementation to satisfy the `stream.Consumer` interface.
// NATS has no support for message acknowledgments (I believe?).
func (c *Consumer) Nack(m streammsg.Message) error {
	return nil
}

// Close closes the consumer connection.
func (c *Consumer) Close() (err error) {
	c.once.Do(func() {
		// TODO: write why we need this, shouldn't `c.nats.Close()` already do this
		//       for us?
		err = c.sub.Unsubscribe()
		if err != nil {
			return
		}

		// This synchronous call closes the Kafka consumer and also sends any
		// still-to-be-committed offsets to the Broker before returning. This is
		// done first, so that no new messages are delivered to us, before we close
		// our own channel.
		c.nats.Close()

		// Trigger the quit channel, which terminates our internal goroutine to
		// process messages, and closes the messages channel.
		c.quit <- true

		// Wait until the WaitGroup counter is zero. This makes sure we block the
		// close call until the reader has been closed, to prevent an application
		// from quiting before we are fully done with all the clean-up.
		c.wg.Wait()

		// Let's flush all logs still in the buffer, since this consumer is no
		// longer useful after this point.
		_ = c.logger.Sync() // nolint: gas
	})

	return err
}

// Config returns a read-only representation of the consumer configuration.
func (c *Consumer) Config() streamconfig.Consumer {
	return c.c
}

func (c *Consumer) consume() {
	defer func() {
		close(c.messages)
		c.wg.Done()
	}()

	c.nats.SetErrorHandler(func(_ *nats.Conn, s *nats.Subscription, err error) {
		c.c.Logger.Fatal(
			"Received error from NATS while consuming messsages.",
			zap.String("subject", s.Subject),
			zap.String("queue", s.Queue),
			zap.Error(err),
		)
	})

	for {
		select {
		case <-c.quit:
			c.logger.Info("Received quit signal. Exiting consumer.")

			return
		case msg := <-c.recvCh:
			c.messages <- *newMessageFromNATS(msg)
		}
	}

	// c.nats.ChanQueueSubscribe(subj string, group string, ch chan *nats.Msg)
	// c.nats.QueueSubscribeSyncWithChan(subj string, queue string, ch chan *nats.Msg)

	// sub, err := c.nats.QueueSubscribe(c.c.NATS.Topic, c.c.NATS.QueueID, func(msg *nats.Msg) {
	// 	c.messages <- *newMessageFromNATS(msg)
	// })

	// if err != nil && err != nats.ErrConnectionClosed {
	// 	c.logger.Fatal("Received error from NATS.", zap.Error(err))
	// }

	// <-c.quit

	// c.logger.Info("Received quit signal. Exiting consumer.")
}

func newConsumer(options []func(*streamconfig.Consumer)) (*Consumer, error) {
	// Construct a full configuration object, based on the provided configuration,
	// the default configurations, and the static configurations.
	config, err := streamconfig.NewConsumer(options...)
	if err != nil {
		return nil, err
	}

	// We create a buffered channel to be able to accept incoming messages
	// _before_ the consumer's `Messages()` channel is actually being read. This
	// is required because NATS does not allow blocking reads, and will instantly
	// trigger an `ErrSlowConsumer` error. Having an internal buffer can lead to
	// lost messages if the consumer is closed before the buffer is emptied, but
	// since this is a non-durable consumer, that makes no difference from simply
	// closing the consumer and not being able to retrieve any messages produced
	// on the configured subject after reconnecting a few minutes later.
	ch := make(chan *nats.Msg, config.NATS.BufferSize)

	config.Logger.Info(
		"Finished parsing NATS client configurations.",
		zap.Any("config", config),
	)

	no := nats.Options{
		Servers:    config.NATS.Servers,
		SubChanLen: config.NATS.BufferSize,
	}

	// Instantiate a new NATS consumer.
	natsconsumer, err := no.Connect()
	if err != nil {
		return nil, err
	}

	sub, err := natsconsumer.ChanQueueSubscribe(config.NATS.Topic, config.NATS.QueueID, ch)
	if err != nil {
		return nil, err
	}

	// TODO: should this be configurable?
	//
	// FIXME: can't be used for channel based subscriptions, do we need it?
	// err = sub.SetPendingLimits(config.NATS.BufferSize, config.NATS.BufferSize*1024)
	// if err != nil {
	// 	return nil, err
	// }

	consumer := &Consumer{
		c:        config,
		logger:   &config.Logger,
		messages: make(chan streammsg.Message),
		nats:     natsconsumer,
		once:     &sync.Once{},
		quit:     make(chan bool),
		recvCh:   ch,
		sub:      sub,
	}

	return consumer, nil
}

func newMessageFromNATS(m *nats.Msg) *streammsg.Message {
	msg := &streammsg.Message{
		Value: m.Data,
		Topic: m.Subject,
	}

	return msg
}
