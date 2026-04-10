package pubsub

import "sync"

type Observable[T any] struct {
	stateMutex sync.RWMutex
	state      T
	store      *subStore[T]
}

func (s *Observable[T]) Get() T {
	s.stateMutex.RLock()
	defer s.stateMutex.RUnlock()
	return s.state
}

func (s *Observable[T]) Set(newState T) {
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

func NewObservable[T any](initialState T) *Observable[T] {
	return &Observable[T]{state: initialState, store: newSubStore[T]()}
}

func (s *Observable[T]) Subscribe() *Subscriber[T] {
	s.stateMutex.RLock()
	ch := make(chan T, 1)
	ch <- s.state
	s.store.add(ch)
	s.stateMutex.RUnlock()
	return &Subscriber[T]{unsubscribe: func() { s.store.removeAndClose(ch) }, ch: ch}
}
