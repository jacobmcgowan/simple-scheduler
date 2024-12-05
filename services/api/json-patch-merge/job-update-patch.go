package jsonPatchMerge

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
)

func PatchJobUpdate(jsonData []byte) (dtos.JobUpdate, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return dtos.JobUpdate{}, fmt.Errorf("failed to decode JSON Patch Merge document: %s", err)
	}

	return dtos.JobUpdate{
			Enabled:             InterfaceToUndefinable[bool](data["enabled"]),
			NextRunAt:           InterfaceToUndefinable[time.Time](data["nextRunAt"]),
			Interval:            InterfaceToUndefinable[int](data["interval"]),
			RunExecutionTimeout: InterfaceToUndefinable[int](data["runExecutionTimeout"]),
			RunStartTimeout:     InterfaceToUndefinable[int](data["RunStartTimeout"]),
			MaxQueueCount:       InterfaceToUndefinable[int](data["maxQueueCount"]),
			AllowConcurrentRuns: InterfaceToUndefinable[bool](data["heartbeatTimeout"]),
		},
		nil
}
