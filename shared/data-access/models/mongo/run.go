package mongoModels

import (
	"time"

	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
	"github.com/jacobmcgowan/simple-scheduler/shared/runStatuses"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Run struct {
	Id          bson.ObjectID `bson:"_id,omitempty"`
	JobName     string        `bson:"jobName"`
	Status      string        `bson:"status"`
	CreatedTime time.Time     `bson:"createdTime"`
	StartTime   time.Time     `bson:"startTime"`
	EndTime     time.Time     `bson:"endTime"`
	Heartbeat   time.Time     `bson:"heartbeat"`
}

func (run Run) ToDto() dtos.Run {
	return dtos.Run{
		Id:          run.Id.Hex(),
		JobName:     run.JobName,
		Status:      runStatuses.RunStatus(run.Status),
		CreatedTime: run.CreatedTime,
		StartTime:   run.StartTime,
		EndTime:     run.EndTime,
		Heartbeat:   run.Heartbeat,
	}
}

func (run *Run) FromDto(dto dtos.Run) {
	id, err := bson.ObjectIDFromHex(dto.Id)
	if err != nil {
		id = bson.NilObjectID
	}

	run.Id = id
	run.JobName = dto.JobName
	run.Status = string(dto.Status)
	run.CreatedTime = dto.CreatedTime
	run.StartTime = dto.StartTime
	run.EndTime = dto.EndTime
	run.Heartbeat = dto.Heartbeat
}
