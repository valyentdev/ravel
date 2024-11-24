package instance

import (
	"time"
)

type EventType string

const (
	InstanceStarted   EventType = "instance.started"
	InstanceExited    EventType = "instance.exited"
	InstanceDestroyed EventType = "instance.destroyed"
)

type Origin string

const (
	OriginUser   Origin = "user"
	OriginSystem Origin = "system"
)

type Event struct {
	Event            EventType            `json:"event"`
	Payload          InstanceEventPayload `json:"payload"`
	InstanceId       string               `json:"instance_id"`
	InstanceMetadata InstanceMetadata     `json:"metadata"`
	Timestamp        time.Time            `json:"timestamp"`
}

type InstancePrepareEventPayload struct {
	Retries int `json:"retries"`
}

type InstancePreparationFailedEventPayload struct {
	Error string
}

type ExitResult struct {
	Success   bool      `json:"success"`
	ExitCode  int       `json:"exit_code,omitempty"`
	Requested bool      `json:"requested"`
	ExitedAt  time.Time `json:"exited_at"`
}

type InstanceEventPayload struct {
	Exited *ExitResult `json:"exited,omitempty"`
}
