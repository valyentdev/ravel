package firecracker

// MachineConfig specifies the microVM machine configuration.
type MachineConfig struct {
	VcpuCount       int  `json:"vcpu_count"`
	MemSizeMib      int  `json:"mem_size_mib"`
	Smt             bool `json:"smt,omitempty"`
	TrackDirtyPages bool `json:"track_dirty_pages,omitempty"`
}

// BootSource specifies the boot source for the microVM.
type BootSource struct {
	KernelImagePath string  `json:"kernel_image_path"`
	BootArgs        *string `json:"boot_args,omitempty"`
	InitrdPath      *string `json:"initrd_path,omitempty"`
}

// Drive specifies a block device configuration.
type Drive struct {
	DriveID      string `json:"drive_id"`
	PathOnHost   string `json:"path_on_host"`
	IsRootDevice bool   `json:"is_root_device"`
	IsReadOnly   bool   `json:"is_read_only"`
}

// NetworkInterface specifies a network interface configuration.
type NetworkInterface struct {
	IfaceID     string  `json:"iface_id"`
	GuestMAC    *string `json:"guest_mac,omitempty"`
	HostDevName string  `json:"host_dev_name"`
}

// VsockDevice specifies a vsock device configuration.
type VsockDevice struct {
	GuestCID int    `json:"guest_cid"`
	UdsPath  string `json:"uds_path"`
}

// InstanceActionInfo describes an action to perform on the instance.
type InstanceActionInfo struct {
	ActionType string `json:"action_type"` // InstanceStart, SendCtrlAltDel, FlushMetrics
}

// SnapshotCreateParams specifies parameters for creating a snapshot.
type SnapshotCreateParams struct {
	SnapshotPath string  `json:"snapshot_path"`
	MemFilePath  string  `json:"mem_file_path"`
	SnapshotType *string `json:"snapshot_type,omitempty"` // Full or Diff
}

// SnapshotLoadParams specifies parameters for loading a snapshot.
type SnapshotLoadParams struct {
	SnapshotPath        string `json:"snapshot_path"`
	MemFilePath         string `json:"mem_file_path"`
	EnableDiffSnapshots bool   `json:"enable_diff_snapshots,omitempty"`
	ResumeVM            bool   `json:"resume_vm,omitempty"`
}

// VMState represents the state of the VM for pause/resume.
type VMState struct {
	State string `json:"state"` // Paused or Resumed
}

// InstanceInfo provides information about the instance state.
type InstanceInfo struct {
	ID      string `json:"id"`
	State   string `json:"state"` // Not started, Running, Paused
	VMState string `json:"vmm_version"`
	AppName string `json:"app_name"`
}

// Error represents a Firecracker API error response.
type Error struct {
	FaultMessage string `json:"fault_message"`
}
