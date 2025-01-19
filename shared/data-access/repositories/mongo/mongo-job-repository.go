package mongoRepos

import (
	"fmt"
	"time"

	mongoModels "github.com/jacobmcgowan/simple-scheduler/shared/data-access/models/mongo"
	repositoryErrors "github.com/jacobmcgowan/simple-scheduler/shared/data-access/repositories/errors"
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readconcern"
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

	filterDoc, findOpts, err := mongoModels.JobLockFilterFromDto(filter)
	if err != nil {
		return nil, fmt.Errorf("invalid filter: %s", err)
	}

	now := time.Now()
	lockUpdate, err := mongoModels.JobLock(filter.ManagerId, now)
	if err != nil {
		return nil, fmt.Errorf("failed to build lock document: %s", err)
	}

	tranOpts := options.Transaction().SetReadConcern(readconcern.Majority())
	sessOpts := options.Session().SetDefaultTransactionOptions(tranOpts)
	session, err := repo.DbContext.client.StartSession(sessOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %s", err)
	}
	defer session.EndSession(repo.DbContext.ctx)

	coll := repo.DbContext.db.Collection(JobsCollection)
	cur, err := coll.Find(repo.DbContext.ctx, filterDoc, findOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to find jobs: %s", err)
	}

	jobIds := []string{}
	for cur.Next(repo.DbContext.ctx) {
		var job mongoModels.Job
		err = cur.Decode(&job)
		if err != nil {
			return nil, fmt.Errorf("failed to parse job: %s", err)
		}

		jobIds = append(jobIds, job.Name)
		jobs = append(jobs, job.ToDto())
	}

	updateFilter := bson.D{{
		Key: "_id",
		Value: bson.D{{
			Key:   "$in",
			Value: jobIds,
		}},
	}}
	_, err = coll.UpdateMany(repo.DbContext.ctx, updateFilter, lockUpdate)
	if err != nil {
		return nil, fmt.Errorf("failed to lock jobs: %s", err)
	}

	err = cur.Close(repo.DbContext.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to close cursor: %s", err)
	}

	return jobs, nil
}

func (repo MongoJobRepository) Unlock(managerId string) error {
	objId, err := bson.ObjectIDFromHex(managerId)
	if err != nil {
		return &repositoryErrors.InvalidIdError{
			Value: managerId,
		}
	}

	filterDoc := bson.D{{
		Key:   "managerId",
		Value: objId,
	}}
	updateDoc := mongoModels.JobUnlock()
	coll := repo.DbContext.db.Collection(JobsCollection)
	if _, err = coll.UpdateMany(repo.DbContext.ctx, filterDoc, updateDoc); err != nil {
		return fmt.Errorf("failed to unlock jobs: %s", err)
	}

	return nil
}
