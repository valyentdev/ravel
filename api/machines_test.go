package api

import (
	"testing"
	"time"
)

func TestHealthCheckDefaults(t *testing.T) {
	hc := &HealthCheck{
		Exec: []string{"test"},
	}

	if hc.Interval == 0 {
		t.Error("Expected Interval to be settable to 0 for custom logic")
	}
}

func TestExecOptionsGetTimeout(t *testing.T) {
	tests := []struct {
		name      string
		timeoutMs int
		want      time.Duration
	}{
		{
			name:      "1 second",
			timeoutMs: 1000,
			want:      1 * time.Second,
		},
		{
			name:      "5 seconds",
			timeoutMs: 5000,
			want:      5 * time.Second,
		},
		{
			name:      "zero timeout",
			timeoutMs: 0,
			want:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &ExecOptions{
				TimeoutMs: tt.timeoutMs,
			}
			if got := opts.GetTimeout(); got != tt.want {
				t.Errorf("GetTimeout() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMachineStatusConstants(t *testing.T) {
	statuses := []MachineStatus{
		MachineStatusCreated,
		MachineStatusPreparing,
		MachineStatusStarting,
		MachineStatusRunning,
		MachineStatusStopping,
		MachineStatusStopped,
		MachineStatusDestroying,
		MachineStatusDestroyed,
	}

	// Just ensure all constants are defined and unique
	seen := make(map[MachineStatus]bool)
	for _, status := range statuses {
		if status == "" {
			t.Error("Empty status constant")
		}
		if seen[status] {
			t.Errorf("Duplicate status: %s", status)
		}
		seen[status] = true
	}
}

func TestHealthStatusConstants(t *testing.T) {
	statuses := []HealthStatus{
		HealthStatusUnknown,
		HealthStatusHealthy,
		HealthStatusUnhealthy,
		HealthStatusStarting,
	}

	seen := make(map[HealthStatus]bool)
	for _, status := range statuses {
		if status == "" {
			t.Error("Empty health status constant")
		}
		if seen[status] {
			t.Errorf("Duplicate health status: %s", status)
		}
		seen[status] = true
	}
}

func TestResourcesOperations(t *testing.T) {
	r1 := Resources{
		CpusMHz:  1000,
		MemoryMB: 512,
	}
	r2 := Resources{
		CpusMHz:  500,
		MemoryMB: 256,
	}

	// Test Add
	sum := r1.Add(r2)
	if sum.CpusMHz != 1500 {
		t.Errorf("Add() CpusMHz = %d, want 1500", sum.CpusMHz)
	}
	if sum.MemoryMB != 768 {
		t.Errorf("Add() MemoryMB = %d, want 768", sum.MemoryMB)
	}

	// Test Sub
	diff := r1.Sub(r2)
	if diff.CpusMHz != 500 {
		t.Errorf("Sub() CpusMHz = %d, want 500", diff.CpusMHz)
	}
	if diff.MemoryMB != 256 {
		t.Errorf("Sub() MemoryMB = %d, want 256", diff.MemoryMB)
	}

	// Test GT
	if !r1.GT(r2) {
		t.Error("GT() should return true when r1 > r2")
	}
	if r2.GT(r1) {
		t.Error("GT() should return false when r2 < r1")
	}

	equal := Resources{CpusMHz: 1000, MemoryMB: 512}
	if r1.GT(equal) {
		t.Error("GT() should return false when resources are equal")
	}
}

func TestStopConfigDefaults(t *testing.T) {
	defaultCfg := GetDefaultStopConfig()

	if defaultCfg.Timeout == nil {
		t.Fatal("Default timeout should not be nil")
	}
	if *defaultCfg.Timeout != DefaultStopTimeout {
		t.Errorf("Default timeout = %d, want %d", *defaultCfg.Timeout, DefaultStopTimeout)
	}

	if defaultCfg.Signal == nil {
		t.Fatal("Default signal should not be nil")
	}
	if *defaultCfg.Signal != DefaultStopSignal {
		t.Errorf("Default signal = %s, want %s", *defaultCfg.Signal, DefaultStopSignal)
	}
}

func TestRestartPolicyConstants(t *testing.T) {
	policies := []RestartPolicy{
		RestartPolicyAlways,
		RestartPolicyOnFailure,
		RestartPolicyNever,
	}

	seen := make(map[RestartPolicy]bool)
	for _, policy := range policies {
		if policy == "" {
			t.Error("Empty restart policy constant")
		}
		if seen[policy] {
			t.Errorf("Duplicate restart policy: %s", policy)
		}
		seen[policy] = true
	}
}
