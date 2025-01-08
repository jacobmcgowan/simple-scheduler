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
	"github.com/jacobmcgowan/simple-scheduler/shared/runStatuses"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
)

func initContainers(t *testing.T, ctx context.Context) (*mongodb.MongoDBContainer, *rabbitmq.RabbitMQContainer) {
	os.Setenv("SIMPLE_SCHEDULER_DB_TYPE", "mongodb")
	os.Setenv("SIMPLE_SCHEDULER_MESSAGEBUS_TYPE", "rabbitmq")
	os.Setenv("SIMPLE_SCHEDULER_DB_NAME", "simpleSchedulerTests")

	dbC, err := mongodb.Run(ctx, "mongodb/mongodb-community-server:latest")
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
	require.NoError(t, err)

	rabbitConnStr, err := rabbitC.AmqpURL(ctx)
	require.NoError(t, err)
	os.Setenv("SIMPLE_SCHEDULER_MESSAGEBUS_CONNECTION_STRING", rabbitConnStr)

	return dbC, rabbitC
}

func TestRecurringJobWithRabbitMQ(t *testing.T) {
	ctx := context.Background()
	dbC, rabbitC := initContainers(t, ctx)
	defer testcontainers.TerminateContainer(dbC)
	defer testcontainers.TerminateContainer(rabbitC)

	dbResources, err := resources.RegisterRepos()
	require.NoError(t, err)

	err = dbResources.Context.Connect(ctx)
	require.NoError(t, err)
	defer dbResources.Context.Disconnect()

	msgBusResources, err := resources.RegisterMessageBus()
	require.NoError(t, err)

	err = msgBusResources.MessageBus.Connect()
	require.NoError(t, err)
	defer msgBusResources.MessageBus.Close()

	job := dtos.Job{
		Name:      "Test Job",
		Enabled:   true,
		NextRunAt: time.Now().Add(time.Second),
		Interval:  1000,
	}
	jobName, err := dbResources.JobRepo.Add(job)
	require.NoError(t, err)

	wg := sync.WaitGroup{}
	manager := workers.ManagerWorker{
		MessageBus:           msgBusResources.MessageBus,
		JobRepo:              dbResources.JobRepo,
		RunRepo:              dbResources.RunRepo,
		CacheRefreshDuration: time.Minute * 1000, // Prevent cache refresh
	}
	manager.Start(&wg)

	completedRuns := []string{}
	failedRuns := []string{}
	client := TestClientWorker{
		Job:        job,
		MessageBus: msgBusResources.MessageBus,
		RunStarted: func(runId string) {
			time.Sleep(time.Millisecond * 50)
			run, err := dbResources.RunRepo.Read(runId)
			require.NoError(t, err)
			require.Equal(t, runStatuses.Running, run.Status)

			if len(completedRuns) > len(failedRuns) {
				failedRuns = append(failedRuns, runId)
			} else {
				completedRuns = append(completedRuns, runId)
			}
		},
	}
	client.Start(&wg)

	time.Sleep(time.Second * 2)

	updatedJob, err := dbResources.JobRepo.Read(jobName)
	require.NoError(t, err)
	require.GreaterOrEqual(t, updatedJob.NextRunAt.Unix(), job.NextRunAt.Unix())

	require.Equal(t, 1, len(completedRuns))
	for _, runId := range completedRuns {
		err = client.CompleteRun(runId)
		require.NoError(t, err)

		time.Sleep(time.Millisecond * 50)

		run, err := dbResources.RunRepo.Read(runId)
		require.NoError(t, err)
		require.Equal(t, runStatuses.Completed, run.Status)
	}

	require.Equal(t, 1, len(failedRuns))
	for _, runId := range failedRuns {
		err = client.FailRun(runId)
		require.NoError(t, err)

		time.Sleep(time.Millisecond * 50)

		run, err := dbResources.RunRepo.Read(runId)
		require.NoError(t, err)
		require.Equal(t, runStatuses.Failed, run.Status)
	}

	runFilter := dtos.RunFilter{
		JobName: common.Undefinable[string]{
			Value:   jobName,
			Defined: true,
		},
	}
	runs, err := dbResources.RunRepo.Browse(runFilter)
	require.NoError(t, err)
	require.Len(t, runs, 2)

	manager.Stop()
	client.Stop()
	wg.Wait()
}
