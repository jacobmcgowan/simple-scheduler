package mongoModels

import (
	"github.com/jacobmcgowan/simple-scheduler/common"
	"github.com/jacobmcgowan/simple-scheduler/dtos"
	"go.mongodb.org/mongo-driver/bson"
)

func JobUpdateFromDto(dto dtos.JobUpdate) bson.D {
	setDoc := bson.D{}
	setDoc = common.AppendBson(setDoc, "enabled", dto.Enabled)
	setDoc = common.AppendBson(setDoc, "nextRunAt", dto.NextRunAt)
	setDoc = common.AppendBson(setDoc, "interval", dto.Interval)
	setDoc = common.AppendBson(setDoc, "runExecutionTimeout", dto.RunExecutionTimeout)
	setDoc = common.AppendBson(setDoc, "runStartTimeout", dto.RunStartTimeout)
	setDoc = common.AppendBson(setDoc, "maxQueueCount", dto.MaxQueueCount)
	setDoc = common.AppendBson(setDoc, "allowConcurrentRuns", dto.AllowConcurrentRuns)
	setDoc = common.AppendBson(setDoc, "heartbeatTimeout", dto.HeartbeatTimeout)

	return bson.D{{
		Key:   "$set",
		Value: setDoc,
	}}
}
