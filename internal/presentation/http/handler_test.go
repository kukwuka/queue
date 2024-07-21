package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/kukwuka/queue/internal/domain"
	appHTTP "github.com/kukwuka/queue/internal/presentation/http"
	mocks "github.com/kukwuka/queue/mocks/domain"
)

const queueName = "test"

type handlerTestSuite struct {
	suite.Suite
}

func (s *handlerTestSuite) TestGetFromQueueHandler_Success() {
	ctx := context.Background()
	req, err := http.NewRequest(http.MethodGet, "/queue/"+queueName, bytes.NewBuffer(nil))
	s.Require().NoError(err)
	req = req.WithContext(ctx)
	response := httptest.NewRecorder()

	queuesInstance := mocks.NewQueues(s.T())
	message := uuid.NewString()
	queuesInstance.
		EXPECT().
		GetMessageFromQueue(ctx, queueName).
		Return(message, nil)
	buffer := bytes.NewBuffer(nil)
	logger := slog.New(slog.NewJSONHandler(buffer, nil))
	mux := appHTTP.NewRouter(queuesInstance, logger)
	mux.ServeHTTP(response, req)
	s.Equal(http.StatusOK, response.Code)
	s.JSONEq(fmt.Sprintf("{\"message\": %q}", message), response.Body.String())
	s.Zero(buffer.String())
}

func (s *handlerTestSuite) TestGetFromQueueHandler_ErrorFromQueues() {
	ctx := context.Background()
	req, err := http.NewRequest(http.MethodGet, "/queue/"+queueName, bytes.NewBuffer(nil))
	s.Require().NoError(err)
	req = req.WithContext(ctx)
	response := httptest.NewRecorder()

	queuesInstance := mocks.NewQueues(s.T())
	message := uuid.NewString()
	queuesInstance.
		EXPECT().
		GetMessageFromQueue(ctx, queueName).
		Return(message, errors.New("some error"))

	buffer := bytes.NewBuffer(nil)
	logger := slog.New(slog.NewJSONHandler(buffer, nil))

	mux := appHTTP.NewRouter(queuesInstance, logger)
	mux.ServeHTTP(response, req)
	s.Equal(http.StatusInternalServerError, response.Code)
	s.Zero(response.Body.String())
	s.logMessageEqual("get from queue handler: some error", buffer.Bytes())
}

func (s *handlerTestSuite) TestGetFromQueueHandler_ErrorContextCanceled() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req, err := http.NewRequest(http.MethodGet, "/queue/"+queueName, bytes.NewBuffer(nil))
	s.Require().NoError(err)
	req = req.WithContext(ctx)
	response := httptest.NewRecorder()

	queuesInstance := mocks.NewQueues(s.T())
	message := uuid.NewString()
	queuesInstance.
		EXPECT().
		GetMessageFromQueue(ctx, queueName).
		Return(message, domain.ErrMessageWaitTimeOut)

	buffer := bytes.NewBuffer(nil)
	logger := slog.New(slog.NewJSONHandler(buffer, nil))

	mux := appHTTP.NewRouter(queuesInstance, logger)
	mux.ServeHTTP(response, req)
	s.Equal(http.StatusNotFound, response.Code)
	s.Equal("didn't wait for the message\n", response.Body.String())
	s.Zero(buffer.String())
}

func (s *handlerTestSuite) TestPutToQueueHandler_Success() {
	ctx := context.Background()
	body := []byte(`{"message": "message"}`)
	req, err := http.NewRequest(http.MethodPut, "/queue/"+queueName, bytes.NewBuffer(body))
	s.Require().NoError(err)
	req = req.WithContext(ctx)
	response := httptest.NewRecorder()

	queuesInstance := mocks.NewQueues(s.T())
	queuesInstance.
		EXPECT().
		PutMessageToQueue(ctx, queueName, "message").
		Return(nil)
	buffer := bytes.NewBuffer(nil)
	logger := slog.New(slog.NewJSONHandler(buffer, nil))
	mux := appHTTP.NewRouter(queuesInstance, logger)
	mux.ServeHTTP(response, req)
	s.Equal(http.StatusOK, response.Code)
	s.Zero(response.Body.String())
	s.Zero(buffer.String())
}

func (s *handlerTestSuite) TestPutToQueueHandler_ErrInvalidJson() {
	ctx := context.Background()
	body := []byte(`{{}}}}}}{"message": "message"}`)
	req, err := http.NewRequest(http.MethodPut, "/queue/"+queueName, bytes.NewBuffer(body))
	s.Require().NoError(err)
	req = req.WithContext(ctx)
	response := httptest.NewRecorder()

	queuesInstance := mocks.NewQueues(s.T())
	buffer := bytes.NewBuffer(nil)
	logger := slog.New(slog.NewJSONHandler(buffer, nil))
	mux := appHTTP.NewRouter(queuesInstance, logger)
	mux.ServeHTTP(response, req)
	s.Equal(http.StatusBadRequest, response.Code)
	s.Equal("invalid character '{' looking for beginning of object key string\n", response.Body.String())
	s.Zero(buffer.String())
}

func (s *handlerTestSuite) TestPutToQueueHandler_ErrFromQueues() {
	ctx := context.Background()
	body := []byte(`{"message": "message"}`)
	req, err := http.NewRequest(http.MethodPut, "/queue/"+queueName, bytes.NewBuffer(body))
	s.Require().NoError(err)
	req = req.WithContext(ctx)
	response := httptest.NewRecorder()

	queuesInstance := mocks.NewQueues(s.T())
	queuesInstance.
		EXPECT().
		PutMessageToQueue(ctx, queueName, "message").
		Return(errors.New("some put error"))
	buffer := bytes.NewBuffer(nil)
	logger := slog.New(slog.NewJSONHandler(buffer, nil))
	mux := appHTTP.NewRouter(queuesInstance, logger)
	mux.ServeHTTP(response, req)
	s.Equal(http.StatusInternalServerError, response.Code)
	s.Zero(response.Body.String())
	s.logMessageEqual("put to queue handler: some put error", buffer.Bytes())
}

func (s *handlerTestSuite) logMessageEqual(expectedMessage string, log []byte) {
	type logSchema struct {
		MSG string `json:"msg"`
	}
	var payload logSchema
	err := json.Unmarshal(log, &payload)
	s.Require().NoError(err)
	s.Equal(expectedMessage, payload.MSG)
}

func TestHandlers(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(handlerTestSuite))
}
