package agent

import (
	"log/slog"
	"sync"
	"time"

	"github.com/gammazero/deque"
	"github.com/valyentdev/ravel/agent/machine/state"
	"github.com/valyentdev/ravel/agent/store"
	"github.com/valyentdev/ravel/api"
)

type eventer struct {
	store      *store.Store
	mutex      sync.RWMutex
	queue      *deque.Deque[api.MachineEvent]
	notify     chan struct{}
	reportFunc func(api.MachineEvent) error
}

var _ state.Eventer = (*eventer)(nil)

func (e *eventer) triggerNotify() {
	select {
	case e.notify <- struct{}{}:
	default:
		return
	}
}
func (e *eventer) nextEvent() (api.MachineEvent, bool) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	if e.queue.Len() == 0 {
		return api.MachineEvent{}, false
	}

	event := e.queue.Front()

	return event, true
}

func newEventer(store *store.Store, reportFunc func(api.MachineEvent) error) *eventer {
	return &eventer{
		store:      store,
		queue:      deque.New[api.MachineEvent](0),
		notify:     make(chan struct{}, 1),
		reportFunc: reportFunc,
	}
}

func (es *eventer) ReportEvent(event api.MachineEvent) {
	es.mutex.Lock()
	es.queue.PushBack(event)
	es.mutex.Unlock()

	es.triggerNotify()

}

func (e *eventer) Start() error {
	events, err := e.store.LoadMachineInstanceEvents()
	if err != nil {
		return err
	}

	for _, event := range events {
		e.queue.PushBack(event)
	}

	e.triggerNotify()

	go e.start()
	return nil
}

func (e *eventer) start() {
	for range e.notify {
		e.reportEvents()
	}
}

func (e *eventer) report(event api.MachineEvent) error {
	err := e.reportFunc(event)
	if err != nil {
		return err
	}

	err = e.store.DeleteEvent(event.Id)
	if err != nil {
		slog.Error("failed to delete event", "error", err)
	}

	return nil
}

func (e *eventer) reportEvents() {
	for event, ok := e.nextEvent(); ok; event, ok = e.nextEvent() {
		err := e.report(event)
		if err != nil {
			slog.Error("failed to report event", "error", err)
			time.Sleep(1 * time.Second)
			continue
		}

		e.mutex.Lock()
		e.queue.PopFront()
		e.mutex.Unlock()
	}
}
