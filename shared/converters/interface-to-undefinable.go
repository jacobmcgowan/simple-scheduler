package converters

import (
	"time"

	"github.com/jacobmcgowan/simple-scheduler/shared/common"
	"github.com/jacobmcgowan/simple-scheduler/shared/runStatuses"
)

func InterfaceToUndefinable[T bool | int | float64 | string | time.Time | runStatuses.RunStatus](i interface{}) common.Undefinable[T] {
	value, ok := i.(T)
	return common.Undefinable[T]{
		Value:   value,
		Defined: ok,
	}
}
