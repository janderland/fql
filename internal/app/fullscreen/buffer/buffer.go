package buffer

import (
	"container/list"
	"sync"
)

type StreamBuffer[T any] struct {
	stream chan T
	buffer *list.List
	mutex  *sync.Mutex
	done   bool
}

func New[T any](in chan T) *StreamBuffer[T] {
	sb := StreamBuffer[T]{
		stream: in,
		buffer: list.New(),
		mutex:  new(sync.Mutex),
	}
	go sb.read()
	return &sb
}

func (x *StreamBuffer[T]) Get() (*list.List, bool) {
	x.mutex.Lock()
	defer x.mutex.Unlock()

	out := x.buffer
	x.buffer = list.New()
	return out, x.done
}

func (x *StreamBuffer[T]) read() {
	for item := range x.stream {
		x.mutex.Lock()
		x.buffer.PushFront(item)
		x.mutex.Unlock()
	}

	x.mutex.Lock()
	x.done = true
	x.mutex.Unlock()
}
