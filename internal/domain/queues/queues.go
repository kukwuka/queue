package queues

import (
	"context"
	"fmt"
	"sync"

	"github.com/kukwuka/queue/internal/domain"
)

type Queues struct {
	queuesByName   map[string]domain.Queue
	factory        domain.QueueFactory
	queueMaxLen    int
	queuesMaxCount int
	rw             *sync.RWMutex
}

func NewQueues(factory domain.QueueFactory, queueMaxLen int, queuesMaxCount int) *Queues {
	return &Queues{
		queuesByName:   make(map[string]domain.Queue, queuesMaxCount),
		factory:        factory,
		queueMaxLen:    queueMaxLen,
		queuesMaxCount: queuesMaxCount,
		rw:             &sync.RWMutex{},
	}
}

func (queues *Queues) Close() {
	for _, queue := range queues.queuesByName {
		queue.Close()
	}
}

func (queues *Queues) GetMessageFromQueue(ctx context.Context, queueName string) (string, error) {
	queue, err := queues.getOrMakeNewTopic(queueName)
	if err != nil {
		return "", err
	}
	message, err := queue.GetMessage(ctx)
	if err != nil {
		return "", fmt.Errorf("get message from queue %s: %w", queueName, err)
	}
	return message, nil
}

func (queues *Queues) PutMessageToQueue(ctx context.Context, queueName string, message string) error {
	queue, err := queues.getOrMakeNewTopic(queueName)
	if err != nil {
		return err
	}
	err = queue.PutMessage(ctx, message)
	if err != nil {
		return fmt.Errorf("put message to queue %s: %w", queueName, err)
	}
	return nil
}

func (queues *Queues) getOrMakeNewTopic(queueName string) (domain.Queue, error) { //nolint:ireturn
	queue, exist := queues.get(queueName)
	if exist {
		return queue, nil
	}
	if queues.getLen() >= queues.queuesMaxCount {
		return nil, domain.ErrMaxCountQueuesCount
	}
	return queues.createNewQueue(queueName), nil
}

func (queues *Queues) get(queueName string) (domain.Queue, bool) { //nolint:ireturn
	queues.rw.RLock()
	queue, exist := queues.queuesByName[queueName]
	queues.rw.RUnlock()
	return queue, exist
}

func (queues *Queues) getLen() int {
	queues.rw.RLock()
	length := len(queues.queuesByName)
	queues.rw.RUnlock()
	return length
}

func (queues *Queues) createNewQueue(queueName string) domain.Queue { //nolint:ireturn
	queue := queues.factory(queues.queueMaxLen)
	queues.rw.Lock()
	queues.queuesByName[queueName] = queue
	queues.rw.Unlock()
	return queue
}
