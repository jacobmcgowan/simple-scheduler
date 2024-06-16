package dtos

import (
	"time"

	"github.com/jacobmcgowan/simple-scheduler/common"
	"github.com/jacobmcgowan/simple-scheduler/enums"
)

type RunUpdate struct {
	Status    common.Undefinable[enums.RunStatus]
	StartTime common.Undefinable[time.Time]
	EndTime   common.Undefinable[time.Time]
}
