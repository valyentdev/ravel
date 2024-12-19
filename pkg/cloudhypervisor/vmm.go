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

func NewVMMClient(socket string) (*VMM, error) {
	client, err := newCHClient(socket)
	if err != nil {
		return nil, err
	}

	vmm := &VMM{
		client: client,
	}

	return vmm, nil
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
