package mongoModels

import (
	"github.com/jacobmcgowan/simple-scheduler/dtos"
	"go.mongodb.org/mongo-driver/bson"
)

func RunUpdateFromDto(dto dtos.RunUpdate) bson.D {
	update := bson.D{}

	if dto.Status.Defined {
		update = append(update, bson.E{
			Key:   "status",
			Value: string(dto.Status.Value),
		})
	}

	if dto.StartTime.Defined {
		update = append(update, bson.E{
			Key:   "startTime",
			Value: dto.StartTime.Value.String(),
		})
	}

	if dto.EndTime.Defined {
		update = append(update, bson.E{
			Key:   "endTime",
			Value: dto.EndTime.Value.String(),
		})
	}

	return update
}
