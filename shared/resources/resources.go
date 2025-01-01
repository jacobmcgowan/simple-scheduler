package resources

import (
	"fmt"
	"net/url"
	"os"

	dbTypes "github.com/jacobmcgowan/simple-scheduler/shared/data-access/db-types"
	"github.com/jacobmcgowan/simple-scheduler/shared/data-access/repositories"
	mongoRepos "github.com/jacobmcgowan/simple-scheduler/shared/data-access/repositories/mongo"
	messageBus "github.com/jacobmcgowan/simple-scheduler/shared/message-bus"
	messageBusTypes "github.com/jacobmcgowan/simple-scheduler/shared/message-bus/message-bus-types"
	"github.com/jacobmcgowan/simple-scheduler/shared/message-bus/rabbitmqMessageBus"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func RegisterRepos() (string, repositories.DbContext, repositories.JobRepository, repositories.RunRepository, error) {
	dbType := os.Getenv("SIMPLE_SCHEDULER_DB_TYPE")
	conStr := os.Getenv("SIMPLE_SCHEDULER_DB_CONNECTION_STRING")
	conStrUrl, err := url.Parse(conStr)
	if err != nil {
		return "", nil, nil, nil, fmt.Errorf("connection string invalid: %s", err)
	}

	switch dbType {
	case string(dbTypes.MongoDb):
		dbName := os.Getenv("SIMPLE_SCHEDULER_DB_NAME")
		dbCtx := mongoRepos.MongoDbContext{
			DbName:  dbName,
			Options: *options.Client().ApplyURI(conStr),
		}
		jobRepo := mongoRepos.MongoJobRepository{
			DbContext: &dbCtx,
		}
		runRepo := mongoRepos.MongoRunRepository{
			DbContext: &dbCtx,
		}

		return dbName + "@" + conStrUrl.Host, &dbCtx, jobRepo, runRepo, nil
	default:
		return conStrUrl.Host, nil, nil, nil, fmt.Errorf("DB type %s not supported", dbType)
	}
}

func RegisterMessageBus() (string, messageBus.MessageBus, error) {
	msgBusType := os.Getenv("SIMPLE_SCHEDULER_MESSAGEBUS_TYPE")
	conStr := os.Getenv("SIMPLE_SCHEDULER_MESSAGEBUS_CONNECTION_STRING")
	conStrUrl, err := url.Parse(conStr)
	if err != nil {
		return "", nil, fmt.Errorf("connection string invalid: %s", err)
	}

	switch msgBusType {
	case string(messageBusTypes.RabbitMq):
		msgBus := rabbitmqMessageBus.RabbitMessageBus{
			ConnectionString: conStr,
		}
		return conStrUrl.Host + conStrUrl.Path, &msgBus, nil
	default:
		return conStrUrl.Host + conStrUrl.Path, nil, fmt.Errorf("message bus type %s not supported", msgBusType)
	}
}
