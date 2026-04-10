// Package jailer provides VM isolation and sandboxing capabilities for Ravel microVMs.
// It implements a security layer that confines Cloud Hypervisor processes using:
//   - Linux namespaces (PID, mount, network)
//   - chroot/pivot_root for filesystem isolation
//   - cgroups for resource limits
//   - rlimits for process constraints
//
// The jailer ensures each microVM runs in a dedicated security sandbox, preventing
// interference between VMs and limiting the blast radius of potential security issues.
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

	"github.com/alexisbouchez/ravel/internal/cgroups"
	"github.com/vishvananda/netns"

	"golang.org/x/sys/execabs"
	"golang.org/x/sys/unix"
)

const defaultFolderPerm = 0o700

// JailerConfig specifies the isolation and resource constraints for a jailed process.
// It defines the security sandbox parameters for running Cloud Hypervisor.
type JailerConfig struct {
	Uid       int      // User ID to run the jailed process as
	Gid       int      // Group ID to run the jailed process as
	NewRoot   string   // New root directory (chroot target)
	Netns     string   // Network namespace to join
	NewPid    bool     // Create a new PID namespace
	Command   []string // Command and arguments to execute in the jail
	NoFiles   int      // Maximum number of open file descriptors (RLIMIT_NOFILE)
	Fsize     int      // Maximum file size (RLIMIT_FSIZE)
	MountProc bool     // Whether to mount /proc in the new root
	Cgroup    string   // Cgroup path for resource limits
}

// setupRLimits applies resource limits (rlimits) to the current process.
// This restricts the number of file descriptors and maximum file size
// that the jailed process can use.
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

// joinNetNs switches the current thread into the specified network namespace.
// This is used to isolate the VM's network from the host and other VMs.
// The calling goroutine is locked to the OS thread to ensure namespace
// changes don't affect other goroutines.
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

// closeFileDescriptors closes all file descriptors except stdin, stdout, and stderr.
// This prevents file descriptor leakage into the jailed process and ensures a
// clean execution environment. Uses the CLOSE_RANGE syscall for efficiency.
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
