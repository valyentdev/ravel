package eventer

import (
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/gammazero/deque"
	"github.com/nats-io/nats.go"
	"github.com/valyentdev/ravel/internal/agent/store"
	"github.com/valyentdev/ravel/pkg/core"
)

type Eventer struct {
	subject string

	toReport *deque.Deque[core.InstanceEvent]
	mutex    sync.RWMutex

	nc     *nats.Conn
	store  *store.Store
	notify chan struct{}
	stopCh chan struct{}
}

func NewEventer(unreportedEvents []core.InstanceEvent, machineId, instanceId string, nc *nats.Conn, store *store.Store) *Eventer {
	toReport := deque.New[core.InstanceEvent](len(unreportedEvents))

	e := &Eventer{
		subject:  "events." + machineId + "." + instanceId,
		toReport: toReport,
		nc:       nc,
		store:    store,
		notify:   make(chan struct{}, 1),
		stopCh:   make(chan struct{}),
	}

	go e.start()

	return e
}

func (e *Eventer) triggerNotify() {
	select {
	case e.notify <- struct{}{}:
	default:
		return
	}
}

func (e *Eventer) Report(event core.InstanceEvent) {
	slog.Debug("reporting event", "event", event)
	e.mutex.Lock()
	e.toReport.PushBack(event)
	e.triggerNotify()
	e.mutex.Unlock()
}

func (e *Eventer) Stop() {
	e.stopCh <- struct{}{}
}

func (e *Eventer) Start() {
	go e.start()
}

func (e *Eventer) start() {
	for {
		select {
		case <-e.stopCh:
			return
		case <-e.notify:
			e.reportEvents()
		}
	}
}

func (e *Eventer) nextEvent() (core.InstanceEvent, bool) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	if e.toReport.Len() == 0 {
		return core.InstanceEvent{}, false
	}

	event := e.toReport.Front()

	return event, true
}

func (e *Eventer) reportEvents() {
	for event, ok := e.nextEvent(); ok; event, ok = e.nextEvent() {
		err := e.report(event)
		if err != nil {
			slog.Error("failed to report event", "error", err)
			time.Sleep(1 * time.Second)
			continue
		}

		e.mutex.Lock()
		e.toReport.PopFront()
		e.mutex.Unlock()
	}
}

func (e *Eventer) report(event core.InstanceEvent) error {
	bytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = e.nc.Request(e.subject, bytes, 1*time.Second)
	if err != nil {
		return err
	}

	err = e.store.SetLastReportedEventId(event.InstanceId, event.Id)
	if err != nil {
		return err
	}

	return nil
}
