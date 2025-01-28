package workers

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jacobmcgowan/simple-scheduler/shared/data-access/repositories"
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
	messageBus "github.com/jacobmcgowan/simple-scheduler/shared/message-bus"
)

type JobCustodian struct {
	MessageBus       messageBus.MessageBus
	JobRepo          repositories.JobRepository
	Duration         time.Duration
	HeartbeatTimeout time.Duration
	quit             chan struct{}
	isRunningLock    sync.Mutex `default:"sync.Mutex{}"`
	isRunning        bool
}

func (worker *JobCustodian) Start(wg *sync.WaitGroup) error {
	worker.isRunningLock.Lock()
	defer worker.isRunningLock.Unlock()

	if worker.isRunning {
		return nil
	}

	log.Printf("Starting job custodian...")
	worker.quit = make(chan struct{})
	go worker.process(wg)
	worker.isRunning = true

	log.Printf("Started job custodian")
	return nil
}

func (worker *JobCustodian) restartStuckJobs() (int64, error) {
	hrtbtBefore := time.Now().Add(-worker.HeartbeatTimeout)
	filter := dtos.JobUnlockFilter{
		IsManaged:       true,
		HeartbeatBefore: &hrtbtBefore,
	}
	count, err := worker.JobRepo.Unlock(filter)
	if err != nil {
		return 0, fmt.Errorf("failed to unlock jobs: %s", err)
	}

	return count, nil
}

func (worker *JobCustodian) Stop() {
	worker.isRunningLock.Lock()
	defer worker.isRunningLock.Unlock()

	if !worker.isRunning {
		return
	}

	log.Printf("Stopping job custodian...")
	worker.quit <- struct{}{}
	worker.isRunning = false
}

func (worker *JobCustodian) process(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	for {
		select {
		case <-worker.quit:
			log.Printf("Stopped job custodian")
			return
		case <-time.After(worker.Duration):
			if count, err := worker.restartStuckJobs(); err != nil {
				log.Printf("Failed to restart stuck jobs: %s", err)
			} else {
				log.Printf("Restarted %d stuck jobs", count)
			}
		}
	}
}
