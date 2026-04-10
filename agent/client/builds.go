package agentclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"

	"github.com/alexisbouchez/ravel/api"
	"github.com/alexisbouchez/ravel/api/errdefs"
	"github.com/alexisbouchez/ravel/internal/httpclient"
)

// CreateBuildOptions contains options for creating a new build
type CreateBuildOptions struct {
	Namespace  string
	ImageName  string
	Tag        string
	Registry   string
	Dockerfile string
	Target     string
	NoCache    bool
	Context    io.Reader // tar.gz stream
}

// CreateBuild starts a new image build on the agent
func (a *AgentClient) CreateBuild(ctx context.Context, opts CreateBuildOptions) (*api.Build, error) {
	// Create multipart form for file only
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add context file
	part, err := writer.CreateFormFile("context", "context.tar.gz")
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, opts.Context); err != nil {
		return nil, fmt.Errorf("failed to copy context: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Build URL with query parameters
	params := url.Values{}
	params.Set("namespace", opts.Namespace)
	params.Set("image_name", opts.ImageName)
	params.Set("registry", opts.Registry)
	if opts.Tag != "" {
		params.Set("tag", opts.Tag)
	}
	if opts.Dockerfile != "" {
		params.Set("dockerfile", opts.Dockerfile)
	}
	if opts.Target != "" {
		params.Set("target", opts.Target)
	}
	if opts.NoCache {
		params.Set("no_cache", "true")
	}
	reqURL := a.client.BaseURL() + "/builds?" + params.Encode()

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	resp, err := a.client.HTTPClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, errdefs.FromHTTPResponse(resp)
	}

	var build api.Build
	if err := json.NewDecoder(resp.Body).Decode(&build); err != nil {
		return nil, err
	}

	return &build, nil
}

// GetBuild gets the status of a build
func (a *AgentClient) GetBuild(ctx context.Context, id string) (*api.Build, error) {
	var build api.Build
	err := a.client.Get(ctx, "/builds/"+id, &build)
	if err != nil {
		return nil, err
	}
	return &build, nil
}

// ListBuilds lists builds on the agent
func (a *AgentClient) ListBuilds(ctx context.Context, namespace string, limit int) ([]*api.Build, error) {
	var builds []*api.Build
	opts := []httpclient.ReqOpt{}
	if namespace != "" {
		opts = append(opts, httpclient.WithQuery("namespace", namespace))
	}
	if limit > 0 {
		opts = append(opts, httpclient.WithQuery("limit", strconv.Itoa(limit)))
	}
	err := a.client.Get(ctx, "/builds", &builds, opts...)
	if err != nil {
		return nil, err
	}
	return builds, nil
}

// GetBuildLogs streams build logs from the agent
func (a *AgentClient) GetBuildLogs(ctx context.Context, id string, follow bool) (io.ReadCloser, error) {
	path := "/builds/" + id + "/logs"
	opts := []httpclient.ReqOpt{}
	if follow {
		opts = append(opts, httpclient.WithQuery("follow", "true"))
	}
	return a.client.RawGet(ctx, path, opts...)
}

// CancelBuild cancels an in-progress build
func (a *AgentClient) CancelBuild(ctx context.Context, id string) error {
	return a.client.Post(ctx, "/builds/"+id+"/cancel", nil)
}
