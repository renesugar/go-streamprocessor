package streamclient

import (
	"github.com/blendle/go-streamprocessor/streamclient/inmemclient"
	"go.uber.org/zap"
)

// Config ...
type Config struct {
	Inmem *inmemclient.Config

	// HandleInterrupt determines whether the client should close itself
	// gracefully when an interrupt signal (^C) is received. This defaults to true
	// to increase first-time ease-of-use, but if the application wants to handle
	// these signals manually, this flag disables the automated implementation.
	HandleInterrupt bool `ignored:"true"`

	// HandleErrors determines whether the client should self-handle any errors
	// returned on the Errors channel. This is enabled by default, meaning any
	// error returned on the Errors channel will result in a fatal termination of
	// the application. If you want more fine-grained control over what to do when
	// an error occurs, you can set this to false, and manually listen to, and act
	// on errors received on the Errors channel.
	HandleErrors bool `ignored:"true"`

	// Logger is the configurable logger instance to log messages. If left
	// undefined, a no-op logger will be used.
	Logger *zap.Logger `ignored:"true"`
}

// NewConfig returns a new configuration struct for all stream clients.
func NewConfig() *Config {
	return &Config{
		Inmem:           inmemclient.NewConfig(),
		HandleInterrupt: true,
		HandleErrors:    true,
		Logger:          zap.NewNop(),
	}
}
