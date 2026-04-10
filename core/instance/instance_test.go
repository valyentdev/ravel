package instance

import (
	"testing"

	"github.com/alexisbouchez/ravel/api"
)

func TestInstanceConfigGetDisks(t *testing.T) {
	tests := []struct {
		name   string
		mounts []Mount
		want   []string
	}{
		{
			name:   "no mounts",
			mounts: []Mount{},
			want:   []string{},
		},
		{
			name: "single mount",
			mounts: []Mount{
				{Disk: "disk1", Path: "/data"},
			},
			want: []string{"disk1"},
		},
		{
			name: "multiple mounts",
			mounts: []Mount{
				{Disk: "disk1", Path: "/data"},
				{Disk: "disk2", Path: "/logs"},
				{Disk: "disk3", Path: "/cache"},
			},
			want: []string{"disk1", "disk2", "disk3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := InstanceConfig{
				Mounts: tt.mounts,
			}
			got := cfg.GetDisks()
			if len(got) != len(tt.want) {
				t.Errorf("GetDisks() length = %d, want %d", len(got), len(tt.want))
				return
			}
			for i, diskID := range got {
				if diskID != tt.want[i] {
					t.Errorf("GetDisks()[%d] = %s, want %s", i, diskID, tt.want[i])
				}
			}
		})
	}
}

func TestInstanceStatusConstants(t *testing.T) {
	statuses := []InstanceStatus{
		InstanceStatusCreated,
		InstanceStatusStopped,
		InstanceStatusStarting,
		InstanceStatusRunning,
		InstanceStatusDestroying,
		InstanceStatusDestroyed,
	}

	seen := make(map[InstanceStatus]bool)
	for _, status := range statuses {
		if status == "" {
			t.Error("Empty instance status constant")
		}
		if seen[status] {
			t.Errorf("Duplicate instance status: %s", status)
		}
		seen[status] = true
	}
}

func TestMountStructure(t *testing.T) {
	mount := Mount{
		Disk: "test-disk",
		Path: "/mnt/data",
	}

	if mount.Disk != "test-disk" {
		t.Errorf("Mount.Disk = %s, want test-disk", mount.Disk)
	}
	if mount.Path != "/mnt/data" {
		t.Errorf("Mount.Path = %s, want /mnt/data", mount.Path)
	}
}

func TestExitResult(t *testing.T) {
	result := ExitResult{
		Success:   true,
		ExitCode:  0,
		Requested: false,
	}

	if !result.Success {
		t.Error("ExitResult.Success should be true")
	}
	if result.ExitCode != 0 {
		t.Errorf("ExitResult.ExitCode = %d, want 0", result.ExitCode)
	}
}

func TestInstanceGuestConfig(t *testing.T) {
	guest := InstanceGuestConfig{
		MemoryMB: 512,
		VCpus:    2,
		CpusMHz:  2000,
	}

	if guest.MemoryMB != 512 {
		t.Errorf("MemoryMB = %d, want 512", guest.MemoryMB)
	}
	if guest.VCpus != 2 {
		t.Errorf("VCpus = %d, want 2", guest.VCpus)
	}
	if guest.CpusMHz != 2000 {
		t.Errorf("CpusMHz = %d, want 2000", guest.CpusMHz)
	}
}

func TestInstanceState(t *testing.T) {
	state := State{
		Status:   InstanceStatusRunning,
		Stopping: false,
	}

	if state.Status != InstanceStatusRunning {
		t.Errorf("Status = %s, want %s", state.Status, InstanceStatusRunning)
	}
	if state.Stopping {
		t.Error("Stopping should be false")
	}
}

func TestInstanceConfigWithHealthCheck(t *testing.T) {
	cfg := InstanceConfig{
		Image: "alpine:latest",
		Guest: InstanceGuestConfig{
			MemoryMB: 256,
			VCpus:    1,
			CpusMHz:  1000,
		},
		Init: api.InitConfig{
			Cmd: []string{"/bin/sh"},
		},
		Env: []string{"FOO=bar"},
		Mounts: []Mount{
			{Disk: "data", Path: "/data"},
		},
	}

	if cfg.Image != "alpine:latest" {
		t.Errorf("Image = %s, want alpine:latest", cfg.Image)
	}
	if len(cfg.Mounts) != 1 {
		t.Errorf("Mounts length = %d, want 1", len(cfg.Mounts))
	}
	if len(cfg.Env) != 1 {
		t.Errorf("Env length = %d, want 1", len(cfg.Env))
	}
}
