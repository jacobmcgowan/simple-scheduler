package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	messageBus "github.com/jacobmcgowan/simple-scheduler/services/scheduler/message-bus"
	messageBusTypes "github.com/jacobmcgowan/simple-scheduler/services/scheduler/message-bus/message-bus-types"
	"github.com/jacobmcgowan/simple-scheduler/services/scheduler/message-bus/rabbitmqMessageBus"
	"github.com/jacobmcgowan/simple-scheduler/services/scheduler/workers"
	dbTypes "github.com/jacobmcgowan/simple-scheduler/shared/data-access/db-types"
	"github.com/jacobmcgowan/simple-scheduler/shared/data-access/repositories"
	mongoRepos "github.com/jacobmcgowan/simple-scheduler/shared/data-access/repositories/mongo"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	appName := "simple-scheduler"
	envFilename := ".env"
	if len(os.Args) > 1 {
		envFilename = os.Args[1]
	}

	log.Printf("Starting %s...", appName)

	if err := godotenv.Load(envFilename); err != nil {
		log.Fatalf("Failed to load environment file, %s", envFilename)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbName, dbCtx, jobRepo, runRepo, err := registerRepos()
	if err != nil {
		log.Fatalf("Failed to register repositories: %s", err)
	}

	log.Printf("Connecting to database %s...", dbName)
	if err = dbCtx.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect to database: %s", err)
	}
	defer dbCtx.Disconnect()
	log.Println("Connected to database")

	msgBusName, msgBus, err := registerMessageBus()
	if err != nil {
		log.Fatalf("Failed to register message bus: %s", err)
	}

	log.Printf("Connecting to message bus %s...", msgBusName)
	if err = msgBus.Connect(); err != nil {
		log.Fatalf("Failed to connect to message bus: %s", err)
	}
	defer msgBus.Close()
	log.Println("Connected to message bus")

	wg := sync.WaitGroup{}
	manager := workers.ManagerWorker{
		MessageBus:           msgBus,
		JobRepo:              jobRepo,
		RunRepo:              runRepo,
		CacheRefreshDuration: time.Minute * 5,
	}

	manager.Start(&wg)

	log.Printf("Started %s", appName)

	<-ctx.Done()
	manager.Stop()
	wg.Wait()
}

func registerRepos() (string, repositories.DbContext, repositories.JobRepository, repositories.RunRepository, error) {
	dbType := os.Getenv("SIMPLE_SCHEDULER_DB_TYPE")
	conStr := os.Getenv("SIMPLE_SCHEDULER_DB_CONNECTION_STRING")
	conStrUrl, err := url.Parse(conStr)
	if err != nil {
		return "", nil, nil, nil, fmt.Errorf("connection string invalid: %s", err)
	}

	switch dbType {
	case string(dbTypes.MongoDb):
		dbName := os.Getenv("SIMPLE_SCHEDULER_DB_NAME")
		dbCtx := mongoRepos.MongoDbContext{
			DbName:  dbName,
			Options: *options.Client().ApplyURI(conStr),
		}
		jobRepo := mongoRepos.MongoJobRepository{
			DbContext: &dbCtx,
		}
		runRepo := mongoRepos.MongoRunRepository{
			DbContext: &dbCtx,
		}

		return dbName + "@" + conStrUrl.Host, &dbCtx, jobRepo, runRepo, nil
	default:
		return conStrUrl.Host, nil, nil, nil, fmt.Errorf("DB type %s not supported", dbType)
	}
}

func registerMessageBus() (string, messageBus.MessageBus, error) {
	msgBusType := os.Getenv("SIMPLE_SCHEDULER_MESSAGEBUS_TYPE")
	conStr := os.Getenv("SIMPLE_SCHEDULER_MESSAGEBUS_CONNECTION_STRING")
	conStrUrl, err := url.Parse(conStr)
	if err != nil {
		return "", nil, fmt.Errorf("connection string invalid: %s", err)
	}

	switch msgBusType {
	case string(messageBusTypes.RabbitMq):
		msgBus := rabbitmqMessageBus.RabbitMessageBus{
			ConnectionString: conStr,
		}
		return conStrUrl.Host + conStrUrl.Path, &msgBus, nil
	default:
		return conStrUrl.Host + conStrUrl.Path, nil, fmt.Errorf("message bus type %s not supported", msgBusType)
	}
}
