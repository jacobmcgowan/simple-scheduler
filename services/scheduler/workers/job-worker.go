package workers

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jacobmcgowan/simple-scheduler/shared/common"
	"github.com/jacobmcgowan/simple-scheduler/shared/data-access/repositories"
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
	messageBus "github.com/jacobmcgowan/simple-scheduler/shared/message-bus"
	"github.com/jacobmcgowan/simple-scheduler/shared/runStatuses"
)

type JobWorker struct {
	Job           dtos.Job
	MessageBus    messageBus.MessageBus
	JobRepo       repositories.JobRepository
	RunRepo       repositories.RunRepository
	quit          chan struct{}
	isRunningLock sync.Mutex `default:"sync.Mutex{}"`
	isRunning     bool
	actionQueue   string
	statusQueue   string
}

func (worker *JobWorker) Start(wg *sync.WaitGroup) error {
	worker.isRunningLock.Lock()
	defer worker.isRunningLock.Unlock()

	if worker.isRunning {
		return nil
	}

	log.Printf("Starting job %s...", worker.Job.Name)
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

	err = worker.MessageBus.Subscribe(
		wg,
		worker.statusQueue,
		worker.statusMessageReceived,
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe to status queue for job %s: %s", worker.Job.Name, err)
	}

	worker.quit = make(chan struct{})
	go worker.process(wg)
	worker.isRunning = true

	log.Printf("Started job %s", worker.Job.Name)
	return nil
}

func (worker *JobWorker) Stop() {
	worker.isRunningLock.Lock()
	defer worker.isRunningLock.Unlock()

	if !worker.isRunning {
		return
	}

	log.Printf("Stopping job %s...", worker.Job.Name)
	worker.MessageBus.Unsubscribe(worker.statusQueue)
	worker.quit <- struct{}{}
	worker.isRunning = false
}

func (worker *JobWorker) statusMessageReceived(body []byte) (error, bool) {
	log.Printf("Job %s status message received: %s", worker.Job.Name, body)
	var statusMsg dtos.JobStatusMessage
	if err := json.Unmarshal(body, &statusMsg); err != nil {
		return fmt.Errorf("failed to deserialize status message for job %s: %s", worker.Job.Name, err), false
	}

	status := runStatuses.RunStatus(statusMsg.Status)
	switch status {
	case runStatuses.Cancelled,
		runStatuses.Cancelling,
		runStatuses.Completed,
		runStatuses.Failed,
		runStatuses.Pending,
		runStatuses.Running:
		worker.updateRunStatus(statusMsg.RunId, status)
	default:
		return fmt.Errorf("unsupported status %s for job %s", status, worker.Job.Name), false
	}

	return nil, false
}

func (worker *JobWorker) setNextRunTime() error {
	elapsed := time.Since(worker.Job.NextRunAt)

	if worker.Job.Interval <= 0 || elapsed.Milliseconds() <= 0 {
		return nil
	}

	intervals := (elapsed.Milliseconds() / int64(worker.Job.Interval)) + 1
	tilNextRun := time.Duration(worker.Job.Interval * int(intervals) * int(time.Millisecond))
	nextRunAt := worker.Job.NextRunAt.Add(tilNextRun)
	update := dtos.JobUpdate{
		NextRunAt: common.Undefinable[time.Time]{
			Value:   nextRunAt,
			Defined: true,
		},
	}

	if err := worker.JobRepo.Edit(worker.Job.Name, update); err != nil {
		return fmt.Errorf("failed to set next run time: %s", err)
	}

	worker.Job.NextRunAt = nextRunAt

	return nil
}

func (worker *JobWorker) startRun() error {
	run := dtos.Run{
		JobName:     worker.Job.Name,
		Status:      runStatuses.Pending,
		CreatedTime: worker.Job.NextRunAt,
	}
	runId, err := worker.RunRepo.Add(run)
	if err != nil {
		return fmt.Errorf("failed to add run for job %s: %s", worker.Job.Name, err)
	}

	body, err := json.Marshal(dtos.JobActionMessage{
		JobName: worker.Job.Name,
		RunId:   runId,
		Action:  "run",
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
		return fmt.Errorf("failed to publish run action %s: %s", runId, err)
	}

	return nil
}

func (worker *JobWorker) updateRunStatus(runId string, status runStatuses.RunStatus) error {
	runUpdate := dtos.RunUpdate{
		Status: common.Undefinable[runStatuses.RunStatus]{
			Value:   status,
			Defined: true,
		},
		StartTime: common.Undefinable[time.Time]{
			Value:   time.Now(),
			Defined: status == runStatuses.Running,
		},
		EndTime: common.Undefinable[time.Time]{
			Value: time.Now(),
			Defined: status == runStatuses.Completed ||
				status == runStatuses.Failed ||
				status == runStatuses.Cancelled,
		},
	}

	if err := worker.RunRepo.Edit(runId, runUpdate); err != nil {
		return fmt.Errorf("failed to edit run %s for job %s: %s", runId, worker.Job.Name, err)
	}

	return nil
}

func (worker *JobWorker) process(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	for {
		select {
		case <-worker.quit:
			log.Printf("Stopped job %s", worker.Job.Name)
			return
		case <-time.After(time.Until(worker.Job.NextRunAt)):
			log.Printf("Starting run for job %s...", worker.Job.Name)

			if err := worker.startRun(); err != nil {
				log.Printf("Failed to start run for job %s: %s", worker.Job.Name, err)
			} else {
				log.Printf("Started run for job %s", worker.Job.Name)
			}

			if err := worker.setNextRunTime(); err != nil {
				log.Printf("Failed to update next run time for job %s: %s", worker.Job.Name, err)
				worker.Stop()
			} else {
				log.Printf("Next run for job %s is at %s", worker.Job.Name, worker.Job.NextRunAt.String())
			}
		}
	}
}
