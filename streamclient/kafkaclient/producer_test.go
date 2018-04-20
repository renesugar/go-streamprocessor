package kafkaclient_test

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/blendle/go-streamprocessor/streamclient/kafkaclient"
	"github.com/blendle/go-streamprocessor/streamconfig"
	"github.com/blendle/go-streamprocessor/streammsg"
	"github.com/blendle/go-streamprocessor/streamutils/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestProducer(t *testing.T) {
	t.Parallel()

	_ = kafkaclient.Producer{}
}

func TestIntegrationNewProducer(t *testing.T) {
	t.Parallel()
	testutils.Integration(t)

	topic := testutils.Random(t)
	options := kafkaclient.TestProducerConfig(t, topic)

	producer, err := kafkaclient.NewProducer(options...)
	require.NoError(t, err)
	defer func() { require.NoError(t, producer.Close()) }()

	assert.Equal(t, "*kafkaclient.Producer", reflect.TypeOf(producer).String())
}

func TestIntegrationNewProducer_WithOptions(t *testing.T) {
	t.Parallel()
	testutils.Integration(t)

	topic := testutils.Random(t)
	options := kafkaclient.TestProducerConfig(t, topic, func(c *streamconfig.Producer) {
		c.Kafka.Debug.Msg = true
		c.Kafka.SSL.KeyPassword = "test"
	})

	producer, err := kafkaclient.NewProducer(options...)
	require.NoError(t, err)
	defer func() { require.NoError(t, producer.Close()) }()

	assert.Equal(t, false, producer.Config().Kafka.Debug.Broker)
	assert.Equal(t, true, producer.Config().Kafka.Debug.Msg)
	assert.Equal(t, "test", producer.Config().Kafka.SSL.KeyPassword)
}

func TestIntegrationProducer_Messages(t *testing.T) {
	t.Parallel()
	testutils.Integration(t)

	topic := testutils.Random(t)
	message := streammsg.Message{Value: []byte("Hello Universe!")}

	producer, closer := kafkaclient.TestProducer(t, topic)
	defer closer()

	select {
	case producer.Messages() <- message:
	case <-time.After(time.Duration(5*kafkaclient.TestTimeoutMultiplier) * time.Second):
		require.Fail(t, "Timeout while waiting for message to be delivered.")
	}

	msg := kafkaclient.TestMessageFromTopic(t, topic)
	assert.EqualValues(t, message.Value, msg.Value)
}

func TestIntegrationProducer_Messages_Ordering(t *testing.T) {
	t.Parallel()
	testutils.Integration(t)

	messageCount := 5000
	topic := testutils.Random(t)

	producer, closer := kafkaclient.TestProducer(t, topic)
	defer closer()

	for i := 0; i < messageCount; i++ {
		select {
		case producer.Messages() <- streammsg.Message{Value: []byte(strconv.Itoa(i))}:
		case <-time.After(1 * time.Second):
			require.Fail(t, "Timeout while waiting for message to be delivered.")
		}
	}

	// We explicitly close the producer here to force flushing of any messages
	// still in the queue.
	closer()

	messages := kafkaclient.TestMessagesFromTopic(t, topic)
	assert.Len(t, messages, messageCount)

	for i, msg := range messages {
		assert.Equal(t, strconv.Itoa(i), string(msg.Value))
	}
}

func BenchmarkIntegrationProducer_Messages(b *testing.B) {
	testutils.Integration(b)

	topic := testutils.Random(b)
	logger, err := zap.NewDevelopment()
	require.NoError(b, err, logger)

	// We use the default (production-like) config in this benchmark, to simulate
	// real-world usage as best as possible.
	options := func(c *streamconfig.Producer) {
		c.Kafka.Brokers = []string{kafkaclient.TestBrokerAddress}
		c.Kafka.Topic = topic
	}

	producer, err := kafkaclient.NewProducer(options)
	require.NoError(b, err)
	defer func() { require.NoError(b, producer.Close()) }()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		msg := streammsg.TestMessage(b, strconv.Itoa(i), fmt.Sprintf(`{"number":%d}`, i))
		msg.Topic = topic

		producer.Messages() <- msg
	}
}