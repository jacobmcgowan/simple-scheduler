package mongoRepos

import (
	"fmt"

	mongoModels "github.com/jacobmcgowan/simple-scheduler/data-access/models/mongo"
	"github.com/jacobmcgowan/simple-scheduler/dtos"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const RunsCollection = "runs"

type RunRepository struct {
	DbContext DbContext
}

func (repo RunRepository) Browse(filter dtos.RunFilter) ([]dtos.Run, error) {
	var runs []dtos.Run
	filterDoc := mongoModels.RunFilterFromDto(filter)
	coll := repo.DbContext.db.Collection(RunsCollection)
	cur, err := coll.Find(repo.DbContext.ctx, filterDoc)
	if err != nil {
		return nil, fmt.Errorf("failed to find runs: %s", err)
	}

	for cur.Next(repo.DbContext.ctx) {
		var run mongoModels.Run
		err = cur.Decode(&run)
		if err != nil {
			return nil, fmt.Errorf("failed to parse run: %s", err)
		}

		runs = append(runs, run.ToDto())
	}

	err = cur.Close(repo.DbContext.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to close cursor: %s", err)
	}

	return runs, nil
}

func (repo RunRepository) Read(id string) (dtos.Run, error) {
	var run mongoModels.Run
	filter := bson.D{{
		Key:   "_id",
		Value: id,
	}}
	coll := repo.DbContext.db.Collection(RunsCollection)
	err := coll.FindOne(repo.DbContext.ctx, filter).Decode(&run)
	if err != nil {
		return dtos.Run{}, fmt.Errorf("failed to find run %s: %s", id, err)
	}

	return run.ToDto(), nil
}

func (repo RunRepository) Edit(id string, update dtos.RunUpdate) error {
	updateDoc := mongoModels.RunUpdateFromDto(update)
	filter := bson.D{{
		Key:   "_id",
		Value: id,
	}}
	coll := repo.DbContext.db.Collection(RunsCollection)
	_, err := coll.UpdateOne(repo.DbContext.ctx, filter, updateDoc)
	if err != nil {
		return fmt.Errorf("failed to edit run %s: %s", id, err)
	}

	return nil
}

func (repo RunRepository) Add(run dtos.Run) (string, error) {
	runDoc := mongoModels.Run{}
	runDoc.FromDto(run)

	coll := repo.DbContext.db.Collection(RunsCollection)
	res, err := coll.InsertOne(repo.DbContext.ctx, runDoc)
	if err != nil {
		return "", fmt.Errorf("failed to add run: %s", err)
	}

	if id, ok := res.InsertedID.(primitive.ObjectID); ok {
		return id.Hex(), nil
	}

	return "", fmt.Errorf("failed to parse id of run: %s", err)
}

func (repo RunRepository) Delete(name string) error {
	filter := bson.D{{
		Key:   "_id",
		Value: name,
	}}
	coll := repo.DbContext.db.Collection(RunsCollection)
	_, err := coll.DeleteOne(repo.DbContext.ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete run %s: %s", name, err)
	}

	return nil
}
