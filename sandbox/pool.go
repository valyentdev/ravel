// Package sandbox provides a pool manager for AI sandbox workloads.
// It maintains a pool of pre-warmed VMs that can be claimed instantly,
// enabling sub-100ms cold starts for code execution.
package sandbox

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/alexisbouchez/ravel/api"
	"github.com/alexisbouchez/ravel/core/cluster"
)

// PoolConfig configures the sandbox pool behavior.
type PoolConfig struct {
	// MinWarm is the minimum number of warm sandboxes to maintain
	MinWarm int `json:"min_warm"`
	// MaxWarm is the maximum number of warm sandboxes
	MaxWarm int `json:"max_warm"`
	// TemplateImage is the OCI image to use for sandboxes
	TemplateImage string `json:"template_image"`
	// SnapshotAfterBoot creates a snapshot after initial boot for faster restore
	SnapshotAfterBoot bool `json:"snapshot_after_boot"`
	// DefaultCPUs is the number of CPUs for sandbox VMs
	DefaultCPUs int `json:"default_cpus"`
	// DefaultMemoryMB is the memory in MB for sandbox VMs
	DefaultMemoryMB int `json:"default_memory_mb"`
	// IdleTimeout is how long an idle sandbox stays warm before being destroyed
	IdleTimeout time.Duration `json:"idle_timeout"`
	// Namespace is the Ravel namespace for sandbox machines
	Namespace string `json:"namespace"`
	// Fleet is the Ravel fleet for sandbox machines
	Fleet string `json:"fleet"`
}

// DefaultPoolConfig returns sensible defaults for AI sandbox workloads.
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MinWarm:           2,
		MaxWarm:           10,
		TemplateImage:     "python:3.11-slim",
		SnapshotAfterBoot: true,
		DefaultCPUs:       1,
		DefaultMemoryMB:   512,
		IdleTimeout:       5 * time.Minute,
		Namespace:         "sandbox",
		Fleet:             "default",
	}
}

// SandboxState represents the state of a pooled sandbox.
type SandboxState string

const (
	SandboxStateWarming  SandboxState = "warming"  // VM is booting
	SandboxStateReady    SandboxState = "ready"    // VM is warm and available
	SandboxStateClaimed  SandboxState = "claimed"  // VM is in use
	SandboxStateReseting SandboxState = "reseting" // VM is restoring from snapshot
)

// PooledSandbox represents a sandbox in the pool.
type PooledSandbox struct {
	ID         string       `json:"id"`
	MachineID  string       `json:"machine_id"`
	State      SandboxState `json:"state"`
	SnapshotID string       `json:"snapshot_id,omitempty"`
	ClaimedAt  *time.Time   `json:"claimed_at,omitempty"`
	CreatedAt  time.Time    `json:"created_at"`
	LastUsedAt time.Time    `json:"last_used_at"`
}

// Pool manages a collection of pre-warmed sandbox VMs.
type Pool struct {
	config    PoolConfig
	agent     cluster.Agent
	mu        sync.RWMutex
	sandboxes map[string]*PooledSandbox
	ready     chan *PooledSandbox
	stopCh    chan struct{}
	wg        sync.WaitGroup
}

// NewPool creates a new sandbox pool.
func NewPool(config PoolConfig, agent cluster.Agent) *Pool {
	return &Pool{
		config:    config,
		agent:     agent,
		sandboxes: make(map[string]*PooledSandbox),
		ready:     make(chan *PooledSandbox, config.MaxWarm),
		stopCh:    make(chan struct{}),
	}
}

// Start begins the pool manager background processes.
func (p *Pool) Start(ctx context.Context) error {
	slog.Info("Starting sandbox pool", "min_warm", p.config.MinWarm, "max_warm", p.config.MaxWarm)

	// Start the warmer goroutine
	p.wg.Add(1)
	go p.warmerLoop(ctx)

	// Start the idle cleanup goroutine
	p.wg.Add(1)
	go p.cleanupLoop(ctx)

	return nil
}

// Stop gracefully shuts down the pool.
func (p *Pool) Stop(ctx context.Context) error {
	slog.Info("Stopping sandbox pool")
	close(p.stopCh)
	p.wg.Wait()

	// Destroy all sandboxes
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, sandbox := range p.sandboxes {
		if err := p.agent.DestroyMachine(ctx, sandbox.MachineID, true); err != nil {
			slog.Error("Failed to destroy sandbox", "id", sandbox.ID, "error", err)
		}
	}

	return nil
}

// Claim gets a ready sandbox from the pool.
// If no sandbox is immediately available, it waits up to timeout.
func (p *Pool) Claim(ctx context.Context, timeout time.Duration) (*PooledSandbox, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	select {
	case sandbox := <-p.ready:
		p.mu.Lock()
		sandbox.State = SandboxStateClaimed
		now := time.Now()
		sandbox.ClaimedAt = &now
		p.mu.Unlock()

		slog.Info("Claimed sandbox from pool", "id", sandbox.ID, "machine_id", sandbox.MachineID)
		return sandbox, nil

	case <-ctx.Done():
		return nil, fmt.Errorf("no sandbox available within timeout")
	}
}

// Release returns a sandbox to the pool.
// If SnapshotAfterBoot is enabled, it restores the sandbox to its clean state.
func (p *Pool) Release(ctx context.Context, sandboxID string) error {
	p.mu.Lock()
	sandbox, ok := p.sandboxes[sandboxID]
	if !ok {
		p.mu.Unlock()
		return fmt.Errorf("sandbox not found: %s", sandboxID)
	}

	if sandbox.State != SandboxStateClaimed {
		p.mu.Unlock()
		return fmt.Errorf("sandbox not claimed: %s", sandboxID)
	}

	sandbox.State = SandboxStateReseting
	p.mu.Unlock()

	// Restore from snapshot if available
	if sandbox.SnapshotID != "" {
		if err := p.agent.MachineRestore(ctx, sandbox.MachineID, sandbox.SnapshotID); err != nil {
			slog.Error("Failed to restore sandbox from snapshot, destroying", "id", sandbox.ID, "error", err)
			p.destroySandbox(ctx, sandbox)
			return err
		}
	}

	p.mu.Lock()
	sandbox.State = SandboxStateReady
	sandbox.ClaimedAt = nil
	sandbox.LastUsedAt = time.Now()
	p.mu.Unlock()

	// Return to ready pool
	select {
	case p.ready <- sandbox:
		slog.Info("Returned sandbox to pool", "id", sandbox.ID)
	default:
		// Pool is full, destroy this sandbox
		slog.Info("Pool full, destroying sandbox", "id", sandbox.ID)
		p.destroySandbox(ctx, sandbox)
	}

	return nil
}

// Stats returns current pool statistics.
func (p *Pool) Stats() PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := PoolStats{
		Total:    len(p.sandboxes),
		Ready:    0,
		Claimed:  0,
		Warming:  0,
		Reseting: 0,
	}

	for _, s := range p.sandboxes {
		switch s.State {
		case SandboxStateReady:
			stats.Ready++
		case SandboxStateClaimed:
			stats.Claimed++
		case SandboxStateWarming:
			stats.Warming++
		case SandboxStateReseting:
			stats.Reseting++
		}
	}

	return stats
}

// PoolStats contains pool statistics.
type PoolStats struct {
	Total    int `json:"total"`
	Ready    int `json:"ready"`
	Claimed  int `json:"claimed"`
	Warming  int `json:"warming"`
	Reseting int `json:"reseting"`
}

// warmerLoop maintains the minimum number of warm sandboxes.
func (p *Pool) warmerLoop(ctx context.Context) {
	defer p.wg.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopCh:
			return
		case <-ticker.C:
			stats := p.Stats()
			needed := p.config.MinWarm - (stats.Ready + stats.Warming)

			for i := 0; i < needed && stats.Total < p.config.MaxWarm; i++ {
				go p.warmNewSandbox(ctx)
				stats.Total++
				stats.Warming++
			}
		}
	}
}

// cleanupLoop removes idle sandboxes that have exceeded IdleTimeout.
func (p *Pool) cleanupLoop(ctx context.Context) {
	defer p.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopCh:
			return
		case <-ticker.C:
			p.cleanupIdleSandboxes(ctx)
		}
	}
}

func (p *Pool) cleanupIdleSandboxes(ctx context.Context) {
	p.mu.Lock()
	var toDestroy []*PooledSandbox
	now := time.Now()

	for _, sandbox := range p.sandboxes {
		if sandbox.State == SandboxStateReady {
			if now.Sub(sandbox.LastUsedAt) > p.config.IdleTimeout {
				// Keep minimum warm
				stats := p.Stats()
				if stats.Ready > p.config.MinWarm {
					toDestroy = append(toDestroy, sandbox)
				}
			}
		}
	}
	p.mu.Unlock()

	for _, sandbox := range toDestroy {
		slog.Info("Destroying idle sandbox", "id", sandbox.ID, "idle_duration", time.Since(sandbox.LastUsedAt))
		p.destroySandbox(ctx, sandbox)
	}
}

func (p *Pool) warmNewSandbox(ctx context.Context) {
	sandboxID := generateSandboxID()

	slog.Info("Warming new sandbox", "id", sandboxID)

	sandbox := &PooledSandbox{
		ID:         sandboxID,
		State:      SandboxStateWarming,
		CreatedAt:  time.Now(),
		LastUsedAt: time.Now(),
	}

	p.mu.Lock()
	p.sandboxes[sandboxID] = sandbox
	p.mu.Unlock()

	// Create the machine (simplified - in production would use full machine creation)
	machineConfig := cluster.Machine{
		Id:        "sandbox-" + sandboxID,
		Namespace: p.config.Namespace,
		FleetId:   p.config.Fleet,
	}

	mi, err := p.agent.PutMachine(ctx, cluster.PutMachineOptions{
		Machine: machineConfig,
		Version: api.MachineVersion{
			Config: api.MachineConfig{
				Image: p.config.TemplateImage,
				Guest: api.GuestConfig{
					Cpus:     p.config.DefaultCPUs,
					MemoryMB: p.config.DefaultMemoryMB,
				},
			},
		},
		Start: true,
	})
	if err != nil {
		slog.Error("Failed to create sandbox machine", "id", sandboxID, "error", err)
		p.mu.Lock()
		delete(p.sandboxes, sandboxID)
		p.mu.Unlock()
		return
	}

	sandbox.MachineID = mi.MachineId

	// Wait for machine to be running
	err = p.agent.WaitForMachineStatus(ctx, sandbox.MachineID, api.MachineStatusRunning, 30)
	if err != nil {
		slog.Error("Sandbox failed to start", "id", sandboxID, "error", err)
		p.destroySandbox(ctx, sandbox)
		return
	}

	// Take initial snapshot if configured
	if p.config.SnapshotAfterBoot {
		snapshotID := sandboxID + "-base"
		if err := p.agent.MachineSnapshot(ctx, sandbox.MachineID, snapshotID); err != nil {
			slog.Warn("Failed to create base snapshot", "id", sandboxID, "error", err)
		} else {
			sandbox.SnapshotID = snapshotID
		}
	}

	p.mu.Lock()
	sandbox.State = SandboxStateReady
	p.mu.Unlock()

	// Add to ready pool
	select {
	case p.ready <- sandbox:
		slog.Info("Sandbox ready", "id", sandboxID, "machine_id", sandbox.MachineID)
	default:
		slog.Warn("Ready pool full, keeping sandbox warm", "id", sandboxID)
	}
}

func (p *Pool) destroySandbox(ctx context.Context, sandbox *PooledSandbox) {
	p.mu.Lock()
	delete(p.sandboxes, sandbox.ID)
	p.mu.Unlock()

	if err := p.agent.DestroyMachine(ctx, sandbox.MachineID, true); err != nil {
		slog.Error("Failed to destroy sandbox machine", "id", sandbox.ID, "error", err)
	}
}

// generateSandboxID creates a unique sandbox identifier.
func generateSandboxID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
