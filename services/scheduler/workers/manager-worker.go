package workers

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jacobmcgowan/simple-scheduler/shared/data-access/repositories"
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
	messageBus "github.com/jacobmcgowan/simple-scheduler/shared/message-bus"
)

type ManagerWorker struct {
	Id                   string
	Hostname             string
	MaxJobs              int
	MessageBus           messageBus.MessageBus
	ManagerRepo          repositories.ManagerRepository
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

func (worker *ManagerWorker) Start(wg *sync.WaitGroup) error {
	worker.isRunningLock.Lock()
	defer worker.isRunningLock.Unlock()

	if worker.isRunning {
		return nil
	}

	log.Println("Starting job manager...")
	if err := worker.registerWorker(); err != nil {
		return fmt.Errorf("failed to start job manager: %s", err)
	}

	worker.jobs = make(map[string]*JobWorker)
	worker.custodians = make(map[string]*RunCustodian)
	worker.quit = make(chan struct{})
	worker.nextCacheRefreshAt = time.Now()

	go worker.process(wg)
	worker.isRunning = true
	log.Printf("Started job manager %s\n", worker.Id)

	return nil
}

func (worker *ManagerWorker) Stop() {
	worker.isRunningLock.Lock()
	defer worker.isRunningLock.Unlock()

	log.Println("Stopping job manager...")
	worker.quit <- struct{}{}
	worker.isRunning = false
}

func (worker *ManagerWorker) registerWorker() error {
	log.Println("Registering job manager...")
	mngr := dtos.Manager{
		Hostname: worker.Hostname,
	}
	id, err := worker.ManagerRepo.Add(mngr)
	if err != nil {
		return fmt.Errorf("failed to register manager: %s", err)
	}

	log.Printf("Registered job manager %s\n", id)

	worker.Id = id
	return nil
}

func (worker *ManagerWorker) refreshCache(wg *sync.WaitGroup) error {
	worker.nextCacheRefreshAt = time.Now().Add(worker.CacheRefreshDuration)
	refreshedJobs := make(map[string]bool)
	filter := dtos.JobLockFilter{
		ManagerId: worker.Id,
		Take:      worker.MaxJobs,
	}
	jobs, err := worker.JobRepo.Lock(filter)
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

	jobErrs := []error{}
	unlockJobNames := []string{}
	for name, job := range worker.jobs {
		_, refreshed := refreshedJobs[name]
		if refreshed {
			if err = job.Start(wg); err != nil {
				jobErrs = append(jobErrs, fmt.Errorf("failed to start job %s: %s", name, err))
			}
		} else {
			unlockJobNames = append(unlockJobNames, job.Job.Name)
			job.Stop()
			delete(worker.jobs, name)
		}
	}

	for name, custodian := range worker.custodians {
		_, refreshed := refreshedJobs[name]
		if refreshed {
			if err = custodian.Start(wg); err != nil {
				jobErrs = append(jobErrs, fmt.Errorf("failed to start custodian for job %s: %s", name, err))
			}
		} else {
			custodian.Stop()
			delete(worker.custodians, name)
		}
	}

	unlockFilter := dtos.JobUnlockFilter{
		ManagerId: worker.Id,
		JobNames:  unlockJobNames,
	}
	if err = worker.JobRepo.Unlock(unlockFilter); err != nil {
		jobErrs = append(jobErrs, fmt.Errorf("failed to unlock jobs: %s", err))
	}

	if len(jobErrs) > 0 {
		return errors.Join(jobErrs...)
	}

	return nil
}

func (worker *ManagerWorker) stopAllJobs() error {
	for name, job := range worker.jobs {
		job.Stop()
		delete(worker.jobs, name)
	}

	for name, custodian := range worker.custodians {
		custodian.Stop()
		delete(worker.custodians, name)
	}

	unlockFilter := dtos.JobUnlockFilter{
		ManagerId: worker.Id,
	}
	if err := worker.JobRepo.Unlock(unlockFilter); err != nil {
		return fmt.Errorf("failed to unlock jobs: %s", err)
	}

	return nil
}

func (worker *ManagerWorker) process(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	for {
		select {
		case <-worker.quit:
			if err := worker.stopAllJobs(); err != nil {
				log.Printf("Error occurred when stopping jobs: %s", err)
			}

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
