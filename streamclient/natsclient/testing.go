package natsclient

import (
	"reflect"
	"testing"
	"time"

	"github.com/blendle/go-streamprocessor/stream"
	"github.com/blendle/go-streamprocessor/streamconfig"
	"github.com/blendle/go-streamprocessor/streammsg"
	"github.com/blendle/go-streamprocessor/streamutils/testutils"
	"github.com/nats-io/gnatsd/server"
	gnatsd "github.com/nats-io/gnatsd/test"
	nats "github.com/nats-io/go-nats"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var (
	// TestServerAddress is the address used to connect to the testing server.
	// This defaults to 127.0.0.1:8222, but can be overwritten if desired.
	TestServerAddress = "127.0.0.1:4222"
)

// TestConsumer returns a new NATS consumer to be used in test cases. It also
// returns a function that should be deferred to clean up resources.
//
// You pass the topic and queue ID of the consumer as a single argument.
func TestConsumer(tb testing.TB, topicAndQueue string, options ...func(c *streamconfig.Consumer)) (stream.Consumer, func()) {
	tb.Helper()

	consumer, err := NewConsumer(TestConsumerConfig(tb, topicAndQueue, options...)...)
	require.NoError(tb, err)

	return consumer, func() { require.NoError(tb, consumer.Close()) }
}

// TestProducer returns a new NATS consumer to be used in test cases. It also
// returns a function that should be deferred to clean up resources.
//
// You pass the topic and group name of the consumer as a single argument.
func TestProducer(tb testing.TB, topic string, options ...func(c *streamconfig.Producer)) (stream.Producer, func()) {
	tb.Helper()

	producer, err := NewProducer(TestProducerConfig(tb, topic, options...)...)
	require.NoError(tb, err)

	return producer, func() { require.NoError(tb, producer.Close()) }
}

// TestServer ....
func TestServer() (*server.Server, func()) {
	opts := gnatsd.DefaultTestOptions
	opts.Port = nats.DefaultPort
	s := gnatsd.RunServer(&opts)

	return s, s.Shutdown
}

// TestMessageFromTopic returns a single message, consumed from the provided
// topic. It has a built-in timeout mechanism to prevent the test from getting
// stuck.
func TestMessageFromTopic(tb testing.TB, topic string) streammsg.Message {
	tb.Helper()

	consumer, closer := testNATSClient(tb)
	defer closer()

	sub, err := consumer.QueueSubscribeSync(topic, topic)
	require.NoError(tb, err)

	m, err := sub.NextMsg(time.Duration(3000*testutils.TimeoutMultiplier) * time.Millisecond)
	require.NoError(tb, err)

	return *newMessageFromNATS(m)
}

// TestMessagesFromTopic returns all messages in a topic.
func TestMessagesFromTopic(tb testing.TB, topic string) []streammsg.Message {
	tb.Helper()

	consumer, closer := testNATSClient(tb)
	defer closer()

	var messages []streammsg.Message

	sub, err := consumer.QueueSubscribeSync(topic, topic)
	require.NoError(tb, err)

	for {
		m, err := sub.NextMsg(100 * time.Millisecond)
		if err == nats.ErrTimeout {
			break
		}

		require.NoError(tb, err)
		messages = append(messages, *newMessageFromNATS(m))
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
// * `streammsg.Message` – The value are set on a new `kafka.Message`.
//
// * `*nats.Msg` – The message is delivered to NATS as-is. If
// `nats.Subscription` is empty, the passed in topic value is used instead.
func TestProduceMessages(tb testing.TB, topic string, values ...interface{}) {
	tb.Helper()

	producer, closer := testNATSClient(tb)
	defer closer()

	sub := &nats.Subscription{Subject: topic}
	for _, v := range values {
		m := nats.Msg{Sub: sub, Subject: topic}

		switch value := v.(type) {
		case string:
			m.Data = []byte(value)
		case streammsg.Message:
			m.Data = value.Value
			m.Subject = value.Topic

			if m.Subject == "" {
				m.Subject = topic
			}
		case nats.Msg:
			m = value
			if m.Sub == nil {
				m.Sub = sub
			}
		default:
			require.Fail(tb, "invalid interface type received.", "type: %s", reflect.TypeOf(value).String())
		}

		require.NoError(tb, producer.PublishMsg(&m))
	}

	require.NoError(tb, producer.FlushTimeout(10*time.Second))
}

// TestConsumerConfig returns sane default options to use during testing of the
// kafkaclient consumer implementation.
func TestConsumerConfig(tb testing.TB, topicAndQueue string, options ...func(c *streamconfig.Consumer)) []func(c *streamconfig.Consumer) {
	var allOptions []func(c *streamconfig.Consumer)

	if testing.Verbose() {
		logger, err := zap.NewDevelopment()
		require.NoError(tb, err)

		verbose := func(c *streamconfig.Consumer) {
			c.Logger = *logger.Named("TestConsumer")
		}

		allOptions = append(allOptions, verbose)
	}

	opts := func(c *streamconfig.Consumer) {
		c.NATS.ID = "testConsumer"
		c.NATS.Servers = []string{"nats://" + TestServerAddress}
		c.NATS.Topic = topicAndQueue
		c.NATS.QueueID = topicAndQueue
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
			c.Logger = *logger.Named("TestProducer")
		}

		allOptions = append(allOptions, verbose)
	}

	opts := func(p *streamconfig.Producer) {
		p.NATS.ID = "testProducer"
		p.NATS.Servers = []string{"nats://" + TestServerAddress}
		p.NATS.Topic = topic
	}

	return append(append(allOptions, opts), options...)
}

func testNATSClient(tb testing.TB) (*nats.Conn, func()) {
	tb.Helper()

	no := nats.Options{Servers: []string{"nats://" + TestServerAddress}}
	consumer, err := no.Connect()
	require.NoError(tb, err)

	return consumer, consumer.Close
}
