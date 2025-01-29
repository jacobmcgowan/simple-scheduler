package integration_tests

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/jacobmcgowan/simple-scheduler/services/scheduler/workers"
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
	"github.com/jacobmcgowan/simple-scheduler/shared/resources"
	"github.com/jacobmcgowan/simple-scheduler/shared/runStatuses"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
)

func TestRecurringJob(t *testing.T) {
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
		Name:      t.Name() + "-job",
		Enabled:   true,
		NextRunAt: time.Now().Add(time.Second),
		Interval:  1000,
	}
	jobName, err := dbResources.JobRepo.Add(job)
	require.NoError(t, err)

	wg := sync.WaitGroup{}
	mngr := workers.ManagerWorker{
		Hostname:             t.Name() + "-manager",
		MaxJobs:              0,
		MessageBus:           msgBusResources.MessageBus,
		ManagerRepo:          dbResources.ManagerRepo,
		JobRepo:              dbResources.JobRepo,
		RunRepo:              dbResources.RunRepo,
		CacheRefreshDuration: time.Minute * 1000, // Prevent cache refresh
		CleanupDuration:      time.Minute * 1000, // Prevent cleanup
		HeartbeatDuration:    time.Minute * 1000, // Prevent heartbeat
	}
	err = mngr.Start(&wg)
	require.NoError(t, err)

	completedRuns := []string{}
	failedRuns := []string{}
	client := TestClientWorker{
		Job:               job,
		MessageBus:        msgBusResources.MessageBus,
		HeartbeatDuration: time.Minute * 1000, // Prevent heartbeat
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
	err = client.Start(&wg)
	require.NoError(t, err)

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
		JobName: &jobName,
	}
	runs, err := dbResources.RunRepo.Browse(runFilter)
	require.NoError(t, err)
	require.Len(t, runs, 2)

	mngr.Stop()
	client.Stop()
	wg.Wait()
}

func TestRunCleanup(t *testing.T) {
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
		Name:                t.Name() + "-job",
		Enabled:             true,
		NextRunAt:           time.Now().Add(time.Second),
		Interval:            1000,
		RunStartTimeout:     1000,
		RunExecutionTimeout: 1000,
	}
	jobName, err := dbResources.JobRepo.Add(job)
	require.NoError(t, err)

	wg := sync.WaitGroup{}
	mngr := workers.ManagerWorker{
		Hostname:             t.Name() + "-manager",
		MaxJobs:              0,
		MessageBus:           msgBusResources.MessageBus,
		ManagerRepo:          dbResources.ManagerRepo,
		JobRepo:              dbResources.JobRepo,
		RunRepo:              dbResources.RunRepo,
		CacheRefreshDuration: time.Minute * 1000, // Prevent cache refresh
		CleanupDuration:      time.Second,
		HeartbeatDuration:    time.Minute * 1000, // Prevent heartbeat
	}
	err = mngr.Start(&wg)
	require.NoError(t, err)

	execExpRuns := []string{}
	client := TestClientWorker{
		Job:               job,
		MessageBus:        msgBusResources.MessageBus,
		HeartbeatDuration: time.Minute * 1000, // Prevent heartbeat
		RunStarted: func(runId string) {
			time.Sleep(time.Millisecond * 50)
			run, err := dbResources.RunRepo.Read(runId)
			require.NoError(t, err)
			require.Equal(t, runStatuses.Running, run.Status)

			execExpRuns = append(execExpRuns, runId)
		},
	}
	err = client.Start(&wg)
	require.NoError(t, err)

	time.Sleep(time.Second)
	client.Stop()

	time.Sleep(time.Second * 2)

	require.Equal(t, 1, len(execExpRuns))
	for _, runId := range execExpRuns {
		run, err := dbResources.RunRepo.Read(runId)
		require.NoError(t, err)
		require.Equal(t, runStatuses.Cancelling, run.Status)
	}

	time.Sleep(time.Second) // wait for a run that the client will not start

	cancellingStatus := runStatuses.Cancelling
	cancelledFilter := dtos.RunFilter{
		JobName: &jobName,
		Status:  &cancellingStatus,
	}
	cancelledRuns, err := dbResources.RunRepo.Browse(cancelledFilter)
	require.NoError(t, err)
	require.Greater(t, len(cancelledRuns), len(execExpRuns))

	mngr.Stop()
	client.Stop()
	wg.Wait()
}

func TestRunHeartbeat(t *testing.T) {
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
		Name:             t.Name() + "-job",
		Enabled:          true,
		NextRunAt:        time.Now().Add(time.Second),
		Interval:         1000,
		HeartbeatTimeout: 1000,
	}
	_, err = dbResources.JobRepo.Add(job)
	require.NoError(t, err)

	wg := sync.WaitGroup{}
	mngr := workers.ManagerWorker{
		Hostname:             t.Name() + "-manager",
		MaxJobs:              0,
		MessageBus:           msgBusResources.MessageBus,
		ManagerRepo:          dbResources.ManagerRepo,
		JobRepo:              dbResources.JobRepo,
		RunRepo:              dbResources.RunRepo,
		CacheRefreshDuration: time.Minute * 1000, // Prevent cache refresh
		CleanupDuration:      time.Second,
		HeartbeatDuration:    time.Minute * 1000, // Prevent heartbeat
	}
	err = mngr.Start(&wg)
	require.NoError(t, err)

	runs := []string{}
	client := TestClientWorker{
		Job:               job,
		MessageBus:        msgBusResources.MessageBus,
		HeartbeatDuration: time.Millisecond * 100,
		RunStarted: func(runId string) {
			time.Sleep(time.Millisecond * 50)
			run, err := dbResources.RunRepo.Read(runId)
			require.NoError(t, err)
			require.Equal(t, runStatuses.Running, run.Status)

			runs = append(runs, runId)
		},
	}
	err = client.Start(&wg)
	require.NoError(t, err)

	time.Sleep(time.Second)

	for _, runId := range runs {
		run, err := dbResources.RunRepo.Read(runId)
		require.NoError(t, err)
		require.Equal(t, runStatuses.Running, run.Status)
	}

	client.Stop()
	time.Sleep(time.Second * 2)

	for _, runId := range runs {
		run, err := dbResources.RunRepo.Read(runId)
		require.NoError(t, err)
		require.Equal(t, runStatuses.Pending, run.Status)
	}

	mngr.Stop()
	wg.Wait()
}

func TestConcurrentManagerWorkers(t *testing.T) {
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

	jobs := []dtos.Job{
		{
			Name:      t.Name() + "-jobA",
			Enabled:   true,
			NextRunAt: time.Now().Add(time.Second),
			Interval:  1000,
		},
		{
			Name:      t.Name() + "-jobB",
			Enabled:   true,
			NextRunAt: time.Now().Add(time.Second),
			Interval:  1000,
		},
		{
			Name:      t.Name() + "-jobC",
			Enabled:   true,
			NextRunAt: time.Now().Add(time.Second),
			Interval:  1000,
		},
		{
			Name:      t.Name() + "-jobD",
			Enabled:   true,
			NextRunAt: time.Now().Add(time.Second),
			Interval:  1000,
		},
	}

	for _, job := range jobs {
		_, err := dbResources.JobRepo.Add(job)
		require.NoError(t, err)
	}

	wg := sync.WaitGroup{}
	mngrA := workers.ManagerWorker{
		Hostname:             t.Name() + "-managerA",
		MaxJobs:              2,
		MessageBus:           msgBusResources.MessageBus,
		ManagerRepo:          dbResources.ManagerRepo,
		JobRepo:              dbResources.JobRepo,
		RunRepo:              dbResources.RunRepo,
		CacheRefreshDuration: time.Minute * 1000, // Prevent cache refresh
		CleanupDuration:      time.Minute * 1000, // Prevent cleanup
		HeartbeatDuration:    time.Minute * 1000, // Prevent heartbeat
	}

	mngrB := workers.ManagerWorker{
		Hostname:             t.Name() + "-managerB",
		MaxJobs:              1,
		MessageBus:           msgBusResources.MessageBus,
		ManagerRepo:          dbResources.ManagerRepo,
		JobRepo:              dbResources.JobRepo,
		RunRepo:              dbResources.RunRepo,
		CacheRefreshDuration: time.Minute * 1000, // Prevent cache refresh
		CleanupDuration:      time.Minute * 1000, // Prevent cleanup
		HeartbeatDuration:    time.Minute * 1000, // Prevent heartbeat
	}

	err = mngrA.Start(&wg)
	require.NoError(t, err)
	err = mngrB.Start(&wg)
	require.NoError(t, err)

	time.Sleep(time.Millisecond * 50)

	updatedJobs, err := dbResources.JobRepo.Browse()
	require.NoError(t, err)

	mngrAJobs := []string{}
	mngrBJobs := []string{}
	unmngedJobs := []string{}

	for _, job := range updatedJobs {
		if job.ManagerId == mngrA.Id {
			mngrAJobs = append(mngrAJobs, job.Name)
		} else if job.ManagerId == mngrB.Id {
			mngrBJobs = append(mngrBJobs, job.Name)
		} else {
			unmngedJobs = append(unmngedJobs, job.Name)
		}
	}

	require.Len(t, mngrAJobs, 2)
	require.Len(t, mngrBJobs, 1)
	require.Len(t, unmngedJobs, 1)

	mngrA.Stop()
	mngrB.Stop()
	wg.Wait()

	time.Sleep(time.Millisecond * 50)

	updatedJobs, err = dbResources.JobRepo.Browse()
	require.NoError(t, err)

	mngrAJobs = []string{}
	mngrBJobs = []string{}
	unmngedJobs = []string{}

	for _, job := range updatedJobs {
		if job.ManagerId == mngrA.Id {
			mngrAJobs = append(mngrAJobs, job.Name)
		} else if job.ManagerId == mngrB.Id {
			mngrBJobs = append(mngrBJobs, job.Name)
		} else {
			unmngedJobs = append(unmngedJobs, job.Name)
		}
	}

	require.Len(t, mngrAJobs, 0)
	require.Len(t, mngrBJobs, 0)
	require.Len(t, unmngedJobs, 4)
}
