package core

import (
	"encoding/json"
	"errors"
	"time"
)

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

type InstanceEvent = BaseInstanceEvent[any]

type InstanceEventI interface {
	ToAny() InstanceEvent
}

type BaseInstanceEvent[P any] struct {
	Id         string            `json:"id"`
	Type       InstanceEventType `json:"type"`
	Origin     Origin            `json:"origin"`
	Payload    P                 `json:"payload"`
	InstanceId string            `json:"instance_id"`
	Status     InstanceStatus    `json:"status"`
	Reported   bool              `json:"reported"`
	Timestamp  time.Time         `json:"time"`
}

func (e *BaseInstanceEvent[P]) ToAny() *BaseInstanceEvent[any] {
	return &BaseInstanceEvent[any]{
		Id:         e.Id,
		Type:       e.Type,
		Origin:     e.Origin,
		Payload:    e.Payload,
		InstanceId: e.InstanceId,
		Status:     e.Status,
		Reported:   e.Reported,
		Timestamp:  e.Timestamp,
	}
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

type InstanceCreatedEvent = BaseInstanceEvent[InstanceCreatedEventPayload]

type InstancePrepareEvent = BaseInstanceEvent[InstancePrepareEventPayload]

type InstancePreparedEvent = BaseInstanceEvent[InstancePreparedEventPayload]

type InstancePreparationFailedEvent = BaseInstanceEvent[InstancePreparationFailedEventPayload]

type InstanceStartEvent = BaseInstanceEvent[InstanceStartEventPayload]

type InstanceStartFailedEvent = BaseInstanceEvent[InstanceStartFailedEventPayload]

type InstanceStartedEvent = BaseInstanceEvent[InstanceStartedEventPayload]

type InstanceStopEvent = BaseInstanceEvent[InstanceStopEventPayload]

type InstanceExitedEvent = BaseInstanceEvent[InstanceExitedEventPayload]

type InstanceDestroyEvent = BaseInstanceEvent[InstanceDestroyEventPayload]

type InstanceDestroyedEvent = BaseInstanceEvent[InstanceDestroyedEventPayload]

func unmarshalGeneric[P any](payload []byte) (P, error) {
	var p P
	err := json.Unmarshal(payload, &p)
	if err != nil {
		return p, err
	}

	return p, nil
}

func UnmarshalEventPayload(eventType InstanceEventType, payload []byte) (any, error) {
	switch eventType {
	case InstancePrepare:
		return unmarshalGeneric[InstancePrepareEventPayload](payload)
	case InstancePrepared:
		return unmarshalGeneric[InstancePreparedEventPayload](payload)
	case InstancePreparationFailed:
		return unmarshalGeneric[InstancePreparationFailedEventPayload](payload)
	case InstanceExited:
		return unmarshalGeneric[InstanceExitedEventPayload](payload)
	case InstanceStart:
		return unmarshalGeneric[InstanceStartEventPayload](payload)
	case InstanceStartFailed:
		return unmarshalGeneric[InstanceStartFailedEventPayload](payload)
	case InstanceStarted:
		return unmarshalGeneric[InstanceStartedEventPayload](payload)
	case InstanceStop:
		return unmarshalGeneric[InstanceStopEventPayload](payload)
	default:
		return nil, errors.New("unknown event type")
	}
}
