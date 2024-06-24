package mongoRepos

import (
	"fmt"

	mongoModels "github.com/jacobmcgowan/simple-scheduler/data-access/models/mongo"
	"github.com/jacobmcgowan/simple-scheduler/dtos"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const JobsCollection = "jobs"

type JobRepository struct {
	DbContext *DbContext
}

func (repo JobRepository) Browse() ([]dtos.Job, error) {
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

func (repo JobRepository) Read(name string) (dtos.Job, error) {
	var job mongoModels.Job
	filter := bson.D{{
		Key:   "_id",
		Value: name,
	}}
	coll := repo.DbContext.db.Collection(JobsCollection)
	err := coll.FindOne(repo.DbContext.ctx, filter).Decode(&job)
	if err != nil {
		return dtos.Job{}, fmt.Errorf("failed to find job %s: %s", name, err)
	}

	return job.ToDto(), nil
}

func (repo JobRepository) Edit(name string, update dtos.JobUpdate) error {
	updateDoc := mongoModels.JobUpdateFromDto(update)
	filter := bson.D{{
		Key:   "_id",
		Value: name,
	}}
	coll := repo.DbContext.db.Collection(JobsCollection)
	_, err := coll.UpdateOne(repo.DbContext.ctx, filter, updateDoc)
	if err != nil {
		return fmt.Errorf("failed to edit job %s: %s", name, err)
	}

	return nil
}

func (repo JobRepository) Add(job dtos.Job) (string, error) {
	jobDoc := mongoModels.Job{}
	jobDoc.FromDto(job)

	coll := repo.DbContext.db.Collection(JobsCollection)
	res, err := coll.InsertOne(repo.DbContext.ctx, jobDoc)
	if err != nil {
		return "", fmt.Errorf("failed to add job: %s", err)
	}

	if id, ok := res.InsertedID.(primitive.ObjectID); ok {
		return id.Hex(), nil
	}

	return "", fmt.Errorf("failed to parse id of job: %s", err)
}

func (repo JobRepository) Delete(name string) error {
	filter := bson.D{{
		Key:   "_id",
		Value: name,
	}}
	coll := repo.DbContext.db.Collection(JobsCollection)
	_, err := coll.DeleteOne(repo.DbContext.ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete job %s: %s", name, err)
	}

	return nil
}
