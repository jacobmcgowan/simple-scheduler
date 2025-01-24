package mongoModels

import (
	repositoryErrors "github.com/jacobmcgowan/simple-scheduler/shared/data-access/repositories/errors"
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func JobAggregatorFromDto(dto dtos.JobLockFilter, collection string) (mongo.Pipeline, error) {
	objId, err := bson.ObjectIDFromHex(dto.ManagerId)
	if err != nil {
		return nil, &repositoryErrors.InvalidIdError{
			Value: dto.ManagerId,
		}
	}

	matchStage := bson.D{{
		Key: "$match",
		Value: bson.D{{
			Key: "$or",
			Value: bson.A{
				bson.D{{
					Key: "managerId",
					Value: bson.D{{
						Key:   "$eq",
						Value: objId,
					}},
				}},
				bson.D{{
					Key: "managerId",
					Value: bson.D{{
						Key:   "$exists",
						Value: false,
					}},
				}},
			},
		}},
	}}
	sortStage := bson.D{{
		Key: "$sort",
		Value: bson.D{{
			Key:   "managerId",
			Value: -1,
		}},
	}}
	limitStage := bson.D{{
		Key:   "$limit",
		Value: dto.Take,
	}}
	setStage := bson.D{{
		Key: "$set",
		Value: bson.D{{
			Key:   "managerId",
			Value: objId,
		}},
	}}
	mergeStage := bson.D{{
		Key: "$merge",
		Value: bson.D{{
			Key:   "into",
			Value: collection,
		}, {
			Key:   "on",
			Value: "_id",
		}, {
			Key:   "whenMatched",
			Value: "merge",
		}},
	}}

	if dto.Take > 0 {
		return mongo.Pipeline{matchStage, sortStage, limitStage, setStage, mergeStage}, nil
	}

	return mongo.Pipeline{matchStage, sortStage, setStage, mergeStage}, nil
}

func JobLockFilterFromDto(dto dtos.JobLockFilter) (bson.D, error) {
	objId, err := bson.ObjectIDFromHex(dto.ManagerId)
	if err != nil {
		return nil, &repositoryErrors.InvalidIdError{
			Value: dto.ManagerId,
		}
	}

	return AppendBsonCondition(bson.D{}, "managerId", "$eq", &objId), nil
}
