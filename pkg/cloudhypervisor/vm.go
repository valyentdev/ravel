package cloudhypervisor

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

func (v *VMM) CreateVM(ctx context.Context, config VmConfig) error {
	res, err := v.client.CreateVMWithResponse(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to create vm: %w", err)
	}

	if res.StatusCode() != http.StatusNoContent {
		return fmt.Errorf("failed to create vm: %s", string(res.Body))
	}

	return nil
}

func (v *VMM) BootVM(ctx context.Context) error {
	res, err := v.client.BootVMWithResponse(ctx)
	if err != nil {
		return fmt.Errorf("failed to boot vm: %w", err)
	}

	if res.StatusCode() != http.StatusNoContent {
		return fmt.Errorf("failed to boot vm: %s", string(res.Body))
	}

	return nil
}

func (v *VMM) ShutdownVM(ctx context.Context) error {
	res, err := v.client.ShutdownVMWithResponse(ctx)
	if err != nil {
		return fmt.Errorf("failed to shutdown vm: %w", err)
	}

	if res.StatusCode() != http.StatusNoContent {
		return fmt.Errorf("failed to shutdown vm: %s", string(res.Body))
	}

	return nil
}

func (v *VMM) TriggerPowerButton(ctx context.Context) error {
	res, err := v.client.PowerButtonVMWithResponse(ctx)
	if err != nil {
		return fmt.Errorf("failed to power button: %w", err)
	}

	if res.StatusCode() != http.StatusNoContent {
		return fmt.Errorf("failed to power button: %s", string(res.Body))
	}

	return nil
}

func (v *VMM) VMInfo(ctx context.Context) (*VmInfo, error) {
	res, err := v.client.GetVmInfoWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get vm info: %w", err)
	}

	if res.JSON200 == nil {
		return nil, errors.New("failed to get vm info")
	}

	return res.JSON200, nil
}
