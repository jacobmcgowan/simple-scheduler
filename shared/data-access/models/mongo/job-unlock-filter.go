package mongoModels

import (
	repositoryErrors "github.com/jacobmcgowan/simple-scheduler/shared/data-access/repositories/errors"
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func JobUnlockFilterFromDto(dto dtos.JobUnlockFilter) (bson.D, error) {
	filter := AppendBsonCondition(bson.D{}, "heartbeat", "$lt", dto.HeartbeatBefore)

	if dto.IsManaged {
		filter = AppendBsonCondition(filter, "managerId", "$exists", &dto.IsManaged)
	}

	if dto.ManagerId != nil {
		objId, err := bson.ObjectIDFromHex(*dto.ManagerId)
		if err != nil {
			return nil, &repositoryErrors.InvalidIdError{
				Value: *dto.ManagerId,
			}
		}

		filter = AppendBsonCondition(filter, "managerId", "$eq", &objId)
	}

	if dto.JobNames != nil && len(dto.JobNames) > 0 {
		filter = append(filter, bson.E{
			Key: "_id",
			Value: bson.M{
				"$in": dto.JobNames,
			},
		})
	}

	return filter, nil
}
