package auth

import (
	"fmt"
	"sync"
	"hexmeet.com/haishen/tuna/logp"
	"net/http"
	"time"
)


type Manager struct {
	dispatchChan   chan WorkerRequest
	freeWorkerChan chan chan interface{}
	doneChan       chan bool
	config         Config
	logger         *logp.Logger
	waitgroup      sync.WaitGroup
	httpClient     *http.Client
}

// WorkerRequest request wrapper
type WorkerRequest struct {
	Type     string
	GUID     string
	Body     interface{}
	RspChan  chan interface{}
	DoneChan chan bool
}

func Run() {
	//init master data
	config, _ := initConfig() //load config file tuna.json
	dispatchChan := make(chan WorkerRequest, 2000)
	freeWorkerChan := make(chan chan interface{}, config.MaxWorker)
	doneChan := make(chan bool)
	logger  := logp.NewLogger(ModuleName)
	fmt.Println("master begin to init config and data ")

	manager := Manager {
		config:           config,
		logger:           logger,
		dispatchChan:     dispatchChan,
		freeWorkerChan:   freeWorkerChan,
		doneChan:         doneChan,
		}

	manager.httpClient = &http.Client{Timeout: time.Second * 2}

	//to select a free worker  to handle task
	go manager.dispatch()

	for i := 0; i < config.MaxWorker; i++ {
		workerID := fmt.Sprintf("worker_%d", i)
		go manager.work(workerID)
	}

	//to liston to port 8088 by default
	manager.webListen()

	//set doneChan to close all tasks
	close(manager.doneChan)

	time.Sleep(time.Duration(5) * time.Second)
}

func (m Manager) dispatch() {
	logger := m.logger.Named("dispatch")

	var req WorkerRequest

	for {
		select {
		case req = <-m.dispatchChan: //receive a request
			logger.Debugf("recv req: %+v", req)
			select {
			case workerChan := <-m.freeWorkerChan: //find a free worker to handle the request
				workerChan <- req
			case <-m.doneChan:
				return
			}
		case <-m.doneChan:
			return
		}
	}
}