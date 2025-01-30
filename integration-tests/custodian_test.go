package integration_tests

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/jacobmcgowan/simple-scheduler/services/custodian/workers"
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
	"github.com/jacobmcgowan/simple-scheduler/shared/resources"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestHeartbeat(t *testing.T) {
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

	nilMngrId := bson.NilObjectID.Hex()
	mngrId := bson.NewObjectID().Hex()
	now := time.Now()
	shouldUnlockJob := dtos.Job{
		Name:      t.Name() + "-shouldUnlockJob",
		ManagerId: mngrId,
		Heartbeat: now.Add(-time.Minute),
	}
	_, err = dbResources.JobRepo.Add(shouldUnlockJob)
	require.NoError(t, err)

	shouldNotUnlockJob := dtos.Job{
		Name:      t.Name() + "-shouldNotUnlockJob",
		ManagerId: mngrId,
		Heartbeat: now.Add(time.Minute),
	}
	_, err = dbResources.JobRepo.Add(shouldNotUnlockJob)
	require.NoError(t, err)

	alreadyUnlockedJob := dtos.Job{
		Name:      t.Name() + "-alreadyUnlockedJob",
		Heartbeat: now.Add(-time.Minute),
	}
	_, err = dbResources.JobRepo.Add(alreadyUnlockedJob)
	require.NoError(t, err)

	wg := sync.WaitGroup{}
	cust := workers.JobCustodian{
		JobRepo:          dbResources.JobRepo,
		Duration:         time.Second,
		HeartbeatTimeout: time.Second,
	}
	err = cust.Start(&wg)
	require.NoError(t, err)

	time.Sleep(time.Second * 2)

	jobs, err := dbResources.JobRepo.Browse()
	require.NoError(t, err)

	for _, job := range jobs {
		switch job.Name {
		case shouldUnlockJob.Name:
			require.Equal(t, nilMngrId, job.ManagerId)
		case shouldNotUnlockJob.Name:
			require.Equal(t, mngrId, job.ManagerId)
		case alreadyUnlockedJob.Name:
			require.Equal(t, nilMngrId, job.ManagerId)
		default:
			require.Failf(t, "Unexpected job %s", job.Name)
		}
	}

	cust.Stop()
	wg.Wait()
}
