package options

type JobOptions struct {
	Name                string
	Enabled             bool
	NextRunAt           string
	Interval            int
	RunExecutionTimeout int
	RunStartTimeout     int
	MaxQueueCount       int
	AllowConcurrentRuns bool
	HeartbeatTimeout    int
}
