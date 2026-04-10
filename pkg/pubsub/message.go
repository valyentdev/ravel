package pubsub

// A message broadcaster is a pubsub system that allows for broadcasting messages to multiple subscribers.
// Well suited for event-like messages.
type MessageBroadcaster[T any] struct {
	subsBufferSize int
	store          *subStore[T]
	stopCh         chan struct{}
	publishCh      chan T
}

type BroadcasterOpts[T any] struct {
	SubsBufferSize int
}

func NewMessageBroadcaster[T any](opts BroadcasterOpts[T]) *MessageBroadcaster[T] {
	return &MessageBroadcaster[T]{
		subsBufferSize: opts.SubsBufferSize,
		store:          newSubStore[T](),
		stopCh:         make(chan struct{}),
		publishCh:      make(chan T, 1),
	}
}

func (b *MessageBroadcaster[T]) unsubscribe(ch chan T) {
	b.store.removeAndClose(ch)
}

func (b *MessageBroadcaster[T]) Subscribe() *Subscriber[T] {
	ch := make(chan T, b.subsBufferSize)
	b.store.add(ch)
	return &Subscriber[T]{unsubscribe: func() { b.unsubscribe(ch) }, ch: ch}
}

func (b *MessageBroadcaster[T]) Start() {
	b.stopCh = make(chan struct{})
	go func() {
		for {
			select {
			case msg := <-b.publishCh:
				b.store.forEach(func(ch chan T) {
					select {
					case ch <- msg:
					default:
						break
					}
				})
			case <-b.stopCh:
				b.store.removeAndCloseAll()
				return
			}
		}
	}()
}

func (b *MessageBroadcaster[T]) Stop() {
	close(b.stopCh)
}

func (b *MessageBroadcaster[T]) Publish(msg T) {
	select {
	case <-b.stopCh:
		return
	default:
		b.publishCh <- msg
	}
}
