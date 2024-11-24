package api

import (
	"time"

	"github.com/oklog/ulid"
)

const MaxStopTimeout = 30    // in seconds
const DefaultStopTimeout = 5 // in seconds
const DefaultStopSignal = "SIGTERM"

func GetDefaultStopConfig() *StopConfig {
	defaultStopTimeout := DefaultStopTimeout
	defaultStopSignal := DefaultStopSignal
	return &StopConfig{
		Timeout: &defaultStopTimeout,
		Signal:  &defaultStopSignal,
	}
}

type Machine struct {
	Id             string        `json:"id"`
	Namespace      string        `json:"namespace"`
	FleetId        string        `json:"fleet"`
	InstanceId     string        `json:"instance_id"`
	MachineVersion ulid.ULID     `json:"machine_version"`
	Region         string        `json:"region"`
	Config         MachineConfig `json:"config"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
	Status         MachineStatus `json:"state"`
}

type MachineStatus string

const (
	MachineStatusCreated    MachineStatus = "created"
	MachineStatusPreparing  MachineStatus = "preparing"
	MachineStatusStarting   MachineStatus = "starting"
	MachineStatusRunning    MachineStatus = "running"
	MachineStatusStopping   MachineStatus = "stopping"
	MachineStatusStopped    MachineStatus = "stopped"
	MachineStatusDestroying MachineStatus = "destroying"
	MachineStatusDestroyed  MachineStatus = "destroyed"
)

type ExecOptions struct {
	Cmd       []string `json:"cmd"`
	TimeoutMs int      `json:"timeout_ms"`
}

func (e *ExecOptions) GetTimeout() time.Duration {
	return time.Duration(e.TimeoutMs) * time.Millisecond
}

type ExecResult struct {
	Stderr   string
	Stdout   string
	ExitCode int
}

const (
	RestartPolicyAlways    RestartPolicy = "always"
	RestartPolicyOnFailure RestartPolicy = "on-failure"
	RestartPolicyNever     RestartPolicy = "never"
)

type (
	MachineConfig struct {
		Image       string      `json:"image"`
		Guest       GuestConfig `json:"guest"`
		Workload    Workload    `json:"workload"`
		StopConfig  *StopConfig `json:"stop_config,omitempty"`
		AutoDestroy bool        `json:"auto_destroy,omitempty"`
	}

	GuestConfig struct {
		CpuKind  string `json:"cpu_kind"`
		MemoryMB int    `json:"memory_mb" minimum:"1"`
		Cpus     int    `json:"cpus" minimum:"1"`
	}

	Workload struct {
		Restart RestartPolicyConfig `json:"restart,omitempty"`
		Env     []string            `json:"env,omitempty"`
		Init    InitConfig          `json:"init,omitempty"`
	}

	InitConfig struct {
		Cmd        []string `json:"cmd,omitempty"`
		Entrypoint []string `json:"entrypoint,omitempty"`
		User       string   `json:"user,omitempty"`
	}

	RestartPolicy string

	RestartPolicyConfig struct {
		Policy     RestartPolicy `json:"policy,omitempty"`
		MaxRetries int           `json:"max_retries,omitempty"`
	}

	StopConfig struct {
		Timeout *int    `json:"timeout,omitempty"` // in seconds
		Signal  *string `json:"signal,omitempty"`
	}
)

type MachineEventType string

const (
	MachineCreated       MachineEventType = "machine.created"
	MachinePrepare       MachineEventType = "machine.prepare"
	MachinePrepared      MachineEventType = "machine.prepared"
	MachinePrepareFailed MachineEventType = "machine.prepare_failed"
	MachineStart         MachineEventType = "machine.start"
	MachineStartFailed   MachineEventType = "machine.start_failed"
	MachineStarted       MachineEventType = "machine.started"
	MachineStop          MachineEventType = "machine.stop"
	MachineStopFailed    MachineEventType = "machine.stop_failed"
	MachineExited        MachineEventType = "machine.exited"
	MachineDestroy       MachineEventType = "machine.destroy"
	MachineDestroyed     MachineEventType = "machine.destroyed"
)

type MachineStartEventPayload struct {
	IsRestart bool `json:"is_restart"`
}

type MachineStopEventPayload struct {
	Config *StopConfig `json:"config,omitempty"`
}

type MachinePrepareFailedEventPayload struct {
	Error string `json:"error"`
}

type MachineStartFailedEventPayload struct {
	Error string `json:"error"`
}

type MachineStartedEventPayload struct {
	StartedAt time.Time `json:"started_at"`
}

type MachineExitedEventPayload struct {
	ExitCode int       `json:"exit_code"`
	ExitedAt time.Time `json:"exited_at"`
}

type MachineDestroyEventPayload struct {
	AutoDestroy bool   `json:"auto_destroy"`
	Reason      string `json:"reason"`
	Force       bool   `json:"force"`
}

type MachineEventPayload struct {
	PrepareFailed *MachinePrepareFailedEventPayload `json:"prepare_failed,omitempty"`
	Stop          *MachineStopEventPayload          `json:"stop,omitempty"`
	Start         *MachineStartEventPayload         `json:"start,omitempty"`
	StartFailed   *MachineStartFailedEventPayload   `json:"start_failed,omitempty"`
	Started       *MachineStartedEventPayload       `json:"started,omitempty"`
	Exited        *MachineExitedEventPayload        `json:"stopped,omitempty"`
	Destroy       *MachineDestroyEventPayload       `json:"destroy,omitempty"`
}

type Origin string

const (
	OriginRavel Origin = "ravel"
	OriginUser  Origin = "user"
)

type MachineEvent struct {
	Id         ulid.ULID           `json:"id"`
	MachineId  string              `json:"machine_id"`
	InstanceId string              `json:"instance_id"`
	Status     MachineStatus       `json:"status"`
	Type       MachineEventType    `json:"type"`
	Origin     Origin              `json:"origin"`
	Payload    MachineEventPayload `json:"payload"`
	Timestamp  time.Time           `json:"timestamp"`
}
