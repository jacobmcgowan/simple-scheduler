package mongoModels

import (
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func JobLockFilterFromDto(dto dtos.JobLockFilter) (bson.D, *options.FindOptionsBuilder, error) {
	jobFilter := dtos.JobFilter{
		IsUnmanaged: true,
		ManagerId:   &dto.ManagerId,
		Take:        dto.Take,
	}

	return JobFilterFromDto(jobFilter)
}
