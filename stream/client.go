package stream

// // Event interface representing a returned value by the consumer.
// type Event interface{}

// Client interface to be implemented by different stream clients.
type Client interface {
	// // Consume is a read-only channel on which the consumer delivers any messages
	// // being read from the stream.
	// //
	// // The channel returns each message as an `Event` interface. This interface
	// // can be one of two types: a `stream.Message` value object, or a
	// // `stream.Error` error state. After receiving an event, you can add a case
	// // statement to determine which type of event was returned, and handle the
	// // event appropriately.
	// Consume() <-chan Event

	// Consume is a read-only channel on which the consumer delivers any messages
	// being read from the stream.
	Consume() <-chan Message

	// Produce is a write-only channel on which you can deliver any messages that
	// need to be produced on the message stream.
	//
	// The channel accepts `stream.Message` value objects.
	Produce() chan<- Message

	// Error is a read-only channel on which any errors occurred during the
	// execution of the client implementation are returned. The listener of this
	// channel can determine what to do with the returned error, but any error
	// returned should be considered a *fatal* error, and most likely the best
	// course of action is to try to close the client, and exit the application,
	// to prevent unintended side-effects caused by the error.
	//
	// Note that this channel does _not_ return any errors occurred when consuming
	// any messages. These errors are returned on the `Consume` channel, in-line
	// with any actually returned messages, to allow for synchronous handling of
	// error states, and preventing race conditions between consuming more
	// messages and dealing with error states.
	Error() <-chan error

	// Ack can be used to acknowledge that a message was processed and should not
	// be delivered again.
	Ack(Message) error

	// Nack is the opposite of `Ack`. It can be used to indicate that a message
	// was _not_ processed, and should be delivered again in the future.
	Nack(Message) error

	// Close closes the client. After calling this method, the client is no longer
	// in a usable state, and subsequent method calls can result in panics.
	//
	// Check the specific implementations to know what exactly happens when
	// calling close, but in general any active connection to the message stream
	// is terminated and the messages channel is closed.
	Close() error

	// Config returns the final configuration used by the client.
	// TODO: see if we can actually move this to a `testing` implementation, as
	//       there's not really a production use-case to have this data, AFAIK.
	// Config() streamconfig.Client
}
