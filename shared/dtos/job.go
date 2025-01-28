package dtos

import (
	"encoding/json"
	"time"
)

type Job struct {
	Name                string    `json:"name" binding:"required"`
	Enabled             bool      `json:"enabled" binding:"required"`
	NextRunAt           time.Time `json:"nextRunAt" binding:"required"`
	Interval            int       `json:"interval"`
	RunExecutionTimeout int       `json:"runExecutionTimeout"`
	RunStartTimeout     int       `json:"runStartTimeout"`
	MaxQueueCount       int       `json:"maxQueueCount"`
	AllowConcurrentRuns bool      `json:"allowConcurrentRuns"`
	HeartbeatTimeout    int       `json:"heartbeatTimeout"`
	ManagerId           string    `json:"managerId,omitempty"`
	Heartbeat           time.Time `json:"heartbeat"`
}

func (job *Job) UnmarshalJSON(data []byte) error {
	var tmp struct {
		Name                string    `json:"name" binding:"required"`
		Enabled             bool      `json:"enabled" binding:"required"`
		NextRunAt           time.Time `json:"nextRunAt" binding:"required"`
		Interval            int       `json:"interval"`
		RunExecutionTimeout int       `json:"runExecutionTimeout"`
		RunStartTimeout     int       `json:"runStartTimeout"`
		MaxQueueCount       int       `json:"maxQueueCount"`
		AllowConcurrentRuns bool      `json:"allowConcurrentRuns"`
		HeartbeatTimeout    int       `json:"heartbeatTimeout"`
	}

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	job.Name = tmp.Name
	job.Enabled = tmp.Enabled
	job.NextRunAt = tmp.NextRunAt
	job.Interval = tmp.Interval
	job.RunExecutionTimeout = tmp.RunExecutionTimeout
	job.RunStartTimeout = tmp.RunStartTimeout
	job.MaxQueueCount = tmp.MaxQueueCount
	job.AllowConcurrentRuns = tmp.AllowConcurrentRuns
	job.HeartbeatTimeout = tmp.HeartbeatTimeout

	return nil
}
