package dtos

import (
	"github.com/jacobmcgowan/simple-scheduler/shared/common"
	"github.com/jacobmcgowan/simple-scheduler/shared/runStatuses"
)

type RunFilter struct {
	JobName common.Undefinable[string]
	Status  common.Undefinable[runStatuses.RunStatus]
}
