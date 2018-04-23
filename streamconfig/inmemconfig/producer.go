package inmemconfig

import (
	"github.com/blendle/go-streamprocessor/streamutil/inmemstore"
)

// Producer is a value-object, containing all user-configurable configuration
// values that dictate how the inmem client's producer will behave.
type Producer struct {
	// Store is the inmem store to which to produce messages. If left undefined,
	// an internal store will be used.
	Store *inmemstore.Store `ignored:"true"`
}

// ProducerDefaults holds the default values for Producer.
var ProducerDefaults = Producer{
	Store: inmemstore.New(),
}
