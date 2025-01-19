package mongoModels

import (
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Manager struct {
	Id       bson.ObjectID `bson:"_id,omitempty"`
	Hostname string        `bson:"hostname"`
}

func (manager Manager) ToDto() dtos.Manager {
	return dtos.Manager{
		Id:       manager.Id.Hex(),
		Hostname: manager.Hostname,
	}
}

func (manager *Manager) FromDto(dto dtos.Manager) {
	id, err := bson.ObjectIDFromHex(dto.Id)
	if err != nil {
		id = bson.NilObjectID
	}

	manager.Id = id
	manager.Hostname = dto.Hostname
}
