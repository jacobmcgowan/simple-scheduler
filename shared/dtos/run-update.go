package dtos

import (
	"time"

	"github.com/jacobmcgowan/simple-scheduler/shared/common"
	"github.com/jacobmcgowan/simple-scheduler/shared/runStatuses"
)

type RunUpdate struct {
	Status    common.Undefinable[runStatuses.RunStatus]
	StartTime common.Undefinable[time.Time]
	EndTime   common.Undefinable[time.Time]
	Heartbeat common.Undefinable[time.Time]
}
