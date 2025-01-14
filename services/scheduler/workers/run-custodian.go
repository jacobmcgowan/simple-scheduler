package workers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jacobmcgowan/simple-scheduler/shared/data-access/repositories"
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
	"github.com/jacobmcgowan/simple-scheduler/shared/jobActions"
	messageBus "github.com/jacobmcgowan/simple-scheduler/shared/message-bus"
	"github.com/jacobmcgowan/simple-scheduler/shared/runStatuses"
)

type RunCustodian struct {
	Job           dtos.Job
	MessageBus    messageBus.MessageBus
	RunRepo       repositories.RunRepository
	Duration      time.Duration
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
	if worker.Job.HeartbeatTimeout <= 0 {
		return nil
	}

	runningStatus := runStatuses.Running
	heartbeatBefore := time.Now().Add(time.Duration(-int(time.Millisecond) * worker.Job.HeartbeatTimeout))
	filter := dtos.RunFilter{
		JobName:         &worker.Job.Name,
		Status:          &runningStatus,
		HeartbeatBefore: &heartbeatBefore,
	}
	runs, err := worker.RunRepo.Browse(filter)
	if err != nil {
		return fmt.Errorf("failed to get runs: %s", err)
	}

	count := 0
	errs := []error{}
	pendingStatus := runStatuses.Pending
	for _, run := range runs {
		runUpdate := dtos.RunUpdate{
			Status: &pendingStatus,
		}
		if err := worker.RunRepo.Edit(run.Id, runUpdate); err != nil {
			errs = append(errs, fmt.Errorf("failed to reset run %s: %s", run.Id, err))
		} else {
			count++
		}
	}

	if count > 0 {
		log.Printf("Reset %d stuck runs for job %s", count, worker.Job.Name)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (worker *RunCustodian) cancelRun(runId string) error {
	cancellingStatus := runStatuses.Cancelling
	runUpdate := dtos.RunUpdate{
		Status: &cancellingStatus,
	}
	if err := worker.RunRepo.Edit(runId, runUpdate); err != nil {
		return fmt.Errorf("failed to cancel run %s: %s", runId, err)
	}

	body, err := json.Marshal(dtos.JobActionMessage{
		RunId:  runId,
		Action: string(jobActions.Cancel),
	})
	if err != nil {
		return fmt.Errorf("failed to serialize run action %s: %s", runId, err)
	}

	err = worker.MessageBus.Publish(
		"scheduler.job."+worker.Job.Name,
		"action",
		body,
	)
	if err != nil {
		return fmt.Errorf("failed to publish cancel action for run %s: %s", runId, err)
	}

	return nil
}

func (worker *RunCustodian) cancelRuns(runs []dtos.Run, reason string) error {
	count := 0
	errs := []error{}
	for _, run := range runs {
		if err := worker.cancelRun(run.Id); err != nil {
			errs = append(errs, err)
		} else {
			count++
		}
	}

	if count > 0 {
		log.Printf("Cancelled %d runs for job %s because of %s", count, worker.Job.Name, reason)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (worker *RunCustodian) cancelTimeoutPendingRuns() error {
	if worker.Job.RunStartTimeout <= 0 {
		return nil
	}

	pendingStatus := runStatuses.Pending
	createdBefore := time.Now().Add(time.Duration(-int(time.Millisecond) * worker.Job.RunStartTimeout))
	filter := dtos.RunFilter{
		JobName:       &worker.Job.Name,
		Status:        &pendingStatus,
		CreatedBefore: &createdBefore,
	}
	runs, err := worker.RunRepo.Browse(filter)
	if err != nil {
		return fmt.Errorf("failed to get runs: %s", err)
	}

	return worker.cancelRuns(runs, "run start timeout")
}

func (worker *RunCustodian) cancelTimeoutRunningRuns() error {
	if worker.Job.RunExecutionTimeout <= 0 {
		return nil
	}

	runningStatus := runStatuses.Running
	startedBefore := time.Now().Add(time.Duration(-int(time.Millisecond) * worker.Job.RunExecutionTimeout))
	filter := dtos.RunFilter{
		JobName:       &worker.Job.Name,
		Status:        &runningStatus,
		StartedBefore: &startedBefore,
	}
	runs, err := worker.RunRepo.Browse(filter)
	if err != nil {
		return fmt.Errorf("failed to get runs: %s", err)
	}

	return worker.cancelRuns(runs, "run execution timeout")
}

func (worker *RunCustodian) clean() error {
	restartErr := worker.restartStuckRuns()
	pendingErr := worker.cancelTimeoutPendingRuns()
	runningErr := worker.cancelTimeoutRunningRuns()

	return errors.Join(restartErr, pendingErr, runningErr)
}

func (worker *RunCustodian) process(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	for {
		select {
		case <-worker.quit:
			log.Printf("Stopped run custodian for job %s", worker.Job.Name)
			return
		case <-time.After(worker.Duration):
			if err := worker.clean(); err != nil {
				log.Printf("Failed to clean runs for job %s: %s", worker.Job.Name, err)
			}
		}
	}
}
