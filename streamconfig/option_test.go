package streamconfig_test

import (
	"testing"

	"github.com/blendle/go-streamprocessor/streamconfig"
	"github.com/stretchr/testify/assert"
)

// optionFunc wraps a func so it satisfies the Option interface.
type optionFunc func(*streamconfig.Consumer, *streamconfig.Producer)

func (f optionFunc) Apply(c *streamconfig.Consumer, p *streamconfig.Producer) {
	f(c, p)
}

func TestConsumer_WithOptions(t *testing.T) {
	t.Parallel()

	consumer := &streamconfig.Consumer{}
	opt := optionFunc(func(c *streamconfig.Consumer, _ *streamconfig.Producer) {
		c.Name = "test"
	})

	consumer2 := consumer.WithOptions(opt)

	assert.Equal(t, "", consumer.Name)
	assert.Equal(t, "test", consumer2.Name)
}
