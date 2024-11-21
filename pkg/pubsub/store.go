package pubsub

import "sync"

type subStore[T any] struct {
	subs  map[chan T]struct{}
	mutex sync.RWMutex
}

func newSubStore[T any]() *subStore[T] {
	return &subStore[T]{
		subs: make(map[chan T]struct{}),
	}
}

func (s *subStore[T]) removeAndClose(ch chan T) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	_, ok := s.subs[ch]
	if !ok {
		return
	}

	delete(s.subs, ch)
	close(ch)
}

func (s *subStore[T]) removeAndCloseAll() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for ch := range s.subs {
		delete(s.subs, ch)
		close(ch)
	}
}

func (s *subStore[T]) add(ch chan T) {
	s.mutex.Lock()
	s.subs[ch] = struct{}{}
	s.mutex.Unlock()
}

func (s *subStore[T]) forEach(f func(ch chan T)) {
	s.mutex.RLock()
	for ch := range s.subs {
		f(ch)
	}
	s.mutex.RUnlock()
}
