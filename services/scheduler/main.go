package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
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

	refreshInterval, err := strconv.Atoi(os.Getenv("SIMPLE_SCHEDULER_CACHE_REFRESH_INTERVAL"))
	if err != nil || refreshInterval < 1 {
		log.Fatalf("Cache refresh interval invalid")
	}

	cleanupInterval, err := strconv.Atoi(os.Getenv("SIMPLE_SCHEDULER_CLEANUP_INTERVAL"))
	if err != nil || cleanupInterval < 1 {
		log.Fatalf("Cleanup interval invalid")
	}

	wg := sync.WaitGroup{}
	manager := workers.ManagerWorker{
		MessageBus:           msgBus,
		JobRepo:              jobRepo,
		RunRepo:              runRepo,
		CacheRefreshDuration: time.Duration(int(time.Millisecond) * refreshInterval),
		CleanupDuration:      time.Duration(int(time.Millisecond) * cleanupInterval),
	}

	manager.Start(&wg)

	log.Printf("Started %s", appName)

	<-ctx.Done()
	manager.Stop()
	wg.Wait()
}
