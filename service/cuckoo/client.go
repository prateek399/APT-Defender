package cuckoo

import (
	"net/http"
	"time"
)

const (
	defaultTimeout = time.Second * 120
)

type Client struct {
	APIKey  string
	BaseURL string
	Client  *http.Client
}

type Config struct {
	APIKey  string
	BaseURL string
	Client  *http.Client
}

func New(c *Config) *Client {
	client := c.Client
	if client == nil {
		client = &http.Client{
			Timeout: defaultTimeout,
		}
	}

	return &Client{
		APIKey:  c.APIKey,
		BaseURL: c.BaseURL,
		Client:  client,
	}
}
