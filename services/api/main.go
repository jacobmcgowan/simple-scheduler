package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	controllers "github.com/jacobmcgowan/simple-scheduler/services/api/contollers"
	"github.com/jacobmcgowan/simple-scheduler/shared/resources"
	"github.com/joho/godotenv"
)

func main() {
	appName := "simple-scheduler-api"
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

	dbName, dbCtx, jobRepo, _, err := resources.RegisterRepos()
	if err != nil {
		log.Fatalf("Failed to register repositories: %s", err)
	}

	log.Printf("Connecting to database %s...", dbName)
	if err = dbCtx.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect to database: %s", err)
	}
	defer dbCtx.Disconnect()
	log.Println("Connected to database")

	router := gin.Default()

	controllers.RegisterControllers(router, jobRepo)

	router.Run()
}
