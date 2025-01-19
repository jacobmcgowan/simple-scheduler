package mongoModels

import (
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func RunUpdateFromDto(dto dtos.RunUpdate) bson.D {
	setDoc := bson.D{}
	setDoc = AppendBson(setDoc, "status", dto.Status)
	setDoc = AppendBson(setDoc, "startTime", dto.StartTime)
	setDoc = AppendBson(setDoc, "endTime", dto.EndTime)

	return bson.D{{
		Key:   "$set",
		Value: setDoc,
	}}
}
