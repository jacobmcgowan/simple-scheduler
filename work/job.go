package work

import (
	"encoding/json"
	"fmt"
	"time"
)

type Job struct {
	Name       string
	NextRunAt  time.Time
	Interval   int
	MessageBus MessageBus
	running    chan bool
}

func (job Job) Start() error {
	fullName := "scheduler.job." + job.Name
	err := job.MessageBus.Register(
		fullName,
		map[string][]string{
			fullName + ".action": {"action"},
			fullName + ".status": {"status"},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to register job %s to message bus: %s", job.Name, err)
	}

	err = job.MessageBus.Subscribe(
		"scheduler.job."+job.Name,
		"status",
		"scheduler.job.status",
		job.messageReceived,
	)
	if err != nil {
		return err
	}

	job.running = make(chan bool)
	go job.run()

	return nil
}

func (job Job) messageReceived(body []byte) bool {
	fmt.Printf("Job status message received: %s", body)
	return true
}

func (job Job) run() {
	for {
		if <-job.running {
			return
		}

		if job.NextRunAt.Compare(time.Now()) >= 0 {
			body, err := json.Marshal(JobActionMessage{
				JobName: job.Name,
				RunId:   "",
				Action:  "run",
			})
			if err != nil {
				fmt.Printf("Failed to serialize run action: %s", err)
			}

			job.MessageBus.Publish(
				"scheduler.job."+job.Name,
				"action",
				body,
			)
		}
	}
}
