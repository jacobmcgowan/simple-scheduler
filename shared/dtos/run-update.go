package dtos

import (
	"time"

	"github.com/jacobmcgowan/simple-scheduler/shared/runStatuses"
)

type RunUpdate struct {
	Status    *runStatuses.RunStatus `json:"status,omitempty"`
	StartTime *time.Time             `json:"startTime,omitempty"`
	EndTime   *time.Time             `json:"endTime,omitempty"`
	Heartbeat *time.Time             `json:"heartbeat,omitempty"`
}
