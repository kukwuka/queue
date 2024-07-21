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

func (f *basicFifo[T]) add(element T) {
	select {
	case f.listener <- element:
	default:
		f.mu.Lock()
		f.messages = append(f.messages, element)
		f.mu.Unlock()
	}

}

func (f *basicFifo[T]) get() T {
	length := f.len()
	if length == 0 {
		elementToReturn := <-f.listener
		return elementToReturn
	}
	return f.getLast()
}

func (f *basicFifo[T]) len() int {
	f.mu.Lock()
	length := len(f.messages)
	f.mu.Unlock()
	return length
}

func (f *basicFifo[T]) getLast() T {
	f.mu.Lock()
	defer f.mu.Unlock()

	elementToReturn := f.messages[0]
	f.messages = f.messages[1:]

	return elementToReturn
}

func (f *basicFifo[T]) removeByFilter(input filter[T]) {
	for i, messages := range f.messages {
		if input(messages) {
			f.messages = removeIndex(f.messages, i)
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
