package mongoModels

import (
	"github.com/jacobmcgowan/simple-scheduler/dtos"
	"go.mongodb.org/mongo-driver/bson"
)

func RunFilterFromDto(dto dtos.RunFilter) bson.D {
	filter := bson.D{}

	if dto.JobName.Defined {
		filter = append(filter, bson.E{
			Key:   "jobName",
			Value: dto.JobName.Value,
		})
	}

	if dto.Status.Defined {
		filter = append(filter, bson.E{
			Key:   "status",
			Value: dto.Status.Value,
		})
	}

	return filter
}
