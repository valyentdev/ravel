package broadcaster

import (
	"sync"
)

type Subscriber[T any] struct {
	bc *Broadcaster[T]
	ch chan T
}

func (s *Subscriber[T]) Ch() <-chan T {
	return s.ch
}

func (s *Subscriber[T]) Unsubscribe() {
	s.bc.unsubscribe(s.ch)
}

type replayMsg[T any] struct {
	msg []T
	to  chan T
}

type Broadcaster[T any] struct {
	subsBufferSize int
	getReplay      func() []T
	subs           map[chan T]struct{}
	subsMutex      sync.RWMutex
	stopCh         chan struct{}
	replayCh       chan replayMsg[T]
	publishCh      chan T
}

type BroadcasterOpts[T any] struct {
	SubsBufferSize int
	GetReplay      func() []T
}

func NewBroadcaster[T any](opts BroadcasterOpts[T]) *Broadcaster[T] {
	return &Broadcaster[T]{
		subsBufferSize: opts.SubsBufferSize,
		getReplay:      opts.GetReplay,
		subs:           make(map[chan T]struct{}),
		stopCh:         make(chan struct{}),
		publishCh:      make(chan T, 1),
		replayCh:       make(chan replayMsg[T], 1),
	}
}

func (b *Broadcaster[T]) unsubscribe(ch chan T) {
	b.subsMutex.Lock()
	defer b.subsMutex.Unlock()
	_, ok := b.subs[ch]
	if !ok {
		return
	}

	delete(b.subs, ch)
	close(ch)
}

func (b *Broadcaster[T]) Subscribe() *Subscriber[T] {
	ch := make(chan T, b.subsBufferSize)
	b.subsMutex.Lock()
	b.subs[ch] = struct{}{}

	toReplay := b.getReplay()
	if len(toReplay) > 0 {
		b.replayCh <- replayMsg[T]{msg: toReplay, to: ch}
	}

	b.subsMutex.Unlock()
	return &Subscriber[T]{bc: b, ch: ch}
}

func (b *Broadcaster[T]) Start() {
	b.stopCh = make(chan struct{})
	go func() {
		for {
			select {
			case msg := <-b.publishCh:
				b.subsMutex.RLock()
				for ch := range b.subs {
					select {
					case ch <- msg:
					default:
						continue
					}
				}
				b.subsMutex.RUnlock()
			case replay := <-b.replayCh:
				b.subsMutex.RLock()
				if _, ok := b.subs[replay.to]; !ok {
					b.subsMutex.RUnlock()
					continue
				}
				for _, msg := range replay.msg {
					select {
					case replay.to <- msg:
					default:
						continue
					}
				}
				b.subsMutex.RUnlock()
			case <-b.stopCh:
				b.subsMutex.Lock()
				for ch := range b.subs {
					close(ch)
					delete(b.subs, ch)
				}
				b.subsMutex.Unlock()
				return
			}
		}
	}()
}

func (b *Broadcaster[T]) Stop() {
	close(b.stopCh)
}

func (b *Broadcaster[T]) Publish(msg T) {
	select {
	case <-b.stopCh:
		return
	default:
		b.publishCh <- msg
	}
}
