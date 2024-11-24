package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/valyentdev/ravel/core/errdefs"
)

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
	body   io.Reader
}

type ReqOpt func(*reqOptions) error

func WithHeader(key, value string) ReqOpt {
	return func(o *reqOptions) error {
		o.header.Set(key, value)
		return nil
	}
}

func WithQuery(key, value string) ReqOpt {
	return func(o *reqOptions) error {
		o.query.Add(key, value)
		return nil
	}
}

func WithBody(body io.Reader) ReqOpt {
	return func(o *reqOptions) error {
		o.body = body
		return nil
	}
}

func WithJSONBody(body any) ReqOpt {
	return func(o *reqOptions) error {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}

		o.body = bytes.NewReader(b)
		o.header.Set("Content-Type", "application/json")

		return nil
	}
}

func buildHttpRequest(ctx context.Context, method, path string, opts ...ReqOpt) (*http.Request, error) {
	o := &reqOptions{
		header: make(http.Header),
		query:  make(url.Values),
	}
	for _, opt := range opts {
		opt(o)
	}

	req, err := http.NewRequestWithContext(ctx, method, path, o.body)
	if err != nil {
		return nil, err
	}

	req.Header = o.header
	req.URL.RawQuery = o.query.Encode()

	return req, nil
}

func handleError(body io.ReadCloser) error {
	var rerr errdefs.RavelError
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
	req, err := buildHttpRequest(ctx, http.MethodGet, c.baseURL+path, opts...)
	if err != nil {
		return err
	}

	return c.do(req, dest)
}

func (c *Client) Post(ctx context.Context, path string, dest any, opts ...ReqOpt) error {
	req, err := buildHttpRequest(ctx, http.MethodPost, c.baseURL+path, opts...)
	if err != nil {
		return err
	}

	return c.do(req, dest)
}

func (c *Client) Delete(ctx context.Context, path string, opts ...ReqOpt) error {
	req, err := buildHttpRequest(ctx, http.MethodDelete, c.baseURL+path, opts...)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) RawGet(ctx context.Context, path string, opts ...ReqOpt) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if !isOk(resp.StatusCode) {
		defer resp.Body.Close()
		return nil, handleError(resp.Body)
	}

	return resp.Body, nil
}
