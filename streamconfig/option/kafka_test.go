package option_test

import (
	"testing"

	"github.com/blendle/go-streamprocessor/streamconfig"
	"github.com/blendle/go-streamprocessor/streamconfig/kafkaconfig"
	"github.com/blendle/go-streamprocessor/streamconfig/option"
	"github.com/stretchr/testify/assert"
)

func TestBroker(t *testing.T) {
	var tests = map[string]struct {
		in       string
		consumer kafkaconfig.Consumer
		producer kafkaconfig.Producer
	}{
		"both": {
			"test",
			kafkaconfig.Consumer{Brokers: []string{"test"}},
			kafkaconfig.Producer{Brokers: []string{"test"}},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := &streamconfig.Consumer{}
			p := &streamconfig.Producer{}

			o := option.KafkaBroker(tt.in)
			o.Apply(c, p)

			assert.EqualValues(t, tt.consumer, c.Kafka)
			assert.EqualValues(t, tt.producer, p.Kafka)
		})
	}
}
