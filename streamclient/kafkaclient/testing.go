package kafkaclient

import (
	"reflect"
	"testing"
	"time"

	"github.com/blendle/go-streamprocessor/stream"
	"github.com/blendle/go-streamprocessor/streamconfig"
	"github.com/blendle/go-streamprocessor/streamconfig/kafkaconfig"
	"github.com/blendle/go-streamprocessor/streammsg"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var (
	// TestBrokerAddress is the address used to connect to the testing broker.
	// This defaults to 127.0.0.1:9092, but can be overwritten if desired.
	TestBrokerAddress = "127.0.0.1:9092"

	// TestTimeoutMultiplier can be used to increase the default timeouts during
	// test runs when waiting for time-sensitive values to be returned. It
	// defaults to a multiplier of 1.
	//
	// This is specifically useful on slower environments like a CI server.
	TestTimeoutMultiplier = 1
)

// TestConsumer returns a new kafka consumer to be used in test cases. It also
// returns a function that should be deferred to clean up resources.
//
// You pass the topic and group name of the consumer as a single argument.
func TestConsumer(tb testing.TB, topicAndGroup string, options ...func(c *streamconfig.Consumer)) (stream.Consumer, func()) {
	tb.Helper()

	consumer, err := NewConsumer(TestConsumerConfig(tb, topicAndGroup, options...)...)
	require.NoError(tb, err)

	return consumer, func() { require.NoError(tb, consumer.Close()) }
}

// TestProducer returns a new kafka consumer to be used in test cases. It also
// returns a function that should be deferred to clean up resources.
//
// You pass the topic and group name of the consumer as a single argument.
func TestProducer(tb testing.TB, topic string, options ...func(c *streamconfig.Producer)) (stream.Producer, func()) {
	tb.Helper()

	producer, err := NewProducer(TestProducerConfig(tb, topic, options...)...)
	require.NoError(tb, err)

	return producer, func() { require.NoError(tb, producer.Close()) }
}

// TestMessageFromConsumer returns a single message, consumed from the provided
// consumer. It has a built-in timeout mechanism to prevent the test from
// getting stuck.
func TestMessageFromConsumer(tb testing.TB, consumer stream.Consumer) streammsg.Message {
	tb.Helper()

	select {
	case m := <-consumer.Messages():
		require.NotNil(tb, m)

		return m
	case <-time.After(time.Duration(5*TestTimeoutMultiplier) * time.Second):
		require.Fail(tb, "Timeout while waiting for message to be returned.")
	}

	return streammsg.Message{}
}

// TestMessageFromTopic returns a single message, consumed from the provided
// topic. It has a built-in timeout mechanism to prevent the test from getting
// stuck.
func TestMessageFromTopic(tb testing.TB, topic string) streammsg.Message {
	tb.Helper()

	consumer, closer := testKafkaConsumer(tb, topic, false)
	defer closer()

	m, err := consumer.ReadMessage(time.Duration(3000*TestTimeoutMultiplier) * time.Millisecond)
	require.NoError(tb, err)

	return *newMessageFromKafka(m)
}

// TestMessagesFromTopic returns all messages in a topic.
func TestMessagesFromTopic(tb testing.TB, topic string) []streammsg.Message {
	tb.Helper()

	consumer, closer := testKafkaConsumer(tb, topic, true)
	defer closer()

	var messages []streammsg.Message
	for event := range consumer.Events() {
		switch e := event.(type) {
		case *kafka.Message:
			messages = append(messages, *newMessageFromKafka(e))
		case kafka.PartitionEOF:
			return messages
		}
	}

	return messages
}

// TestProduceMessages accepts a string to use as the topic, and an arbitrary
// number of argument to generate messages on the provided Kafka topic.
//
// The provided extra arguments can be of several different types:
//
// * `string` – The value is used as the kafka message value.
//
// * `[]string` – The first value is used as the kafka message key, the second
// as the message value, all other values are ignored.
//
// * `streammsg.Message` – The value (and, if applicable, the key) are set on a
//  new `kafka.Message`.
//
// * `*kafka.Message` – The message is delivered to Kafka as-is. If
// `kafka.TopicPartition` is empty, the passed in topic value is used instead.
func TestProduceMessages(tb testing.TB, topic string, values ...interface{}) {
	tb.Helper()

	producer, closer := testKafkaProducer(tb)
	defer closer()

	tp := kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny} // nolint: gotype
	for _, v := range values {
		m := kafka.Message{TopicPartition: tp}

		switch value := v.(type) {
		case string:
			m.Value = []byte(value)
		case []string:
			require.Len(tb, value, 2)

			m.Key = []byte(value[0])
			m.Value = []byte(value[1])
		case streammsg.Message:
			m.Value = value.Value
			m.Key = value.Key
		case kafka.Message:
			m = value
			if m.TopicPartition == (kafka.TopicPartition{}) {
				m.TopicPartition = tp
			}
		default:
			require.Fail(tb, "invalid interface type received.", "type: %s", reflect.TypeOf(value).String())
		}

		require.NoError(tb, producer.Produce(&m, nil))

		select {
		case <-producer.Events():
		case <-time.After(time.Duration(5*TestTimeoutMultiplier) * time.Second):
			require.Fail(tb, "Timeout while waiting for message to be delivered.")
		}
	}

	require.Zero(tb, producer.Flush(10000), "Messages remain in queue after Flush()")
}

// TestOffsets returns a list of `kafka.TopicPartition`s.
func TestOffsets(tb testing.TB, message streammsg.Message) []kafka.TopicPartition {
	tb.Helper()

	consumer, closer := testKafkaConsumer(tb, message.Topic, false)
	defer closer()

	o := streammsg.MessageOpqaue(&message).(opaque)
	offsets, err := consumer.Committed([]kafka.TopicPartition{*o.toppar}, 1000*TestTimeoutMultiplier)
	require.NoError(tb, err)

	return offsets
}

// TestConsumerConfig returns sane default options to use during testing of the
// kafkaclient consumer implementation.
func TestConsumerConfig(tb testing.TB, topicAndGroup string, options ...func(c *streamconfig.Consumer)) []func(c *streamconfig.Consumer) {
	var allOptions []func(c *streamconfig.Consumer)

	if testing.Verbose() {
		logger, err := zap.NewDevelopment()
		require.NoError(tb, err)

		verbose := func(c *streamconfig.Consumer) {
			c.Kafka.Logger = *logger.Named("TestConsumer")
			c.Kafka.Debug.All = true
		}

		allOptions = append(allOptions, verbose)
	}

	opts := func(c *streamconfig.Consumer) {
		c.Kafka.ID = "testConsumer"
		c.Kafka.SessionTimeout = time.Duration(1000*TestTimeoutMultiplier) * time.Millisecond
		c.Kafka.HeartbeatInterval = time.Duration(150*TestTimeoutMultiplier) * time.Millisecond
		c.Kafka.CommitInterval = time.Duration(500*TestTimeoutMultiplier) * time.Millisecond
		c.Kafka.Brokers = []string{TestBrokerAddress}
		c.Kafka.GroupID = topicAndGroup
		c.Kafka.Topics = []string{topicAndGroup}
		c.Kafka.InitialOffset = kafkaconfig.OffsetBeginning
	}

	return append(append(allOptions, opts), options...)
}

// TestProducerConfig returns sane default options to use during testing of the
// kafkaclient producer implementation.
func TestProducerConfig(tb testing.TB, topic string, options ...func(c *streamconfig.Producer)) []func(c *streamconfig.Producer) {
	var allOptions []func(c *streamconfig.Producer)

	if testing.Verbose() {
		logger, err := zap.NewDevelopment()
		require.NoError(tb, err)

		verbose := func(c *streamconfig.Producer) {
			c.Kafka.Logger = *logger.Named("TestProducer")
			c.Kafka.Debug.All = true
		}

		allOptions = append(allOptions, verbose)
	}

	opts := func(p *streamconfig.Producer) {
		p.Kafka.ID = "testProducer"
		p.Kafka.SessionTimeout = time.Duration(1000*TestTimeoutMultiplier) * time.Millisecond
		p.Kafka.HeartbeatInterval = time.Duration(150*TestTimeoutMultiplier) * time.Millisecond
		p.Kafka.Brokers = []string{TestBrokerAddress}
		p.Kafka.Topic = topic
	}

	return append(append(allOptions, opts), options...)
}

func testKafkaProducer(tb testing.TB) (*kafka.Producer, func()) {
	tb.Helper()

	config := &kafka.ConfigMap{
		"client.id":            "testKafkaProducer",
		"metadata.broker.list": TestBrokerAddress,
		"go.batch.producer":    false,
		"default.topic.config": kafka.ConfigMap{"acks": 1},
	}

	producer, err := kafka.NewProducer(config)
	require.NoError(tb, err)

	closer := func() {
		i := producer.Flush(1000 * TestTimeoutMultiplier)
		require.Zero(tb, i, "expected all messages to be flushed")

		producer.Close()
	}

	return producer, closer
}

func testKafkaConsumer(tb testing.TB, topic string, events bool) (*kafka.Consumer, func()) {
	tb.Helper()

	config, err := streamconfig.NewConsumer(TestConsumerConfig(tb, topic)...)
	require.NoError(tb, err)

	kconfig, err := config.Kafka.ConfigMap()
	require.NoError(tb, err)

	_ = kconfig.SetKey("client.id", "testKafkaConsumer")
	_ = kconfig.SetKey("enable.partition.eof", true)
	_ = kconfig.SetKey("go.events.channel.enable", events)
	_ = kconfig.SetKey("go.application.rebalance.enable", false)

	consumer, err := kafka.NewConsumer(kconfig)
	require.NoError(tb, err)

	err = consumer.Subscribe(topic, nil)
	require.NoError(tb, err)

	return consumer, func() { require.NoError(tb, consumer.Close()) }
}