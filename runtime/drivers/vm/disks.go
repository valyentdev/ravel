package vm

import (
	"slices"

	"github.com/valyentdev/ravel/core/instance"
	"github.com/valyentdev/ravel/initd"
	"github.com/valyentdev/ravel/pkg/cloudhypervisor"
	"github.com/valyentdev/ravel/runtime/disks"
)

func getAdditionalDisks(disks []disks.Disk) []cloudhypervisor.DiskConfig {
	additionalDisks := make([]cloudhypervisor.DiskConfig, 0, len(disks))
	for _, d := range disks {
		additionalDisks = append(additionalDisks, cloudhypervisor.DiskConfig{
			Path: d.Path,
		})
	}

	return additionalDisks
}

func getAdditionalMounts(mounts []instance.Mount) []initd.Mounts {
	if len(mounts) == 0 {
		return nil
	}
	additionalMounts := make([]initd.Mounts, 0, len(mounts))
	for i, m := range mounts {
		additionalMounts = append(additionalMounts, initd.Mounts{
			DevicePath: getVirtioDiskPath(i + 1), // rootfs is at index 0 so we start at 1
			MountPath:  m.Path,
		})
	}

	return additionalMounts
}

func getVirtioDiskPath(index int) string {
	var suffix []rune

	for index >= 0 {
		char := 'a' + (index % 26)
		suffix = append(suffix, rune(char))
		index = (index / 26) - 1
	}

	slices.Reverse(suffix)

	return "/dev/vd" + string(suffix)
}
