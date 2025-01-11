package integration_tests

import (
	"context"
	"fmt"
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

type ContainerResources struct {
	DbContainer         *mongodb.MongoDBContainer
	MessageBusContainer *rabbitmq.RabbitMQContainer
	DbEnv               resources.DbEnv
	MessageBusEnv       resources.MessageBusEnv
}

func initContainers(t *testing.T, ctx context.Context) ContainerResources {
	dbEnv := resources.DbEnv{
		Type: "mongodb",
		Name: "simpleSchedulerTests",
	}
	msgBusEnv := resources.MessageBusEnv{
		Type: "rabbitmq",
	}
	dbC, err := mongodb.Run(ctx, "mongodb/mongodb-community-server:latest")
	require.NoError(t, err)

	dbHost, err := dbC.Host(ctx)
	require.NoError(t, err)
	dbPort, err := dbC.MappedPort(ctx, "27017")
	require.NoError(t, err)
	dbEnv.ConnectionString = fmt.Sprintf("mongodb://%s:%s", dbHost, dbPort.Port())

	rabbitC, err := rabbitmq.Run(
		ctx,
		"rabbitmq:3.12.11-management-alpine",
		rabbitmq.WithAdminUsername("guest"),
		rabbitmq.WithAdminPassword("guest"),
	)
	require.NoError(t, err)

	rabbitConnStr, err := rabbitC.AmqpURL(ctx)
	require.NoError(t, err)
	msgBusEnv.ConnectionString = rabbitConnStr

	return ContainerResources{
		DbContainer:         dbC,
		MessageBusContainer: rabbitC,
		DbEnv:               dbEnv,
		MessageBusEnv:       msgBusEnv,
	}
}

func TestRecurringJobWithRabbitMQ(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cRes := initContainers(t, ctx)
	defer testcontainers.TerminateContainer(cRes.DbContainer)
	defer testcontainers.TerminateContainer(cRes.MessageBusContainer)

	dbResources, err := resources.RegisterRepos(cRes.DbEnv)
	require.NoError(t, err)

	err = dbResources.Context.Connect(ctx)
	require.NoError(t, err)
	defer dbResources.Context.Disconnect()

	msgBusResources, err := resources.RegisterMessageBus(cRes.MessageBusEnv)
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
		CleanupDuration:      time.Minute * 1000, // Prevent cleanup
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

func TestRunCleanupWithRabbitMQ(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cRes := initContainers(t, ctx)
	defer testcontainers.TerminateContainer(cRes.DbContainer)
	defer testcontainers.TerminateContainer(cRes.MessageBusContainer)

	dbResources, err := resources.RegisterRepos(cRes.DbEnv)
	require.NoError(t, err)

	err = dbResources.Context.Connect(ctx)
	require.NoError(t, err)
	defer dbResources.Context.Disconnect()

	msgBusResources, err := resources.RegisterMessageBus(cRes.MessageBusEnv)
	require.NoError(t, err)

	err = msgBusResources.MessageBus.Connect()
	require.NoError(t, err)
	defer msgBusResources.MessageBus.Close()

	job := dtos.Job{
		Name:                "Test Job",
		Enabled:             true,
		NextRunAt:           time.Now().Add(time.Second),
		Interval:            1000,
		RunStartTimeout:     1000,
		RunExecutionTimeout: 1000,
		HeartbeatTimeout:    1000,
	}
	jobName, err := dbResources.JobRepo.Add(job)
	require.NoError(t, err)

	wg := sync.WaitGroup{}
	manager := workers.ManagerWorker{
		MessageBus:           msgBusResources.MessageBus,
		JobRepo:              dbResources.JobRepo,
		RunRepo:              dbResources.RunRepo,
		CacheRefreshDuration: time.Minute * 1000, // Prevent cache refresh
		CleanupDuration:      time.Second,
	}
	manager.Start(&wg)

	execExpRuns := []string{}
	hrtbtExpRuns := []string{}
	client := TestClientWorker{
		Job:        job,
		MessageBus: msgBusResources.MessageBus,
		RunStarted: func(runId string) {
			time.Sleep(time.Millisecond * 50)
			run, err := dbResources.RunRepo.Read(runId)
			require.NoError(t, err)
			require.Equal(t, runStatuses.Running, run.Status)

			if len(execExpRuns) > len(hrtbtExpRuns) {
				hrtbtExpRuns = append(hrtbtExpRuns, runId)
			} else {
				execExpRuns = append(execExpRuns, runId)
			}
		},
	}
	client.Start(&wg)
	time.Sleep(time.Second * 2) // wait for the client to start a couple runs
	client.Stop()

	time.Sleep(time.Second)

	require.Equal(t, 1, len(execExpRuns))
	for _, runId := range execExpRuns {
		run, err := dbResources.RunRepo.Read(runId)
		require.NoError(t, err)
		require.Equal(t, runStatuses.Cancelling, run.Status)
	}

	require.Equal(t, 1, len(hrtbtExpRuns))
	for _, runId := range hrtbtExpRuns {
		run, err := dbResources.RunRepo.Read(runId)
		require.NoError(t, err)
		require.Equal(t, runStatuses.Pending, run.Status)
	}

	time.Sleep(time.Second * 2) // wait for a run that the client will not start

	for _, runId := range hrtbtExpRuns {
		run, err := dbResources.RunRepo.Read(runId)
		require.NoError(t, err)
		require.Equal(t, runStatuses.Cancelling, run.Status)
	}

	cancelledFilter := dtos.RunFilter{
		JobName: common.Undefinable[string]{
			Value:   jobName,
			Defined: true,
		},
		Status: common.Undefinable[runStatuses.RunStatus]{
			Value:   runStatuses.Cancelling,
			Defined: true,
		},
	}
	cancelledRuns, err := dbResources.RunRepo.Browse(cancelledFilter)
	require.NoError(t, err)
	require.Greater(t, len(cancelledRuns), len(execExpRuns))
	require.Greater(t, len(cancelledRuns), len(hrtbtExpRuns))

	manager.Stop()
	client.Stop()
	wg.Wait()
}
