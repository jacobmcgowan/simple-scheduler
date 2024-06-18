package dtos

import (
	"time"

	"github.com/jacobmcgowan/simple-scheduler/common"
	"github.com/jacobmcgowan/simple-scheduler/runStatuses"
)

type RunUpdate struct {
	Status    common.Undefinable[runStatuses.RunStatus]
	StartTime common.Undefinable[time.Time]
	EndTime   common.Undefinable[time.Time]
}
