package firecracker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
)

// Client is a Firecracker HTTP API client.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new Firecracker API client connected via Unix socket.
func NewClient(socketPath string) *Client {
	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
	}

	return &Client{
		httpClient: httpClient,
		baseURL:    "http://localhost",
	}
}

// doRequest performs an HTTP request to the Firecracker API.
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	return c.httpClient.Do(req)
}

// checkResponse checks if the response indicates success.
func checkResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	var apiErr Error
	if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.FaultMessage != "" {
		return fmt.Errorf("firecracker API error: %s", apiErr.FaultMessage)
	}

	return fmt.Errorf("firecracker API error: status %d, body: %s", resp.StatusCode, string(body))
}

// GetInstanceInfo retrieves information about the instance.
func (c *Client) GetInstanceInfo(ctx context.Context) (*InstanceInfo, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var info InstanceInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &info, nil
}

// PutMachineConfig sets the machine configuration.
func (c *Client) PutMachineConfig(ctx context.Context, config MachineConfig) error {
	resp, err := c.doRequest(ctx, http.MethodPut, "/machine-config", config)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return checkResponse(resp)
}

// PutBootSource sets the boot source configuration.
func (c *Client) PutBootSource(ctx context.Context, bootSource BootSource) error {
	resp, err := c.doRequest(ctx, http.MethodPut, "/boot-source", bootSource)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return checkResponse(resp)
}

// PutDrive adds or updates a block device.
func (c *Client) PutDrive(ctx context.Context, drive Drive) error {
	resp, err := c.doRequest(ctx, http.MethodPut, "/drives/"+drive.DriveID, drive)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return checkResponse(resp)
}

// PutNetworkInterface adds or updates a network interface.
func (c *Client) PutNetworkInterface(ctx context.Context, iface NetworkInterface) error {
	resp, err := c.doRequest(ctx, http.MethodPut, "/network-interfaces/"+iface.IfaceID, iface)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return checkResponse(resp)
}

// PutVsock sets the vsock device configuration.
func (c *Client) PutVsock(ctx context.Context, vsock VsockDevice) error {
	resp, err := c.doRequest(ctx, http.MethodPut, "/vsock", vsock)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return checkResponse(resp)
}

// CreateAction performs an instance action (e.g., InstanceStart).
func (c *Client) CreateAction(ctx context.Context, action InstanceActionInfo) error {
	resp, err := c.doRequest(ctx, http.MethodPut, "/actions", action)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return checkResponse(resp)
}

// PatchVMState changes the VM state (Paused or Resumed).
func (c *Client) PatchVMState(ctx context.Context, state VMState) error {
	resp, err := c.doRequest(ctx, http.MethodPatch, "/vm", state)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return checkResponse(resp)
}

// CreateSnapshot creates a snapshot of the microVM.
func (c *Client) CreateSnapshot(ctx context.Context, params SnapshotCreateParams) error {
	resp, err := c.doRequest(ctx, http.MethodPut, "/snapshot/create", params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return checkResponse(resp)
}

// LoadSnapshot loads a snapshot into the microVM.
func (c *Client) LoadSnapshot(ctx context.Context, params SnapshotLoadParams) error {
	resp, err := c.doRequest(ctx, http.MethodPut, "/snapshot/load", params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return checkResponse(resp)
}
