package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jacobmcgowan/simple-scheduler/services/scheduler/workers"
	"github.com/jacobmcgowan/simple-scheduler/shared/resources"
	"github.com/joho/godotenv"
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

	dbName, dbCtx, jobRepo, runRepo, err := resources.RegisterRepos()
	if err != nil {
		log.Fatalf("Failed to register repositories: %s", err)
	}

	log.Printf("Connecting to database %s...", dbName)
	if err = dbCtx.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect to database: %s", err)
	}
	defer dbCtx.Disconnect()
	log.Println("Connected to database")

	msgBusName, msgBus, err := resources.RegisterMessageBus()
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
