package option

import (
	"github.com/blendle/go-streamprocessor/streamconfig"
)

// ClientID sets `ClientID`.
func ClientID(id string) streamconfig.Option {
	return optionFunc(func(c *streamconfig.Consumer, p *streamconfig.Producer) {
		c.Kafka.ID = id
		p.Kafka.ID = id
	})
}
