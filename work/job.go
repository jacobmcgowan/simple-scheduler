package work

import (
	"time"
)

type Job struct {
	Name       string
	NextRunAt  time.Time
	Interval   int
	MessageBus MessageBus
}

func (job Job) Start() (<-chan bool, error) {
	running := make(chan bool)

	go func() {
		if <-running {
			return
		}

		if job.NextRunAt.Compare(time.Now()) >= 0 {
			job.MessageBus.Publish(
				"scheduler.job."+job.Name,
				"run",
				"{}",
			)
		}
	}()

	return running, nil
}
