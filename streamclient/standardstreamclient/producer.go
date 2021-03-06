package standardstreamclient

import (
	"bytes"
	"os"
	"sync"

	"github.com/blendle/go-streamprocessor/stream"
	"github.com/blendle/go-streamprocessor/streamconfig"
	"github.com/blendle/go-streamprocessor/streamutil"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// producer implements the stream.Producer interface for the standard stream
// client.
type producer struct {
	// c represents the configuration passed into the producer on
	// initialization.
	c streamconfig.Producer

	logger   *zap.Logger
	wg       sync.WaitGroup
	errors   chan error
	messages chan<- stream.Message
	signals  chan os.Signal
	once     *sync.Once
}

var _ stream.Producer = (*producer)(nil)

// NewProducer returns a new standard stream producer.
func NewProducer(options ...streamconfig.Option) (stream.Producer, error) {
	ch := make(chan stream.Message)

	p, err := newProducer(ch, options)
	if err != nil {
		return nil, err
	}

	// add one to the WaitGroup. We remove this one only after all writes (below)
	// are completed and the write channel is closed.
	p.wg.Add(1)

	// We start a goroutine to listen for errors on the errors channel, and log a
	// fatal error (terminating the application in the process) when an error is
	// received.
	//
	// This functionality is enabled by default, but can be disabled through a
	// configuration flag. If the auto-error functionality is disabled, the user
	// needs to manually listen to the `Errors()` channel and act accordingly.
	if p.c.HandleErrors {
		go streamutil.HandleErrors(p.errors, p.logger.Fatal)
	}

	// We listen to the produce channel in a goroutine. For every message
	// delivered to this producer we add a newline (if missing), and send the
	// message value to the configured io.Writer. If the producer is closed, the
	// close is blocked until the channel is closed.
	go p.produce(ch)

	// Finally, we monitor for any interrupt signals. Ideally, the user handles
	// these cases gracefully, but just in case, we try to close the producer if
	// any such interrupt signal is intercepted. If closing the producer fails, we
	// exit 1, and log a fatal message explaining what happened.
	//
	// This functionality is enabled by default, but can be disabled through a
	// configuration flag.
	if p.c.HandleInterrupt {
		go streamutil.HandleInterrupts(p.signals, p.Close, p.logger)
	}

	return p, nil
}

// Messages returns the write channel for messages to be produced.
func (p *producer) Messages() chan<- stream.Message {
	return p.messages
}

// Errors returns the read channel for the errors that are returned by the
// stream.
func (p *producer) Errors() <-chan error {
	return streamutil.ErrorsChan(p.errors, p.c.HandleErrors)
}

// Close closes the producer connection. This function blocks until all messages
// still in the channel have been processed, and the channel is properly closed.
func (p *producer) Close() error {
	p.once.Do(func() {
		close(p.messages)

		// Wait until the WaitGroup counter is zero. This makes sure we block the
		// close call until all messages have been delivered, to prevent data-loss.
		p.wg.Wait()

		// At this point, no more errors are expected, so we can close the errors
		// channel.
		close(p.errors)

		// Let's flush all logs still in the buffer, since this producer is no
		// longer useful after this point. We ignore any errors returned by sync, as
		// it is known to return unexpected errors. See: https://git.io/vpJFk
		_ = p.logger.Sync() // nolint: gas

		// Finally, close the signals channel, as it's no longer needed
		close(p.signals)
	})

	return nil
}

// Config returns a read-only representation of the producer configuration as an
// interface. To access the underlying configuration struct, cast the interface
// to `streamconfig.Producer`.
func (p *producer) Config() interface{} {
	return p.c
}

func (p *producer) produce(ch <-chan stream.Message) {
	defer p.wg.Done()

	for msg := range ch {
		message := msg.Value

		// If the original message does not contain a newline at the end, we add
		// it, as this is used as the message delimiter.
		if !bytes.HasSuffix(message, []byte("\n")) {
			message = append(message, "\n"...)
		}

		_, err := p.c.Standardstream.Writer.Write(message)
		if err != nil {
			p.errors <- errors.Wrap(err, "unable to write message to stream")
		}
	}
}

func newProducer(ch chan stream.Message, options []streamconfig.Option) (*producer, error) {
	config, err := streamconfig.NewProducer(options...)
	if err != nil {
		return nil, err
	}

	p := &producer{
		c:        config,
		logger:   config.Logger,
		errors:   make(chan error),
		messages: ch,
		once:     &sync.Once{},
		signals:  make(chan os.Signal, 3),
	}

	return p, nil
}
