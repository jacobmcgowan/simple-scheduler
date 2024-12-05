package dtos

import (
	"time"

	"github.com/jacobmcgowan/simple-scheduler/shared/runStatuses"
)

type Run struct {
	Id        string
	JobName   string
	Status    runStatuses.RunStatus
	StartTime time.Time
	EndTime   time.Time
}
