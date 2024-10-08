package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/tmaxmax/go-sse"
	"github.com/valyentdev/ravel/pkg/core"
)

func getBody(body any) (io.Reader, error) {
	if body == nil {
		return nil, nil
	}

	if b, ok := body.(io.Reader); ok {
		return b, nil
	}

	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(b), nil
}

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient(baseURL string, c *http.Client, defaults ...ReqOpt) *Client {
	return &Client{client: c, baseURL: baseURL}
}

type reqOptions struct {
	header http.Header
	query  url.Values
}

type ReqOpt func(*reqOptions)

func WithHeader(key, value string) ReqOpt {
	return func(o *reqOptions) {
		o.header.Set(key, value)
	}
}

func WithQuery(key, value string) ReqOpt {
	return func(o *reqOptions) {
		o.query.Add(key, value)
	}
}

func buildHttpRequest(ctx context.Context, method, path string, body any, opts ...ReqOpt) (*http.Request, error) {
	b, err := getBody(body)
	if err != nil {
		return nil, err
	}

	o := &reqOptions{
		header: make(http.Header),
		query:  make(url.Values),
	}

	for _, opt := range opts {
		opt(o)
	}

	req, err := http.NewRequestWithContext(ctx, method, path, b)
	if err != nil {
		return nil, err
	}

	req.Header = o.header
	req.URL.RawQuery = o.query.Encode()

	return req, nil
}

func handleError(body io.ReadCloser) error {
	var rerr core.RavelError
	err := json.NewDecoder(body).Decode(&rerr)
	if err != nil {
		return err
	}

	return &rerr
}

func isOk(status int) bool {
	return status >= 200 && status <= 204
}

func (c *Client) do(req *http.Request, dest any) error {
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if !isOk(resp.StatusCode) {
		return handleError(resp.Body)
	}

	if dest != nil {
		err = json.NewDecoder(resp.Body).Decode(dest)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) Get(ctx context.Context, path string, dest any, opts ...ReqOpt) error {
	req, err := buildHttpRequest(ctx, http.MethodGet, c.baseURL+path, nil, opts...)
	if err != nil {
		return err
	}

	return c.do(req, dest)
}

func (c *Client) Post(ctx context.Context, path string, body, dest any, opts ...ReqOpt) error {
	req, err := buildHttpRequest(ctx, http.MethodPost, c.baseURL+path, body, opts...)
	if err != nil {
		return err
	}

	return c.do(req, dest)
}

func (c *Client) Delete(ctx context.Context, path string, opts ...ReqOpt) error {
	req, err := buildHttpRequest(ctx, http.MethodDelete, c.baseURL+path, nil, opts...)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) SSE(ctx context.Context, method string, path string, opts ...ReqOpt) (*sse.Connection, error) {
	req, err := buildHttpRequest(ctx, method, c.baseURL+path, nil, opts...)
	if err != nil {
		return nil, err
	}

	client := sse.Client{
		HTTPClient: c.client,
	}

	return client.NewConnection(req), nil
}
