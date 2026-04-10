package firecracker

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/alexisbouchez/ravel/api"
	"github.com/alexisbouchez/ravel/core/instance"
	initdclient "github.com/alexisbouchez/ravel/initd/client"
	"github.com/alexisbouchez/ravel/pkg/firecracker"
	"github.com/alexisbouchez/ravel/runtime/drivers"
	"github.com/alexisbouchez/ravel/runtime/drivers/common"
)

// firecrackerVM implements drivers.InstanceTask for Firecracker.
type firecrackerVM struct {
	id                   string
	cmd                  *exec.Cmd
	vmm                  *firecracker.VMM
	runResult            *RunResult
	initClient           *initdclient.InternalClient
	successfullyShutdown atomic.Bool
	vmConfig             VMConfig
	stopRequested        bool
	waitChan             chan struct{}
}

var _ drivers.InstanceTask = (*firecrackerVM)(nil)

func (vm *firecrackerVM) Id() string {
	return vm.id
}

func newVM(id string, cmd *exec.Cmd, vmConfig VMConfig) *firecrackerVM {
	vmm := firecracker.NewVMMClient(getAPISocketPath(id))
	client := initdclient.NewInternalClient(getVsockPath(id))

	return &firecrackerVM{
		id:         id,
		cmd:        cmd,
		vmConfig:   vmConfig,
		vmm:        vmm,
		waitChan:   make(chan struct{}),
		initClient: client,
	}
}

// Start implements drivers.InstanceTask.
func (vm *firecrackerVM) Start(ctx context.Context) error {
	err := vm.cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start firecracker for machine %q: %w", vm.Id(), err)
	}
	defer func() {
		if err != nil {
			vm.shutdown(ctx)
		}
	}()

	err = vm.vmm.WaitReady(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait for firecracker to be ready for machine %q: %w", vm.Id(), err)
	}

	// Configure the VM
	if err = vm.configureVM(ctx); err != nil {
		return fmt.Errorf("failed to configure VM for machine %q: %w", vm.Id(), err)
	}

	// Start the VM
	if err = vm.vmm.Start(ctx); err != nil {
		return fmt.Errorf("failed to start VM for machine %q: %w", vm.Id(), err)
	}

	go vm.run()

	return nil
}

// configureVM configures the Firecracker VM via the API.
func (vm *firecrackerVM) configureVM(ctx context.Context) error {
	// Set machine config (CPU + memory)
	if err := vm.vmm.SetMachineConfig(ctx, vm.vmConfig.VCpus, vm.vmConfig.MemoryMB); err != nil {
		return fmt.Errorf("failed to set machine config: %w", err)
	}

	// Set boot source (kernel + initrd)
	if err := vm.vmm.SetBootSource(ctx, vm.vmConfig.KernelPath, vm.vmConfig.InitrdPath, vm.vmConfig.BootArgs); err != nil {
		return fmt.Errorf("failed to set boot source: %w", err)
	}

	// Add root drive
	if err := vm.vmm.AddDrive(ctx, "rootfs", vm.vmConfig.RootfsPath, true, false); err != nil {
		return fmt.Errorf("failed to add root drive: %w", err)
	}

	// Add additional drives
	for i, diskPath := range vm.vmConfig.AdditionalDisks {
		driveID := fmt.Sprintf("disk%d", i+1)
		if err := vm.vmm.AddDrive(ctx, driveID, diskPath, false, false); err != nil {
			return fmt.Errorf("failed to add drive %s: %w", driveID, err)
		}
	}

	// Add network interface
	if err := vm.vmm.AddNetworkInterface(ctx, "eth0", vm.vmConfig.TapDevice); err != nil {
		return fmt.Errorf("failed to add network interface: %w", err)
	}

	// Set vsock
	if err := vm.vmm.SetVsock(ctx, 3, vm.vmConfig.VsockPath); err != nil {
		return fmt.Errorf("failed to set vsock: %w", err)
	}

	return nil
}

// StartFromSnapshot implements drivers.InstanceTask.
func (vm *firecrackerVM) StartFromSnapshot(ctx context.Context, globalSnapshotPath, jailSnapshotPath string) error {
	bootStart := time.Now()
	slog.Info("StartFromSnapshot called", "id", vm.id, "globalPath", globalSnapshotPath, "jailPath", jailSnapshotPath)

	// Copy snapshot from global storage into the jail
	jailHostPath := getInstanceDir(vm.id) + jailSnapshotPath
	slog.Info("copying snapshot", "from", globalSnapshotPath, "to", jailHostPath)
	if err := copyDirWithLinks(globalSnapshotPath, jailHostPath); err != nil {
		slog.Error("failed to copy snapshot", "error", err)
		return fmt.Errorf("failed to copy snapshot to jail: %w", err)
	}

	// Chown to ravel-jailer user so Firecracker can read
	jailerUid, jailerGid, err := common.SetupRavelJailerUser()
	if err != nil {
		return fmt.Errorf("failed to get jailer user: %w", err)
	}
	if err := chownRecursive(jailHostPath, jailerUid, jailerGid); err != nil {
		return fmt.Errorf("failed to chown snapshot directory: %w", err)
	}

	slog.Debug("snapshot copied", "from", globalSnapshotPath, "to", jailHostPath)

	err = vm.cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start firecracker for machine %q: %w", vm.Id(), err)
	}
	defer func() {
		if err != nil {
			vm.shutdown(ctx)
		}
	}()

	err = vm.vmm.WaitReady(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait for firecracker to be ready for machine %q: %w", vm.Id(), err)
	}

	// Load the snapshot
	snapshotFile := jailSnapshotPath + "/snapshot"
	memFile := jailSnapshotPath + "/mem"
	if err = vm.vmm.LoadSnapshot(ctx, snapshotFile, memFile, true); err != nil {
		return fmt.Errorf("failed to load snapshot for machine %q: %w", vm.Id(), err)
	}

	bootDuration := time.Since(bootStart).Seconds()
	slog.Info("VM restored from snapshot", "id", vm.id, "duration_ms", bootDuration*1000)

	go vm.run()

	return nil
}

// Snapshot implements drivers.InstanceTask.
func (vm *firecrackerVM) Snapshot(ctx context.Context, path string) error {
	// Create the directory on the host
	hostPath := getInstanceDir(vm.id) + path
	if err := os.MkdirAll(hostPath, 0755); err != nil {
		return fmt.Errorf("failed to create snapshot directory: %w", err)
	}

	// Chown to ravel-jailer user
	jailerUid, jailerGid, err := common.SetupRavelJailerUser()
	if err != nil {
		return fmt.Errorf("failed to get jailer user: %w", err)
	}
	if err := os.Chown(hostPath, jailerUid, jailerGid); err != nil {
		return fmt.Errorf("failed to chown snapshot directory: %w", err)
	}

	// Create snapshot (Firecracker requires path relative to its chroot)
	snapshotFile := path + "/snapshot"
	memFile := path + "/mem"
	if err := vm.vmm.Snapshot(ctx, snapshotFile, memFile); err != nil {
		return fmt.Errorf("failed to create snapshot: %w", err)
	}

	return nil
}

// Restore implements drivers.InstanceTask.
func (vm *firecrackerVM) Restore(ctx context.Context, path string) error {
	snapshotFile := path + "/snapshot"
	memFile := path + "/mem"
	if err := vm.vmm.LoadSnapshot(ctx, snapshotFile, memFile, true); err != nil {
		return fmt.Errorf("failed to restore from snapshot: %w", err)
	}

	return nil
}

// Signal implements drivers.InstanceTask.
func (vm *firecrackerVM) Signal(ctx context.Context, signal string) error {
	sig := syscallSignal(signal)

	err := vm.initClient.Signal(ctx, int(sig))
	if err != nil {
		return fmt.Errorf("failed to send signal to init: %w", err)
	}

	return nil
}

// Exec implements drivers.InstanceTask.
func (vm *firecrackerVM) Exec(ctx context.Context, cmd []string, timeout time.Duration) (*api.ExecResult, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	res, err := vm.initClient.Exec(timeoutCtx, api.ExecOptions{
		Cmd:       cmd,
		TimeoutMs: int(timeout.Milliseconds()),
	})
	if err != nil {
		return nil, err
	}

	return &api.ExecResult{
		Stderr:   res.Stderr,
		Stdout:   res.Stdout,
		ExitCode: int(res.ExitCode),
	}, nil
}

// Stop implements drivers.InstanceTask.
func (vm *firecrackerVM) Stop(ctx context.Context, signal string) error {
	vm.stopRequested = true
	err := vm.Signal(ctx, signal)
	if err != nil {
		return fmt.Errorf("failed to send signal to init: %w", err)
	}

	return nil
}

// Shutdown implements drivers.InstanceTask.
func (vm *firecrackerVM) Shutdown(ctx context.Context) error {
	return vm.shutdown(ctx)
}

func (vm *firecrackerVM) shutdown(ctx context.Context) error {
	if vm.successfullyShutdown.Load() {
		return nil
	}

	// Try graceful shutdown via Ctrl+Alt+Del
	if err := vm.vmm.SendCtrlAltDel(ctx); err != nil {
		slog.Debug("failed to send Ctrl+Alt+Del", "error", err)
	}

	// Kill the process if it's still running
	if vm.cmd != nil && vm.cmd.Process != nil {
		if err := vm.cmd.Process.Kill(); err != nil {
			slog.Debug("failed to kill firecracker process", "error", err)
		}
	}

	vm.successfullyShutdown.Store(true)
	return nil
}

// Run implements drivers.InstanceTask.
func (vm *firecrackerVM) Run() instance.ExitResult {
	<-vm.waitChan
	return instance.ExitResult{
		Success:   vm.runResult.ExitCode == 0,
		ExitCode:  vm.runResult.ExitCode,
		ExitedAt:  vm.runResult.ExitedAt,
		Requested: vm.runResult.HasBeenStopped,
	}
}

// WaitExit implements drivers.InstanceTask.
func (vm *firecrackerVM) WaitExit(ctx context.Context) bool {
	for {
		select {
		case <-ctx.Done():
			return false
		case <-vm.waitChan:
			return true
		}
	}
}

// RunResult holds the result of a VM run.
type RunResult struct {
	HasBeenStopped bool
	VMExited       bool
	InitFailed     bool
	ProcessExited  bool
	ExitCode       int
	ExitedAt       time.Time
}

func (vm *firecrackerVM) run() {
	slog.Debug("vm run")
	defer close(vm.waitChan)
	defer vm.Shutdown(context.Background())
	result := &RunResult{
		ExitCode: -1,
	}
	defer func() {
		if vm.stopRequested {
			result.HasBeenStopped = true
		}
		vm.runResult = result
	}()

	started := false

	for i := 0; i < 5; i++ {
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		err := vm.initClient.HealthCheck(timeoutCtx)
		cancel()
		if err == nil {
			started = true
			break
		}
		slog.Debug("waiting for init to start", "err", err, "id", vm.id)
		time.Sleep(100 * time.Millisecond)
	}

	if !started {
		result.InitFailed = true
		result.ExitedAt = time.Now()
		vm.runResult = result
		return
	}

	slog.Debug("vm has started")

	retryCount := 0

RETRY:
	exitResult, err := vm.initClient.Wait(context.Background())
	if err != nil {
		err := vm.initClient.HealthCheck(context.Background())
		if err != nil {
			result.VMExited = true
			result.ExitedAt = time.Now()
			return
		}
		if retryCount < 10 {
			slog.Debug("retrying to wait for init to exit")
			retryCount++
		} else {
			slog.Debug("failed to wait for init to exit", "err", err)
			result.InitFailed = true
			result.ExitedAt = time.Now()
			return
		}
		goto RETRY
	}

	result.ProcessExited = true
	result.ExitCode = exitResult.ExitCode
	slog.Debug("init process exited", "exitCode", exitResult.ExitCode)
}

func (vm *firecrackerVM) recover() bool {
	ok := false

	// Check if Firecracker process is still running
	info, err := vm.vmm.GetInfo(context.Background())
	if err != nil {
		slog.Debug("failed to get VM info", "err", err)
		return ok
	}

	if info.State != "Running" {
		slog.Debug("VM is not running", "state", info.State)
		vm.shutdown(context.Background())
		return ok
	}

	// Reinitialize the init client
	vm.initClient = initdclient.NewInternalClient(getVsockPath(vm.id))
	vm.vmm = firecracker.NewVMMClient(getAPISocketPath(vm.id))

	go vm.run()

	return true
}

// Helper functions

func copyDirWithLinks(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := src + "/" + entry.Name()
		dstPath := dst + "/" + entry.Name()

		if entry.IsDir() {
			if err := copyDirWithLinks(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Try hard link first (instant, no copy), fall back to copy
			if err := os.Link(srcPath, dstPath); err != nil {
				if err := copyFile(srcPath, dstPath); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = dstFile.ReadFrom(srcFile)
	return err
}

func chownRecursive(path string, uid, gid int) error {
	return filepath.Walk(path, func(name string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return os.Chown(name, uid, gid)
	})
}
