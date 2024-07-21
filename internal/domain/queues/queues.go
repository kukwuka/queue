package queues

import (
	"context"
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
		queuesByName:   make(map[string]domain.Queue, 10),
		factory:        factory,
		queueMaxLen:    queueMaxLen,
		queuesMaxCount: queuesMaxCount,
		rw:             &sync.RWMutex{},
	}
}

func (q *Queues) Close() {
	for _, queue := range q.queuesByName {
		queue.Close()
	}
}

func (q *Queues) GetMessageFromQueue(ctx context.Context, queueName string) (string, error) {
	queue, err := q.getOrMakeNewTopic(queueName)
	if err != nil {
		return "", err
	}
	return queue.GetMessage(ctx)
}

func (q *Queues) PutMessageToQueue(ctx context.Context, queueName string, message string) error {
	queue, err := q.getOrMakeNewTopic(queueName)
	if err != nil {
		return err
	}
	return queue.PutMessage(ctx, message)
}

func (q *Queues) getOrMakeNewTopic(queueName string) (domain.Queue, error) {
	queue, exist := q.get(queueName)
	if exist {
		return queue, nil
	}
	if q.getLen() >= q.queuesMaxCount {
		return nil, domain.ErrMaxCountQueuesCount
	}
	return q.createNewQueue(queueName), nil
}

func (q *Queues) get(queueName string) (domain.Queue, bool) {
	q.rw.RLock()
	queue, exist := q.queuesByName[queueName]
	q.rw.RUnlock()
	return queue, exist
}

func (q *Queues) getLen() int {
	q.rw.RLock()
	length := len(q.queuesByName)
	q.rw.RUnlock()
	return length
}

func (q *Queues) createNewQueue(queueName string) domain.Queue {
	queue := q.factory(q.queueMaxLen)
	q.rw.Lock()
	q.queuesByName[queueName] = queue
	q.rw.Unlock()
	return queue
}
