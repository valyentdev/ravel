package container

import (
	"fmt"
)

func socketPath(id string) string {
	return fmt.Sprintf("/tmp/ravel/%s.sock", id)
}

func vsockPath(id string) string {
	return fmt.Sprintf("/tmp/ravel/%s.vsock", id)
}

func getInstancePath(instanceId string) string {
	return "/etc/ravel/instances/" + instanceId
}

func getInitrdPath(instanceId string) string {
	return getInstancePath(instanceId) + "/initrd"
}
