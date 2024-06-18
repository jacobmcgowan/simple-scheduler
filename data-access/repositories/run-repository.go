package repositories

import "github.com/jacobmcgowan/simple-scheduler/dtos"

type RunRepository interface {
	Browse() ([]dtos.Run, error)
	Read(name string) (dtos.Run, error)
	Edit(name string, update dtos.RunUpdate) error
	Add(run dtos.Run) (string, error)
	Delete(name string) error
}
