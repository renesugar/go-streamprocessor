package inmemclient

import (
	"sync"

	"github.com/blendle/go-streamprocessor/stream"
	"github.com/blendle/go-streamprocessor/streamconfig"
	"github.com/blendle/go-streamprocessor/streamutils"
)

// Client implements the stream.Client interface for the inmemclient package.
type Client struct {
	chConsume chan stream.Message
	chError   chan error
	chProduce chan stream.Message

	onceClose   sync.Once
	onceConsume sync.Once
	onceProduce sync.Once

	wgConsume sync.WaitGroup
	wgProduce sync.WaitGroup

	config *streamconfig.Client
	quit   chan struct{}
}

var _ stream.Client = (*Client)(nil)

// New instantiates a new inmemclient.
func New(opts ...streamconfig.Option) (*Client, error) {
	c := &Client{
		chConsume: make(chan stream.Message),
		chError:   make(chan error),
		chProduce: make(chan stream.Message),

		onceClose:   sync.Once{},
		onceConsume: sync.Once{},
		onceProduce: sync.Once{},

		wgConsume: sync.WaitGroup{},
		wgProduce: sync.WaitGroup{},

		config: streamconfig.New(opts...),

		// The quit channel is a buffered channel, accepting a single message in its
		// buffer. This is done to allow calling `Close()` – which sends a message
		// to this quit channel to signal a running consumer to close – to function,
		// even if no consumer was started, and thus nothing is listening to this
		// channel.
		quit: make(chan struct{}, 1),
	}

	cc := Config{}

	cc.ConsumeOnce = true

	// We monitor for any interrupt signals. Ideally, the user handles these cases
	// gracefully, but just in case, we try to close the consumer if any such
	// interrupt signal is intercepted. If closing the consumer fails, we exit 1,
	// and log a fatal message explaining what happened.
	//
	// This functionality is enabled by default, but can be disabled through a
	// configuration flag.
	if c.config.HandleInterrupt {
		go streamutils.HandleInterrupts(c.Close, c.config.Logger)
	}

	return c, nil
}

// Consume returns the read channel for the messages that are returned while
// consuming from the stream.
func (c *Client) Consume() <-chan stream.Message {
	// The first time `Consume()` is called, we start a goroutine that ingests any
	// messages from the configured consumer.
	c.onceConsume.Do(func() {
		// Add one to the WaitGroup. We remove this one only after the consumer
		// channel is closed.
		c.wgConsume.Add(1)

		// We start a goroutine to consume any messages currently stored in the
		// inmem storage. We deliver these messages on a blocking channel.
		go c.consume()
	})

	return c.chConsume
}

// Produce returns the write channel for messages to be produced.
func (c *Client) Produce() chan<- stream.Message {
	// The first time `Produce` is called, we start a goroutine that waits for any
	// messages to be produced, and delivers them as expected.
	c.onceProduce.Do(func() {
		go c.produce(c.chProduce)
	})

	return c.chProduce
}

// Error returns the read channel for any non-consumer errors occurred during
// message processing.
func (c *Client) Error() <-chan error {
	return c.chError
}

// Ack is a no-op implementation to satisfy the stream.Client interface.
func (c *Client) Ack(_ stream.Message) error { return nil }

// Nack is a no-op implementation to satisfy the stream.Client interface.
func (c *Client) Nack(_ stream.Message) error { return nil }

// Close closes the client.
func (c *Client) Close() (err error) {
	c.onceClose.Do(func() {
		if !c.config.Inmem.ConsumeOnce {
			// Trigger the quit channel, which terminates our internal goroutine to
			// process messages, and closes the messages channel (if one is running).
			c.quit <- struct{}{}
		}

		// Wait until the WaitGroup counter is zero. This makes sure we block the
		// close call until the reader has been closed, to prevent reading errors.
		c.wgConsume.Wait()

		// Next, close the producer (if it was ever started).
		close(c.chProduce)

		// Wait until the WaitGroup counter is zero. This makes sure we block the
		// close call until all messages have been delivered, to prevent data-loss.
		c.wgProduce.Wait()

		// Let's flush all logs still in the buffer, since this client is no longer
		// useful after this point. We ignore any errors returned by sync, as it is
		// known to return unexpected errors. See: https://git.io/vpJFk
		_ = c.config.Logger.Sync() // nolint: gas
	})

	return err
}

// Config returns a read-only representation of the consumer configuration.
func (c *Client) Config() streamconfig.Client {
	return streamconfig.Client{}
}
