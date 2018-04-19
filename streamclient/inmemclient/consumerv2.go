package inmemclient

func (c *Client) consume() {
	defer func() {
		close(c.chConsume)
		c.wgConsume.Done()
	}()

	// If `ConsumeOnce` is set to true, we simply loop over all the existing
	// messages in the inmem store and send them to the consumer channel. This
	// will result in this `consume()` method to return once all messages are
	// delivered to the channel.
	if c.config.Inmem.ConsumeOnce {
		for _, msg := range c.config.Inmem.Store.MessagesV2() {
			c.chConsume <- msg
		}

		return
	}

	// If `ConsumeOnce` is set to true, we'll start an infinite loop that listens
	// to new messages in the inmem store, and send them to the consumer channel.
	// Once a message is read from the store, it's also deleted from the store, so
	// that that message is not delivered twice.
	//
	// TODO: we might consider implementing `Ack` to actually delete the message
	//       from the store, which would create a more true-to-spirit
	//       implementation of a stream client, instead of having the side-effect
	//       of actually removing the message from the store happening in this
	//       method.
	for {
		select {
		case <-c.quit:
			c.config.Logger.Info("Received quit signal. Exiting consumer.")

			return
		default:
			for _, msg := range c.config.Inmem.Store.MessagesV2() {
				c.chConsume <- msg
				c.config.Inmem.Store.DeleteV2(msg)
			}
		}
	}
}
