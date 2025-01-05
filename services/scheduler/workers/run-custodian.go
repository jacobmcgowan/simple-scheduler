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
	"github.com/jacobmcgowan/simple-scheduler/shared/runStatuses"
)

type RunCustodian struct {
	Job        dtos.Job
	MessageBus messageBus.MessageBus
	RunRepo    repositories.RunRepository
	Interval
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

func (worker *RunCustodian) restartStuckRuns() error {
	filter := dtos.RunFilter{
		JobName:         worker.Job.Name,
		Status:          runStatuses.Running,
		HeartbeatBefore: time.Now().Add(time.Duration(-int(time.Millisecond) * worker.Job.HeartbeatTimeout)),
	}
	runs, err := worker.RunRepo.Browse(filter)
	if err != nil {
		return fmt.Errorf("failed to get runs for job %s: %s", worker.Job.Name, err)
	}

	for _, run := range runs {
		runUpdate := dtos.RunUpdate{
			Status: runStatuses.Pending,
		}
		if err := worker.RunRepo.Edit(run.Id, runUpdate); err != nil {
			return fmt.Errorf("failed to reset run %s for job %s: %s", run.Id, worker.Job.Name, err)
		}
	}

	return nil
}

func (worker *RunCustodian) cancelRun(runId string) error {
	runUpdate := dtos.RunUpdate{
		Status: runStatuses.Cancelling,
	}
	if err := worker.RunRepo.Edit(runId, runUpdate); err != nil {
		return fmt.Errorf("failed to cancel run %s for job %s: %s", runId, worker.Job.Name, err)
	}

	msg := dtos.RunActionMessage{
		RunId:  runId,
		Action: runStatuses.Cancel,
	}
	if err := worker.MessageBus.Publish(worker.actionQueue, msg); err != nil {
		return fmt.Errorf("failed to publish cancel action for run %s: %s", runId, err)
	}

	return nil
}

func (worker *RunCustodian) cancelTimeoutPendingRuns() error {
	filter := dtos.RunFilter{
		JobName:       worker.Job.Name,
		Status:        runStatuses.Pending,
		CreatedBefore: time.Now().Add(time.Duration(-int(time.Millisecond) * worker.Job.RunStartTimeout)),
	}
	runs, err := worker.RunRepo.Browse(filter)
	if err != nil {
		return fmt.Errorf("failed to get runs for job %s: %s", worker.Job.Name, err)
	}

	errs := []error{}
	for _, run := range runs {
		if err := worker.cancelRun(run.Id); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs)
	}

	return nil
}

func (worker *RunCustodian) cancelTimeoutRunningRuns() error {
	filter := dtos.RunFilter{
		JobName:       worker.Job.Name,
		Status:        runStatuses.Running,
		StartedBefore: time.Now().Add(time.Duration(-int(time.Millisecond) * worker.Job.RunExecutionTimeout)),
	}
	runs, err := worker.RunRepo.Browse(filter)
	if err != nil {
		return fmt.Errorf("failed to get runs for job %s: %s", worker.Job.Name, err)
	}

	errs := []error{}
	for _, run := range runs {
		if err := worker.cancelRun(run.Id); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs)
	}

	return nil
}

func (worker *RunCustodian) clean() error {
	restartErr := worker.restartStuckRuns()
	cancelTimeoutPendingErr := worker.cancelTimeoutPendingRuns()
	cancelTimeoutRunningErr := worker.cancelTimeoutRunningRuns()

	return errors.Join(restartErr, cancelTimeoutPendingErr, cancelTimeoutRunningErr)
}

func (worker *RunCustodian) process(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	for {
		select {
		case <-worker.quit:
			log.Printf("Stopped run custodian for job %s", worker.Job.Name)
			return
		case <-time.After(time.Duration(int(time.Millisecond) * worker.Interval)):
			if err := worker.clean(); err != nil {
				log.Printf("Failed to clean runs for job %s: %s", worker.Job.Name, err)
			}
		}
	}
}
