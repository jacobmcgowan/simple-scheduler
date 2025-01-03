package mongoModels

import (
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
	"go.mongodb.org/mongo-driver/bson"
)

func RunFilterFromDto(dto dtos.RunFilter) bson.D {
	filter := bson.D{}

	if dto.JobName.Defined {
		filter = append(filter, bson.E{
			Key: "jobName",
			Value: bson.D{{
				Key:   "$eq",
				Value: dto.JobName.Value,
			}},
		})
	}

	if dto.Status.Defined {
		filter = append(filter, bson.E{
			Key: "status",
			Value: bson.D{{
				Key:   "$eq",
				Value: dto.Status.Value,
			}},
		})
	}

	return filter
}
