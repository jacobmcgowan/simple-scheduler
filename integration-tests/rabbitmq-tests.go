package integration_tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func RabbitMqTest(t *testing.T) {
	ctx := context.Background()
	dbReq := testcontainers.ContainerRequest{
		Image:        "mongodb/mongodb-community-server:latest",
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor:   wait.ForLog("mongod startup complete"),
	}
	rabbitReq := testcontainers.ContainerRequest{
		Image:        "rabbitmq:3.13-management",
		ExposedPorts: []string{"15672/tcp"},
		WaitingFor:   wait.ForLog("Server startup complete"),
	}

	dbC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: dbReq,
		Started:          true,
	})
	defer testcontainers.TerminateContainer(dbC)
	require.NoError(t, err)

	rabbitC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: rabbitReq,
		Started:          true,
	})
	defer testcontainers.TerminateContainer(rabbitC)
	require.NoError(t, err)

}
