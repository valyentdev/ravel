package pubsub

type Subscriber[T any] struct {
	unsubscribe func()
	ch          chan T
}

func (s *Subscriber[T]) Unsubscribe() {
	s.unsubscribe()
}

func (s *Subscriber[T]) Ch() <-chan T {
	return s.ch
}
