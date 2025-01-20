package mongoModels

import (
	repositoryErrors "github.com/jacobmcgowan/simple-scheduler/shared/data-access/repositories/errors"
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func JobUnlockFilterFromDto(dto dtos.JobUnlockFilter) (bson.D, error) {
	objId, err := bson.ObjectIDFromHex(dto.ManagerId)
	if err != nil {
		return nil, &repositoryErrors.InvalidIdError{
			Value: dto.ManagerId,
		}
	}

	filter := bson.D{{
		Key: "managerId",
		Value: bson.D{{
			Key:   "$eq",
			Value: objId,
		}},
	}}

	if dto.JobNames != nil && len(dto.JobNames) > 0 {
		filter = append(filter, bson.E{
			Key: "_id",
			Value: bson.D{{
				Key:   "$in",
				Value: dto.JobNames,
			}},
		})
	}

	return filter, nil
}
