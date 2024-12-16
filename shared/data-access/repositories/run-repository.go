package repositories

import "github.com/jacobmcgowan/simple-scheduler/shared/dtos"

type RunRepository interface {
	Browse(filter dtos.RunFilter) ([]dtos.Run, error)
	Read(id string) (dtos.Run, error)
	Edit(id string, update dtos.RunUpdate) error
	Add(run dtos.Run) (string, error)
	Delete(id string) error
}
