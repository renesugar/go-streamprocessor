package inmemclient

import "github.com/blendle/go-streamprocessor/stream"

func (c *Client) produce(ch <-chan stream.Message) {
	defer c.wgProduce.Done()

	for msg := range ch {
		c.config.Inmem.Store.AddV2(msg)
	}
}
