package workers

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jacobmcgowan/simple-scheduler/common"
	"github.com/jacobmcgowan/simple-scheduler/data-access/repositories"
	"github.com/jacobmcgowan/simple-scheduler/dtos"
	messageBus "github.com/jacobmcgowan/simple-scheduler/message-bus"
	"github.com/jacobmcgowan/simple-scheduler/runStatuses"
)

type JobWorker struct {
	Job           dtos.Job
	MessageBus    messageBus.MessageBus
	JobRepo       repositories.JobRepository
	RunRepo       repositories.RunRepository
	quit          chan struct{}
	isRunningLock sync.Mutex `default:"sync.Mutex{}"`
	isRunning     bool
}

func (worker *JobWorker) Start(wg *sync.WaitGroup) error {
	worker.isRunningLock.Lock()
	defer worker.isRunningLock.Unlock()

	if worker.isRunning {
		return nil
	}

	log.Printf("Starting job %s...", worker.Job.Name)
	fullName := "scheduler.job." + worker.Job.Name
	err := worker.MessageBus.Register(
		fullName,
		map[string][]string{
			fullName + ".action": {"action"},
			fullName + ".status": {"status"},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to register job %s to message bus: %s", worker.Job.Name, err)
	}

	err = worker.MessageBus.Subscribe(
		wg,
		fullName+".status",
		worker.statusMessageReceived,
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe to status queue for job %s: %s", worker.Job.Name, err)
	}

	wg.Add(1)
	worker.quit = make(chan struct{})
	go worker.process(wg)
	worker.isRunning = true

	log.Printf("Started job %s", worker.Job.Name)
	return nil
}

func (worker *JobWorker) Stop() {
	worker.isRunningLock.Lock()
	defer worker.isRunningLock.Unlock()

	log.Printf("Stopping job %s...", worker.Job.Name)
	worker.quit <- struct{}{}
	worker.isRunning = false
}

func (worker *JobWorker) statusMessageReceived(body []byte) bool {
	log.Printf("Job %s status message received: %s", worker.Job.Name, body)
	return true
}

func (worker *JobWorker) setNextRunTime() error {
	elapsed := time.Since(worker.Job.NextRunAt)

	if worker.Job.Interval <= 0 || elapsed.Milliseconds() <= 0 {
		return nil
	}

	intervals := (elapsed.Milliseconds() / int64(worker.Job.Interval)) + 1
	nextRunAt := worker.Job.NextRunAt.Add(time.Duration(worker.Job.Interval * int(intervals)))
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
		JobName:   worker.Job.Name,
		Status:    runStatuses.Pending,
		StartTime: worker.Job.NextRunAt,
	}
	runId, err := worker.RunRepo.Add(run)
	if err != nil {
		return fmt.Errorf("failed to add run for job %s: %s", worker.Job.Name, err)
	}

	body, err := json.Marshal(JobActionMessage{
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

	if err = worker.setNextRunTime(); err != nil {
		return fmt.Errorf("failed to update next run time for job %s: %s", worker.Job.Name, err)
	}

	return nil
}

func (worker *JobWorker) process(wg *sync.WaitGroup) {
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
		}
	}
}
