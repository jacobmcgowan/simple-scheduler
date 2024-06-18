package mongoModels

import (
	"github.com/jacobmcgowan/simple-scheduler/common"
	"github.com/jacobmcgowan/simple-scheduler/dtos"
	"go.mongodb.org/mongo-driver/bson"
)

func JobUpdateFromDto(dto dtos.JobUpdate) bson.D {
	update := bson.D{}
	update = common.AppendBson(update, "enabled", dto.Enabled)
	update = common.AppendBson(update, "nextRunAt", dto.NextRunAt)
	update = common.AppendBson(update, "interval", dto.Interval)
	update = common.AppendBson(update, "runExecutionTimeout", dto.RunExecutionTimeout)
	update = common.AppendBson(update, "runStartTimeout", dto.RunStartTimeout)
	update = common.AppendBson(update, "maxQueueCount", dto.MaxQueueCount)
	update = common.AppendBson(update, "allowConcurrentRuns", dto.AllowConcurrentRuns)
	update = common.AppendBson(update, "heartbeatTimeout", dto.HeartbeatTimeout)

	return update
}
