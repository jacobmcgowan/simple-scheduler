package workers

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jacobmcgowan/simple-scheduler/shared/data-access/repositories"
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
)

type JobCustodian struct {
	JobRepo          repositories.JobRepository
	Duration         time.Duration
	HeartbeatTimeout time.Duration
	quit             chan struct{}
	isRunningLock    sync.Mutex `default:"sync.Mutex{}"`
	isRunning        bool
	stopOnce         sync.Once
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
	worker.stopOnce.Do(func() {
		worker.isRunningLock.Lock()
		defer worker.isRunningLock.Unlock()

		if !worker.isRunning {
			return
		}

		log.Printf("Stopping job custodian...")
		close(worker.quit)
	})
}

func (worker *JobCustodian) stopped() {
	worker.isRunningLock.Lock()
	defer worker.isRunningLock.Unlock()
	worker.isRunning = false
}

func (worker *JobCustodian) process(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	ticker := time.NewTicker(worker.Duration)
	defer ticker.Stop()

	for {
		select {
		case <-worker.quit:
			log.Printf("Stopped job custodian")
			worker.stopped()
			return
		case <-ticker.C:
			if count, err := worker.restartStuckJobs(); err != nil {
				log.Printf("Failed to restart stuck jobs: %s", err)
			} else {
				log.Printf("Restarted %d stuck jobs", count)
			}
		}
	}
}
