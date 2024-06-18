package runStatuses

type RunStatus string

const (
	Pending    RunStatus = "pending"
	Running    RunStatus = "running"
	Cancelling RunStatus = "cancelling"
	Cancelled  RunStatus = "cancelled"
	Failed     RunStatus = "failed"
	Completed  RunStatus = "completed"
)
