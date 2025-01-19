package dtos

import "time"

type JobFilter struct {
	IsUnmanaged     bool
	ManagerId       *string
	HeartbeatBefore *time.Time
	Take            *int
}
