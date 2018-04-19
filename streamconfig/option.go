package streamconfig

import (
	"os"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

// Option configures a stream Consumer and/or Producer.
type Option interface {
	Apply(*Client)
}

// WithOptions clones the current Consumer, applies the supplied Options, and
// returns the resulting Consumer. It's safe to use concurrently.
func (c *Client) WithOptions(opts ...Option) *Client {
	cc := c.clone()
	for _, opt := range opts {
		opt.Apply(cc)
	}
	return cc
}

// WithEnvironmentVariables clones the current Consumer, applies the supplied
// environment variables as options, and returns the resulting Consumer. It is
// safe to use concurrently.
func (c *Client) WithEnvironmentVariables(envs ...string) (*Client, error) {
	cc := c.clone()

	for i := range envs {
		v := strings.SplitN(envs[i], "=", 2)
		if len(v) < 2 {
			continue
		}

		os.Setenv(v[0], v[1])
	}

	return cc, envconfig.Process(cc.Name, cc)
}

func (c *Client) clone() *Client {
	copy := *c
	return &copy
}
