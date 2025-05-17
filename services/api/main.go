package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jacobmcgowan/simple-scheduler/services/api/auth"
	controllers "github.com/jacobmcgowan/simple-scheduler/services/api/contollers"
	"github.com/jacobmcgowan/simple-scheduler/services/api/middleware"
	"github.com/jacobmcgowan/simple-scheduler/shared/resources"
	envVars "github.com/jacobmcgowan/simple-scheduler/shared/resources/env-vars"
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

	authCache := &auth.AuthCache{
		Issuer: os.Getenv(envVars.OidcIssuer),
	}
	router := gin.Default()
	router.Use(middleware.ErrorHandler())
	router.Use(gin.Recovery())
	controllers.RegisterControllers(router, authCache, dbResources.JobRepo, dbResources.RunRepo)

	srv := &http.Server{
		Addr:    os.Getenv(envVars.ApiUrl),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen: %s\n", err)
		}
	}()

	<-ctx.Done()
	stop()
	log.Printf("Stopping %s...", appName)

	stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(stopCtx); err != nil {
		log.Fatal("Forced to shutdown: ", err)
	}

	log.Printf("Stopped %s", appName)
}
