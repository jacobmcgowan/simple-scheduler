package mongoRepos

import (
	"fmt"
	"time"

	mongoModels "github.com/jacobmcgowan/simple-scheduler/shared/data-access/models/mongo"
	repositoryErrors "github.com/jacobmcgowan/simple-scheduler/shared/data-access/repositories/errors"
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const JobsCollection = "jobs"

type MongoJobRepository struct {
	DbContext *MongoDbContext
}

func (repo MongoJobRepository) Browse() ([]dtos.Job, error) {
	var jobs []dtos.Job
	coll := repo.DbContext.db.Collection(JobsCollection)
	cur, err := coll.Find(repo.DbContext.ctx, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("failed to find jobs: %s", err)
	}

	for cur.Next(repo.DbContext.ctx) {
		var job mongoModels.Job
		err = cur.Decode(&job)
		if err != nil {
			return nil, fmt.Errorf("failed to parse job: %s", err)
		}

		jobs = append(jobs, job.ToDto())
	}

	err = cur.Close(repo.DbContext.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to close cursor: %s", err)
	}

	return jobs, nil
}

func (repo MongoJobRepository) Read(name string) (dtos.Job, error) {
	var job mongoModels.Job
	filter := bson.D{{
		Key: "_id",
		Value: bson.D{{
			Key:   "$eq",
			Value: name,
		}},
	}}
	coll := repo.DbContext.db.Collection(JobsCollection)
	err := coll.FindOne(repo.DbContext.ctx, filter).Decode(&job)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return dtos.Job{}, &repositoryErrors.NotFoundError{
				Message: fmt.Sprintf("failed to find job %s: %s", name, err),
			}
		}

		return dtos.Job{}, fmt.Errorf("failed to find job %s: %s", name, err)
	}

	return job.ToDto(), nil
}

func (repo MongoJobRepository) Edit(name string, update dtos.JobUpdate) error {
	updateDoc := mongoModels.JobUpdateFromDto(update)
	filter := bson.D{{
		Key: "_id",
		Value: bson.D{{
			Key:   "$eq",
			Value: name,
		}},
	}}
	coll := repo.DbContext.db.Collection(JobsCollection)
	_, err := coll.UpdateOne(repo.DbContext.ctx, filter, updateDoc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &repositoryErrors.NotFoundError{
				Message: fmt.Sprintf("failed to find job %s: %s", name, err),
			}
		}

		return fmt.Errorf("failed to edit job %s: %s", name, err)
	}

	return nil
}

func (repo MongoJobRepository) Add(job dtos.Job) (string, error) {
	jobDoc := mongoModels.Job{}
	jobDoc.FromDto(job)

	coll := repo.DbContext.db.Collection(JobsCollection)
	res, err := coll.InsertOne(repo.DbContext.ctx, jobDoc)
	if err != nil {
		return "", fmt.Errorf("failed to add job: %s", err)
	}

	if name, ok := res.InsertedID.(string); ok {
		return name, nil
	}

	return "", fmt.Errorf("failed to parse id of job: %s", err)
}

func (repo MongoJobRepository) Delete(name string) error {
	filter := bson.D{{
		Key: "_id",
		Value: bson.D{{
			Key:   "$eq",
			Value: name,
		}},
	}}
	coll := repo.DbContext.db.Collection(JobsCollection)
	_, err := coll.DeleteOne(repo.DbContext.ctx, filter)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &repositoryErrors.NotFoundError{
				Message: fmt.Sprintf("failed to find job %s: %s", name, err),
			}
		}

		return fmt.Errorf("failed to delete job %s: %s", name, err)
	}

	return nil
}

func (repo MongoJobRepository) Lock(filter dtos.JobLockFilter) ([]dtos.Job, error) {
	var jobs []dtos.Job

	aggr, err := mongoModels.JobAggregatorFromDto(filter, JobsCollection)
	if err != nil {
		return nil, fmt.Errorf("invalid filter: %s", err)
	}

	coll := repo.DbContext.db.Collection(JobsCollection)
	_, err = coll.Aggregate(repo.DbContext.ctx, aggr)
	if err != nil {
		return nil, fmt.Errorf("failed to locks jobs: %s", err)
	}

	lockFilter, err := mongoModels.JobLockFilterFromDto(filter)
	if err != nil {
		return nil, fmt.Errorf("invalid filter: %s", err)
	}

	cur, err := coll.Find(repo.DbContext.ctx, lockFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to find locked jobs: %s", err)
	}

	for cur.Next(repo.DbContext.ctx) {
		var job mongoModels.Job
		err = cur.Decode(&job)
		if err != nil {
			return nil, fmt.Errorf("failed to parse job: %s", err)
		}

		jobs = append(jobs, job.ToDto())
	}

	err = cur.Close(repo.DbContext.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to close cursor: %s", err)
	}

	return jobs, nil
}

func (repo MongoJobRepository) Unlock(filter dtos.JobUnlockFilter) (int64, error) {
	filterDoc, err := mongoModels.JobUnlockFilterFromDto(filter)
	if err != nil {
		return 0, fmt.Errorf("invalid filter: %s", err)
	}

	updateDoc := mongoModels.JobUnlock()
	coll := repo.DbContext.db.Collection(JobsCollection)
	cur, err := coll.UpdateMany(repo.DbContext.ctx, filterDoc, updateDoc)
	if err != nil {
		return 0, fmt.Errorf("failed to unlock jobs: %s", err)
	}

	return cur.ModifiedCount, nil
}

func (repo MongoJobRepository) Heartbeat(mngrId string) error {
	filterDoc := mongoModels.AppendBsonCondition(bson.D{}, "managerId", "$eq", &mngrId)
	updateDoc := bson.D{{
		Key: "$set",
		Value: bson.D{{
			Key:   "heartbeat",
			Value: time.Now(),
		}},
	}}
	coll := repo.DbContext.db.Collection(JobsCollection)
	_, err := coll.UpdateMany(repo.DbContext.ctx, filterDoc, updateDoc)
	if err != nil {
		return fmt.Errorf("failed to set heartbeat: %s", err)
	}

	return nil
}
