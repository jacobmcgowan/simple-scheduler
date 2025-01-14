package dtos

import (
	"time"
)

type JobUpdate struct {
	Enabled             *bool      `json:"enabled,omitempty"`
	NextRunAt           *time.Time `json:"nextRunAt,omitempty"`
	Interval            *int       `json:"interval,omitempty"`
	RunExecutionTimeout *int       `json:"runExecutionTimeout,omitempty"`
	RunStartTimeout     *int       `json:"runStartTimeout,omitempty"`
	MaxQueueCount       *int       `json:"maxQueueCount,omitempty"`
	AllowConcurrentRuns *bool      `json:"allowConcurrentRuns,omitempty"`
	HeartbeatTimeout    *int       `json:"heartbeatTimeout,omitempty"`
}
