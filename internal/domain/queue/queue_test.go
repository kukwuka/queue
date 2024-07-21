package queue_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/kukwuka/queue/internal/domain"
	"github.com/kukwuka/queue/internal/domain/queue"
)

type queueTestSuite struct {
	suite.Suite
}

func (s *queueTestSuite) TestPushGet_3Request1Cancel() {
	queueInstance := queue.NewQueue[string](3)
	defer queueInstance.Close()

	resultFrom1, resultFrom3 := make(chan string), make(chan string)
	go func() {
		messageFromQueue, err := queueInstance.GetMessage(context.Background())
		s.NoError(err)
		resultFrom1 <- messageFromQueue
	}()

	time.Sleep(100 * time.Millisecond)
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()
		messageFromQueue, err := queueInstance.GetMessage(ctx)
		s.ErrorIs(err, domain.ErrMessageWaitTimeOut)
		s.Zero(messageFromQueue)
	}()

	time.Sleep(100 * time.Millisecond)
	go func() {
		messageFromQueue, err := queueInstance.GetMessage(context.Background())
		s.NoError(err)
		resultFrom3 <- messageFromQueue
	}()
	time.Sleep(100 * time.Millisecond)

	messageToSend1, messageToSend2 := uuid.NewString(), uuid.NewString()
	err := queueInstance.PutMessage(context.Background(), messageToSend1)
	s.Require().NoError(err)
	err = queueInstance.PutMessage(context.Background(), messageToSend2)
	s.Require().NoError(err)

	messageFrom1Request, messageFrom3Request := <-resultFrom1, <-resultFrom3
	s.Equal(messageToSend1, messageFrom1Request)
	s.Equal(messageToSend2, messageFrom3Request)
}

func (s *queueTestSuite) TestPushGet_MessageWaitSomeRequest() {
	queueInstance := queue.NewQueue[string](2)
	defer queueInstance.Close()

	messageToSend := uuid.NewString()
	result := make(chan string)
	go func() {
		time.Sleep(100 * time.Millisecond)
		messageFromQueue, err := queueInstance.GetMessage(context.Background())
		s.NoError(err)
		result <- messageFromQueue
	}()
	err := queueInstance.PutMessage(context.Background(), messageToSend)
	s.Require().NoError(err)

	s.Equal(messageToSend, <-result)
}

func (s *queueTestSuite) TestPushGet_ErrPutMessageTimeout() {
	queueInstance := queue.NewQueue[string](2)
	defer queueInstance.Close()

	ctx, cancel := context.WithCancel(context.Background())
	for range 3 {
		err := queueInstance.PutMessage(ctx, uuid.NewString())
		s.Require().NoError(err)
	}
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()
	err := queueInstance.PutMessage(ctx, uuid.NewString())
	s.Require().NoError(err)
}

func TestQueue(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(queueTestSuite))
}
