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
	envVars "github.com/jacobmcgowan/simple-scheduler/shared/resources/env-vars"
	"github.com/joho/godotenv"
)

func main() {
	appName := "simple-scheduler"

	log.Printf("Starting %s...", appName)

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("Failed to get hostname: %s", err)
	}

	envFilename := ".env"
	if len(os.Args) > 1 {
		envFilename = os.Args[1]
	}

	if err := godotenv.Load(envFilename); err != nil {
		log.Fatalf("Failed to load environment file, %s", envFilename)
	}

	maxJobsStr := os.Getenv(envVars.MaxJobs)
	maxJobs, err := strconv.Atoi(maxJobsStr)
	if err != nil {
		log.Fatalf("Invalid value for %s, %s", envVars.MaxJobs, maxJobsStr)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbEnv := resources.LoadDbEnv()
	dbResources, err := resources.RegisterRepos(dbEnv)
	if err != nil {
		log.Fatalf("Failed to register repositories: %s", err)
	}

	log.Printf("Connecting to database %s...", dbResources.Name)
	if err = dbResources.Context.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect to database: %s", err)
	}
	defer dbResources.Context.Disconnect()
	log.Println("Connected to database")

	msgBusEnv := resources.LoadMessageBusEnv()
	msgBusResources, err := resources.RegisterMessageBus(msgBusEnv)
	if err != nil {
		log.Fatalf("Failed to register message bus: %s", err)
	}

	log.Printf("Connecting to message bus %s...", msgBusResources.Name)
	if err = msgBusResources.MessageBus.Connect(); err != nil {
		log.Fatalf("Failed to connect to message bus: %s", err)
	}
	defer msgBusResources.MessageBus.Close()
	log.Println("Connected to message bus")

	refreshInterval, err := strconv.Atoi(os.Getenv(envVars.CacheRefreshInterval))
	if err != nil || refreshInterval < 1 {
		log.Fatalf("Cache refresh interval invalid")
	}

	cleanupInterval, err := strconv.Atoi(os.Getenv(envVars.CleanupInterval))
	if err != nil || cleanupInterval < 1 {
		log.Fatalf("Cleanup interval invalid")
	}

	wg := sync.WaitGroup{}
	manager := workers.ManagerWorker{
		Hostname:             hostname,
		MaxJobs:              maxJobs,
		MessageBus:           msgBusResources.MessageBus,
		ManagerRepo:          dbResources.ManagerRepo,
		JobRepo:              dbResources.JobRepo,
		RunRepo:              dbResources.RunRepo,
		CacheRefreshDuration: time.Duration(int(time.Millisecond) * refreshInterval),
		CleanupDuration:      time.Duration(int(time.Millisecond) * cleanupInterval),
	}

	manager.Start(&wg)

	log.Printf("Started %s", appName)

	<-ctx.Done()
	manager.Stop()
	wg.Wait()
}
