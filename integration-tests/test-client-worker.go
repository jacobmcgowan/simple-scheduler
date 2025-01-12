package integration_tests

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
	"github.com/jacobmcgowan/simple-scheduler/shared/jobActions"
	messageBus "github.com/jacobmcgowan/simple-scheduler/shared/message-bus"
	"github.com/jacobmcgowan/simple-scheduler/shared/runStatuses"
)

type TestClientWorker struct {
	Job               dtos.Job
	MessageBus        messageBus.MessageBus
	HeartbeatDuration time.Duration
	RunStarted        func(runId string)
	RunCanceled       func(runId string)
	quit              chan struct{}
	isRunningLock     sync.Mutex `default:"sync.Mutex{}"`
	isRunning         bool
	actionQueue       string
	statusQueue       string
	heartbeatQueue    string
	runsLock          sync.RWMutex `default:"sync.RWMutex{}"`
	runs              map[string]bool
}

func (worker *TestClientWorker) Start(wg *sync.WaitGroup) error {
	worker.isRunningLock.Lock()
	defer worker.isRunningLock.Unlock()

	if worker.isRunning {
		return nil
	}

	log.Printf("Starting client for job %s...", worker.Job.Name)
	worker.clearRuns()

	fullName := "scheduler.job." + worker.Job.Name
	worker.actionQueue = fullName + ".action"
	worker.statusQueue = fullName + ".status"
	worker.heartbeatQueue = fullName + ".heartbeat"
	err := worker.MessageBus.Register(
		fullName,
		map[string][]string{
			worker.actionQueue:    {"action"},
			worker.statusQueue:    {"status"},
			worker.heartbeatQueue: {"heartbeat"},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to register job %s to message bus: %s", worker.Job.Name, err)
	}

	err = worker.MessageBus.Subscribe(
		wg,
		worker.actionQueue,
		worker.actionMessageReceived,
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe to action queue for job %s: %s", worker.Job.Name, err)
	}

	worker.quit = make(chan struct{})
	go worker.process(wg)
	worker.isRunning = true

	log.Printf("Started client for job %s", worker.Job.Name)
	return nil
}

func (worker *TestClientWorker) Stop() {
	worker.isRunningLock.Lock()
	defer worker.isRunningLock.Unlock()

	if !worker.isRunning {
		return
	}

	log.Printf("Stopping client for job %s...", worker.Job.Name)
	worker.MessageBus.Unsubscribe(worker.actionQueue)
	worker.quit <- struct{}{}
	worker.isRunning = false
}

func (worker *TestClientWorker) CompleteRun(runId string) error {
	return worker.updateRunStatus(runId, runStatuses.Completed)
}

func (worker *TestClientWorker) FailRun(runId string) error {
	return worker.updateRunStatus(runId, runStatuses.Failed)
}

func (worker *TestClientWorker) clearRuns() {
	worker.runsLock.Lock()
	defer worker.runsLock.Unlock()
	worker.runs = map[string]bool{}
}

func (worker *TestClientWorker) startRun(runId string) {
	worker.runsLock.Lock()
	defer worker.runsLock.Unlock()
	worker.runs[runId] = true
}

func (worker *TestClientWorker) stopRun(runId string) {
	worker.runsLock.Lock()
	defer worker.runsLock.Unlock()
	worker.runs[runId] = false
}

func (worker *TestClientWorker) updateRunStatus(runId string, status runStatuses.RunStatus) error {
	worker.isRunningLock.Lock()
	defer worker.isRunningLock.Unlock()

	if !worker.isRunning {
		return nil
	}

	switch status {
	case runStatuses.Cancelled, runStatuses.Completed, runStatuses.Failed:
		worker.stopRun(runId)
	case runStatuses.Pending, runStatuses.Cancelling:
		// do nothing
	case runStatuses.Running:
		worker.startRun(runId)
	default:
		return fmt.Errorf("unsupported status %s for run %s", status, runId)
	}

	body, err := json.Marshal(dtos.JobStatusMessage{
		JobName: worker.Job.Name,
		RunId:   runId,
		Status:  string(status),
	})
	if err != nil {
		return fmt.Errorf("failed to serialize job status %s for run %s: %s", status, runId, err)
	}

	err = worker.MessageBus.Publish(
		"scheduler.job."+worker.Job.Name,
		"status",
		body,
	)
	if err != nil {
		return fmt.Errorf("failed to publish job status %s for run %s: %s", status, runId, err)
	}

	return nil
}

func (worker *TestClientWorker) updateRunHeartbeat(runId string) error {
	worker.isRunningLock.Lock()
	defer worker.isRunningLock.Unlock()

	if !worker.isRunning {
		return nil
	}

	body, err := json.Marshal(dtos.JobHeartbeatMessage{
		JobName: worker.Job.Name,
		RunId:   runId,
	})
	if err != nil {
		return fmt.Errorf("failed to serialize job heartbeat for run %s: %s", runId, err)
	}

	err = worker.MessageBus.Publish(
		"scheduler.job."+worker.Job.Name,
		"heartbeat",
		body,
	)
	if err != nil {
		return fmt.Errorf("failed to publish job heartbeat for run %s: %s", runId, err)
	}

	return nil
}

func (worker *TestClientWorker) updateRunHeartbeats() error {
	worker.runsLock.RLock()
	defer worker.runsLock.RUnlock()

	errs := []error{}
	for runId, running := range worker.runs {
		if running {
			if err := worker.updateRunHeartbeat(runId); err != nil {
				errs = append(errs, err)
			}
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (worker *TestClientWorker) actionMessageReceived(body []byte) (error, bool) {
	log.Printf("Client for job %s action message received: %s", worker.Job.Name, body)
	var actionMsg dtos.JobActionMessage
	if err := json.Unmarshal(body, &actionMsg); err != nil {
		return fmt.Errorf("failed to parse action message: %s", err), false
	}

	action := jobActions.JobAction(actionMsg.Action)
	switch action {
	case jobActions.Cancel:
		worker.updateRunStatus(actionMsg.RunId, runStatuses.Cancelled)
		if worker.RunCanceled != nil {
			worker.RunCanceled(actionMsg.RunId)
		}
	case jobActions.Run:
		worker.updateRunStatus(actionMsg.RunId, runStatuses.Running)
		if worker.RunStarted != nil {
			worker.RunStarted(actionMsg.RunId)
		}
	default:
		return fmt.Errorf("unsupported action: %s", action), false
	}

	return nil, false
}

func (worker *TestClientWorker) process(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	for {
		select {
		case <-worker.quit:
			log.Printf("Stopped client for job %s", worker.Job.Name)
			return
		case <-time.After(worker.HeartbeatDuration):
			if err := worker.updateRunHeartbeats(); err != nil {
				log.Printf("Failed to update run heartbeats for job %s: %s", worker.Job.Name, err)
			}
		}
	}
}
