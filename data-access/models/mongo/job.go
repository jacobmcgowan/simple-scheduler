package mongoModels

import (
	"time"

	"github.com/jacobmcgowan/simple-scheduler/dtos"
)

type Job struct {
	Name                string `bson:"_id"`
	Enabled             bool
	NextRunAt           string
	Interval            int
	RunExecutionTimeout int
	RunStartTimeout     int
	MaxQueueCount       int
	AllowConcurrentRuns int
	HeartbeatTimeout    int
}

func (job Job) ToDto() dtos.Job {
	nextRunAt, err := time.Parse(time.RFC3339, job.NextRunAt)
	if err != nil {
		nextRunAt = time.Time{}
	}

	return dtos.Job{
		Name:                job.Name,
		Enabled:             job.Enabled,
		NextRunAt:           nextRunAt,
		Interval:            job.Interval,
		RunExecutionTimeout: job.RunExecutionTimeout,
		RunStartTimeout:     job.RunStartTimeout,
		MaxQueueCount:       job.MaxQueueCount,
		AllowConcurrentRuns: job.AllowConcurrentRuns,
		HeartbeatTimeout:    job.HeartbeatTimeout,
	}
}

func (job *Job) FromDto(dto dtos.Job) {
	job.Name = dto.Name
	job.Enabled = dto.Enabled
	job.NextRunAt = dto.NextRunAt.String()
	job.Interval = dto.Interval
	job.RunExecutionTimeout = dto.RunExecutionTimeout
	job.RunStartTimeout = dto.RunStartTimeout
	job.MaxQueueCount = dto.MaxQueueCount
	job.AllowConcurrentRuns = dto.AllowConcurrentRuns
	job.HeartbeatTimeout = dto.HeartbeatTimeout
}
