package stream_test

import (
	"testing"

	"github.com/blendle/go-streamprocessor/stream"
	"github.com/blendle/go-streamprocessor/streamconfig"
	"github.com/blendle/go-streamprocessor/streammsg"
)

func TestProducer(t *testing.T) {
	t.Parallel()

	var _ stream.Producer = (*fakeProducer)(nil)
}

type fakeMessage struct {
	value []byte
}

func (m *fakeMessage) Value() []byte {
	return m.value
}

func (m *fakeMessage) SetValue(v []byte) {
	m.value = v
}

func (m *fakeMessage) Ack() {}

type fakeProducer struct {
	messages chan<- streammsg.Message
}

func (p *fakeProducer) NewMessage(value []byte) streammsg.Message {
	return &fakeMessage{value: value}
}

func (p *fakeProducer) Messages() chan<- streammsg.Message {
	return p.messages
}

func (p *fakeProducer) Close() error {
	return nil
}

func (p fakeProducer) Config() streamconfig.Producer {
	return streamconfig.Producer{}
}
