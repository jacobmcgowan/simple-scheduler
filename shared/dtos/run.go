package dtos

import (
	"time"

	"github.com/jacobmcgowan/simple-scheduler/shared/runStatuses"
)

type Run struct {
	Id          string                `json:"id"`
	JobName     string                `json:"jobName"`
	Status      runStatuses.RunStatus `json:"status"`
	CreatedTime time.Time             `json:"createdTime"`
	StartTime   time.Time             `json:"startTime"`
	EndTime     time.Time             `json:"endTime"`
	Heartbeat   time.Time             `json:"heartbeat"`
}
