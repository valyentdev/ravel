package tap

import (
	"errors"
	"fmt"

	"github.com/valyentdev/ravel/pkg/core"
)

func CleanupMachineTapDevice(instanceId string, machine core.Instance) error {
	errs := []error{}

	err := cleanupTapDeviceConfig(TapName(instanceId), machine)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to cleanup tap device config: %w", err))
	}

	err = deleteTap(TapName(instanceId))
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to delete tap: %w", err))
	}

	return errors.Join(errs...)
}
