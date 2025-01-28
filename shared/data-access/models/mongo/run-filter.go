package mongoModels

import (
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func RunFilterFromDto(dto dtos.RunFilter) bson.D {
	filter := bson.D{}
	filter = AppendBsonCondition(filter, "jobName", "$eq", dto.JobName)
	filter = AppendBsonCondition(filter, "status", "$eq", dto.Status)
	filter = AppendBsonCondition(filter, "createdTime", "$lt", dto.CreatedBefore)
	filter = AppendBsonCondition(filter, "startTime", "$lt", dto.StartedBefore)
	filter = AppendBsonCondition(filter, "heartbeat", "$lt", dto.HeartbeatBefore)

	return filter
}
