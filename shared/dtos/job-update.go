package dtos

import (
	"time"

	"github.com/jacobmcgowan/simple-scheduler/shared/common"
)

type JobUpdate struct {
	Enabled             common.Undefinable[bool]
	NextRunAt           common.Undefinable[time.Time]
	Interval            common.Undefinable[int]
	RunExecutionTimeout common.Undefinable[int]
	RunStartTimeout     common.Undefinable[int]
	MaxQueueCount       common.Undefinable[int]
	AllowConcurrentRuns common.Undefinable[bool]
	HeartbeatTimeout    common.Undefinable[int]
}
