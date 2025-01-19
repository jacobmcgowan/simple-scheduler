package repositories

import "github.com/jacobmcgowan/simple-scheduler/shared/dtos"

type ManagerRepository interface {
	Browse() ([]dtos.Manager, error)
	Read(id string) (dtos.Manager, error)
	Add(mngr dtos.Manager) (string, error)
	Delete(id string) error
}
