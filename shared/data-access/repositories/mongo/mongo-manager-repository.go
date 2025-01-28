package mongoRepos

import (
	"fmt"

	mongoModels "github.com/jacobmcgowan/simple-scheduler/shared/data-access/models/mongo"
	repositoryErrors "github.com/jacobmcgowan/simple-scheduler/shared/data-access/repositories/errors"
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const ManagersCollection = "managers"

type MongoManagerRepository struct {
	DbContext *MongoDbContext
}

func (repo MongoManagerRepository) Browse() ([]dtos.Manager, error) {
	var mngrs []dtos.Manager
	coll := repo.DbContext.db.Collection(ManagersCollection)
	cur, err := coll.Find(repo.DbContext.ctx, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("failed to find managers: %s", err)
	}

	for cur.Next(repo.DbContext.ctx) {
		var mngr mongoModels.Manager
		err = cur.Decode(&mngr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse manager: %s", err)
		}

		mngrs = append(mngrs, mngr.ToDto())
	}

	err = cur.Close(repo.DbContext.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to close cursor: %s", err)
	}

	return mngrs, nil
}

func (repo MongoManagerRepository) Read(id string) (dtos.Manager, error) {
	objId, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return dtos.Manager{}, &repositoryErrors.InvalidIdError{
			Value: id,
		}
	}

	var mngr mongoModels.Manager
	filter := bson.D{{
		Key: "_id",
		Value: bson.D{{
			Key:   "$eq",
			Value: objId,
		}},
	}}
	coll := repo.DbContext.db.Collection(ManagersCollection)
	err = coll.FindOne(repo.DbContext.ctx, filter).Decode(&mngr)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return dtos.Manager{}, &repositoryErrors.NotFoundError{
				Message: fmt.Sprintf("failed to find manager %s: %s", id, err),
			}
		}

		return dtos.Manager{}, fmt.Errorf("failed to find manager %s: %s", id, err)
	}

	return mngr.ToDto(), nil
}

func (repo MongoManagerRepository) Add(mngr dtos.Manager) (string, error) {
	mngrDoc := mongoModels.Manager{}
	mngrDoc.FromDto(mngr)

	coll := repo.DbContext.db.Collection(ManagersCollection)
	res, err := coll.InsertOne(repo.DbContext.ctx, mngrDoc)
	if err != nil {
		return "", fmt.Errorf("failed to add manager: %s", err)
	}

	if id, ok := res.InsertedID.(bson.ObjectID); ok {
		return id.Hex(), nil
	}

	return "", fmt.Errorf("failed to parse id of manager: %s", err)
}

func (repo MongoManagerRepository) Delete(id string) error {
	objId, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return &repositoryErrors.InvalidIdError{
			Value: id,
		}
	}

	filter := bson.D{{
		Key: "_id",
		Value: bson.D{{
			Key:   "$eq",
			Value: objId,
		}},
	}}
	coll := repo.DbContext.db.Collection(ManagersCollection)
	_, err = coll.DeleteOne(repo.DbContext.ctx, filter)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &repositoryErrors.NotFoundError{
				Message: fmt.Sprintf("failed to find manager %s: %s", id, err),
			}
		}

		return fmt.Errorf("failed to delete manager %s: %s", id, err)
	}

	return nil
}
