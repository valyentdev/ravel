package tap

import (
	"github.com/valyentdev/ravel/core/instance"
)

func PrepareInstanceTapDevice(id string, config instance.NetworkingConfig, uid, gid int) (string, error) {
	tapName := config.TapDevice
	err := createTap(tapName, uint32(uid), uint32(gid))
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
