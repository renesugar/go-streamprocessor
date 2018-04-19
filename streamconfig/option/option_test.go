package option_test

import (
	"fmt"
	"testing"

	"github.com/blendle/go-streamprocessor/streamconfig"
	"github.com/blendle/go-streamprocessor/streamconfig/option"
	"github.com/stretchr/testify/assert"
)

func TestOption(t *testing.T) {
	var tests = map[string]struct {
		opt      streamconfig.Option
		consumer streamconfig.Consumer
		producer streamconfig.Producer
	}{
		"both": {
			testName("hello world"),
			streamconfig.Consumer{Name: "hello world"},
			streamconfig.Producer{Name: "hello world"},
		},

		"consumer (one)": {
			option.Consumer(testName("hello world")),
			streamconfig.Consumer{Name: "hello world"},
			streamconfig.Producer{},
		},

		"consumer (multiple)": {
			option.Consumer(testName("hello world"), testHandleInterrupt(true)),
			streamconfig.Consumer{Name: "hello world", HandleInterrupt: true},
			streamconfig.Producer{},
		},

		"producer (one)": {
			option.Producer(testName("hello world")),
			streamconfig.Consumer{},
			streamconfig.Producer{Name: "hello world"},
		},

		"producer (multiple)": {
			option.Producer(testName("hello world"), testHandleInterrupt(true)),
			streamconfig.Consumer{},
			streamconfig.Producer{Name: "hello world", HandleInterrupt: true},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := &streamconfig.Consumer{}
			p := &streamconfig.Producer{}

			tt.opt.Apply(c, p)

			fmt.Printf("%#v\n", c)

			assert.EqualValues(t, &tt.consumer, c)
			assert.EqualValues(t, &tt.producer, p)
		})
	}
}

// optionFunc wraps a func so it satisfies the Option interface.
type optionFunc func(*streamconfig.Consumer, *streamconfig.Producer)

func (f optionFunc) Apply(c *streamconfig.Consumer, p *streamconfig.Producer) {
	f(c, p)
}

func testName(s string) streamconfig.Option {
	return optionFunc(func(c *streamconfig.Consumer, p *streamconfig.Producer) {
		c.Name = s
		p.Name = s
	})
}

func testHandleInterrupt(b bool) streamconfig.Option {
	return optionFunc(func(c *streamconfig.Consumer, p *streamconfig.Producer) {
		c.HandleInterrupt = b
		p.HandleInterrupt = b
	})
}
