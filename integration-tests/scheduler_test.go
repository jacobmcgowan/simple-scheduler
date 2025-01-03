package integration_tests

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jacobmcgowan/simple-scheduler/services/scheduler/workers"
	"github.com/jacobmcgowan/simple-scheduler/shared/common"
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
	"github.com/jacobmcgowan/simple-scheduler/shared/resources"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
)

func initTests() {
	os.Setenv("SIMPLE_SCHEDULER_DB_TYPE", "mongodb")
	os.Setenv("SIMPLE_SCHEDULER_MESSAGEBUS_TYPE", "rabbitmq")
	os.Setenv("SIMPLE_SCHEDULER_DB_NAME", "simpleSchedulerTests")
}

func TestRecurringJobWithRabbitMQ(t *testing.T) {
	initTests()
	ctx := context.Background()

	dbC, err := mongodb.Run(ctx, "mongodb/mongodb-community-server:latest")
	defer testcontainers.TerminateContainer(dbC)
	require.NoError(t, err)

	dbHost, err := dbC.Host(ctx)
	require.NoError(t, err)
	dbPort, err := dbC.MappedPort(ctx, "27017")
	require.NoError(t, err)
	dbConnStr := fmt.Sprintf("mongodb://%s:%s", dbHost, dbPort.Port())
	os.Setenv("SIMPLE_SCHEDULER_DB_CONNECTION_STRING", dbConnStr)

	rabbitC, err := rabbitmq.Run(
		ctx,
		"rabbitmq:3.12.11-management-alpine",
		rabbitmq.WithAdminUsername("guest"),
		rabbitmq.WithAdminPassword("guest"),
	)
	defer testcontainers.TerminateContainer(rabbitC)
	require.NoError(t, err)

	rabbitConnStr, err := rabbitC.AmqpURL(ctx)
	require.NoError(t, err)
	os.Setenv("SIMPLE_SCHEDULER_MESSAGEBUS_CONNECTION_STRING", rabbitConnStr)
	fmt.Println(rabbitConnStr)

	_, dbCtx, jobRepo, runRepo, err := resources.RegisterRepos()
	require.NoError(t, err)

	err = dbCtx.Connect(ctx)
	require.NoError(t, err)
	defer dbCtx.Disconnect()

	_, msgBus, err := resources.RegisterMessageBus()
	require.NoError(t, err)

	time.Sleep(time.Second) // Give a second for the message bus to start
	err = msgBus.Connect()
	require.NoError(t, err)
	defer msgBus.Close()

	job := dtos.Job{
		Name:      "Test Job",
		Enabled:   true,
		NextRunAt: time.Now().Add(time.Second),
		Interval:  1000,
	}
	jobName, err := jobRepo.Add(job)
	require.NoError(t, err)

	wg := sync.WaitGroup{}
	manager := workers.ManagerWorker{
		MessageBus:           msgBus,
		JobRepo:              jobRepo,
		RunRepo:              runRepo,
		CacheRefreshDuration: time.Minute * 1000, // Prevent cache refresh
	}
	manager.Start(&wg)

	time.Sleep(time.Second * 2)

	updatedJob, err := jobRepo.Read(jobName)
	require.NoError(t, err)
	require.GreaterOrEqual(t, updatedJob.NextRunAt.Unix(), job.NextRunAt.Unix())

	runFilter := dtos.RunFilter{
		JobName: common.Undefinable[string]{
			Value:   jobName,
			Defined: true,
		},
	}
	runs, err := runRepo.Browse(runFilter)
	require.NoError(t, err)
	require.Len(t, runs, 2)

	manager.Stop()
	wg.Wait()
}
