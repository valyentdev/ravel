package cgroups

import (
	"os"

	"github.com/containerd/cgroups/v3/cgroup2"
)

func JoinCgroup(group string) error {
	m, err := cgroup2.Load(group)
	if err != nil {
		return err
	}

	return m.AddProc(uint64(os.Getpid()))
}
