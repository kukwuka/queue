package domain

import (
	"context"
	"errors"
)

var (
	ErrMessageWaitTimeOut  = errors.New("didn't wait for the message")
	ErrMaxCountQueuesCount = errors.New("maximum of count queues")
)

// Queue -абстракция отвечающая за логику работы внутри 1 очереди.
type Queue interface {
	GetMessage(ctx context.Context) (string, error)
	PutMessage(ctx context.Context, message string) error
	Close()
}

// Queues -абстракция отвечающая оркестрацию всех очередей.
type Queues interface {
	GetMessageFromQueue(ctx context.Context, queueName string) (string, error)
	PutMessageToQueue(ctx context.Context, queueName string, message string) error
	Close()
}

// QueueFactory Фабрика для очередей нужной для оркестрации, длину как параметр вынес в домен.
type QueueFactory func(maxLen int) Queue
