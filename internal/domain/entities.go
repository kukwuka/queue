package domain

import (
	"context"
	"errors"
)

var (
	ErrMessageWaitTimeOut  = errors.New("didn't wait for the message")
	ErrMaxCountQueuesCount = errors.New("maximum of count queues")
)

type Queue interface {
	GetMessage(ctx context.Context) (string, error)
	PutMessage(ctx context.Context, message string) error
	Close()
}

type Queues interface {
	GetMessageFromQueue(ctx context.Context, queueName string) (string, error)
	PutMessageToQueue(ctx context.Context, queueName string, message string) error
	Close()
}

type QueueFactory func(maxLen int) Queue
