package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	dbTypes "github.com/jacobmcgowan/simple-scheduler/data-access/db-types"
	"github.com/jacobmcgowan/simple-scheduler/data-access/repositories"
	mongoRepos "github.com/jacobmcgowan/simple-scheduler/data-access/repositories/mongo"
	messageBus "github.com/jacobmcgowan/simple-scheduler/message-bus"
	messageBusTypes "github.com/jacobmcgowan/simple-scheduler/message-bus/message-bus-types"
	"github.com/jacobmcgowan/simple-scheduler/message-bus/rabbitmqMessageBus"
	"github.com/jacobmcgowan/simple-scheduler/workers"
	"github.com/joho/godotenv"
)

func main() {
	envFilename := ".env"
	if len(os.Args) > 1 {
		envFilename = os.Args[1]
	}

	if err := godotenv.Load(envFilename); err != nil {
		log.Fatalf("Failed to load environment file, %s", envFilename)
	}

	ctx := context.Background()
	dbCtx, jobRepo, runRepo, err := registerRepos()
	if err != nil {
		log.Fatalf("Failed to register repositories: %s", err)
	}

	if err = dbCtx.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect to database: %s", err)
	}
	defer dbCtx.Disconnect()

	msgBus, err := registerMessageBus()
	if err != nil {
		log.Fatalf("Failed to register message bus: %s", err)
	}

	if err = msgBus.Connect(os.Getenv("SIMPLE_SCHEDULER_MESSAGEBUS_CONNECTION_STRING")); err != nil {
		log.Fatalf("Failed to connect to message bus: %s", err)
	}
	defer msgBus.Close()

	wg := sync.WaitGroup{}
	manager := workers.ManagerWorker{
		MessageBus:          msgBus,
		JobRepo:             jobRepo,
		RunRepo:             runRepo,
		CacheRefreshMinutes: 5,
	}
	manager.Start(&wg)
	defer manager.Stop()

	wg.Wait()
}

func registerRepos() (repositories.DbContext, repositories.JobRepository, repositories.RunRepository, error) {
	dbType := os.Getenv("SIMPLE_SCHEDULER_DB_TYPE")

	switch dbType {
	case string(dbTypes.MongoDb):
		dbCtx := mongoRepos.DbContext{}
		jobRepo := mongoRepos.JobRepository{
			DbContext: &dbCtx,
		}
		runRepo := mongoRepos.RunRepository{
			DbContext: &dbCtx,
		}

		return &dbCtx, jobRepo, runRepo, nil
	default:
		return nil, nil, nil, fmt.Errorf("DB type %s not supported", dbType)
	}
}

func registerMessageBus() (messageBus.MessageBus, error) {
	msgBusType := os.Getenv("SIMPLE_SCHEDULER_MESSAGEBUS_TYPE")

	switch msgBusType {
	case string(messageBusTypes.RabbitMq):
		msgBus := rabbitmqMessageBus.MessageBus{}
		return &msgBus, nil
	default:
		return nil, fmt.Errorf("Message Bus type % not supported", msgBusType)
	}
}
