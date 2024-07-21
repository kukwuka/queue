package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/kukwuka/queue/internal/domain"
)

const timeOutQueryParamKey = "timeout"

type messageSchemas struct {
	Message string `json:"message"`
}

func NewGetFromQueueHandler(queues domain.Queues) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel, err := makeCtx(r)
		defer cancel()
		r.WithContext(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		queueName := r.PathValue("queue")
		message, err := queues.GetMessageFromQueue(ctx, queueName)
		if err != nil {
			if errors.Is(err, domain.ErrMessageWaitTimeOut) {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
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

func NewPutToQueueHandler(queues domain.Queues) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		queueName := r.PathValue("queue")
		var schema messageSchemas
		err := json.NewDecoder(r.Body).Decode(&schema)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = queues.PutMessageToQueue(r.Context(), queueName, "message")
		if err != nil {
			if errors.Is(err, domain.ErrMaxCountQueuesCount) {
				http.Error(w, err.Error(), http.StatusTooManyRequests)
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}
