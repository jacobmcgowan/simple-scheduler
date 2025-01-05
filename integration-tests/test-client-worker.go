package integration_tests

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
	"github.com/jacobmcgowan/simple-scheduler/shared/jobActions"
	messageBus "github.com/jacobmcgowan/simple-scheduler/shared/message-bus"
	"github.com/jacobmcgowan/simple-scheduler/shared/runStatuses"
)

type TestClientWorker struct {
	Job           dtos.Job
	MessageBus    messageBus.MessageBus
	RunStarted    func(runId string)
	RunCanceled   func(runId string)
	quit          chan struct{}
	isRunningLock sync.Mutex `default:"sync.Mutex{}"`
	isRunning     bool
	actionQueue   string
	statusQueue   string
}

func (worker *TestClientWorker) Start(wg *sync.WaitGroup) error {
	worker.isRunningLock.Lock()
	defer worker.isRunningLock.Unlock()

	if worker.isRunning {
		return nil
	}

	log.Printf("Starting client for job %s...", worker.Job.Name)
	fullName := "scheduler.job." + worker.Job.Name
	worker.actionQueue = fullName + ".action"
	worker.statusQueue = fullName + ".status"
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

func (worker *TestClientWorker) updateRunStatus(runId string, status runStatuses.RunStatus) error {
	worker.isRunningLock.Lock()
	defer worker.isRunningLock.Unlock()

	if !worker.isRunning {
		return nil
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
	<-worker.quit
	log.Printf("Stopped client for job %s", worker.Job.Name)
}
