package mongoModels

import (
	"github.com/jacobmcgowan/simple-scheduler/shared/common"
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
	"go.mongodb.org/mongo-driver/bson"
)

func RunUpdateFromDto(dto dtos.RunUpdate) bson.D {
	setDoc := bson.D{}
	setDoc = common.AppendBson(setDoc, "status", dto.Status)
	setDoc = common.AppendBson(setDoc, "startTime", dto.StartTime)
	setDoc = common.AppendBson(setDoc, "endTime", dto.EndTime)

	return bson.D{{
		Key:   "$set",
		Value: setDoc,
	}}
}
