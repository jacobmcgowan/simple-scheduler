package integration_tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/jacobmcgowan/simple-scheduler/shared/resources"
	"github.com/stretchr/testify/require"
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
