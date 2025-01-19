package mongoModels

import (
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func JobUpdateFromDto(dto dtos.JobUpdate) bson.D {
	setDoc := bson.D{}
	setDoc = AppendBson(setDoc, "enabled", dto.Enabled)
	setDoc = AppendBson(setDoc, "nextRunAt", dto.NextRunAt)
	setDoc = AppendBson(setDoc, "interval", dto.Interval)
	setDoc = AppendBson(setDoc, "runExecutionTimeout", dto.RunExecutionTimeout)
	setDoc = AppendBson(setDoc, "runStartTimeout", dto.RunStartTimeout)
	setDoc = AppendBson(setDoc, "maxQueueCount", dto.MaxQueueCount)
	setDoc = AppendBson(setDoc, "allowConcurrentRuns", dto.AllowConcurrentRuns)
	setDoc = AppendBson(setDoc, "heartbeatTimeout", dto.HeartbeatTimeout)

	return bson.D{{
		Key:   "$set",
		Value: setDoc,
	}}
}
