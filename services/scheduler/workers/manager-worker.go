package workers

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jacobmcgowan/simple-scheduler/shared/data-access/repositories"
	messageBus "github.com/jacobmcgowan/simple-scheduler/shared/message-bus"
)

type ManagerWorker struct {
	MessageBus           messageBus.MessageBus
	JobRepo              repositories.JobRepository
	RunRepo              repositories.RunRepository
	CacheRefreshDuration time.Duration
	CleanupDuration      time.Duration
	nextCacheRefreshAt   time.Time
	jobs                 map[string]*JobWorker
	custodians           map[string]*RunCustodian
	quit                 chan struct{}
	isRunningLock        sync.Mutex `default:"sync.Mutex{}"`
	isRunning            bool
}

func (worker *ManagerWorker) Start(wg *sync.WaitGroup) {
	worker.isRunningLock.Lock()
	defer worker.isRunningLock.Unlock()

	if worker.isRunning {
		return
	}

	log.Println("Starting job manager...")
	worker.jobs = make(map[string]*JobWorker)
	worker.custodians = make(map[string]*RunCustodian)
	worker.quit = make(chan struct{})
	worker.nextCacheRefreshAt = time.Now()

	go worker.process(wg)
	worker.isRunning = true
	log.Println("Started job manager")
}

func (worker *ManagerWorker) Stop() {
	worker.isRunningLock.Lock()
	defer worker.isRunningLock.Unlock()

	log.Println("Stopping job manager...")
	worker.quit <- struct{}{}
	worker.isRunning = false
}

func (worker *ManagerWorker) refreshCache(wg *sync.WaitGroup) error {
	worker.nextCacheRefreshAt = time.Now().Add(worker.CacheRefreshDuration)
	refreshedJobs := make(map[string]bool)
	jobs, err := worker.JobRepo.Browse()
	if err != nil {
		return fmt.Errorf("failed to get jobs: %s", err)
	}

	for _, job := range jobs {
		jobWorker, found := worker.jobs[job.Name]
		if found {
			jobWorker.Job = job
		} else {
			worker.jobs[job.Name] = &JobWorker{
				Job:        job,
				MessageBus: worker.MessageBus,
				JobRepo:    worker.JobRepo,
				RunRepo:    worker.RunRepo,
			}
		}

		runCustodian, found := worker.custodians[job.Name]
		if found {
			runCustodian.Job = job
		} else {
			worker.custodians[job.Name] = &RunCustodian{
				Job:        job,
				MessageBus: worker.MessageBus,
				RunRepo:    worker.RunRepo,
				Duration:   worker.CleanupDuration,
			}
		}

		refreshedJobs[job.Name] = true
	}

	for name, job := range worker.jobs {
		_, refreshed := refreshedJobs[name]
		if refreshed {
			job.Start(wg)
		} else {
			job.Stop()
			delete(worker.jobs, name)
		}
	}

	for name, custodian := range worker.custodians {
		_, refreshed := refreshedJobs[name]
		if refreshed {
			custodian.Start(wg)
		} else {
			custodian.Stop()
			delete(worker.custodians, name)
		}
	}

	return nil
}

func (worker *ManagerWorker) stopAllJobs() {
	for name, job := range worker.jobs {
		job.Stop()
		delete(worker.jobs, name)
	}

	for name, custodian := range worker.custodians {
		custodian.Stop()
		delete(worker.custodians, name)
	}
}

func (worker *ManagerWorker) process(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	for {
		select {
		case <-worker.quit:
			worker.stopAllJobs()
			log.Println("Stopped job manager")
			return
		case <-time.After(time.Until(worker.nextCacheRefreshAt)):
			log.Println("Refreshing jobs cache...")
			if err := worker.refreshCache(wg); err != nil {
				log.Printf("Failed to refresh jobs cache: %s", err)
			} else {
				log.Printf("Refreshed jobs cache, %d loaded", len(worker.jobs))
			}
		}
	}
}
