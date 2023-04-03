package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Client is the OpenAI API client.
type Client struct {
	hc  *http.Client
	cfg ClientConfig
}

// ClientConfig is the configuration for the OpenAI API client.
type ClientConfig struct {
	HTTPClient *http.Client
	APIKey     string
	RootURL    string
}

// NewClient creates a new OpenAI API client.
func NewClient(cfg ClientConfig) (*Client, error) {
	hc := cfg.HTTPClient
	if hc == nil {
		hc = http.DefaultClient
	}

	if cfg.RootURL == "" {
		cfg.RootURL = APIURLv1
	}

	return &Client{
		hc:  cfg.HTTPClient,
		cfg: cfg,
	}, nil
}

// ChatCompletionRequest is the request for the chat completion endpoint.
func (c *Client) ChatCompletion(req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	hr, err := http.NewRequest(http.MethodPost,
		c.cfg.RootURL+chatCompletionEndpoint, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	hr.Header.Add("Content-Type", "application/json")
	hr.Header.Add("Authorization", "Bearer "+c.cfg.APIKey)

	resp, err := c.hc.Do(hr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var data ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &data, nil
}
