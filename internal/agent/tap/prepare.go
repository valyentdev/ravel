package tap

import (
	"github.com/valyentdev/ravel/internal/networking"
	"github.com/valyentdev/ravel/pkg/core"
)

func PrepareInstanceTapDevice(instance core.Instance, config networking.LocalConfig) (string, error) {
	tapName := TapName(instance.Id)
	err := createTap(tapName)
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			deleteTap(tapName)
		}
	}()

	if err := configureTapDevice(tapName, config); err != nil {
		return "", err
	}

	return tapName, nil
}
