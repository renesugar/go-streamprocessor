package option

import (
	"github.com/blendle/go-streamprocessor/streamclient"
	"github.com/blendle/go-streamprocessor/streamconfig"
)

// optionFunc wraps a func so it satisfies the Option interface.
type optionFunc func(*streamclient.Config)

func (f optionFunc) Apply(c *streamclient.Config) {
	if c == nil {
		return
	}

	f(c)
}

// Consumer takes a list of options, and wraps them into a new Option interface,
// which ignores the provided producer, and only configures the provided
// consumer based on the provided options.
func Consumer(opts ...streamconfig.Option) streamconfig.Option {
	return optionFunc(func(c *streamconfig.Consumer, _ *streamconfig.Producer) {
		// create producer no-op
		p, _ := streamconfig.NewProducer() // nolint: gas

		for i := range opts {
			opts[i].Apply(c, &p)
		}
	})
}

// Producer takes a list of options, and wraps them into a new Option interface,
// which ignores the provided consumer, and only configures the provided
// producer based on the provided options.
func Producer(opts ...streamconfig.Option) streamconfig.Option {
	return optionFunc(func(_ *streamconfig.Consumer, p *streamconfig.Producer) {
		// create consumer no-op
		c, _ := streamconfig.NewConsumer() // nolint: gas

		for i := range opts {
			opts[i].Apply(&c, p)
		}
	})
}
