package common

import (
	"slices"

	"github.com/alexisbouchez/ravel/core/instance"
	"github.com/alexisbouchez/ravel/initd"
)

// GetVirtioDiskPath returns the virtio block device path for the given index.
// Index 0 = /dev/vda, index 1 = /dev/vdb, etc.
func GetVirtioDiskPath(index int) string {
	var suffix []rune

	for index >= 0 {
		char := 'a' + (index % 26)
		suffix = append(suffix, rune(char))
		index = (index / 26) - 1
	}

	slices.Reverse(suffix)

	return "/dev/vd" + string(suffix)
}

// GetAdditionalMounts converts instance mounts to initd mount configurations.
func GetAdditionalMounts(mounts []instance.Mount) []initd.Mounts {
	if len(mounts) == 0 {
		return nil
	}
	additionalMounts := make([]initd.Mounts, 0, len(mounts))
	for i, m := range mounts {
		additionalMounts = append(additionalMounts, initd.Mounts{
			DevicePath: GetVirtioDiskPath(i + 1), // rootfs is at index 0 so we start at 1
			MountPath:  m.Path,
		})
	}

	return additionalMounts
}
