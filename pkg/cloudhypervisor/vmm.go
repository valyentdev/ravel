package cloudhypervisor

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

var ErrVMMUnavailable = errors.New("vmm is not available")

type VMM struct {
	client *ClientWithResponses
}

type VMMConfig struct {
	CloudHypervisorBinaryPath string
	Socket                    string
	AdditionalArgs            []string
}

func NewVMMClient(socket string) *VMM {
	client, _ := newCHClient(socket) // no error is possible here
	vmm := &VMM{
		client: client,
	}

	return vmm
}

func (v *VMM) ShutdownVMM(ctx context.Context) error {
	res, err := v.client.ShutdownVMMWithResponse(ctx)
	if err != nil {
		return fmt.Errorf("failed to shutdown vmm: %w", err)
	}

	if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to shutdown vmm: %s", string(res.Body))
	}

	return nil
}

func (v *VMM) WaitReady(ctx context.Context) error {
	var err error
	for i := 0; i < 50; i++ {
		ping, err := v.client.GetVmmPingWithResponse(ctx)
		if err == nil && ping.JSON200 != nil {
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return fmt.Errorf("vmm is not ready after 500ms, err: %w", err)
}

func (v *VMM) PingVMM(ctx context.Context) (VmmPingResponse, error) {
	res, err := v.client.GetVmmPingWithResponse(ctx)
	if err != nil {
		return VmmPingResponse{}, ErrVMMUnavailable
	}

	if res.JSON200 == nil {
		return VmmPingResponse{}, fmt.Errorf("failed to ping vmm: %s", string(res.Body))
	}

	return *res.JSON200, nil
}

// PauseVM pauses the virtual machine.
func (v *VMM) PauseVM(ctx context.Context) (*http.Response, error) {
	res, err := v.client.PauseVMWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to pause VM: %w", err)
	}

	if res.StatusCode() != http.StatusOK && res.StatusCode() != http.StatusNoContent {
		return nil, fmt.Errorf("failed to pause VM: %s", string(res.Body))
	}

	return res.HTTPResponse, nil
}

// ResumeVM resumes the virtual machine.
func (v *VMM) ResumeVM(ctx context.Context) (*http.Response, error) {
	res, err := v.client.ResumeVMWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to resume VM: %w", err)
	}

	if res.StatusCode() != http.StatusOK && res.StatusCode() != http.StatusNoContent {
		return nil, fmt.Errorf("failed to resume VM: %s", string(res.Body))
	}

	return res.HTTPResponse, nil
}

// PutVmSnapshot creates a snapshot of the VM for fast restore.
func (v *VMM) PutVmSnapshot(ctx context.Context, config VmSnapshotConfig) (*http.Response, error) {
	res, err := v.client.PutVmSnapshotWithResponse(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to snapshot VM: %w", err)
	}

	if res.StatusCode() != http.StatusOK && res.StatusCode() != http.StatusNoContent {
		return nil, fmt.Errorf("failed to snapshot VM: %s", string(res.Body))
	}

	return res.HTTPResponse, nil
}

// PutVmRestore restores the VM from a snapshot.
func (v *VMM) PutVmRestore(ctx context.Context, config RestoreConfig) (*http.Response, error) {
	res, err := v.client.PutVmRestoreWithResponse(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to restore VM: %w", err)
	}

	if res.StatusCode() != http.StatusOK && res.StatusCode() != http.StatusNoContent {
		return nil, fmt.Errorf("failed to restore VM: %s", string(res.Body))
	}

	return res.HTTPResponse, nil
}

func newCHClient(socket string) (*ClientWithResponses, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socket)
			},
		},
	}

	client, err := NewClientWithResponses("http://localhost/api/v1", WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}

	return client, nil
}
