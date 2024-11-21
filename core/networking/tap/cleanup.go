package tap

import (
	"errors"
	"fmt"

	"github.com/valyentdev/ravel/core/instance"
)

func CleanupInstanceTapDevice(id string, config instance.NetworkingConfig) error {
	tapName := config.TapDevice
	errs := []error{}

	err := cleanupTapDeviceConfig(config)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to cleanup tap device config: %w", err))
	}

	err = deleteTap(tapName)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to delete tap: %w", err))
	}

	return errors.Join(errs...)
}
