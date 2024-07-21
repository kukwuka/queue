package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/kukwuka/queue/internal/domain"
)

// Я обычно использую echo, непривычно с чистым http работать.

func NewRouter(queues domain.Queues, logger *slog.Logger) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /queue/{queue}", newPutToQueueHandler(queues, logger))
	mux.HandleFunc("GET /queue/{queue}", newGetFromQueueHandler(queues, logger))
	return mux
}

const timeOutQueryParamKey = "timeout"

type messageSchemas struct {
	Message string `json:"message"`
}

func newGetFromQueueHandler(queues domain.Queues, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel, err := makeCtx(r)
		defer cancel()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		queueName := r.PathValue("queue")
		message, err := queues.GetMessageFromQueue(ctx, queueName) //nolint:contextcheck
		if err != nil {
			if errors.Is(err, domain.ErrMessageWaitTimeOut) {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			logger.Error(fmt.Errorf("get from queue handler: %w", err).Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = json.NewEncoder(w).Encode(messageSchemas{Message: message})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func noop() {}

func makeCtx(r *http.Request) (context.Context, func(), error) {
	ctx := r.Context()
	var cancel context.CancelFunc = noop
	hasTimeout := r.URL.Query().Has(timeOutQueryParamKey)
	if hasTimeout {
		timeoutSecond, err := strconv.Atoi(r.URL.Query().Get(timeOutQueryParamKey))
		if err != nil {
			return ctx, func() {}, fmt.Errorf("parse timout second value: %w", err)
		}
		ctx, cancel = context.WithTimeout(ctx, time.Second*time.Duration(timeoutSecond))
	}
	return ctx, cancel, nil
}

func newPutToQueueHandler(queues domain.Queues, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		queueName := r.PathValue("queue")
		var schema messageSchemas
		err := json.NewDecoder(r.Body).Decode(&schema)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = queues.PutMessageToQueue(r.Context(), queueName, schema.Message)
		if err != nil {
			if errors.Is(err, domain.ErrMaxCountQueuesCount) {
				http.Error(w, err.Error(), http.StatusTooManyRequests)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error(fmt.Errorf("put to queue handler: %w", err).Error())
			return
		}
	}
}
