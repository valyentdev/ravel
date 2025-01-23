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
	Can         func(*S, E) bool
	BeforeEvent func(*S, E) error
	Apply       func(*S, E)
	AfterEvent  func(*S, E) error
}

type Transitions[S any, ET comparable, E Event[ET]] map[ET]TransitionsSet[S, ET, E]

type StateMachine[S any, ET comparable, E Event[ET]] struct {
	lock   sync.Mutex
	state  *pubsub.Observable[*S]
	config Config[S, ET, E]
}

func (s *StateMachine[S, ET, E]) Mutate(mutfn func(*S)) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	newState := s.config.Copy(s.state.Get())

	mutfn(newState)

	s.state.Set(newState)
	if s.config.AfterMutate != nil {
		return s.config.AfterMutate(newState)
	}

	return nil
}

func (s *StateMachine[S, ET, E]) State() *S {
	return s.state.Get()
}

var ErrUnknownEvent = errors.New("unknown event")
var ErrCannotTransition = errors.New("cannot transition")

func (sm *StateMachine[S, ET, E]) PushEvent(event E) (prev, next *S, _ error) {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	prev = sm.state.Get()

	transition, ok := sm.config.Transitions[event.EventType()]
	if !ok {
		return prev, nil, ErrUnknownEvent
	}

	if transition.Can != nil && !transition.Can(sm.state.Get(), event) {
		return prev, nil, ErrCannotTransition
	}

	if sm.config.BeforeAllEvent != nil {
		if err := sm.config.BeforeAllEvent(sm.state.Get(), event); err != nil {
			return prev, nil, err
		}
	}

	if transition.BeforeEvent != nil {
		if err := transition.BeforeEvent(sm.state.Get(), event); err != nil {
			return prev, nil, err
		}
	}

	next = sm.config.Copy(prev)

	if transition.Apply != nil {
		transition.Apply(next, event)
	}

	if sm.config.ApplyAll != nil {
		sm.config.ApplyAll(next, event)
	}

	if transition.AfterEvent != nil {
		if err := transition.AfterEvent(next, event); err != nil {
			return prev, nil, err
		}
	}

	if sm.config.AfterAllEvent != nil {
		if err := sm.config.AfterAllEvent(next, event); err != nil {
			return prev, nil, err
		}
	}

	sm.state.Set(next)

	return prev, next, nil
}

type Config[S any, ET comparable, E Event[ET]] struct {
	Copy           func(*S) *S
	Transitions    Transitions[S, ET, E]
	ApplyAll       func(*S, E)
	BeforeAllEvent func(*S, E) error
	AfterAllEvent  func(*S, E) error
	AfterMutate    func(*S) error
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
