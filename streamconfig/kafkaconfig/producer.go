package kafkaconfig

import (
	"errors"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/zap"
)

// Producer is a value-object, containing all user-configurable configuration
// values that dictate how the Kafka client's producer will behave.
type Producer struct {
	// Brokers is a list of host/port pairs to use for establishing the initial
	// connection to the Kafka cluster. The client will make use of all servers
	// irrespective of which servers are specified here for bootstrapping — this
	// list only impacts the initial hosts used to discover the full set of
	// servers. Since these servers are just used for the initial connection to
	// discover the full cluster membership (which may change dynamically), this
	// list need not contain the full set of servers (you may want more than one,
	// though, in case a server is down).
	Brokers []string `kafka:"metadata.broker.list,omitempty"`

	// Debug allows tweaking of the default debug values.
	Debug Debug `kafka:"debug,omitempty"`

	// HeartbeatInterval represents The expected time between heartbeats to the
	// consumer coordinator when using Kafka's group management facilities.
	// Heartbeats are used to ensure that the consumer's session stays active and
	// to facilitate rebalancing when new consumers join or leave the group. The
	// value must be set lower than `SessionTimeout`, but typically should be set
	// no higher than 1/3 of that value. It can be adjusted even lower to control
	// the expected time for normal rebalances.
	HeartbeatInterval time.Duration `kafka:"heartbeat.interval.ms,omitempty"`

	// ID is an id string to pass to the server when making requests. The purpose
	// of this is to be able to track the source of requests beyond just IP/port
	// by allowing a logical application name to be included in server-side
	// request logging.
	ID string `kafka:"client.id,omitempty"`

	// Logger is the configurable logger instance to log messages. If left
	// undefined, a no-op logger will be used.
	Logger zap.Logger

	// MaxDeliveryRetries dictates how many times to retry sending a failing
	// MessageSet. Note: retrying may cause reordering. Defaults to 0 retries to
	// prevent accidental message reordering.
	MaxDeliveryRetries int `kafka:"message.send.max.retries"`

	// MaxQueueBufferDuration is the delay to wait for messages in the producer
	// queue to accumulate before constructing message batches (MessageSets) to
	// transmit to brokers. A higher value allows larger and more effective (less
	// overhead, improved compression) batches of messages to accumulate at the
	// expense of increased message delivery latency. Defaults to 0.
	MaxQueueBufferDuration time.Duration `kafka:"queue.buffering.max.ms,omitempty"`

	// MaxQueueSizeKBytes is the maximum total message size sum allowed on the
	// producer queue. This property has higher priority than MaxQueueSize.
	MaxQueueSizeKBytes int `kafka:"queue.buffering.max.kbytes,omitempty"`

	// MaxQueueSize dictates the maximum number of messages allowed on the
	// producer queue.
	MaxQueueSize int `kafka:"queue.buffering.max.messages,omitempty"`

	// RequiredAcks indicates how many acknowledgments the leader broker must
	// receive from ISR brokers before responding to the request:
	//
	// AckNone: Broker does not send any response/ack to client
	// AckLeader: Only the leader broker will need to ack the message,
	// AckAll: broker will block until message is committed by all in sync
	// replicas (ISRs).
	//
	// Defaults to `AckLeader`.
	RequiredAcks Ack `kafka:"{topic}.request.required.acks"`

	// SecurityProtocol is the protocol used to communicate with brokers.
	SecurityProtocol Protocol `kafka:"security.protocol,omitempty"`

	// SessionTimeout represents the timeout used to detect consumer failures when
	// using Kafka's group management facility. The consumer sends periodic
	// heartbeats to indicate its liveness to the broker. If no heartbeats are
	// received by the broker before the expiration of this session timeout, then
	// the broker will remove this consumer from the group and initiate a
	// rebalance. Note that the value must be in the allowable range as configured
	// in the broker configuration by `group.min.session.timeout.ms` and
	// `group.max.session.timeout.ms`.
	SessionTimeout time.Duration `kafka:"session.timeout.ms,omitempty"`

	// SSL contains all configuration values for Kafka SSL connections. Defaults
	// to an empty struct, meaning no SSL configuration is required to connect to
	// the brokers.
	SSL SSL `kafka:"ssl,omitempty"`

	// Topic is the topic used to deliver messages to. This value is used as a
	// default value, if the provided message does not define a topic of its own.
	Topic string `kafka:"-"`
}

// staticProducer is a private struct, used to define default configuration
// values that can't be altered in any way. Some of these can eventually become
// public if need be, but to reduce the configuration API surface, they
// currently aren't.
type staticProducer struct {
	// QueueBackpressureThreshold sets the threshold of outstanding not yet
	// transmitted requests needed to back-pressure the producer's message
	// accumulator. A lower number yields larger and more effective batches.
	// Set to 100, and non-configurable for now.
	QueueBackpressureThreshold int `kafka:"queue.buffering.backpressure.threshold"`

	// CompressionCodec sets the compression codec to use for compressing message
	// sets. This is the default value for all topics, may be overridden by the
	// topic configuration property compression.codec. Set tot `Snappy`,
	// non-configurable.
	CompressionCodec Compression `kafka:"compression.codec"`

	// BatchMessageSize sets the maximum number of messages batched in one
	// MessageSet. The total MessageSet size is also limited by message.max.bytes.
	BatchMessageSize int `kafka:"batch.num.messages"`
}

// ProducerDefaults holds the default values for Producer.
var ProducerDefaults = Producer{
	Debug:                  Debug{},
	HeartbeatInterval:      10 * time.Second,
	Logger:                 *zap.NewNop(),
	MaxDeliveryRetries:     0,
	MaxQueueBufferDuration: time.Duration(0),
	MaxQueueSizeKBytes:     2097151,
	MaxQueueSize:           10000000,
	RequiredAcks:           AckLeader,
	SessionTimeout:         30 * time.Second,
	SSL:                    SSL{},
}

var staticProducerDefaults = &staticProducer{
	QueueBackpressureThreshold: 100,
	CompressionCodec:           CompressionSnappy,
	BatchMessageSize:           100000,
}

// ConfigMap converts the current configuration into a format known to the
// rdkafka library.
func (p *Producer) ConfigMap() (*kafka.ConfigMap, error) {
	return configMap(p, staticProducerDefaults), p.validate()
}

func (p *Producer) validate() error {
	if len(p.Brokers) == 0 {
		return errors.New("required config Kafka.Brokers empty")
	}

	if len(p.Topic) == 0 {
		return errors.New("required config Kafka.Topic missing")
	}

	return nil
}
