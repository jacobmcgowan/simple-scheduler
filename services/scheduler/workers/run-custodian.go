package workers

import (
	"fmt"
	"log"
	"sync"

	"github.com/jacobmcgowan/simple-scheduler/shared/data-access/repositories"
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
	messageBus "github.com/jacobmcgowan/simple-scheduler/shared/message-bus"
)

type RunCustodian struct {
	Job           dtos.Job
	MessageBus    messageBus.MessageBus
	RunRepo       repositories.RunRepository
	quit          chan struct{}
	isRunningLock sync.Mutex `default:"sync.Mutex{}"`
	isRunning     bool
	actionQueue   string
	statusQueue   string
}

func (worker *RunCustodian) Start(wg *sync.WaitGroup) error {
	worker.isRunningLock.Lock()
	defer worker.isRunningLock.Unlock()

	if worker.isRunning {
		return nil
	}

	log.Printf("Starting run custodian for job %s...", worker.Job.Name)
	fullName := "scheduler.job." + worker.Job.Name
	worker.statusQueue = fullName + ".status"
	worker.actionQueue = fullName + ".action"
	err := worker.MessageBus.Register(
		fullName,
		map[string][]string{
			worker.actionQueue: {"action"},
			worker.statusQueue: {"status"},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to register job %s to message bus: %s", worker.Job.Name, err)
	}

	worker.quit = make(chan struct{})
	go worker.process(wg)
	worker.isRunning = true

	log.Printf("Started run custodian for job %s", worker.Job.Name)
	return nil
}

func (worker *RunCustodian) Stop() {
	worker.isRunningLock.Lock()
	defer worker.isRunningLock.Unlock()

	if !worker.isRunning {
		return
	}

	log.Printf("Stopping run custodian for job %s...", worker.Job.Name)
	worker.quit <- struct{}{}
	worker.isRunning = false
}

func (worker *RunCustodian) process(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	for {
		select {
		case <-worker.quit:
			log.Printf("Stopped run custodian for job %s", worker.Job.Name)
			return
		}
	}
}
