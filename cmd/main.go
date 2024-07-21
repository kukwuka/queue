package main

import (
	"encoding/json"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/kukwuka/queue/internal/domain"
	"github.com/kukwuka/queue/internal/domain/queue"
	"github.com/kukwuka/queue/internal/domain/queues"
	appHTTP "github.com/kukwuka/queue/internal/presentation/http"
)

// Есть библиотеки для удобной работы, но ладно и так пойдет, задачу решает.
func makeConfig() config {
	const (
		timeOutFlag           = "timeout"
		portFlag              = "port"
		queueMaxSizeFlag      = "queueMaxSize"
		queuesMaxCountFlag    = "queuesMaxCount"
		defaultQueueMaxSize   = 2
		defaultQueuesMaxCount = 2
	)
	var configInstance config
	flag.DurationVar(&configInstance.TimeOut, timeOutFlag, time.Second, "timeout for handlers")
	flag.StringVar(&configInstance.Port, portFlag, "8080", "port for server")
	flag.IntVar(&configInstance.QueueMaxSize, queueMaxSizeFlag, defaultQueueMaxSize, "queue max size")
	flag.IntVar(&configInstance.QueuesMaxCount, queuesMaxCountFlag, defaultQueuesMaxCount, "max count of queues")
	flag.Parse()
	return configInstance
}

func main() {
	configInstance := makeConfig()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	configPayload, err := json.Marshal(configInstance)
	if err != nil {
		logger.Error(err.Error())
	}
	logger.Info("start with", "config", string(configPayload))

	// По слоенной архитектуре еще должны быть юзкейсы, ну стал из делать
	// Т.к. В данном случае они бесполезны и буду просто вызывать доменный сервис.
	queuesInstance := queues.NewQueues(newQ, configInstance.QueueMaxSize, configInstance.QueuesMaxCount)
	defer queuesInstance.Close()

	mux := appHTTP.NewRouter(queuesInstance, logger)

	var handler http.Handler = mux
	// Эта мидлвара должна быть в pkg нашей команды, но ладно, пока так.
	handler = appHTTP.NewTimeoutMiddleware(handler, configInstance.TimeOut)

	err = http.ListenAndServe(":"+configInstance.Port, handler) //nolint:gosec
	if err != nil {
		log.Println(err.Error())
	}
}

func newQ(maxLen int) domain.Queue { //nolint:ireturn
	return queue.NewQueue[string](maxLen)
}

type config struct {
	Port           string        `json:"port"`
	TimeOut        time.Duration `json:"timeOut"`
	QueueMaxSize   int           `json:"queueMaxSize"`
	QueuesMaxCount int           `json:"queuesMaxCount"`
}
