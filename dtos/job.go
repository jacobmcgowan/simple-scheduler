package dtos

import "time"

type Job struct {
	Name                string
	Enabled             bool
	NextRunAt           time.Time
	Interval            int
	RunExecutionTimeout int
	RunStartTimeout     int
	MaxQueueCount       int
	AllowConcurrentRuns bool
	HeartbeatTimeout    int
}
