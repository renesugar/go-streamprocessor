package streamconfig

import (
	"github.com/blendle/go-streamprocessor/streamconfig/inmemconfig"
	"github.com/blendle/go-streamprocessor/streamconfig/kafkaconfig"
	"github.com/blendle/go-streamprocessor/streamconfig/pubsubconfig"
	"github.com/blendle/go-streamprocessor/streamconfig/standardstreamconfig"
)

// NewClient returns a new Client configuration struct, containing the values
// passed into the function. If any error occurs during configuration
// validation, an error is returned as the second argument.
func NewClient(options ...func(*Client)) (Client, error) {
	config := &Client{
		Inmem:          inmemconfig.ClientDefaults(),
		Kafka:          kafkaconfig.ClientDefaults(),
		Pubsub:         pubsubconfig.ClientDefaults(),
		Standardstream: standardstreamconfig.ClientDefaults(),
	}

	for _, option := range options {
		option(config)
	}

	// We pass the config by-value, to prevent any race-condition where the
	// original config struct is modified after the fact.
	return *config, nil
}

// NewConsumer returns a new Consumer configuration struct, containing the
// values passed into the function. If any error occurs during configuration
// validation, an error is returned as the second argument.
func NewConsumer(c Client, options ...func(*Consumer)) (Consumer, error) {
	config := &Consumer{
		Inmem:          inmemconfig.ConsumerDefaults(c.Inmem),
		Kafka:          kafkaconfig.ConsumerDefaults(c.Kafka),
		Pubsub:         pubsubconfig.ConsumerDefaults(c.Pubsub),
		Standardstream: standardstreamconfig.ConsumerDefaults(c.Standardstream),
	}

	for _, option := range options {
		option(config)
	}

	// We pass the config by-value, to prevent any race-condition where the
	// original config struct is modified after the fact.
	return *config, nil
}

// NewProducer returns a new Producer configuration struct, containing the
// values passed into the function. If any error occurs during configuration
// validation, an error is returned as the second argument.
func NewProducer(c Client, options ...func(*Producer)) (Producer, error) {
	config := &Producer{
		Inmem:          inmemconfig.ProducerDefaults(c.Inmem),
		Kafka:          kafkaconfig.ProducerDefaults(c.Kafka),
		Pubsub:         pubsubconfig.ProducerDefaults(c.Pubsub),
		Standardstream: standardstreamconfig.ProducerDefaults(c.Standardstream),
	}

	for _, option := range options {
		option(config)
	}

	// We pass the config by-value, to prevent any race-condition where the
	// original config struct is modified after the fact.
	return *config, nil
}