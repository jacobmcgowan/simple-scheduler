package mongoModels

import (
	"time"

	"github.com/jacobmcgowan/simple-scheduler/dtos"
)

type Job struct {
	Name                string    `bson:"_id"`
	Enabled             bool      `bson:"enabled"`
	NextRunAt           time.Time `bson:"nextRunAt"`
	Interval            int       `bson:"interval"`
	RunExecutionTimeout int       `bson:"runExecutionTimeout"`
	RunStartTimeout     int       `bson:"runStartTimeout"`
	MaxQueueCount       int       `bson:"maxQueueCount"`
	AllowConcurrentRuns bool      `bson:"allowConcurrentRuns"`
	HeartbeatTimeout    int       `bson:"heartbeatTimeout"`
}

func (job Job) ToDto() dtos.Job {
	return dtos.Job{
		Name:                job.Name,
		Enabled:             job.Enabled,
		NextRunAt:           job.NextRunAt,
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
	job.NextRunAt = dto.NextRunAt
	job.Interval = dto.Interval
	job.RunExecutionTimeout = dto.RunExecutionTimeout
	job.RunStartTimeout = dto.RunStartTimeout
	job.MaxQueueCount = dto.MaxQueueCount
	job.AllowConcurrentRuns = dto.AllowConcurrentRuns
	job.HeartbeatTimeout = dto.HeartbeatTimeout
}
