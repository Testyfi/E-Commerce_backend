package http

import (
	"net/http"
)

type Client struct {
	client         *http.Client
	defaultHeaders map[string]string
}

func NewHttpClient() *Client {
	return &Client{
		client: &http.Client{},
		defaultHeaders: map[string]string{
			"Content-Type": "application/json",
			"accept":       "application/json",
		},
	}
}

func (hc *Client) applyDefaultHeaders(req *http.Request) {
	for key, value := range hc.defaultHeaders {
		// Only set default header if it's not already set
		if req.Header.Get(key) == "" {
			req.Header.Set(key, value)
		}
	}
}

type RequestOption func(*http.Request)

func WithHeader(key, value string) RequestOption {
	return func(r *http.Request) {
		r.Header.Set(key, value) // We use Set() to overwrite existing headers
	}
}
