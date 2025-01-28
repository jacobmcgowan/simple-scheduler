package dtos

import "time"

type JobUnlockFilter struct {
	IsManaged       bool
	HeartbeatBefore *time.Time
	ManagerId       *string
	JobNames        []string
}
