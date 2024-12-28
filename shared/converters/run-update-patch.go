package converters

import (
	"encoding/json"
	"fmt"

	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
	"github.com/jacobmcgowan/simple-scheduler/shared/runStatuses"
)

func PatchToRunUpdate(jsonData []byte) (dtos.RunUpdate, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return dtos.RunUpdate{}, fmt.Errorf("failed to decode JSON Patch Merge document: %s", err)
	}

	return dtos.RunUpdate{
			Status: InterfaceToUndefinable[runStatuses.RunStatus](data["status"]),
		},
		nil
}
