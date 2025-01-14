package dtos

import (
	"time"

	"github.com/jacobmcgowan/simple-scheduler/shared/runStatuses"
)

type RunFilter struct {
	JobName         *string                `json:"jobName,omitempty"`
	Status          *runStatuses.RunStatus `json:"status,omitempty"`
	CreatedBefore   *time.Time             `json:"createdBefore,omitempty"`
	StartedBefore   *time.Time             `json:"startedBefore,omitempty"`
	HeartbeatBefore *time.Time             `json:"heartbeatBefore,omitempty"`
}
