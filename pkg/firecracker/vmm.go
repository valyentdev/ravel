package firecracker

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var ErrVMMUnavailable = errors.New("vmm is not available")

// VMM provides high-level operations for managing a Firecracker microVM.
type VMM struct {
	client *Client
}

// NewVMMClient creates a new VMM client connected to the given socket.
func NewVMMClient(socketPath string) *VMM {
	return &VMM{
		client: NewClient(socketPath),
	}
}

// WaitReady waits for the Firecracker API to become available.
func (v *VMM) WaitReady(ctx context.Context) error {
	var lastErr error
	for i := 0; i < 50; i++ {
		_, err := v.client.GetInstanceInfo(ctx)
		if err == nil {
			return nil
		}
		lastErr = err
		time.Sleep(10 * time.Millisecond)
	}
	return fmt.Errorf("firecracker not ready after 500ms: %w", lastErr)
}

// Ping checks if the Firecracker API is available.
func (v *VMM) Ping(ctx context.Context) (*InstanceInfo, error) {
	info, err := v.client.GetInstanceInfo(ctx)
	if err != nil {
		return nil, ErrVMMUnavailable
	}
	return info, nil
}

// SetMachineConfig configures the VM's CPU and memory.
func (v *VMM) SetMachineConfig(ctx context.Context, vcpus int, memMB int) error {
	config := MachineConfig{
		VcpuCount:       vcpus,
		MemSizeMib:      memMB,
		TrackDirtyPages: true, // Required for snapshots
	}
	if err := v.client.PutMachineConfig(ctx, config); err != nil {
		return fmt.Errorf("failed to set machine config: %w", err)
	}
	return nil
}

// SetBootSource configures the kernel and initrd.
func (v *VMM) SetBootSource(ctx context.Context, kernelPath, initrdPath, bootArgs string) error {
	bootSource := BootSource{
		KernelImagePath: kernelPath,
	}
	if initrdPath != "" {
		bootSource.InitrdPath = &initrdPath
	}
	if bootArgs != "" {
		bootSource.BootArgs = &bootArgs
	}

	if err := v.client.PutBootSource(ctx, bootSource); err != nil {
		return fmt.Errorf("failed to set boot source: %w", err)
	}
	return nil
}

// AddDrive adds a block device to the VM.
func (v *VMM) AddDrive(ctx context.Context, driveID, path string, isRoot, readOnly bool) error {
	drive := Drive{
		DriveID:      driveID,
		PathOnHost:   path,
		IsRootDevice: isRoot,
		IsReadOnly:   readOnly,
	}
	if err := v.client.PutDrive(ctx, drive); err != nil {
		return fmt.Errorf("failed to add drive %s: %w", driveID, err)
	}
	return nil
}

// AddNetworkInterface adds a network interface to the VM.
func (v *VMM) AddNetworkInterface(ctx context.Context, ifaceID, hostDevName string) error {
	iface := NetworkInterface{
		IfaceID:     ifaceID,
		HostDevName: hostDevName,
	}
	if err := v.client.PutNetworkInterface(ctx, iface); err != nil {
		return fmt.Errorf("failed to add network interface %s: %w", ifaceID, err)
	}
	return nil
}

// SetVsock configures the vsock device for guest-host communication.
func (v *VMM) SetVsock(ctx context.Context, guestCID int, udsPath string) error {
	vsock := VsockDevice{
		GuestCID: guestCID,
		UdsPath:  udsPath,
	}
	if err := v.client.PutVsock(ctx, vsock); err != nil {
		return fmt.Errorf("failed to set vsock: %w", err)
	}
	return nil
}

// Start boots the microVM.
func (v *VMM) Start(ctx context.Context) error {
	action := InstanceActionInfo{
		ActionType: "InstanceStart",
	}
	if err := v.client.CreateAction(ctx, action); err != nil {
		return fmt.Errorf("failed to start instance: %w", err)
	}
	return nil
}

// Pause pauses the microVM.
func (v *VMM) Pause(ctx context.Context) error {
	state := VMState{State: "Paused"}
	if err := v.client.PatchVMState(ctx, state); err != nil {
		return fmt.Errorf("failed to pause VM: %w", err)
	}
	return nil
}

// Resume resumes a paused microVM.
func (v *VMM) Resume(ctx context.Context) error {
	state := VMState{State: "Resumed"}
	if err := v.client.PatchVMState(ctx, state); err != nil {
		return fmt.Errorf("failed to resume VM: %w", err)
	}
	return nil
}

// Snapshot creates a snapshot of the microVM state.
// snapshotPath is the path to save the VM state (config + registers)
// memFilePath is the path to save the memory contents
func (v *VMM) Snapshot(ctx context.Context, snapshotPath, memFilePath string) error {
	// Pause VM before snapshot
	if err := v.Pause(ctx); err != nil {
		return fmt.Errorf("failed to pause before snapshot: %w", err)
	}

	snapshotType := "Full"
	params := SnapshotCreateParams{
		SnapshotPath: snapshotPath,
		MemFilePath:  memFilePath,
		SnapshotType: &snapshotType,
	}
	if err := v.client.CreateSnapshot(ctx, params); err != nil {
		// Try to resume on failure
		v.Resume(ctx)
		return fmt.Errorf("failed to create snapshot: %w", err)
	}

	// Resume VM after snapshot
	if err := v.Resume(ctx); err != nil {
		return fmt.Errorf("failed to resume after snapshot: %w", err)
	}

	return nil
}

// LoadSnapshot loads a snapshot into the microVM.
// This should be called on a fresh Firecracker instance (not yet started).
func (v *VMM) LoadSnapshot(ctx context.Context, snapshotPath, memFilePath string, resume bool) error {
	params := SnapshotLoadParams{
		SnapshotPath: snapshotPath,
		MemFilePath:  memFilePath,
		ResumeVM:     resume,
	}
	if err := v.client.LoadSnapshot(ctx, params); err != nil {
		return fmt.Errorf("failed to load snapshot: %w", err)
	}
	return nil
}

// SendCtrlAltDel sends Ctrl+Alt+Del to the guest.
func (v *VMM) SendCtrlAltDel(ctx context.Context) error {
	action := InstanceActionInfo{
		ActionType: "SendCtrlAltDel",
	}
	if err := v.client.CreateAction(ctx, action); err != nil {
		return fmt.Errorf("failed to send Ctrl+Alt+Del: %w", err)
	}
	return nil
}

// GetInfo returns information about the microVM instance.
func (v *VMM) GetInfo(ctx context.Context) (*InstanceInfo, error) {
	info, err := v.client.GetInstanceInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance info: %w", err)
	}
	return info, nil
}
