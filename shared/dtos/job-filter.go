package dtos

import "time"

type JobFilter struct {
	IsUnmanaged     bool
	Take            int
	ManagerId       *string
	HeartbeatBefore *time.Time
}
