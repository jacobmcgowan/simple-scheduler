package validators

import "github.com/jacobmcgowan/simple-scheduler/shared/runStatuses"

func ValidateRunStatus(val string, allowNone bool) bool {
	switch val {
	case string(runStatuses.Cancelled),
		string(runStatuses.Cancelling),
		string(runStatuses.Completed),
		string(runStatuses.Failed),
		string(runStatuses.Pending),
		string(runStatuses.Running):
		return true
	case "":
		return allowNone
	default:
		return false
	}
}
