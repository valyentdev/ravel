package tap

import "github.com/valyentdev/ravel/pkg/core"

func PrepareMachineTapDevice(instanceId string, machine core.Instance) (string, error) {
	tapName := TapName(instanceId)
	err := createTap(tapName)
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			deleteTap(tapName)
		}
	}()

	if err := configureTapDevice(tapName, machine); err != nil {
		return "", err
	}

	return tapName, nil
}
