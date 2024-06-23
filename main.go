package main

import (
	"context"
	"log"
	"os"

	dbTypes "github.com/jacobmcgowan/simple-scheduler/data-access/db-types"
	"github.com/jacobmcgowan/simple-scheduler/data-access/repositories"
	mongoRepos "github.com/jacobmcgowan/simple-scheduler/data-access/repositories/mongo"
	"github.com/joho/godotenv"
)

func main() {
	envFilename := ".env"
	if len(os.Args) > 1 {
		envFilename = os.Args[1]
	}

	if err := godotenv.Load(envFilename); err != nil {
		log.Fatal("Failed to load environment file, %s", envFilename)
	}

	ctx := context.Background()
}

func registerRepos(ctx context.Context) (repositories.JobRepository, repositories.RunRepository) {
	switch os.Getenv("SIMPLE_SCHEDULER_DB_TYPE") {
	case string(dbTypes.MongoDb):
		dbContext := mongoRepos.DbContext{}
		dbContext.Connect(ctx)
	}
}
