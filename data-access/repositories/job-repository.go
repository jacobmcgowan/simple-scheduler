package repositories

import "github.com/jacobmcgowan/simple-scheduler/dtos"

type JobRepository interface {
	Browse() ([]dtos.Job, error)
	Read(name string) (dtos.Job, error)
	Edit(name string, update dtos.JobUpdate) error
	Add(job dtos.Job) (string, error)
	Delete(name string) error
}
