package mongoRepos

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DbContext struct {
	Db      *mongo.Database
	Context context.Context
	client  *mongo.Client
	options options.ClientOptions
	dbName  string
}

func (dbContext *DbContext) Connect() error {
	if dbContext.client != nil {
		return nil
	}

	client, err := mongo.Connect(dbContext.Context, &dbContext.options)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %s", err)
	}

	dbContext.client = client
	dbContext.Db = dbContext.client.Database(dbContext.dbName)
	return nil
}

func (dbContext *DbContext) Disconnect() error {
	if dbContext.client == nil {
		return nil
	}

	if err := dbContext.client.Disconnect(dbContext.Context); err != nil {
		return fmt.Errorf("failed to disconnect from MongoDB: %s", err)
	}

	return nil
}
