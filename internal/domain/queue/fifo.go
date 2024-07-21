// nolint:ireturn
package queue

import "sync"

type basicFifo[T any] struct {
	listener chan T
	messages []T
	mu       *sync.RWMutex
}

func newBasicFifo[T any]() *basicFifo[T] {
	return &basicFifo[T]{
		listener: make(chan T),
		messages: make([]T, 0),
		mu:       &sync.RWMutex{},
	}
}

func (fifo *basicFifo[T]) add(element T) {
	select {
	case fifo.listener <- element:
	default:
		fifo.mu.Lock()
		fifo.messages = append(fifo.messages, element)
		fifo.mu.Unlock()
	}
}

func (fifo *basicFifo[T]) get() T {
	length := fifo.len()
	if length == 0 {
		elementToReturn := <-fifo.listener
		return elementToReturn
	}
	return fifo.getLast()
}

func (fifo *basicFifo[T]) len() int {
	fifo.mu.Lock()
	length := len(fifo.messages)
	fifo.mu.Unlock()
	return length
}

func (fifo *basicFifo[T]) getLast() T {
	fifo.mu.Lock()
	defer fifo.mu.Unlock()

	elementToReturn := fifo.messages[0]
	fifo.messages = fifo.messages[1:]

	return elementToReturn
}

func (fifo *basicFifo[T]) removeByFilter(input filter[T]) {
	for i, messages := range fifo.messages {
		if input(messages) {
			fifo.messages = removeIndex(fifo.messages, i)
			return
		}
	}
}

type filter[T any] func(input T) bool

func removeIndex[T any](s []T, index int) []T {
	ret := make([]T, 0, len(s)-1)
	ret = append(ret, s[:index]...)
	return append(ret, s[index+1:]...)
}
