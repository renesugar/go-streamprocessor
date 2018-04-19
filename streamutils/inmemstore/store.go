package inmemstore

import (
	"reflect"
	"sync"

	"github.com/blendle/go-streamprocessor/stream"
	"github.com/blendle/go-streamprocessor/streammsg"
)

// Store hold the in-memory representation of a data storage service.
type Store struct {
	store   []streammsg.Message
	storev2 []stream.Message
	mutex   sync.RWMutex
}

// New initializes a new store struct.
func New() *Store {
	return &Store{store: make([]streammsg.Message, 0)}
}

// Add adds a streammsg.Message to the store.
func (s *Store) Add(msg streammsg.Message) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.store = append(s.store, msg)
}

// Delete deletes a streammsg.Message from the store.
func (s *Store) Delete(msg streammsg.Message) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	ks := s.store
	for i := range s.store {
		if reflect.DeepEqual(s.store[i], msg) {
			ks = append(s.store[:i], s.store[i+1:]...)
			break
		}
	}
	s.store = ks
}

// Messages returns all messages in the store.
func (s *Store) Messages() []streammsg.Message {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.store
}

// AddV2 adds a stream.Message to the store.
func (s *Store) AddV2(msg stream.Message) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.storev2 = append(s.storev2, msg)
}

// DeleteV2 deletes a streammsg.Message from the store.
func (s *Store) DeleteV2(msg stream.Message) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	ks := s.storev2
	for i := range s.storev2 {
		if reflect.DeepEqual(s.storev2[i], msg) {
			ks = append(s.storev2[:i], s.storev2[i+1:]...)
			break
		}
	}
	s.storev2 = ks
}

// MessagesV2 returns all messages in the store.
func (s *Store) MessagesV2() []stream.Message {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.storev2
}
