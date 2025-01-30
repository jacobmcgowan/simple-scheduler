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

	"github.com/jacobmcgowan/simple-scheduler/services/custodian/workers"
	"github.com/jacobmcgowan/simple-scheduler/shared/resources"
	envVars "github.com/jacobmcgowan/simple-scheduler/shared/resources/env-vars"
	"github.com/joho/godotenv"
)

func main() {
	appName := "simple-scheduler-custodian"

	log.Printf("Starting %s...", appName)

	envFilename := ".env"
	if len(os.Args) > 1 {
		envFilename = os.Args[1]
	}

	if err := godotenv.Load(envFilename); err != nil {
		log.Fatalf("Failed to load environment file, %s", envFilename)
	}

	refreshInterval, err := strconv.Atoi(os.Getenv(envVars.CacheRefreshInterval))
	if err != nil || refreshInterval < 1 {
		log.Fatalf("Cache refresh interval invalid")
	}

	hrtbtTimeout, err := strconv.Atoi(os.Getenv(envVars.HeartbeatTimeout))
	if err != nil || hrtbtTimeout < 1 {
		log.Fatalf("Heartbeat timeout invalid")
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

	wg := sync.WaitGroup{}
	cust := workers.JobCustodian{
		JobRepo:          dbResources.JobRepo,
		Duration:         time.Duration(refreshInterval) * time.Second,
		HeartbeatTimeout: time.Duration(hrtbtTimeout) * time.Second,
	}

	cust.Start(&wg)

	log.Printf("Started %s", appName)

	<-ctx.Done()
	cust.Stop()
	wg.Wait()
}
