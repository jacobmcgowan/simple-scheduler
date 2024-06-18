package mongoModels

import (
	"github.com/jacobmcgowan/simple-scheduler/common"
	"github.com/jacobmcgowan/simple-scheduler/dtos"
	"go.mongodb.org/mongo-driver/bson"
)

func RunUpdateFromDto(dto dtos.RunUpdate) bson.D {
	update := bson.D{}
	update = common.AppendBson(update, "status", dto.Status)
	update = common.AppendBson(update, "startTime", dto.StartTime)
	update = common.AppendBson(update, "endTime", dto.EndTime)

	return update
}
