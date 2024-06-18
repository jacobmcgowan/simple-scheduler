package workers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jacobmcgowan/simple-scheduler/common"
	"github.com/jacobmcgowan/simple-scheduler/data-access/repositories"
	"github.com/jacobmcgowan/simple-scheduler/dtos"
	"github.com/jacobmcgowan/simple-scheduler/runStatuses"
)

type JobWorker struct {
	Job        dtos.Job
	Interval   int
	MessageBus MessageBus
	JobRepo    repositories.JobRepository
	RunRepo    repositories.RunRepository
	running    chan bool
}

func (worker JobWorker) Start() error {
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
		"scheduler.job."+worker.Job.Name,
		"status",
		"scheduler.job.status",
		worker.messageReceived,
	)
	if err != nil {
		return err
	}

	worker.running = make(chan bool)
	go worker.process()

	return nil
}

func (worker JobWorker) messageReceived(body []byte) bool {
	fmt.Printf("Job status message received: %s", body)
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

func (worker *JobWorker) process() {
	for {
		if <-worker.running {
			return
		}

		if worker.Job.NextRunAt.Compare(time.Now()) >= 0 {
			if err := worker.startRun(); err != nil {
				fmt.Printf("Failed to start run for job %s: %s", worker.Job.Name, err)
			}
		}
	}
}
