# go-streamprocessor

This package contains shared logic for Go-based stream processors.

## TODO (remove this before committing):

- make consumed messages immutable?

## Usage

Import in your `main.go`:

```golang
import (
  "github.com/blendle/go-streamprocessor/stream"
  "github.com/blendle/go-streamprocessor/streamclient"
)
```

Next, instantiate a new consumer and producer (or one of both):

```golang
consumer, producer, err := streamclient.NewConsumerAndProducer()
if err != nil {
  panic(err)
}
```

Don't forget to close the consumer and producer once you're done:

```golang
defer producer.Close()
defer consumer.Close()
```

Next, loop over consumed messages, and produce new messages:

```golang
for msg := range consumer.Messages() {

  // Initialize a new message.
  message := producer.NewMessage()

  // Set the value of the message to the consumed message, prepended with
  // "processed: ".
  value := append([]byte("processed: "), msg.Value()...)
  message.SetValue(value)

  // Send message to the configured producer.
  producer.Messages() <- message

  // Acknowledge the consumed message. This lets the client know it should not
  // retry this message if the application is restarted.
  msg.Ack()
}
```

Now, run your example in your terminal:

```shell
echo -e "hello\nworld" | env PRODUCER=standardstream go run main.go
# processed: hello
# processed: world
```

You can find the above example in the [examples directory](examples/).

## Using Interfaces for Advanced Usage

In the above example, we used a very simple example, that only dealt with the
so-called "value" of a message. This is usually the most important attribute of a
message, but not the _only_ attribute of a message.

A message can contain many more attributes, and this list of attributes depends
on what client implementation you are using.

In the above example, we used the `standardstream` client implementation. This
implementation takes a string over stdin for consumers, and produces a string
over stdout for producers. Because of this, this client only supports reading and
writing the `value` of a message.

However, if we where to use the `inmem` client implementation, we would have
access to _all_ message attributes, since this client is (mostly) used for
testing purposes, and thus has to support the combination of all available
message attributes across all supported client implementations.

To work with these other attributes, you can use a set of available interfaces.
These interfaces are:

* `ValueReader`
* `ValueWriter`
* `TimestampReader`
* `TimestampReader`
* `TimestampWriter`
* `KeyReader`
* `KeyWriter`
* `TopicReader`
* `TopicWriter`
* `OffsetReader`
* `PartitionReader`
* `TagsReader`
* `TagsWriter`
* `Acker`
* `Nacker`

All clients' messages must adhere to a generic interface that implements some of
the above interfaces, but not all, called the `Message` interface. This interface
is a combination of the following interfaces:

* `ValueReader`
* `ValueWriter`
* `Acker`

This means that for each message, out of the box, you can access its "value",
set the value, and acknowledge the message.

Note that not all clients actually have a useful _implementation_ for
acknowledging messages. For example, using the `kafka` client implementation,
acknowledging a message has a concrete meaning, as in this prevents a message
from being delivered a second time, after you already processed it. But when
using the `standardstream` client implementation, even though it implements the
`Acker` interface, it doesn't actually do anything with your acknowledgment.
In this case, the client is using a so-called _no-op_ implementation, to adhere
to the interface, without actually supporting it.

With that information out of the way, let's look at an example.

### Example

Follow the basic usage example, but this time, we'll handle the received and
produced messages in a different way:

```golang
i := 0
for msg := range consumer.Messages() {

  // Initialize a new message.
  message := producer.NewMessage()

  // Set the value of the message to the consumed message, prepended with
  // "processed: ".
  value := append([]byte("processed: "), msg.Value()...)
  message.SetValue(value)

  // If the producer supports it, set the key of the new message.
  if m, ok := message.(streammsg.KeyWriter); ok {
    b := []byte(fmt.Sprintf("%d", i))
    message.SetKey(b)
  }

  // Send message to the configured producer.
  producer.Messages() <- message

  // Acknowledge the consumed message. This lets the client know it should not
  // retry this message if the application is restarted.
  msg.Ack()
}
```

Here, we've used a non-standard message attribute (the message key), not
supported by all client implementations. For the best interoperability of our
processor, we handle this situation by only acting on this attribute if the used
client has implemented this interface (`KeyWriter`).

This way, we can run our processor with different client implementations, and it
will work as expected:

```golang
```

Obviously, if your implementation _needs_ a specific interface implementation to
do its job as expected, you don't have to handle non-conforming clients
gracefully, and can simply terminate the processor in such a situation.




## Client implementations

In the above example, we used the generic `streamclient.NewConsumerAndProducer`.
This creates a consumer and producer, based on the environment you run the
program in (you can also call `NewConsumer` or `NewProducer` individually).

In our example, we send data to the program over `stdin`, this triggered the
application to initialize the `standardstream` consumer, which listens on the
`stdin` stream for any messages (split by newlines).

We also set the `DRY_RUN` environment variable to `true`, this causes the
`standardstream` _producer_ to be started as well, which produces messages on
the `stdout` stream.

There are multiple stream processor clients available. You can initialize the
one you want by creating a consumer or producer from the correct client.

### standardstream

This stream client's consumer listens to stdin, while the producer sends
messages to stdout.

```golang
client := standardstream.NewClient()
consumer := client.NewConsumer()
producer := client.NewProducer()
```

You can tell the consumer or producer to listen or write to another file
descriptor as well:

```golang
options := func(c *standardstream.Client) {
  c.ConsumerFD, _ = os.Open("/var/my/file")
  c.ProducerFD = os.Stderr
}

client := standardstream.NewClient(options)
```

### kafka

This stream client's consumer and producer listen and produce to Kafka streams.

```golang
client := kafka.NewClient()
consumer := client.NewConsumer()
producer := client.NewProducer()
```

If you set the `KAFKA_CONSUMER_URL` and/or `KAFKA_PRODUCER_URL` environment
variables, the client will configure the broker details using those values.

For example:

    KAFKA_CONSUMER_URL="kafka://localhost:9092/my-topic?group=my-group"
    KAFKA_PRODUCER_URL="kafka://localhost:9092/my-other-topic"

You can also define these values using client options:

```golang
options := func(c *kafka.Client) {
  c.ConsumerBrokers = []string{"localhost"}
  c.ConsumerTopics = []string{"my-topic"}
  c.ConsumerGroup = "my-group"

  c.ProducerBrokers = []string{"localhost"}
  c.ProducerTopics  = []string{"my-other-topic"}
}

client := kafka.NewClient(options)
```

### Inmem

This client receives the messages from an in-memory store, and produces it to
the same (or another) in-memory store.

This client is mostly useful during testing, so you reduce the dependency on
file descriptors, or Kafka itself.

```golang
client := inmem.NewClient()
consumer := client.NewConsumer()
producer := client.NewProducer()
```

Here are some more use-cases:

```golang
// create a new inmemory store, so we can access it and pass it to our new inmem
// client.
store := inmem.NewStore()

// create a new topic in the store
topic := store.NewTopic("input-topic")

// add a single message to the topic.
topic.NewMessage([]byte("hello world!"), []byte("my-partition-key"))

// set custom configuration.
options := func(c *Client) {
  c.ConsumerTopic = "input-topic"
  c.ProducerTopic = "output-topic"
}

client := inmem.NewClientWithStore(store, options)
consumer := client.NewConsumer()

for msg := range consumer.Messages() {
  fmt.Printf("received: %s", string(msg))
}

// received: hello world
```

## Logging

You can pass a pointer to a `zap.Logger` instance to each of the above described
clients.

```golang
logger := func(c *kafka.Client) {
  c.Logger, _ = zap.NewProduction()
}

client := kafka.NewClient(logger)
```

## Messages

A message is implemented as such:

```golang
type Message struct {
  Topic     string
  Key       []byte
  Value     []byte
  Timestamp time.Time
}
```

You can set the value of the message, and a timestamp (depending on the client,
this could be used, or not).

You can also set a "partition key" for a message. Again, depending on the type
of streamclient you are using, this is either used, or it isn't.

You can optionally provide a topic name, it will be used to override the
globally configured topic for the processor on a per-message basis.

Here's an example:

```golang
producer.PartitionKey(func(msg *stream.Message) []byte {
  return getKey(msg.Value)
})
```

In the above example, we implement our own `getKey` function that takes `[]byte`
as its input, and returns the key string as `[]byte`.









## TODO

### Type Casting

All stream client consumers/producers accept and return `interface{}` values. To
be able to use a message, you need to cast it to the appropriate type.

You can do this as follows:

```golang
for msg := range consumer.Message() {
  switch m := msg.(type) {
    case stream.Message:
    case stream.KafkaMessage:
    case stream.PubsubMessage:
    case stream.StandardclientMessage:
    case stream.inmemMessage:
  }
}
```


work with the `stream.Message` struct.
However, for more advanced functionality, you need to cast the return value.


===

    // TODO: how to handle this for producers? We don't want the stream
    //       processors to actually have to define a specific client they want
    //       to use.
    //
    //       Perhaps it makes sense to have a big "catch-all" struct with all
    //       the values for all the different clients. In the documentation of
    //       that struct we can then define what works for which client
    //       implementation. The consumer still returns an interface, so that
    //       reading the values of a message is clean (and also the
    //       implementation) of Ack of Nack is either present and functional, or
    //       completely missing.
    //
    //       For acking, you'd do something like this:
    //
    //       if m, ok := msg.(Acker); ok {
    //         m.Ack()
    //       }
    msg := streammsg.Message{}

    msg := &standardstreamclient.Message{}

    msg := streammsg.New()

    msg := 

    msg.SetValue([]byte())

===

// M represents the message struct with all available fields.
type M struct {
    // Value is the actual value of the message. This field is implemented by all
    // stream clients.
    //
    // Implemented by:
    //   * inmemclient
    //   * kafkaclient
    //   * pubsubclient
    //   * standardstreamclient
    Value []byte

    // Timestamp is the timestamp of the message.
    //
    // Implemented by:
    //   * kafkaclient
    //   * pubsubclient (read-only)
    Timestamp time.Time

    // Key is the key of the message.
    //
    // Implemented by:
    //   * kafkaclient
    //   * pubsubclient (read-only)
    Key []byte

    // Attributes represents the key-value pairs the current message is labeled
    // with.
    //
    // Implemented by:
    //   * kafkaclient
    //   * pubsubclient
    Attributes map[string]string

    // Topic is the topic of the message.
    //
    // Implemented by:
    //   * kafkaclient
    Topic string

    // Offset is the offset of the message stored on the broker.
    //
    // Implemented by:
    //   * kafkaclient
    Offset int64

    // Partition is the partition that the message was sent to.
    //
    // Implemented by:
    //   * kafkaclient
    Partition int32
}
