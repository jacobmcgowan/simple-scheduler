package dtos

type JobStatusMessage struct {
	JobName string `json:"jobName"`
	RunId   string `json:"runId"`
	Status  string `json:"status"`
}
