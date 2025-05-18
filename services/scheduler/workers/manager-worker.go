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
	HeartbeatDuration    time.Duration
	nextCacheRefreshAt   time.Time
	jobsLock             sync.Mutex `default:"sync.Mutex{}"`
	jobs                 map[string]*JobWorker
	custodians           map[string]*RunCustodian
	quit                 chan struct{}
	isRunningLock        sync.Mutex `default:"sync.Mutex{}"`
	isRunning            bool
	stopOnce             sync.Once
}

func (worker *ManagerWorker) Start(wg *sync.WaitGroup) error {
	worker.isRunningLock.Lock()
	defer worker.isRunningLock.Unlock()

	if worker.isRunning {
		return nil
	}

	log.Printf("Starting job manager @%s...\n", worker.Hostname)
	if err := worker.registerWorker(); err != nil {
		return fmt.Errorf("failed to start job manager @%s: %s", worker.Hostname, err)
	}

	worker.jobsLock.Lock()
	defer worker.jobsLock.Unlock()

	worker.jobs = make(map[string]*JobWorker)
	worker.custodians = make(map[string]*RunCustodian)
	worker.quit = make(chan struct{})
	worker.nextCacheRefreshAt = time.Now()

	go worker.process(wg)
	worker.isRunning = true
	log.Printf("Started job manager %s@%s", worker.Id, worker.Hostname)

	return nil
}

func (worker *ManagerWorker) Stop() {
	worker.stopOnce.Do(func() {
		log.Printf("Stopping job manager %s@%s...", worker.Id, worker.Hostname)
		close(worker.quit)
	})
}

func (worker *ManagerWorker) stopped() {
	worker.isRunningLock.Lock()
	defer worker.isRunningLock.Unlock()
	worker.isRunning = false
}

func (worker *ManagerWorker) registerWorker() error {
	log.Printf("Registering job manager @%s...", worker.Hostname)
	mngr := dtos.Manager{
		Hostname: worker.Hostname,
	}
	id, err := worker.ManagerRepo.Add(mngr)
	if err != nil {
		return fmt.Errorf("failed to register manager @%s: %s", worker.Hostname, err)
	}

	log.Printf("Registered job manager %s@%s", id, worker.Hostname)

	worker.Id = id
	return nil
}

func (worker *ManagerWorker) refreshCache(wg *sync.WaitGroup) error {
	worker.jobsLock.Lock()
	defer worker.jobsLock.Unlock()

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
		log.Printf("Locked job %s for manager %s@%s", job.Name, worker.Id, worker.Hostname)
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

	if len(unlockJobNames) > 0 {
		unlockFilter := dtos.JobUnlockFilter{
			ManagerId: &worker.Id,
			JobNames:  unlockJobNames,
		}

		if unlockCount, err := worker.JobRepo.Unlock(unlockFilter); err != nil {
			jobErrs = append(jobErrs, fmt.Errorf("failed to unlock jobs: %s", err))
		} else {
			log.Printf("Unlocked %d jobs for manager %s@%s\n", unlockCount, worker.Id, worker.Hostname)
		}
	}

	if len(jobErrs) > 0 {
		return errors.Join(jobErrs...)
	}

	return nil
}

func (worker *ManagerWorker) setHeartbeat() error {
	if err := worker.JobRepo.Heartbeat(worker.Id); err != nil {
		return fmt.Errorf("failed to set heartbeat: %s", err)
	}

	return nil
}

func (worker *ManagerWorker) stopAllJobs() error {
	worker.jobsLock.Lock()
	defer worker.jobsLock.Unlock()

	for name, job := range worker.jobs {
		job.Stop()
		delete(worker.jobs, name)
	}

	for name, custodian := range worker.custodians {
		custodian.Stop()
		delete(worker.custodians, name)
	}

	unlockFilter := dtos.JobUnlockFilter{
		ManagerId: &worker.Id,
	}
	if _, err := worker.JobRepo.Unlock(unlockFilter); err != nil {
		return fmt.Errorf("failed to unlock jobs: %s", err)
	}

	return nil
}

func (worker *ManagerWorker) process(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	cacheRefreshTimer := time.NewTimer(time.Until(worker.nextCacheRefreshAt))
	defer cacheRefreshTimer.Stop()

	hrtbtTicker := time.NewTicker(worker.HeartbeatDuration)
	defer hrtbtTicker.Stop()

	for {
		select {
		case <-worker.quit:
			if err := worker.stopAllJobs(); err != nil {
				log.Printf("Error occurred when stopping jobs for manager %s@%s: %s", worker.Id, worker.Hostname, err)
			}

			log.Printf("Stopped job manager %s@%s\n", worker.Id, worker.Hostname)
			worker.stopped()
			return
		case <-cacheRefreshTimer.C:
			log.Printf("Refreshing jobs cache for manager %s@%s...", worker.Id, worker.Hostname)
			if err := worker.refreshCache(wg); err != nil {
				log.Printf("Failed to refresh jobs cache for manager %s@%s: %s", worker.Id, worker.Hostname, err)
			} else {
				log.Printf("Refreshed jobs cache, %d loaded, for manager %s@%s", len(worker.jobs), worker.Id, worker.Hostname)
			}
			cacheRefreshTimer.Reset(time.Until(worker.nextCacheRefreshAt))
		case <-hrtbtTicker.C:
			log.Printf("Setting heartbeat of jobs for manager %s@%s...", worker.Id, worker.Hostname)
			if err := worker.setHeartbeat(); err != nil {
				log.Printf("Failed to set heartbeat for manager %s@%s: %s", worker.Id, worker.Hostname, err)
			} else {
				log.Printf("Set heartbeat of jobs for manager %s@%s", worker.Id, worker.Hostname)
			}
		}
	}
}
