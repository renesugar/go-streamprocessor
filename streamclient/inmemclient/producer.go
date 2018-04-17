package inmemclient

import (
	"sync"

	"github.com/blendle/go-streamprocessor/stream"
	"github.com/blendle/go-streamprocessor/streamconfig"
	"github.com/blendle/go-streamprocessor/streammsg"
	"github.com/blendle/go-streamprocessor/streamutils"
	"go.uber.org/zap"
)

// Producer implements the stream.Producer interface for the inmem client.
type Producer struct {
	// c represents the configuration passed into the producer on
	// initialization.
	c streamconfig.Producer

	logger   *zap.Logger
	wg       sync.WaitGroup
	messages chan<- streammsg.Message
	once     *sync.Once
}

var _ stream.Producer = (*Producer)(nil)

// NewProducer returns a new inmem producer.
func NewProducer(options ...func(*streamconfig.Producer)) (stream.Producer, error) {
	ch := make(chan streammsg.Message)

	producer, err := newProducer(ch, options)
	if err != nil {
		return nil, err
	}

	// add one to the WaitGroup. We remove this one only after all writes (below)
	// are completed and the write channel is closed.
	producer.wg.Add(1)

	// We listen to the produce channel in a goroutine. Every message delivered to
	// this producer gets stored in the inmem store. If the producer is closed,
	// the close is blocked until the channel is closed.
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
func (p *Producer) Close() error {
	p.once.Do(func() {
		close(p.messages)

		// Wait until the WaitGroup counter is zero. This makes sure we block the
		// close call until all messages have been delivered, to prevent data-loss.
		p.wg.Wait()
	})

	return nil
}

// Config returns a read-only representation of the producer configuration.
func (p *Producer) Config() streamconfig.Producer {
	return p.c
}

func (p *Producer) produce(ch <-chan streammsg.Message) {
	defer p.wg.Done()

	for msg := range ch {
		p.c.Inmem.Store.Add(msg)
	}
}

func newProducer(ch chan streammsg.Message, options []func(*streamconfig.Producer)) (*Producer, error) {
	config, err := streamconfig.NewProducer(options...)
	if err != nil {
		return nil, err
	}

	producer := &Producer{
		c:        config,
		logger:   &config.Logger,
		messages: ch,
		once:     &sync.Once{},
	}

	return producer, nil
}