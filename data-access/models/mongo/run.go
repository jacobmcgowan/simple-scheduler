package mongoModels

import (
	"time"

	"github.com/jacobmcgowan/simple-scheduler/dtos"
	"github.com/jacobmcgowan/simple-scheduler/runStatuses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Run struct {
	Id        primitive.ObjectID `bson:"_id"`
	JobName   string             `bson:"jobName"`
	Status    string             `bson:"status"`
	StartTime string             `bson:"startTime"`
	EndTime   string             `bson:"endTime"`
}

func (run Run) ToDto() dtos.Run {
	startTime, err := time.Parse(time.RFC3339, run.StartTime)
	if err != nil {
		startTime = time.Time{}
	}

	endTime, err := time.Parse(time.RFC3339, run.EndTime)
	if err != nil {
		endTime = time.Time{}
	}

	return dtos.Run{
		Id:        run.Id.Hex(),
		JobName:   run.JobName,
		Status:    runStatuses.RunStatus(run.Status),
		StartTime: startTime,
		EndTime:   endTime,
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
	run.StartTime = dto.StartTime.String()
	run.EndTime = dto.EndTime.String()
}
