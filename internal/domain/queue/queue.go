// nolint:ireturn
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

// Queue Реализация Самой очереди сообщений.
type Queue[T any] struct {
	messages chan T
	requests *basicFifo[*request[T]]
}

func (queue *Queue[T]) GetMessage(ctx context.Context) (T, error) {
	r := &request[T]{
		id:     uuid.NewString(),
		result: make(chan T, 1),
	}
	defer close(r.result)
	queue.requests.add(r)
	var result T
	select {
	case <-ctx.Done():
		queue.requests.removeByFilter(
			func(input *request[T]) bool {
				return input.id == r.id
			},
		)
		return result, domain.ErrMessageWaitTimeOut
	case result = <-r.result:
		return result, nil
	}
}

func (queue *Queue[T]) PutMessage(ctx context.Context, message T) error {
	select {
	case queue.messages <- message:
	case <-ctx.Done():
		// В условиях задачи не сказано что делаем если очередь переполнилась.
		// Сейчас реализовано так, что клиент ждет пока не запишет.
	}
	return nil
}

func (queue *Queue[T]) Close() {
	close(queue.messages)
}

func (queue *Queue[T]) handle() {
	for message := range queue.messages {
		r := queue.requests.get()
		r.result <- message
	}
}

type request[T any] struct {
	id     string
	result chan T
}
