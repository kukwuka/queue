package http_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	appHTTP "github.com/kukwuka/queue/internal/presentation/http"
)

type middlewaresTestSuite struct {
	suite.Suite
}

func (s *handlerTestSuite) TestTimeOutMiddleware() {
	var handler http.HandlerFunc = func(_ http.ResponseWriter, request *http.Request) {
		time.Sleep(200 * time.Millisecond)
		s.Equal(context.DeadlineExceeded, request.Context().Err())
	}
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/queue/"+queueName,
		bytes.NewBuffer(nil),
	)
	s.Require().NoError(err)
	response := httptest.NewRecorder()
	appHTTP.NewTimeoutMiddleware(handler, 100*time.Millisecond).ServeHTTP(response, req)
}

func TestMiddlewares(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(middlewaresTestSuite))
}
