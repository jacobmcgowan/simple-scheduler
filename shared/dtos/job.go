package dtos

import "time"

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
}
