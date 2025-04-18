package genericpubsub

import (
	"context"
	"sync"
	"sync/atomic"
)

var nextSubId uint64

func getNextSubId() uint64 {
	return atomic.AddUint64(&nextSubId, 1)
}

type Sub[T any] struct {
	receiver chan T
}

type PubSub[T any] struct {
	pubChan chan T
	subs    sync.Map
	ctx     context.Context
}

type Subscription[T any] struct {
	Receiver chan T
	Ctx      context.Context
}

func (s *PubSub[T]) Send(value T) {
	s.pubChan <- value
}

func (ps *PubSub[T]) start() {
	go func() {
		for {
			select {
			case <-ps.ctx.Done():
				// Cleanup and delete all subs to free memory.
				ps.subs.Range(func(subKey, v interface{}) bool {
					ps.subs.Delete(subKey)
					return true
				})
				return
			case m := <-ps.pubChan:
				ps.subs.Range(func(subid, v interface{}) bool {
					sub, _ := v.(*Sub[T])
					// do this in a go routine so that a bad sub doesn't lock
					// the system up.
					go func() {
						sub.receiver <- m
					}()
					return true
				})
			}
		}
	}()

}

// Register provides a channel which the main PubSub will send values to
// Accepts a cancellation context for when the subscriber needs to close its connection.
// Returns a channel the registered receiver receives values from
// The channel will be closed when messages are no longer going to be sent.
func (ps *PubSub[T]) Register(callerCtx context.Context, bufferSize int) chan T {
	id := getNextSubId()
	recv := make(chan T, bufferSize)
	go func() {
		select {
		case <-callerCtx.Done():
			ps.cleanupSubscription(id)
			break
		case <-ps.ctx.Done():
			ps.cleanupSubscription(id)
			break
		}
	}()

	sub := Sub[T]{
		receiver: recv,
	}
	ps.subs.Store(id, &sub)
	return recv
}

func (ps *PubSub[T]) cleanupSubscription(id uint64) {
	if val, ok := ps.subs.Load(id); ok == true {
		if sub, ok := val.(*Sub[T]); ok == true {
			close(sub.receiver)
		}
		ps.subs.Delete(id)

	}
}

func New[T any](ctx context.Context, bufferSize int) *PubSub[T] {
	pubSub := &PubSub[T]{
		pubChan: make(chan T, bufferSize),
		subs:    sync.Map{},
		ctx:     ctx,
	}
	pubSub.start()
	return pubSub
}
