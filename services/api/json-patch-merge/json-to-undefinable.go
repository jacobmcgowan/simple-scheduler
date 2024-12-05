package jsonPatchMerge

import (
	"time"

	"github.com/jacobmcgowan/simple-scheduler/shared/common"
)

func InterfaceToUndefinable[T bool | time.Time | int](i interface{}) common.Undefinable[T] {
	value, ok := i.(T)
	return common.Undefinable[T]{
		Value:   value,
		Defined: ok,
	}
}
