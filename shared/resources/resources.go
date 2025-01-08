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

type DbResources struct {
	Name    string
	Context repositories.DbContext
	JobRepo repositories.JobRepository
	RunRepo repositories.RunRepository
}

func RegisterRepos() (DbResources, error) {
	dbType := os.Getenv("SIMPLE_SCHEDULER_DB_TYPE")
	conStr := os.Getenv("SIMPLE_SCHEDULER_DB_CONNECTION_STRING")
	conStrUrl, err := url.Parse(conStr)
	if err != nil {
		dbResources := DbResources{
			Name:    "",
			Context: nil,
			JobRepo: nil,
			RunRepo: nil,
		}
		return dbResources, fmt.Errorf("connection string invalid: %s", err)
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

		dbResources := DbResources{
			Name:    dbName + "@" + conStrUrl.Host,
			Context: &dbCtx,
			JobRepo: jobRepo,
			RunRepo: runRepo,
		}
		return dbResources, nil
	default:
		dbResources := DbResources{
			Name:    conStrUrl.Host,
			Context: nil,
			JobRepo: nil,
			RunRepo: nil,
		}
		return dbResources, fmt.Errorf("DB type %s not supported", dbType)
	}
}

type MessageBusResources struct {
	Name       string
	MessageBus messageBus.MessageBus
}

func RegisterMessageBus() (MessageBusResources, error) {
	msgBusType := os.Getenv("SIMPLE_SCHEDULER_MESSAGEBUS_TYPE")
	conStr := os.Getenv("SIMPLE_SCHEDULER_MESSAGEBUS_CONNECTION_STRING")
	conStrUrl, err := url.Parse(conStr)
	if err != nil {
		msgBusResources := MessageBusResources{
			Name:       "",
			MessageBus: nil,
		}
		return msgBusResources, fmt.Errorf("connection string invalid: %s", err)
	}

	switch msgBusType {
	case string(messageBusTypes.RabbitMq):
		msgBus := rabbitmqMessageBus.RabbitMessageBus{
			ConnectionString: conStr,
		}
		msgBusResources := MessageBusResources{
			Name:       conStrUrl.Host + conStrUrl.Path,
			MessageBus: &msgBus,
		}
		return msgBusResources, nil
	default:
		msgBusResources := MessageBusResources{
			Name:       conStrUrl.Host + conStrUrl.Path,
			MessageBus: nil,
		}
		return msgBusResources, fmt.Errorf("message bus type %s not supported", msgBusType)
	}
}
