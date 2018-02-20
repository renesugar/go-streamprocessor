package standardstreamclient

// The following interfaces are implemented by standardstreamclient:
//
//   streammsg.Message

type message struct {
	value []byte
}

// Value returns the value of the message.
func (m *message) Value() []byte {
	return m.value
}

// SetValue sets the value of the message.
func (m *message) SetValue(v []byte) {
	m.value = v
}

// Ack is a no-op implementation of the required interface.
func (m *message) Ack() error {
	return nil
}