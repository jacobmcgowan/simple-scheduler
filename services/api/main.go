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
	controllers "github.com/jacobmcgowan/simple-scheduler/services/api/contollers"
	"github.com/jacobmcgowan/simple-scheduler/services/api/middleware"
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

	router := gin.Default()
	router.Use(middleware.ErrorHandler())
	controllers.RegisterControllers(router, jobRepo, runRepo)

	srv := &http.Server{
		Addr:    os.Getenv("SIMPLE_SCHEDULER_API_URL"),
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
