package queue

import (
	"context"
	"github.com/google/uuid"
	"github.com/kukwuka/queue/internal/domain"
)

func NewQueue[T any](maxLen int) *Queue[T] {
	q := &Queue[T]{
		messages: make(chan T, maxLen),
		requests: newBasicFifo[*request[T]](),
	}
	go q.handle()
	return q
}

type Queue[T any] struct {
	messages chan T
	requests *basicFifo[*request[T]]
}

func (q *Queue[T]) GetMessage(ctx context.Context) (T, error) {
	r := &request[T]{
		id:     uuid.NewString(),
		result: make(chan T, 1),
	}
	defer close(r.result)
	q.requests.add(r)
	var result T
	select {
	case <-ctx.Done():
		q.requests.removeByFilter(
			func(input *request[T]) bool {
				return input.id == r.id
			},
		)
		return result, domain.ErrMessageWaitTimeOut
	case result = <-r.result:
		return result, nil
	}
}
func (q *Queue[T]) PutMessage(ctx context.Context, message T) error {
	select {
	case q.messages <- message:
	case <-ctx.Done():
	}
	return nil
}
func (q *Queue[T]) Close() {
	close(q.messages)
}

func (q *Queue[T]) handle() {
	for message := range q.messages {
		r := q.requests.get()
		r.result <- message
	}
}

type request[T any] struct {
	id     string
	result chan T
}
