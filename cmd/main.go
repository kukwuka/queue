package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/kukwuka/queue/internal/domain"
	"github.com/kukwuka/queue/internal/domain/queue"
	"github.com/kukwuka/queue/internal/domain/queues"
	appHTTP "github.com/kukwuka/queue/internal/presentation/http"
)

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
	return configInstance
}

func main() {
	configInstance := makeConfig()
	queuesInstance := queues.NewQueues(newQ, configInstance.QueueMaxSize, configInstance.QueuesMaxCount)
	defer queuesInstance.Close()
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /queue/{queue}", appHTTP.NewPutToQueueHandler(queuesInstance))
	mux.HandleFunc("GET /queue/{queue}", appHTTP.NewGetFromQueueHandler(queuesInstance))

	var handler http.Handler = mux

	err := http.ListenAndServe("localhost:8090", handler) //nolint:gosec
	if err != nil {
		log.Println(err.Error())
	}
}

func newQ(maxLen int) domain.Queue { //nolint:ireturn
	return queue.NewQueue[string](maxLen)
}

type config struct {
	Port           string
	TimeOut        time.Duration
	QueueMaxSize   int
	QueuesMaxCount int
}
