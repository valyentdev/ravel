package firecracker

import (
	"syscall"

	"github.com/alexisbouchez/ravel/core/instance"
	"github.com/alexisbouchez/ravel/runtime/disks"
)

// VMConfig holds the configuration for a Firecracker VM.
type VMConfig struct {
	VCpus           int
	MemoryMB        int
	KernelPath      string
	InitrdPath      string
	BootArgs        string
	RootfsPath      string
	AdditionalDisks []string
	TapDevice       string
	VsockPath       string
}

// getVMConfig creates a VMConfig from an instance configuration.
func (d *Driver) getVMConfig(inst *instance.Instance, rootfs string, instanceDisks []disks.Disk) VMConfig {
	config := inst.Config

	additionalDisks := make([]string, 0, len(instanceDisks))
	for _, disk := range instanceDisks {
		additionalDisks = append(additionalDisks, disk.Path)
	}

	return VMConfig{
		VCpus:           config.Guest.VCpus,
		MemoryMB:        config.Guest.MemoryMB,
		KernelPath:      linuxKernelPath,
		InitrdPath:      initRamfsPath,
		BootArgs:        "ro console=ttyS0 reboot=k panic=1 pci=off rdinit=ravel-init",
		RootfsPath:      rootfs,
		AdditionalDisks: additionalDisks,
		TapDevice:       inst.Network.TapDevice,
		VsockPath:       vsockPath,
	}
}

// syscallSignal converts a signal name to a syscall signal number.
func syscallSignal(signal string) syscall.Signal {
	switch signal {
	case "SIGKILL":
		return syscall.SIGKILL
	case "SIGTERM":
		return syscall.SIGTERM
	case "SIGINT":
		return syscall.SIGINT
	case "SIGHUP":
		return syscall.SIGHUP
	case "SIGUSR1":
		return syscall.SIGUSR1
	case "SIGUSR2":
		return syscall.SIGUSR2
	default:
		return syscall.SIGTERM
	}
}
