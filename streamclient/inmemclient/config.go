package inmemclient

import "github.com/blendle/go-streamprocessor/streamutils/inmemstore"

// Config is the configuration struct for the inmemclient
type Config struct {
	*ConsumerConfig
	*ProducerConfig

	// Store is the inmem store from which to consume or to which to produce
	// messages. If left undefined, an internal store will be used.
	Store *inmemstore.Store `ignored:"true"`
}

// ConsumerConfig contains all consumer-relevant configuration fields.
type ConsumerConfig struct {
	// ConsumeOnce dictates whether the inmem consumer should request all messages
	// in the configured `inmemstore` once, or if it should keep listening for any
	// new messages being added to the store. This toggle is useful for different
	// test-cases where you either want to fetch all messages once and continue
	// the test, or you want to start the consumer in a separate goroutine in the
	// test-setup and add messages at different intervals _after_ you started the
	// consumer. Note that there's a major side-effect right now to setting
	// `ConsumeOnce` to `false`, which is that the consumer will remove any
	// consumed messages from the configured `inmemstore` to prevent consuming
	// duplicate messages. This behavior is not the case when `ConsumeOnce` is
	// kept at the default `true`, meaning messages will stay in the store, even
	// when consumed.
	//
	// Defaults to `true`, meaning the consumer will auto-close after consuming
	// all existing messages in the `inmemstore`.
	ConsumeOnce bool
}

// ProducerConfig contains all producer-relevant configuration fields.
type ProducerConfig struct{}

// NewConfig returns a new configuration struct.
func NewConfig() *Config {
	return &Config{
		Store: inmemstore.New(),

		ConsumerConfig: &ConsumerConfig{
			ConsumeOnce: true,
		},

		ProducerConfig: &ProducerConfig{},
	}
}
