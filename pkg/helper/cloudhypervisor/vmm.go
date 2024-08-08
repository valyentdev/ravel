package cloudhypervisor

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"syscall"
	"time"
)

var ErrVMMUnavailable = errors.New("vmm is not available")

type VMM struct {
	options    *vmmOpts
	client     *ClientWithResponses
	httpClient *http.Client
}

type vmmOpts struct {
	sysProcAttr               *syscall.SysProcAttr
	cloudHypervisorBinaryPath string
	args                      []string
}

type VMMOpt func(*vmmOpts) error

func WithSysProcAttr(sysProcAttr *syscall.SysProcAttr) VMMOpt {
	return func(o *vmmOpts) error {
		o.sysProcAttr = sysProcAttr
		return nil
	}
}

func WithCloudHypervisorBinaryPath(path string) VMMOpt {
	return func(o *vmmOpts) error {
		o.cloudHypervisorBinaryPath = path
		return nil
	}
}

func (vmm *VMM) StartVMM(ctx context.Context) error {
	cmd := exec.Command(vmm.options.cloudHypervisorBinaryPath, vmm.options.args...)
	if vmm.options.sysProcAttr != nil {
		cmd.SysProcAttr = vmm.options.sysProcAttr
	}

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start cloud-hypervisor process: %w", err)
	}

	return nil
}

func NewVMM(socket string, opts ...VMMOpt) (*VMM, error) {
	client, conn, err := newCHClient(socket)
	if err != nil {
		return nil, err
	}

	options := vmmOpts{
		cloudHypervisorBinaryPath: "cloud-hypervisor",
		args:                      []string{"--api-socket", socket},
	}
	for _, opt := range opts {
		if err := opt(&options); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	vmm := &VMM{
		httpClient: conn,
		client:     client,
		options:    &options,
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

func newCHClient(socket string) (*ClientWithResponses, *http.Client, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socket)
			},
		},
	}

	client, err := NewClientWithResponses("http://localhost/api/v1", WithHTTPClient(httpClient))
	if err != nil {
		return nil, nil, err
	}

	return client, httpClient, nil
}
