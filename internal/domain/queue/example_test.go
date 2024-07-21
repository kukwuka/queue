package queue_test

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type queueTestSuite struct {
	suite.Suite
}

func (s *queueTestSuite) TestPushGet_Success() {

}

func TestQueue(t *testing.T) {
	suite.Run(t, new(queueTestSuite))
}
