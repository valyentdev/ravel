// This package is used to interact with the corrosion API.
package corroclient

import "net/http"

type Config struct {
	URL    string
	Bearer string
}

type CorroClient struct {
	c      *http.Client
	url    string
	bearer string
}

func (c *CorroClient) getURL(path string) string {
	return c.url + path
}

func NewCorroClient(config Config) *CorroClient {
	client := &http.Client{}
	corroClient := &CorroClient{
		c:      client,
		url:    config.URL,
		bearer: config.Bearer,
	}

	return corroClient
}
