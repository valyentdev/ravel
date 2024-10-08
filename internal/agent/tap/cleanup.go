package tap

import (
	"errors"
	"fmt"

	"github.com/valyentdev/ravel/internal/networking"
)

func CleanupInstanceTapDevice(instanceId string, config networking.LocalConfig) error {
	errs := []error{}

	err := cleanupTapDeviceConfig(TapName(instanceId), config)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to cleanup tap device config: %w", err))
	}

	err = deleteTap(TapName(instanceId))
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to delete tap: %w", err))
	}

	return errors.Join(errs...)
}
