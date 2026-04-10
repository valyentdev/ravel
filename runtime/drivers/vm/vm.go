package vm

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/alexisbouchez/ravel/api"
	"github.com/alexisbouchez/ravel/core/instance"
	"github.com/alexisbouchez/ravel/initd/client"
	"github.com/alexisbouchez/ravel/pkg/cloudhypervisor"
	"github.com/alexisbouchez/ravel/runtime/drivers"
	"github.com/alexisbouchez/ravel/runtime/drivers/common"
)

type vm struct {
	id                     string
	cmd                    *exec.Cmd
	vmm                    *cloudhypervisor.VMM
	runResult              *RunResult
	initClient             *client.InternalClient
	successFullyShutdowned atomic.Bool
	vmConfig               cloudhypervisor.VmConfig
	stopRequested          bool
	waitChan               chan struct{}
}

var _ drivers.InstanceTask = (*vm)(nil)

func (vm *vm) Id() string {
	return vm.id
}

func newVM(id string, cmd *exec.Cmd, vmConfig cloudhypervisor.VmConfig) *vm {
	vmm := cloudhypervisor.NewVMMClient(getAPISocketPath(id))

	client := client.NewInternalClient(getVsockPath(id))

	return &vm{
		id:         id,
		cmd:        cmd,
		vmConfig:   vmConfig,
		vmm:        vmm,
		waitChan:   make(chan struct{}),
		initClient: client,
	}
}

func (vm *vm) Start(ctx context.Context) error {
	err := vm.cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start vmm for machine %q: %w", vm.Id(), err)
	}
	defer func() {
		if err != nil {
			vm.vmm.ShutdownVMM(ctx)
		}
	}()

	err = vm.vmm.WaitReady(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait for vmm to be ready for machine %q: %w", vm.Id(), err)
	}

	err = vm.vmm.CreateVM(ctx, vm.vmConfig)
	if err != nil {
		return fmt.Errorf("failed to create vm for machine %q: %w", vm.Id(), err)
	}

	err = vm.vmm.BootVM(ctx)
	if err != nil {
		return fmt.Errorf("failed to boot vm for machine %q: %w", vm.Id(), err)
	}

	go vm.run()

	return nil

}

// StartFromSnapshot starts the VM by restoring from a snapshot instead of cold booting.
// This enables sub-100ms cold starts for AI sandbox workloads.
// globalSnapshotPath is the source path on the host (e.g., /var/lib/ravel/global-snapshots/instance-id/snap-1)
// jailSnapshotPath is the jail-relative path (e.g., /snapshots/snap-1)
func (vm *vm) StartFromSnapshot(ctx context.Context, globalSnapshotPath, jailSnapshotPath string) error {
	bootStart := time.Now()
	slog.Info("StartFromSnapshot called", "id", vm.id, "globalPath", globalSnapshotPath, "jailPath", jailSnapshotPath)

	// Copy snapshot from global storage into the jail
	jailHostPath := getInstanceDir(vm.id) + jailSnapshotPath
	slog.Info("copying snapshot", "from", globalSnapshotPath, "to", jailHostPath)
	if err := copyDirWithLinks(globalSnapshotPath, jailHostPath); err != nil {
		slog.Error("failed to copy snapshot", "error", err)
		return fmt.Errorf("failed to copy snapshot to jail: %w", err)
	}
	slog.Info("snapshot copied successfully")

	// Patch config.json with current device paths (rootfs, tap device)
	if err := vm.patchSnapshotConfig(jailHostPath); err != nil {
		return fmt.Errorf("failed to patch snapshot config: %w", err)
	}

	// Chown to ravel-jailer user so CloudHypervisor can read
	jailerUid, jailerGid, err := common.SetupRavelJailerUser()
	if err != nil {
		return fmt.Errorf("failed to get jailer user: %w", err)
	}
	if err := chownRecursive(jailHostPath, jailerUid, jailerGid); err != nil {
		return fmt.Errorf("failed to chown snapshot directory: %w", err)
	}

	slog.Debug("snapshot copied and patched", "from", globalSnapshotPath, "to", jailHostPath)

	err = vm.cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start vmm for machine %q: %w", vm.Id(), err)
	}
	defer func() {
		if err != nil {
			vm.vmm.ShutdownVMM(ctx)
		}
	}()

	err = vm.vmm.WaitReady(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait for vmm to be ready for machine %q: %w", vm.Id(), err)
	}

	// Restore from snapshot instead of create+boot
	snapshotUrl := "file://" + jailSnapshotPath
	prefault := true
	_, err = vm.vmm.PutVmRestore(ctx, cloudhypervisor.RestoreConfig{
		SourceUrl: snapshotUrl,
		Prefault:  &prefault,
	})
	if err != nil {
		return fmt.Errorf("failed to restore vm from snapshot for machine %q: %w", vm.Id(), err)
	}

	// Resume the restored VM
	_, err = vm.vmm.ResumeVM(ctx)
	if err != nil {
		return fmt.Errorf("failed to resume vm after restore for machine %q: %w", vm.Id(), err)
	}

	bootDuration := time.Since(bootStart).Seconds()
	slog.Info("VM restored from snapshot", "id", vm.id, "duration_ms", bootDuration*1000)

	go vm.run()

	return nil
}

// patchSnapshotConfig updates the snapshot config.json with current device paths
func (vm *vm) patchSnapshotConfig(snapshotPath string) error {
	configPath := snapshotPath + "/config.json"

	// Read the config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config.json: %w", err)
	}

	// Parse as generic JSON to preserve structure
	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config.json: %w", err)
	}

	// Update disk path - get the current rootfs from vmConfig
	if disks, ok := config["disks"].([]interface{}); ok && len(disks) > 0 {
		if disk, ok := disks[0].(map[string]interface{}); ok {
			// Get the new rootfs path from vmConfig
			if vm.vmConfig.Disks != nil && len(*vm.vmConfig.Disks) > 0 {
				newRootfs := (*vm.vmConfig.Disks)[0].Path
				oldRootfs := disk["path"]
				disk["path"] = newRootfs
				slog.Debug("patched rootfs path", "old", oldRootfs, "new", newRootfs)
			}
		}
	}

	// Update tap device name
	if nets, ok := config["net"].([]interface{}); ok && len(nets) > 0 {
		if net, ok := nets[0].(map[string]interface{}); ok {
			// Get the new tap device from vmConfig
			if vm.vmConfig.Net != nil && len(*vm.vmConfig.Net) > 0 {
				newTap := (*vm.vmConfig.Net)[0].Tap
				if newTap != nil {
					oldTap := net["tap"]
					net["tap"] = *newTap
					slog.Debug("patched tap device", "old", oldTap, "new", *newTap)
				}
			}
		}
	}

	// Write back the modified config
	newData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config.json: %w", err)
	}

	if err := os.WriteFile(configPath, newData, 0600); err != nil {
		return fmt.Errorf("failed to write config.json: %w", err)
	}

	return nil
}

// copyDirWithLinks copies a directory from src to dst, using hard links for large files
// config.json is always copied (not linked) since we need to modify it
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
			// Always copy config.json since we need to modify it
			if entry.Name() == "config.json" {
				if err := copyFile(srcPath, dstPath); err != nil {
					return err
				}
				continue
			}

			// Try hard link first (instant, no copy), fall back to copy
			if err := os.Link(srcPath, dstPath); err != nil {
				// Hard link failed (maybe cross-device), fall back to copy
				if err := copyFile(srcPath, dstPath); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// copyFile copies a file from src to dst
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

// chownRecursive changes ownership of a directory and all its contents
func chownRecursive(path string, uid, gid int) error {
	return filepath.Walk(path, func(name string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return os.Chown(name, uid, gid)
	})
}

func (vm *vm) Signal(ctx context.Context, signal string) error {
	sig := syscallSignal(signal)

	err := vm.initClient.Signal(ctx, int(sig))
	if err != nil {
		return fmt.Errorf("failed to send signal to init: %w", err)
	}

	return nil
}

type RunResult struct {
	HasBeenStopped bool
	VMExited       bool
	InitFailed     bool
	ProcessExited  bool
	ExitCode       int
	ExitedAt       time.Time
}

func (vm *vm) run() {
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

func (vm *vm) Shutdown(ctx context.Context) error {
	if vm.successFullyShutdowned.Load() {
		return nil
	}
	err := vm.vmm.ShutdownVMM(ctx)
	if err != nil {
		return fmt.Errorf("failed to shutdown VMM: %w", err)
	}
	vm.successFullyShutdowned.Store(true)
	return nil
}

func (vm *vm) Stop(ctx context.Context, signal string) error {
	vm.stopRequested = true
	err := vm.Signal(ctx, signal)
	if err != nil {
		return fmt.Errorf("failed to send signal to init: %w", err)
	}

	return nil
}

func (vm *vm) Run() instance.ExitResult {
	<-vm.waitChan
	return instance.ExitResult{
		Success:   vm.runResult.ExitCode == 0,
		ExitCode:  vm.runResult.ExitCode,
		ExitedAt:  vm.runResult.ExitedAt,
		Requested: vm.runResult.HasBeenStopped,
	}
}

func (vm *vm) WaitExit(ctx context.Context) (exited bool) {
	for {
		select {
		case <-ctx.Done():
			return false
		case <-vm.waitChan:
			return true
		}
	}
}

func (vm *vm) recover() bool {
	ok := false
	state := vm.determinateState()

	if !state.isVMMRunning {
		return ok
	}
	if !state.isVMRunning {
		err := vm.Shutdown(context.Background())
		if err != nil {
			slog.Error("failed to shutdown VMM", "err", err)
		}
		return ok
	}

	go vm.run()

	return true
}

type internalState struct {
	isVMMRunning bool
	isVMRunning  bool
	vminfo       *cloudhypervisor.VmInfo
}

func (vm *vm) determinateState() internalState {
	s := internalState{
		isVMMRunning: false,
		isVMRunning:  false,
	}

	if _, err := vm.vmm.PingVMM(context.Background()); err != nil {
		slog.Debug("failed to ping VMM", "err", err)
		return s
	}
	s.isVMMRunning = true

	vminfo, err := vm.vmm.VMInfo(context.Background())
	if err != nil {
		slog.Debug("failed to get VM info", "err", err)
		return s
	}

	if vminfo.State != cloudhypervisor.Running {
		slog.Debug("VM is not running", "state", vminfo.State)
		return s
	}
	s.isVMRunning = true

	s.vminfo = vminfo

	return s
}

func (vm *vm) Exec(ctx context.Context, cmd []string, timeout time.Duration) (*api.ExecResult, error) {
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

// Snapshot saves the VM state to a file for fast restore.
// This enables sub-100ms cold starts by restoring from a pre-booted snapshot.
func (vm *vm) Snapshot(ctx context.Context, path string) error {
	// The path is relative to the jail chroot (e.g., /snapshots/snap-1)
	// We need to create the directory on the host at /var/lib/ravel/instances/{id}/snapshots/snap-1
	hostPath := getInstanceDir(vm.id) + path
	if err := os.MkdirAll(hostPath, 0755); err != nil {
		return fmt.Errorf("failed to create snapshot directory: %w", err)
	}

	// Chown to ravel-jailer user so CloudHypervisor (running as jailer) can write
	jailerUid, jailerGid, err := common.SetupRavelJailerUser()
	if err != nil {
		return fmt.Errorf("failed to get jailer user: %w", err)
	}
	if err := os.Chown(hostPath, jailerUid, jailerGid); err != nil {
		return fmt.Errorf("failed to chown snapshot directory: %w", err)
	}

	// Pause the VM before taking snapshot for consistency
	_, err = vm.vmm.PauseVM(ctx)
	if err != nil {
		return fmt.Errorf("failed to pause VM before snapshot: %w", err)
	}

	// CloudHypervisor expects file:// URL format with path relative to its chroot
	snapshotUrl := "file://" + path

	// Take the snapshot
	_, err = vm.vmm.PutVmSnapshot(ctx, cloudhypervisor.VmSnapshotConfig{
		DestinationUrl: &snapshotUrl,
	})
	if err != nil {
		// Try to resume even if snapshot failed
		vm.vmm.ResumeVM(ctx)
		return fmt.Errorf("failed to create snapshot: %w", err)
	}

	// Resume the VM after snapshot
	_, err = vm.vmm.ResumeVM(ctx)
	if err != nil {
		return fmt.Errorf("failed to resume VM after snapshot: %w", err)
	}

	return nil
}

// Restore restores the VM state from a snapshot file.
// This provides sub-100ms cold starts for AI sandbox workloads.
func (vm *vm) Restore(ctx context.Context, path string) error {
	// CloudHypervisor expects file:// URL format
	snapshotUrl := "file://" + path

	prefault := true
	_, err := vm.vmm.PutVmRestore(ctx, cloudhypervisor.RestoreConfig{
		SourceUrl: snapshotUrl,
		Prefault:  &prefault,
	})
	if err != nil {
		return fmt.Errorf("failed to restore from snapshot: %w", err)
	}

	// Resume the VM after restore
	_, err = vm.vmm.ResumeVM(ctx)
	if err != nil {
		return fmt.Errorf("failed to resume VM after restore: %w", err)
	}

	return nil
}
