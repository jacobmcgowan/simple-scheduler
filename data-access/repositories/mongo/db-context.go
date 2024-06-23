package mongoRepos

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DbContext struct {
	db      *mongo.Database
	ctx     context.Context
	client  *mongo.Client
	options options.ClientOptions
	dbName  string
}

func (dbContext *DbContext) Connect(ctx context.Context) error {
	if dbContext.client != nil {
		return nil
	}

	dbContext.ctx = ctx
	client, err := mongo.Connect(dbContext.ctx, &dbContext.options)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %s", err)
	}

	dbContext.client = client
	dbContext.db = dbContext.client.Database(dbContext.dbName)
	return nil
}

func (dbContext *DbContext) Disconnect() error {
	if dbContext.client == nil {
		return nil
	}

	if err := dbContext.client.Disconnect(dbContext.ctx); err != nil {
		return fmt.Errorf("failed to disconnect from MongoDB: %s", err)
	}

	return nil
}
