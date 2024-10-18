package core

import (
	"fmt"
	"time"

	"github.com/oklog/ulid"
)

func InstanceEventSubject(m, i string) string {
	return fmt.Sprintf("events.machines.%s.%s", m, i)
}

type InstanceEventType string

const (
	InstanceCreated           InstanceEventType = "instance.created"
	InstancePrepare           InstanceEventType = "instance.prepare"
	InstancePrepared          InstanceEventType = "instance.prepared"
	InstancePreparationFailed InstanceEventType = "instance.preparation_failed"
	InstanceStart             InstanceEventType = "instance.start"
	InstanceStartFailed       InstanceEventType = "instance.start_failed"
	InstanceStarted           InstanceEventType = "instance.started"
	InstanceStop              InstanceEventType = "instance.stop"
	InstanceExited            InstanceEventType = "instance.exited"
	InstanceDestroy           InstanceEventType = "instance.destroy"
	InstanceDestroyed         InstanceEventType = "instance.destroyed"
)

type InstanceEvent struct {
	Id         ulid.ULID            `json:"id"`
	Type       InstanceEventType    `json:"type"`
	Origin     Origin               `json:"origin"`
	Payload    InstanceEventPayload `json:"payload"`
	MachineId  string               `json:"machine_id"`
	InstanceId string               `json:"instance_id"`
	Status     InstanceStatus       `json:"status"`
	Timestamp  time.Time            `json:"timestamp"`
}

type InstanceCreatedEventPayload struct {
}

type InstancePrepareEventPayload struct {
	Retries int `json:"retries"`
}

type InstancePreparedEventPayload struct{}

type InstancePreparationFailedEventPayload struct {
	Error string
}

type InstanceExitedEventPayload struct {
	Success   bool      `json:"success"`
	ExitCode  *int64    `json:"exit_code,omitempty"`
	Requested bool      `json:"requested"`
	ExitedAt  time.Time `json:"exited_at"`
}

type InstanceStartEventPayload struct {
	IsRestart bool `json:"is_restart"`
}

type InstanceStartFailedEventPayload struct {
	Error string `json:"error"`
}

type InstanceStartedEventPayload struct {
}

type InstanceStopEventPayload struct {
	Signal string `json:"signal"`
}

type InstanceStopFailedEventPayload struct {
	Error string `json:"error"`
}

type InstanceDestroyEventPayload struct {
	Force  bool   `json:"force"`
	Reason string `json:"reason"`
}

type InstanceDestroyedEventPayload struct {
}

type InstanceEventPayload struct {
	Created           *InstanceCreatedEventPayload           `json:"created,omitempty"`
	Prepare           *InstancePrepareEventPayload           `json:"prepare,omitempty"`
	Prepared          *InstancePreparedEventPayload          `json:"prepared,omitempty"`
	PreparationFailed *InstancePreparationFailedEventPayload `json:"preparation_failed,omitempty"`
	Exited            *InstanceExitedEventPayload            `json:"exited,omitempty"`
	Start             *InstanceStartEventPayload             `json:"start,omitempty"`
	StartFailed       *InstanceStartFailedEventPayload       `json:"start_failed,omitempty"`
	Started           *InstanceStartedEventPayload           `json:"started,omitempty"`
	Stop              *InstanceStopEventPayload              `json:"stop,omitempty"`
	StopFailed        *InstanceStopFailedEventPayload        `json:"stop_failed,omitempty"`
	Destroy           *InstanceDestroyEventPayload           `json:"destroy,omitempty"`
	Destroyed         *InstanceDestroyedEventPayload         `json:"destroyed,omitempty"`
}
