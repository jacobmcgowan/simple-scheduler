package mongoRepos

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoDbContext struct {
	DbName  string
	Options options.ClientOptions
	db      *mongo.Database
	ctx     context.Context
	client  *mongo.Client
}

func (dbContext *MongoDbContext) Connect(ctx context.Context) error {
	if dbContext.client != nil {
		return nil
	}

	dbContext.ctx = ctx
	client, err := mongo.Connect(&dbContext.Options)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %s", err)
	}

	dbContext.client = client
	dbContext.db = dbContext.client.Database(dbContext.DbName)
	return nil
}

func (dbContext *MongoDbContext) Disconnect() error {
	if dbContext.client == nil {
		return nil
	}

	if err := dbContext.client.Disconnect(dbContext.ctx); err != nil {
		return fmt.Errorf("failed to disconnect from MongoDB: %s", err)
	}

	return nil
}
