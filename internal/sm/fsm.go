package sm

import (
	"errors"
	"sync"

	"github.com/valyentdev/ravel/pkg/pubsub"
)

type Event[ET comparable] interface {
	EventType() ET
}

type TransitionsSet[S any, ET comparable, E Event[ET]] struct {
	Can         func(*S) bool
	BeforeEvent func(*S, E) error
	Apply       func(*S, E)
	AfterEvent  func(*S, E) error
}

type Transitions[S any, ET comparable, E Event[ET]] map[ET]TransitionsSet[S, ET, E]

type StateMachine[S any, ET comparable, E Event[ET]] struct {
	globalLock sync.Mutex
	lock       sync.Mutex
	state      *pubsub.Observable[*S]
	config     Config[S, ET, E]
}

func (s *StateMachine[S, ET, E]) State() *S {
	return s.state.Get()
}

func (s *StateMachine[S, ET, E]) Can(event ET) bool {
	transition, ok := s.config.Transitions[event]
	if !ok {
		return false
	}

	if transition.Can != nil {
		return transition.Can(s.state.Get())
	}

	return true
}

func (s *StateMachine[S, ET, E]) Lock() {
	s.globalLock.Lock()
}

func (s *StateMachine[S, ET, E]) Unlock() {
	s.globalLock.Unlock()
}

func (s *StateMachine[S, ET, E]) TryLock() bool {
	return s.globalLock.TryLock()
}

var ErrUnknownEvent = errors.New("unknown event")
var ErrCannotTransition = errors.New("cannot transition")

func (sm *StateMachine[S, ET, E]) PushEvent(event E) error {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	transition, ok := sm.config.Transitions[event.EventType()]
	if !ok {
		return ErrUnknownEvent
	}

	if transition.Can != nil && !transition.Can(sm.state.Get()) {
		return ErrCannotTransition
	}

	if sm.config.BeforeEvent != nil {
		if err := sm.config.BeforeEvent(sm.state.Get(), event); err != nil {
			return err
		}
	}

	if transition.BeforeEvent != nil {
		if err := transition.BeforeEvent(sm.state.Get(), event); err != nil {
			return err
		}
	}

	newState := sm.config.Copy(sm.state.Get())

	if transition.Apply != nil {
		transition.Apply(newState, event)
	}

	if sm.config.Apply != nil {
		sm.config.Apply(newState, event)
	}

	if transition.AfterEvent != nil {
		if err := transition.AfterEvent(newState, event); err != nil {
			return err
		}
	}

	if sm.config.AfterEvent != nil {
		if err := sm.config.AfterEvent(newState, event); err != nil {
			return err
		}
	}

	sm.state.Set(newState)

	return nil
}

type Config[S any, ET comparable, E Event[ET]] struct {
	Copy        func(*S) *S
	Transitions Transitions[S, ET, E]
	Apply       func(*S, E)
	BeforeEvent func(*S, E) error
	AfterEvent  func(*S, E) error
}

func NewStateMachine[S any, ET comparable, E Event[ET]](initialState *S, config Config[S, ET, E]) *StateMachine[S, ET, E] {
	return &StateMachine[S, ET, E]{
		state:  pubsub.NewObservable(initialState),
		config: config,
	}
}

func (s *StateMachine[S, ET, E]) Subscribe() *pubsub.Subscriber[*S] {
	return s.state.Subscribe()
}
