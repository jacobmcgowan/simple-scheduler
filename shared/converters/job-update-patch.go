package converters

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jacobmcgowan/simple-scheduler/shared/common"
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
)

func PatchToJobUpdate(jsonData []byte) (dtos.JobUpdate, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return dtos.JobUpdate{}, fmt.Errorf("failed to decode JSON Patch Merge document: %s", err)
	}

	interval := InterfaceToUndefinable[float64](data["interval"])
	runExecutionTimeout := InterfaceToUndefinable[float64](data["runExecutionTimeout"])
	runStartTimeout := InterfaceToUndefinable[float64](data["runStartTimeout"])
	maxQueueCount := InterfaceToUndefinable[float64](data["maxQueueCount"])

	jobUpdate := dtos.JobUpdate{
		Enabled: InterfaceToUndefinable[bool](data["enabled"]),
		Interval: common.Undefinable[int]{
			Value:   int(interval.Value),
			Defined: interval.Defined,
		},
		RunExecutionTimeout: common.Undefinable[int]{
			Value:   int(runExecutionTimeout.Value),
			Defined: runExecutionTimeout.Defined,
		},
		RunStartTimeout: common.Undefinable[int]{
			Value:   int(runStartTimeout.Value),
			Defined: runStartTimeout.Defined,
		},
		MaxQueueCount: common.Undefinable[int]{
			Value:   int(maxQueueCount.Value),
			Defined: maxQueueCount.Defined,
		},
		AllowConcurrentRuns: InterfaceToUndefinable[bool](data["heartbeatTimeout"]),
	}

	nextRunAt := InterfaceToUndefinable[string](data["nextRunAt"])
	if nextRunAt.Defined {
		nextRunAtTime, err := time.Parse(time.RFC3339, nextRunAt.Value)
		if err != nil {
			return dtos.JobUpdate{}, fmt.Errorf("nextRunAt, %s, is not a valid RFC3339 datetime", err)
		}

		jobUpdate.NextRunAt = common.Undefinable[time.Time]{
			Value:   nextRunAtTime,
			Defined: true,
		}
	} else {
		jobUpdate.NextRunAt = common.Undefinable[time.Time]{
			Defined: false,
		}
	}

	return jobUpdate, nil
}

func JobUpdateToPatch(jobUpdate dtos.JobUpdate) ([]byte, error) {
	data := map[string]any{}

	if jobUpdate.Enabled.Defined {
		data["enabled"] = jobUpdate.Enabled.Value
	}

	if jobUpdate.NextRunAt.Defined {
		data["nextRunAt"] = jobUpdate.NextRunAt.Value
	}

	if jobUpdate.Interval.Defined {
		data["interval"] = jobUpdate.Interval.Value
	}

	if jobUpdate.RunExecutionTimeout.Defined {
		data["runExecutionTimeout"] = jobUpdate.RunExecutionTimeout.Value
	}

	if jobUpdate.RunStartTimeout.Defined {
		data["RunStartTimeout"] = jobUpdate.RunStartTimeout.Value
	}

	if jobUpdate.MaxQueueCount.Defined {
		data["maxQueueCount"] = jobUpdate.MaxQueueCount.Value
	}

	if jobUpdate.HeartbeatTimeout.Defined {
		data["heartbeatTimeout"] = jobUpdate.HeartbeatTimeout.Value
	}

	return json.Marshal(data)
}
