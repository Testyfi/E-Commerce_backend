package http

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

func (hc *Client) Get(url string, opts ...RequestOption) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	hc.applyDefaultHeaders(req)

	for _, opt := range opts {
		opt(req)
	}

	resp, err := hc.client.Do(req)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Something went wrong while closing response")
		}
	}(resp.Body)

	if resp.StatusCode >= 400 {
		return "", errors.New(resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
