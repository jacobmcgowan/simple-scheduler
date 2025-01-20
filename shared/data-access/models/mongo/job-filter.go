package mongoModels

import (
	repositoryErrors "github.com/jacobmcgowan/simple-scheduler/shared/data-access/repositories/errors"
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func JobFilterFromDto(dto dtos.JobFilter) (bson.D, *options.FindOptionsBuilder, error) {
	filter := bson.D{}
	filter = AppendBsonCondition(filter, "heartbeat", "$lt", dto.HeartbeatBefore)
	mngrIdFilter, err := managerIdFilterFromDto(dto)
	if err != nil {
		return nil, nil, err
	}
	unmngFilter := unmanagedFilterFromDto(dto)

	opts := options.Find()
	if dto.Take > 0 {
		opts = opts.SetLimit(int64(dto.Take))
	}

	if mngrIdFilter != nil && unmngFilter != nil {
		filter = append(filter, bson.E{
			Key: "$or",
			Value: bson.A{
				bson.D{*mngrIdFilter},
				bson.D{*unmngFilter},
			},
		})
		opts = opts.SetSort(bson.D{{
			Key:   "managerId",
			Value: -1,
		}})
	} else if mngrIdFilter != nil {
		filter = append(filter, *mngrIdFilter)
	} else if unmngFilter != nil {
		filter = append(filter, *unmngFilter)
	}

	return filter, opts, nil
}

func managerIdFilterFromDto(dto dtos.JobFilter) (*bson.E, error) {
	if dto.ManagerId == nil {
		return nil, nil
	}

	objId, err := bson.ObjectIDFromHex(*dto.ManagerId)
	if err != nil {
		return nil, &repositoryErrors.InvalidIdError{
			Value: *dto.ManagerId,
		}
	}

	filter := bson.E{
		Key: "managerId",
		Value: bson.D{{
			Key:   "$eq",
			Value: objId,
		}},
	}

	return &filter, nil
}

func unmanagedFilterFromDto(dto dtos.JobFilter) *bson.E {
	if !dto.IsUnmanaged {
		return nil
	}

	filter := bson.E{
		Key: "managerId",
		Value: bson.D{{
			Key:   "$eq",
			Value: bson.NilObjectID,
		}},
	}

	return &filter
}
