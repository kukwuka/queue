package queues_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/kukwuka/queue/internal/domain"
	"github.com/kukwuka/queue/internal/domain/queues"
	mocks "github.com/kukwuka/queue/mocks/domain"
)

const (
	maxLen   = 2
	maxCount = 3
)

type queuesTestSuite struct {
	suite.Suite
}

func (s *queuesTestSuite) TestPushGet_Success() {
	queueName, messageToPut := uuid.NewString(), uuid.NewString()
	ctx := context.Background()

	queueInstance := mocks.NewQueue(s.T())
	queueInstance.
		EXPECT().
		PutMessage(ctx, messageToPut).
		Return(nil).
		Once()
	queueInstance.
		EXPECT().
		Close().
		Once()

	factory := mocks.NewQueueFactory(s.T())
	factory.
		EXPECT().
		Execute(maxLen).
		Return(queueInstance).
		Once()

	queuesInstance := queues.NewQueues(factory.Execute, maxLen, maxCount)
	defer queuesInstance.Close()
	err := queuesInstance.PutMessageToQueue(ctx, queueName, messageToPut)
	s.Require().NoError(err)

	messageFromQueue := uuid.NewString()
	queueInstance.
		EXPECT().
		GetMessage(ctx).
		Return(messageFromQueue, nil)

	resultMessage, err := queuesInstance.GetMessageFromQueue(ctx, queueName)
	s.Require().NoError(err)
	s.Equal(messageFromQueue, resultMessage)
}

func (s *queuesTestSuite) TestPush_ErrMaxQueueCrowded() {
	messageToPut := uuid.NewString()
	ctx := context.Background()

	factory := mocks.NewQueueFactory(s.T())
	factory.
		EXPECT().
		Execute(maxLen).
		RunAndReturn(
			func(_ int) domain.Queue {
				queueInstance := mocks.NewQueue(s.T())
				queueInstance.
					EXPECT().
					PutMessage(ctx, messageToPut).
					Return(nil)
				return queueInstance
			},
		).Times(maxCount)

	queuesInstance := queues.NewQueues(factory.Execute, maxLen, maxCount)

	for range maxCount {
		err := queuesInstance.PutMessageToQueue(ctx, uuid.NewString(), messageToPut)
		s.Require().NoError(err)
	}

	err := queuesInstance.PutMessageToQueue(ctx, uuid.NewString(), messageToPut)
	s.ErrorIs(err, domain.ErrMaxCountQueuesCount)
}

func (s *queuesTestSuite) TestGet_ErrMaxQueueCrowded() {
	ctx := context.Background()

	factory := mocks.NewQueueFactory(s.T())
	factory.
		EXPECT().
		Execute(maxLen).
		RunAndReturn(
			func(_ int) domain.Queue {
				queueInstance := mocks.NewQueue(s.T())
				queueInstance.
					EXPECT().
					GetMessage(ctx).
					Return(uuid.NewString(), nil)
				return queueInstance
			},
		).Times(maxCount)

	queuesInstance := queues.NewQueues(factory.Execute, maxLen, maxCount)

	wg := &sync.WaitGroup{}
	for range maxCount {
		wg.Add(1)
		go func() {
			defer wg.Done()
			message, err := queuesInstance.GetMessageFromQueue(ctx, uuid.NewString())
			s.NoError(err)
			s.NotEmpty(message)
		}()
	}
	wg.Wait()
	message, err := queuesInstance.GetMessageFromQueue(ctx, uuid.NewString())
	s.Require().ErrorIs(err, domain.ErrMaxCountQueuesCount)
	s.Zero(message)
}

func (s *queuesTestSuite) TestPush_QueueError() {
	queueName, messageToPut := "put_test_queue", uuid.NewString()
	ctx := context.Background()

	queueInstance := mocks.NewQueue(s.T())
	queueInstance.
		EXPECT().
		PutMessage(ctx, messageToPut).
		Return(errors.New("some put error")).
		Once()
	queueInstance.
		EXPECT().
		Close().
		Once()

	factory := mocks.NewQueueFactory(s.T())
	factory.
		EXPECT().
		Execute(maxLen).
		Return(queueInstance).
		Once()

	queuesInstance := queues.NewQueues(factory.Execute, maxLen, maxCount)
	defer queuesInstance.Close()
	err := queuesInstance.PutMessageToQueue(ctx, queueName, messageToPut)
	s.Require().EqualError(err, "put message to queue put_test_queue: some put error")
}

func (s *queuesTestSuite) TestGet_QueueError() {
	ctx := context.Background()

	queueInstance := mocks.NewQueue(s.T())
	queueInstance.
		EXPECT().
		GetMessage(ctx).
		Return("", errors.New("some put error")).
		Once()
	queueInstance.
		EXPECT().
		Close().
		Once()

	factory := mocks.NewQueueFactory(s.T())
	factory.
		EXPECT().
		Execute(maxLen).
		Return(queueInstance).
		Once()

	queuesInstance := queues.NewQueues(factory.Execute, maxLen, maxCount)
	defer queuesInstance.Close()
	queueName := "get_test_queue"
	message, err := queuesInstance.GetMessageFromQueue(ctx, queueName)
	s.Require().EqualError(err, "get message from queue get_test_queue: some put error")
	s.Zero(message)
}

func TestQueues(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(queuesTestSuite))
}
