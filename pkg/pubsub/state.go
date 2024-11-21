package pubsub

import "sync"

// A state broadcaster is a pubsub system that allows for broadcasting state changes to multiple subscribers.
// Well suited to observe state changes.
type StateBroadcaster[T any] struct {
	stateMutex sync.RWMutex
	state      T
	store      *subStore[T]
}

func (s *StateBroadcaster[T]) GetState() T {
	s.stateMutex.RLock()
	defer s.stateMutex.RUnlock()
	return s.state
}

func (s *StateBroadcaster[T]) SetState(newState T) {
	s.stateMutex.Lock()
	s.state = newState
	s.store.forEach(func(ch chan T) {
		select {
		case <-ch:
		default:
			break
		}

		ch <- newState
	})
	defer s.stateMutex.Unlock()
}

func NewStateBroadcaster[T any](initialState T) *StateBroadcaster[T] {
	return &StateBroadcaster[T]{state: initialState, store: newSubStore[T]()}
}

func (s *StateBroadcaster[T]) Subscribe() (T, *Subscriber[T]) {
	ch := make(chan T, 1)
	s.store.add(ch)
	return s.GetState(), &Subscriber[T]{unsubscribe: func() { s.store.removeAndClose(ch) }, ch: ch}
}
