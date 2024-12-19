package jailer

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/valyentdev/ravel/internal/cgroups"
	"github.com/vishvananda/netns"

	"golang.org/x/sys/execabs"
	"golang.org/x/sys/unix"
)

const defaultFolderPerm = 0o700

type JailerConfig struct {
	Uid       int
	Gid       int
	NewRoot   string
	Netns     string
	NewPid    bool
	Command   []string
	NoFiles   int
	Fsize     int
	MountProc bool
	Cgroup    string
}

func setupRLimits(config *JailerConfig) error {
	if config.NoFiles != 0 {
		err := unix.Setrlimit(unix.RLIMIT_NOFILE, &unix.Rlimit{
			Cur: uint64(config.NoFiles),
			Max: uint64(config.NoFiles),
		})
		if err != nil {
			return fmt.Errorf("failed to set rlimit nofile: %w", err)
		}
	}

	if config.Fsize != 0 {
		err := unix.Setrlimit(unix.RLIMIT_FSIZE, &unix.Rlimit{
			Cur: uint64(config.Fsize),
			Max: uint64(config.Fsize),
		})
		if err != nil {
			return fmt.Errorf("failed to set rlimit fsize: %w", err)
		}
	}

	return nil
}

func joinNetNs(ns string) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	h, err := netns.GetFromName(ns)
	if err != nil {
		return err
	}

	err = netns.Set(h)
	if err != nil {
		return err
	}

	return nil
}

func closeFileDescriptors() error {
	_, _, err := unix.Syscall(unix.SYS_CLOSE_RANGE, 3, math.MaxUint32, unix.CLOSE_RANGE_UNSHARE)
	if err != 0 {
		return err
	}
	return nil
}

func sanitizeProcess() error {
	os.Clearenv()
	err := closeFileDescriptors()
	if err != nil {
		return err
	}

	return nil
}

func mountProc(uid, gid int) error {
	err := mkdirAndChown("/proc", uid, gid, defaultFolderPerm)
	if err != nil {
		return fmt.Errorf("failed to create /proc: %w", err)
	}

	err = unix.Mount("proc", "/proc", "proc", 0, "")
	if err != nil {
		return err
	}

	return nil
}

// for cloud-hypervisor
func populateVirtualNetDevices(gid, uid int, newRoot string) error {
	tocreate := [...]string{"/sys", "/sys/class", "/sys/class/net"}

	source := "/sys/devices/virtual/net"
	destPath := path.Join(newRoot, "/sys/class/net")
	var err error
	for _, dir := range tocreate {
		err := mkdirAndChown(path.Join(newRoot, dir), uid, gid, defaultFolderPerm)
		if err != nil {
			return fmt.Errorf("failed to create %s: %w", dir, err)
		}
	}

	entries, err := os.ReadDir(source)
	if err != nil {
		return fmt.Errorf("failed to read /sys/devices/virtual/net: %w", err)
	}

	for _, entry := range entries {
		err = mkdirAndChown(path.Join(destPath, entry.Name()), uid, gid, defaultFolderPerm)
		if err != nil {
			return fmt.Errorf("failed to create %s: %w", entry.Name(), err)
		}
	}

	return nil
}

func runJailed(config *JailerConfig) error {
	newRootAbsolute, err := filepath.Abs(config.NewRoot)
	if err != nil {
		return fmt.Errorf("failed to get absolute path of new root: %w", err)
	}

	err = populateVirtualNetDevices(config.Gid, config.Uid, newRootAbsolute)
	if err != nil {
		return fmt.Errorf("failed to populate virtual net devices: %w", err)
	}

	err = chroot(newRootAbsolute)
	if err != nil {
		return fmt.Errorf("failed to chroot: %w", err)
	}

	if config.MountProc {
		err = mountProc(config.Uid, config.Gid)
		if err != nil {
			return fmt.Errorf("failed to mount proc: %w", err)
		}
	}

	null := os.NewFile(uintptr(unix.Stdin), "/dev/null")
	fifoPath := "vm.logs"

	err = mkFifo(fifoPath, config.Uid, config.Gid, 0o600)
	if err != nil {
		return fmt.Errorf("failed to create fifo: %w", err)
	}

	fd, err := unix.Open(fifoPath, unix.O_RDWR|unix.O_NONBLOCK, 0)
	if err != nil {
		return fmt.Errorf("failed to open fifo: %w", err)
	}

	fifo := os.NewFile(uintptr(fd), fifoPath)
	defer fifo.Close()

	cmd := execabs.Cmd{
		Path:   config.Command[0],
		Args:   config.Command,
		Env:    []string{},
		Stdin:  null,
		Stderr: fifo,
		Stdout: fifo,
		SysProcAttr: &unix.SysProcAttr{
			Credential: &syscall.Credential{
				Uid: uint32(config.Uid),
				Gid: uint32(config.Gid),
			},
		},
	}

	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func execJailed(config *JailerConfig) error {
	err := sanitizeProcess()
	if err != nil {
		return fmt.Errorf("failed to sanitize process: %w", err)
	}

	err = setupRLimits(config)
	if err != nil {
		return fmt.Errorf("failed to setup rlimits: %w", err)
	}

	if config.Netns != "" {
		err := joinNetNs(config.Netns)
		if err != nil {
			return fmt.Errorf("failed to join network namespace: %w", err)
		}
	}

	cloneFlags := syscall.CLONE_NEWNS
	if config.NewPid {
		cloneFlags |= syscall.CLONE_NEWPID
	}

	if config.Cgroup != "" {
		err := cgroups.JoinCgroup(config.Cgroup)
		if err != nil {
			return fmt.Errorf("failed to join cgroup: %w", err)
		}
	}

	err = reexec(uintptr(cloneFlags))
	if err != nil {
		return err
	}

	return nil
}

func reexec(cloneFlags uintptr) error {
	cmd := exec.Command("/proc/self/exe", append([]string{"run"}, os.Args[2:]...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: cloneFlags,
	}

	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
