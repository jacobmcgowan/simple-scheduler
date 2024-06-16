package dtos

import (
	"github.com/jacobmcgowan/simple-scheduler/common"
)

type RunFilter struct {
	JobName common.Undefinable[string]
	Status  common.Undefinable[string]
}
