package jailer

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"syscall"

	"golang.org/x/sys/unix"
)

type Jail struct {
	jailer string
	*jailDir
	options *options
}

func (j *Jail) Command(cmd string, args ...string) *exec.Cmd {
	return makeCmd(j.jailer, append([]string{cmd}, args...), j.options)
}

type JailConfig struct {
	Uid     int
	Gid     int
	NewRoot string
}

type Device struct {
	Path string // Path to the device in the jail
	Mode uint32 // Mode of the device
	Dev  uint64 // Device number
}

type file struct {
	Src      string // Absolute path to the source file in the host
	Dst      string // Destination file relative in the jail
	Mode     uint32
	Readonly bool
}

type options struct {
	JailConfig
	setRlimits bool
	noFiles    int
	fsize      int
	devices    []Device
	cgroup     string
	netns      string
	copyFiles  []file
	hardLinks  []file
	newPidNS   bool
	mountProc  bool
}

func WithCopyFile(src string, dst string, mode uint32) Opt {
	return func(o *options) error {
		o.copyFiles = append(o.copyFiles, file{Src: src, Dst: dst, Mode: mode})
		return nil
	}
}

func WithBinary(src string, dst string) Opt {
	return func(o *options) error {
		o.copyFiles = append(o.copyFiles, file{Src: src, Dst: src, Mode: 0700})
		return nil
	}
}

func WithHardLink(src string, dst string, readonly bool) Opt {
	return func(o *options) error {
		o.hardLinks = append(o.hardLinks, file{Src: src, Dst: dst})
		return nil
	}
}

func WithNewPidNS() Opt {
	return func(o *options) error {
		o.newPidNS = true
		return nil
	}
}

func WithNetNs(netns string) Opt {
	return func(o *options) error {
		o.netns = netns
		return nil
	}
}

func WithCgroup(cgroup string) Opt {
	return func(o *options) error {
		o.cgroup = cgroup
		return nil
	}
}

func WithResourceLimits(noFiles, fsize int) Opt {
	return func(o *options) error {
		o.setRlimits = true
		o.noFiles = noFiles
		o.fsize = fsize
		return nil
	}
}

func WithKVM() Opt {
	return func(o *options) error {
		o.devices = append(o.devices, Device{
			Path: devKVMPath,
			Mode: devKVMMode,
			Dev:  unix.Mkdev(devKVMMajor, devKVMMinor),
		})
		return nil
	}
}

func WithTUN() Opt {
	return func(o *options) error {
		o.devices = append(o.devices, Device{
			Path: devTUNPath,
			Mode: devTUNMode,
			Dev:  unix.Mkdev(devTUNMajor, devTUNMinor),
		})
		return nil
	}
}

func WithBlockDevice(device string) Opt {
	return func(o *options) error {
		stat, err := os.Stat(device)
		if err != nil {
			return err
		}
		rdev := stat.Sys().(*syscall.Stat_t).Rdev

		o.devices = append(o.devices, Device{
			Path: device,
			Mode: 0600 | unix.S_IFBLK,
			Dev:  rdev,
		})

		return nil
	}
}

func WithURandom() Opt {
	return func(o *options) error {
		o.devices = append(o.devices, Device{
			Path: devURandomPath,
			Mode: devURandomMode,
			Dev:  unix.Mkdev(devURandomMajor, devURandomMinor),
		})

		return nil
	}
}

func WithMountProc() Opt {
	return func(o *options) error {
		o.mountProc = true
		return nil
	}
}

type Opt func(*options) error

func makeOptions(jailConfig JailConfig, opts ...Opt) (*options, error) {
	o := &options{
		JailConfig: jailConfig,
	}

	for _, opt := range opts {
		err := opt(o)
		if err != nil {
			return nil, err
		}
	}

	return o, nil
}

func (j *Jail) setupDevices() error {
	err := j.Mkdir("/dev")
	if err != nil {
		return fmt.Errorf("failed to create /dev: %w", err)
	}

	for _, dev := range j.options.devices {
		dir := path.Dir(dev.Path)
		err = j.MkdirAll(dir)
		if err != nil {
			return fmt.Errorf("failed to create %s: %w", dir, err)
		}

		err = j.MknodAndOwn(dev.Path, dev.Mode, dev.Dev)
		if err != nil {
			return fmt.Errorf("failed to mknod %s: %w", dev.Path, err)
		}

	}
	return nil
}

func (j *Jail) setupFiles() error {
	for _, file := range j.options.copyFiles {
		dir := path.Dir(file.Dst)
		err := j.MkdirAll(dir)
		if err != nil {
			return fmt.Errorf("failed to create %s: %w", dir, err)
		}

		err = j.CopyFile(file.Src, file.Dst, file.Mode)
		if err != nil {
			return fmt.Errorf("failed to copy file %s to %s: %w", file.Src, file.Dst, err)
		}
	}

	for _, file := range j.options.hardLinks {
		dir := path.Dir(file.Dst)
		err := j.MkdirAll(dir)
		if err != nil {
			return fmt.Errorf("failed to create %s: %w", dir, err)
		}
		err = j.HardLink(file.Src, file.Dst, file.Readonly)
		if err != nil {
			return fmt.Errorf("failed to hard link %s to %s: %w", file.Src, file.Dst, err)
		}
	}

	return nil
}

func CleanupJailDir(dir string) error {
	return os.RemoveAll(dir)
}

func CreateJail(jailer string, jailConfig JailConfig, opts ...Opt) (*Jail, error) {
	o, err := makeOptions(jailConfig, opts...)
	if err != nil {
		return nil, err
	}

	dir, err := createJailDir(jailConfig.NewRoot, jailConfig.Uid, jailConfig.Gid)
	if err != nil {
		return nil, fmt.Errorf("failed to create jail dir: %w", err)
	}

	j := &Jail{
		jailDir: dir,
		jailer:  jailer,
		options: o,
	}

	err = j.setupDevices()
	if err != nil {
		return nil, fmt.Errorf("failed to setup devices: %w", err)
	}

	err = j.setupFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to setup files: %w", err)
	}

	return j, nil
}
