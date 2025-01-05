package dtos

type JobActionMessage struct {
	JobName string `json:"jobName"`
	RunId   string `json:"runId"`
	Action  string `json:"action"`
}
