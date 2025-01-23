package eventer

import (
	"sync"
	"time"

	"github.com/gammazero/deque"
)

type Eventer[E any] struct {
	mutex  sync.RWMutex
	queue  *deque.Deque[*E]
	notify chan struct{}
	opts   Options[E]
}

func (e *Eventer[E]) triggerNotify() {
	select {
	case e.notify <- struct{}{}:
	default:
		return
	}
}
func (e *Eventer[E]) nextEvent() (event *E, ok bool) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	if e.queue.Len() == 0 {

		return event, false
	}

	event = e.queue.Front()

	return event, true
}

type Options[E any] struct {
	Report    func(e *E) error // required, report the event
	OnSuccess func(e *E)       // required, action to take on success (like removing the event from the persistent store)
	OnError   func(e *E) bool  // required, action to take on error (return true to retry, false to ignore)
	Backoff   time.Duration    // default: 1s (time to wait before retrying an event)
}

func NewEventer[E any](options Options[E]) *Eventer[E] {
	if options.Backoff == 0 {
		options.Backoff = time.Second
	}

	if options.OnSuccess == nil || options.OnError == nil || options.Report == nil {
		panic("missing required options")
	}

	return &Eventer[E]{
		queue:  &deque.Deque[*E]{},
		notify: make(chan struct{}, 1),
		opts:   options,
	}
}

func (es *Eventer[E]) ReportEvent(event *E) {
	es.mutex.Lock()
	es.queue.PushBack(event)
	es.mutex.Unlock()

	es.triggerNotify()

}

func (e *Eventer[E]) Start(existing []E) {
	e.queue.Grow(len(existing))
	for _, event := range existing {
		e.queue.PushFront(&event)
	}

	e.triggerNotify()

	go e.start()
}

func (e *Eventer[E]) start() {
	for range e.notify {
		e.reportEvents()
	}
}

func (e *Eventer[E]) reportEvents() {
	for event, ok := e.nextEvent(); ok; event, ok = e.nextEvent() {
		err := e.opts.Report(event)
		if err != nil {
			if e.opts.OnError(event) {
				time.Sleep(e.opts.Backoff)
				continue
			}
		}
		e.opts.OnSuccess(event)
		e.mutex.Lock()
		e.queue.PopFront()
		e.mutex.Unlock()
	}
}
