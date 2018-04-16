package natsconfig

// Producer is a value-object, containing all user-configurable configuration
// values that dictate how the NATS client's producer will behave.
type Producer struct {
	// Servers is a configured set of servers which this client will use when
	// attempting to connect. The URL can contain username/password
	// semantics (e.g. `nats://derek:pass@localhost:4222`).
	Servers []string

	// ID is an id string to pass to the server when making requests. The purpose
	// of this is to be able to track the source of requests beyond just IP/port
	// by allowing a logical application name to be included in server-side
	// request logging.
	ID string

	// Topic is the topic used to deliver messages to. This value is used as a
	// default value, if the provided message does not define a topic of its own.
	Topic string
}

// staticProducer is a private struct, used to define default configuration
// values that can't be altered in any way. Some of these can eventually become
// public if need be, but to reduce the configuration API surface, they
// currently aren't.
type staticProducer struct{}

// ProducerDefaults holds the default values for Producer.
var ProducerDefaults = Producer{}

var staticProducerDefaults = &staticProducer{}
