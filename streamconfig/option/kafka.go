package option

import (
	"time"

	"github.com/blendle/go-streamprocessor/streamconfig"
	"github.com/blendle/go-streamprocessor/streamconfig/kafkaconfig"
)

// KafkaBroker adds a single Kafka broker to `Brokers`.
func KafkaBroker(uri string) streamconfig.Option {
	return optionFunc(func(c *streamconfig.Consumer, p *streamconfig.Producer) {
		c.Kafka.Brokers = append(c.Kafka.Brokers, uri)
		p.Kafka.Brokers = append(p.Kafka.Brokers, uri)
	})
}

// CommitInterval sets `CommitInterval`.
func CommitInterval(d time.Duration) streamconfig.Option {
	return optionFunc(func(c *streamconfig.Consumer, _ *streamconfig.Producer) {
		c.Kafka.CommitInterval = d
	})
}

// EnableDebug sets `Debug.All` to `true`.
func EnableDebug() streamconfig.Option {
	return optionFunc(func(c *streamconfig.Consumer, p *streamconfig.Producer) {
		c.Kafka.Debug.All = true
		p.Kafka.Debug.All = true
	})
}

// GroupID sets `GroupID`.
func GroupID(id string) streamconfig.Option {
	return optionFunc(func(c *streamconfig.Consumer, _ *streamconfig.Producer) {
		c.Kafka.GroupID = id
	})
}

// HeartbeatInterval sets `HeartbeatInterval`.
func HeartbeatInterval(d time.Duration) streamconfig.Option {
	return optionFunc(func(c *streamconfig.Consumer, p *streamconfig.Producer) {
		c.Kafka.HeartbeatInterval = d
		p.Kafka.HeartbeatInterval = d
	})
}

// InitialOffset sets `InitialOffset`.
func InitialOffset(offset kafkaconfig.Offset) streamconfig.Option {
	return optionFunc(func(c *streamconfig.Consumer, _ *streamconfig.Producer) {
		c.Kafka.InitialOffset = offset
	})
}

// SecurityProtocol sets `SecurityProtocol`.
func SecurityProtocol(protocol kafkaconfig.Protocol) streamconfig.Option {
	return optionFunc(func(c *streamconfig.Consumer, p *streamconfig.Producer) {
		c.Kafka.SecurityProtocol = protocol
		p.Kafka.SecurityProtocol = protocol
	})
}

// SessionTimeout sets `SessionTimeout`.
func SessionTimeout(d time.Duration) streamconfig.Option {
	return optionFunc(func(c *streamconfig.Consumer, p *streamconfig.Producer) {
		c.Kafka.SessionTimeout = d
		p.Kafka.SessionTimeout = d
	})
}

// SSLCAPath sets `SSL.CAPath`.
func SSLCAPath(s string) streamconfig.Option {
	return optionFunc(func(c *streamconfig.Consumer, p *streamconfig.Producer) {
		c.Kafka.SSL.CAPath = s
		p.Kafka.SSL.CAPath = s
	})
}

// SSLCRLPath sets `SSL.CRLPath`.
func SSLCRLPath(s string) streamconfig.Option {
	return optionFunc(func(c *streamconfig.Consumer, p *streamconfig.Producer) {
		c.Kafka.SSL.CRLPath = s
		p.Kafka.SSL.CRLPath = s
	})
}

// SSLCertPath sets `SSL.CertPath`.
func SSLCertPath(s string) streamconfig.Option {
	return optionFunc(func(c *streamconfig.Consumer, p *streamconfig.Producer) {
		c.Kafka.SSL.CertPath = s
		p.Kafka.SSL.CertPath = s
	})
}

// SSLKeyPassword sets `SSL.KeyPassword`.
func SSLKeyPassword(s string) streamconfig.Option {
	return optionFunc(func(c *streamconfig.Consumer, p *streamconfig.Producer) {
		c.Kafka.SSL.KeyPassword = s
		p.Kafka.SSL.KeyPassword = s
	})
}

// SSLKeyPath sets `SSL.KeyPath`.
func SSLKeyPath(s string) streamconfig.Option {
	return optionFunc(func(c *streamconfig.Consumer, p *streamconfig.Producer) {
		c.Kafka.SSL.KeyPath = s
		p.Kafka.SSL.KeyPath = s
	})
}

// SSLKeystorePassword sets `SSL.KeystorePassword`.
func SSLKeystorePassword(s string) streamconfig.Option {
	return optionFunc(func(c *streamconfig.Consumer, p *streamconfig.Producer) {
		c.Kafka.SSL.KeystorePassword = s
		p.Kafka.SSL.KeystorePassword = s
	})
}

// SSLKeystorePath sets `SSL.KeystorePath`.
func SSLKeystorePath(s string) streamconfig.Option {
	return optionFunc(func(c *streamconfig.Consumer, p *streamconfig.Producer) {
		c.Kafka.SSL.KeystorePath = s
		p.Kafka.SSL.KeystorePath = s
	})
}

// Topic adds a topic in the case of a consumer, or sets the topic in case of
// the producer.
func Topic(s string) streamconfig.Option {
	return optionFunc(func(c *streamconfig.Consumer, p *streamconfig.Producer) {
		c.Kafka.Topics = append(c.Kafka.Topics, s)
		p.Kafka.Topic = s
	})
}
