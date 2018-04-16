package natsconfig

import nats "github.com/nats-io/go-nats"

// Consumer is a value-object, containing all user-configurable configuration
// values that dictate how the NATS client's consumer will behave.
type Consumer struct {
	// Servers is a configured set of servers which this client will use when
	// attempting to connect. The URL can contain username/password
	// semantics (e.g. `nats://derek:pass@localhost:4222`).
	Servers []string

	// ID is an id string to pass to the server when making requests. The purpose
	// of this is to be able to track the source of requests beyond just IP/port
	// by allowing a logical application name to be included in server-side
	// request logging.
	ID string

	// QueueID is the id used for the queue group this consumer should be part of.
	// If multiple consumers use the same queue ID, they will join the same queue
	// group, and any messages delivered to the subscribed subject, will be
	// randomly delivered to one of the consumers in the queue group.
	QueueID string

	// BufferSize is the size of the buffered channel used between the socket Go
	// routine and the message delivery for SyncSubscriptions.
	BufferSize int

	// NoRandomize configures whether we will randomize the
	// server pool.
	// NoRandomize bool

	// Name is an optional name label which will be sent to the server
	// on CONNECT to identify the client.
	// Name string

	// Verbose signals the server to send an OK ack for commands
	// successfully processed by the server.
	// Verbose bool

	// Pedantic signals the server whether it should be doing further
	// validation of subjects.
	// Pedantic bool

	// Secure enables TLS secure connections that skip server
	// verification by default. NOT RECOMMENDED.
	// Secure bool

	// TLSConfig is a custom TLS configuration to use for secure
	// transports.
	// TLSConfig *tls.Config

	// AllowReconnect enables reconnection logic to be used when we encounter a
	// disconnect from the current server.
	// AllowReconnect bool

	// MaxReconnect sets the number of reconnect attempts that will be tried
	// before giving up. If negative, then it will never give up trying to
	// reconnect.
	// MaxReconnect int

	// ReconnectWait sets the time to backoff after attempting a reconnect to a
	// server that we were already connected to previously.
	// ReconnectWait time.Duration

	// Timeout sets the timeout for a Dial operation on a connection.
	// Timeout time.Duration

	// FlusherTimeout is the maximum time to wait for the flusher loop to be able
	// to finish writing to the underlying connection.
	// FlusherTimeout time.Duration

	// PingInterval is the period at which the client will be sending ping
	// commands to the server, disabled if 0 or negative.
	// PingInterval time.Duration

	// MaxPingsOut is the maximum number of pending ping commands that can be
	// awaiting a response before raising an ErrStaleConnection error.
	// MaxPingsOut int

	// ClosedCB sets the closed handler that is called when a client will no
	// longer be connected.
	// ClosedCB ConnHandler

	// DisconnectedCB sets the disconnected handler that is called whenever the
	// connection is disconnected.
	// DisconnectedCB ConnHandler

	// ReconnectedCB sets the reconnected handler called whenever the connection
	// is successfully reconnected.
	// ReconnectedCB ConnHandler

	// DiscoveredServersCB sets the callback that is invoked whenever a new server
	// has joined the cluster.
	// DiscoveredServersCB ConnHandler

	// AsyncErrorCB sets the async error handler (e.g. slow consumer errors)
	// AsyncErrorCB ErrHandler

	// ReconnectBufSize is the size of the backing bufio during reconnect.
	// Once this has been exhausted publish operations will return an error.
	// ReconnectBufSize int

	// SubChanLen is the size of the buffered channel used between the socket
	// Go routine and the message delivery for SyncSubscriptions.
	// NOTE: This does not affect AsyncSubscriptions which are
	// dictated by PendingLimits()
	// SubChanLen int

	// User sets the username to be used when connecting to the server.
	// User string

	// Password sets the password to be used when connecting to a server.
	// Password string

	// Token sets the token to be used when connecting to a server.
	// Token string

	// Dialer allows a custom net.Dialer when forming connections.
	// DEPRECATED: should use CustomDialer instead.
	// Dialer *net.Dialer

	// CustomDialer allows to specify a custom dialer (not necessarily
	// a *net.Dialer).
	// CustomDialer CustomDialer

	// UseOldRequestStyle forces the old method of Requests that utilize
	// a new Inbox and a new Subscription for each request.
	// UseOldRequestStyle bool

	// SSL contains all configuration values for Kafka SSL connections. Defaults
	// to an empty struct, meaning no SSL configuration is required to connect to
	// the brokers.
	// SSL SSL

	// Topic is the topic to which to subscribe for this consumer.
	Topic string
}

// staticConsumer is a private struct, used to define default configuration
// values that can't be altered in any way. Some of these can eventually become
// public if need be, but to reduce the configuration API surface, they
// currently aren't.
type staticConsumer struct{}

// ConsumerDefaults holds the default values for Consumer.
var ConsumerDefaults = Consumer{
	BufferSize: nats.DefaultMaxChanLen,
}

var staticConsumerDefaults = &staticConsumer{}
