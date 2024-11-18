package mongoModels

import (
	"time"

	"github.com/jacobmcgowan/simple-scheduler/dtos"
	"github.com/jacobmcgowan/simple-scheduler/runStatuses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Run struct {
	Id        primitive.ObjectID `bson:"_id,omitempty"`
	JobName   string             `bson:"jobName"`
	Status    string             `bson:"status"`
	StartTime time.Time          `bson:"startTime"`
	EndTime   time.Time          `bson:"endTime"`
}

func (run Run) ToDto() dtos.Run {
	return dtos.Run{
		Id:        run.Id.Hex(),
		JobName:   run.JobName,
		Status:    runStatuses.RunStatus(run.Status),
		StartTime: run.StartTime,
		EndTime:   run.EndTime,
	}
}

func (run *Run) FromDto(dto dtos.Run) {
	id, err := primitive.ObjectIDFromHex(dto.Id)
	if err != nil {
		id = primitive.NilObjectID
	}

	run.Id = id
	run.JobName = dto.JobName
	run.Status = string(dto.Status)
	run.StartTime = dto.StartTime
	run.EndTime = dto.EndTime
}
