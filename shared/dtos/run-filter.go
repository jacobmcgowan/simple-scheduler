package dtos

import (
	"time"

	"github.com/jacobmcgowan/simple-scheduler/shared/common"
	"github.com/jacobmcgowan/simple-scheduler/shared/runStatuses"
)

type RunFilter struct {
	JobName         common.Undefinable[string]
	Status          common.Undefinable[runStatuses.RunStatus]
	CreatedBefore   common.Undefinable[time.Time]
	StartedBefore   common.Undefinable[time.Time]
	HeartbeatBefore common.Undefinable[time.Time]
}
