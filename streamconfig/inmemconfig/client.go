package inmemconfig

// Client is a value-object, containing all user-configurable configuration
// values that dictate how the inmemclient will behave.
type Client struct {
	Consumer *Consumer
	Producer *Producer
}
