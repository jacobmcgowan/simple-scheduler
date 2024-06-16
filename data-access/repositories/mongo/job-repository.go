package mongoRepos

import (
	"fmt"
	"time"

	mongoModels "github.com/jacobmcgowan/simple-scheduler/data-access/models/mongo"
	"github.com/jacobmcgowan/simple-scheduler/dtos"
	"go.mongodb.org/mongo-driver/bson"
)

const JobsCollection = "jobs"

type JobRepository struct {
	DbContext DbContext
}

func (repo JobRepository) Browse() ([]dtos.Job, error) {
	var jobs []dtos.Job
	coll := repo.DbContext.Db.Collection(JobsCollection)
	cur, err := coll.Find(repo.DbContext.Context, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("failed to find jobs: %s", err)
	}

	for cur.Next(repo.DbContext.Context) {
		var job mongoModels.Job
		err = cur.Decode(&job)
		if err != nil {
			return nil, fmt.Errorf("failed to parse job: %s", err)
		}

		jobs = append(jobs, job.ToDto())
	}

	err = cur.Close(repo.DbContext.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to close cursor: %s", err)
	}

	return jobs, nil
}

func (repo JobRepository) Read(name string) (dtos.Job, error) {
	var job mongoModels.Job
	filter := bson.D{{
		Key:   "name",
		Value: name,
	}}
	coll := repo.DbContext.Db.Collection(JobsCollection)
	err := coll.FindOne(repo.DbContext.Context, filter).Decode(&job)
	if err != nil {
		return dtos.Job{}, fmt.Errorf("failed to find job %s: %s", name, err)
	}

	return job.ToDto(), nil
}

func (repo JobRepository) Edit(job dtos.Job) error {
	jobDoc := mongoModels.Job{}
	jobDoc.FromDto(job)

	filter := bson.D{{
		Key:   "name",
		Value: job.Name,
	}}
	coll := repo.DbContext.Db.Collection(JobsCollection)
	_, err := coll.ReplaceOne(repo.DbContext.Context, filter, jobDoc)
	if err != nil {
		return fmt.Errorf("failed to edit job %s: %s", job.Name, err)
	}

	return nil
}

func (repo JobRepository) SetNextRunTime(job dtos.Job) error {
	elapsed := time.Since(job.NextRunAt)

	if job.Interval <= 0 || elapsed.Milliseconds() <= 0 {
		return nil
	}

	intervals := (elapsed.Milliseconds() / int64(job.Interval)) + 1
	nextRunAt := job.NextRunAt.Add(time.Duration(job.Interval * int(intervals)))
	filter := bson.D{{
		Key:   "name",
		Value: job.Name,
	}}
	update := bson.D{{
		Key: "$set",
		Value: bson.D{{
			Key:   "NextRunAt",
			Value: nextRunAt.String(),
		}},
	}}
	coll := repo.DbContext.Db.Collection(JobsCollection)
	_, err := coll.UpdateOne(repo.DbContext.Context, filter, update)
	if err != nil {
		return fmt.Errorf("failed to edit job %s: %s", job.Name, err)
	}

	return nil
}
