package dtos

type JobHeartbeatMessage struct {
	JobName string `json:"jobName"`
	RunId   string `json:"runId"`
}
