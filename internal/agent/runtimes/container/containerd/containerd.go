package containerd

import (
	"github.com/containerd/containerd/v2/client"
)

func NewContainerdClient() (*client.Client, error) {
	c, err := client.New("/var/run/containerd/containerd.sock", client.WithDefaultNamespace("default"))
	if err != nil {
		return nil, err
	}

	return c, nil
}
