package dtos

import (
	"time"

	"github.com/jacobmcgowan/simple-scheduler/enums"
)

type Run struct {
	Id        string
	JobName   string
	Status    enums.RunStatus
	StartTime time.Time
	EndTime   time.Time
}
