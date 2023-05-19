package buffer

import (
	"container/list"
	"sync"
)

type StreamBuffer interface {
	Get() (*list.List, bool)
}

type streamBuffer[T any] struct {
	stream chan T
	buffer *list.List
	mutex  *sync.Mutex
	done   bool
}

func New[T any](in chan T) StreamBuffer {
	sb := streamBuffer[T]{
		stream: in,
		buffer: list.New(),
		mutex:  new(sync.Mutex),
	}
	go sb.read()
	return &sb
}

func (x *streamBuffer[T]) Get() (*list.List, bool) {
	x.mutex.Lock()
	defer x.mutex.Unlock()

	out := x.buffer
	x.buffer = list.New()
	return out, x.done
}

func (x *streamBuffer[T]) read() {
	for item := range x.stream {
		x.mutex.Lock()
		x.buffer.PushFront(item)
		x.mutex.Unlock()
	}

	x.mutex.Lock()
	x.done = true
	x.mutex.Unlock()
}
