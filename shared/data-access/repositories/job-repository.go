package repositories

import "github.com/jacobmcgowan/simple-scheduler/shared/dtos"

type JobRepository interface {
	Browse() ([]dtos.Job, error)
	Read(name string) (dtos.Job, error)
	Edit(name string, update dtos.JobUpdate) error
	Add(job dtos.Job) (string, error)
	Delete(name string) error
	Lock(filter dtos.JobLockFilter) ([]dtos.Job, error)
	Unlock(filter dtos.JobUnlockFilter) (int64, error)
}
