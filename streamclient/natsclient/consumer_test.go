package natsclient_test

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/blendle/go-streamprocessor/streamclient"
	"github.com/blendle/go-streamprocessor/streamclient/natsclient"
	"github.com/blendle/go-streamprocessor/streamconfig"
	"github.com/blendle/go-streamprocessor/streammsg"
	"github.com/blendle/go-streamprocessor/streamutils/testutils"
	nats "github.com/nats-io/go-nats"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestConsumer(t *testing.T) {
	t.Parallel()

	_ = natsclient.Consumer{}
}

func TestIntegrationNewConsumer(t *testing.T) {
	_, closer := natsclient.TestServer()
	defer closer()

	topicAndGroup := testutils.Random(t)
	options := natsclient.TestConsumerConfig(t, topicAndGroup)

	consumer, err := natsclient.NewConsumer(options...)
	require.NoError(t, err)
	defer func() { require.NoError(t, consumer.Close()) }()

	assert.Equal(t, "*natsclient.Consumer", reflect.TypeOf(consumer).String())
}

func TestIntegrationNewConsumer_WithOptions(t *testing.T) {
	_, closer := natsclient.TestServer()
	defer closer()

	topicAndGroup := testutils.Random(t)
	options := natsclient.TestConsumerConfig(t, topicAndGroup, func(c *streamconfig.Consumer) {
		c.NATS.ID = "HelloUniverse"
		c.NATS.QueueID = "HelloWorld"
	})

	consumer, err := natsclient.NewConsumer(options...)
	require.NoError(t, err)
	defer func() { require.NoError(t, consumer.Close()) }()

	assert.Equal(t, "HelloUniverse", consumer.Config().NATS.ID)
	assert.Equal(t, "HelloWorld", consumer.Config().NATS.QueueID)
}

func TestIntegrationConsumer_Messages(t *testing.T) {
	_, closer := natsclient.TestServer()
	defer closer()

	topicAndGroup := testutils.Random(t)

	consumer, closer := natsclient.TestConsumer(t, topicAndGroup)
	defer closer()

	var actual streammsg.Message
	ch := make(chan bool)
	go func() {
		actual = <-consumer.Messages()
		ch <- true
	}()

	natsclient.TestProduceMessages(t, topicAndGroup, "hello world")

	select {
	case <-ch:
		assert.EqualValues(t, []byte("hello world"), actual.Value)
	case <-time.After(time.Duration(5*testutils.TimeoutMultiplier) * time.Second):
		require.Fail(t, "Timeout while waiting for message to be returned.")
	}
}

func TestIntegrationConsumer_Messages_Ordering(t *testing.T) {
	_, closer := natsclient.TestServer()
	defer closer()

	messageCount := 5000
	topicAndGroup := testutils.Random(t)

	messages := []interface{}{}
	for i := 0; i < messageCount; i++ {
		message := streammsg.TestMessage(t, "", "hello world"+strconv.Itoa(i))
		message.Topic = ""
		messages = append(messages, message)
	}

	consumer, closer := natsclient.TestConsumer(t, topicAndGroup)
	defer closer()

	i := 0
	ch := make(chan bool)
	go func() {
		run := true
		timeout := time.NewTimer(5000 * time.Millisecond)

		for run {
			select {
			case msg := <-consumer.Messages():
				timeout.Reset(5000 * time.Millisecond)

				require.Equal(t, "hello world"+strconv.Itoa(i), string(msg.Value))

				err := consumer.Ack(msg)
				require.NoError(t, err)
				i++

				if i == messageCount {
					run = false
				}
			case <-timeout.C:
				require.Fail(t, "Timeout while waiting for message to be returned.")
			}
		}

		ch <- true
	}()

	natsclient.TestProduceMessages(t, topicAndGroup, messages...)

	<-ch
	assert.Equal(t, messageCount, i)
}

func TestIntegrationConsumer_Ack(t *testing.T) {
	_, closer := natsclient.TestServer()
	defer closer()

	topicOrGroup := testutils.Random(t)

	consumer, closer := natsclient.TestConsumer(t, topicOrGroup)
	defer closer()

	var message streammsg.Message
	ch := make(chan bool)
	go func() {
		message = streamclient.TestMessageFromConsumer(t, consumer)
		ch <- true
	}()

	natsclient.TestProduceMessages(t, topicOrGroup, "hello world", "hello universe!")
	<-ch

	// Ack is a no-op, so it should never error.
	assert.NoError(t, consumer.Ack(message))
}

func BenchmarkIntegrationConsumer_Messages(b *testing.B) {
	_, closer := natsclient.TestServer()
	defer closer()

	topicAndGroup := testutils.Random(b)
	line := `{"number":%d}` + "\n"

	opts := func(c *streamconfig.Consumer) {
		c.NATS.ID = topicAndGroup
		c.NATS.Servers = []string{nats.DefaultURL}
		c.NATS.Topic = topicAndGroup
		c.NATS.BufferSize = b.N

		if testing.Verbose() {
			logger, err := zap.NewDevelopment()
			require.NoError(b, err)
			c.Logger = *logger
		}
	}

	consumer, err := natsclient.NewConsumer(opts)
	require.NoError(b, err)
	defer func() { require.NoError(b, consumer.Close()) }()

	no := nats.Options{
		Servers:    []string{nats.DefaultURL},
		SubChanLen: b.N,
	}

	// Instantiate a new NATS consumer.
	producer, err := no.Connect()
	require.NoError(b, err)

	for i := 1; i <= b.N; i++ {
		_ = producer.Publish(topicAndGroup, []byte(fmt.Sprintf(line, i)))
	}

	require.NoError(b, producer.FlushTimeout(10*time.Second))
	producer.Close()

	i := 0
	b.ResetTimer()
	for {
		select {
		case <-consumer.Messages():
			i++

			if i == b.N {
				return
			}
		case <-time.After(5 * time.Second):
			assert.Fail(b, "timeout waiting for messages to be delivered", "got %d of %d messages", i, b.N)
		}
	}
}
