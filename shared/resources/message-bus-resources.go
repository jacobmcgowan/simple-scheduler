package resources

import (
	"fmt"
	"net/url"
	"os"

	messageBus "github.com/jacobmcgowan/simple-scheduler/shared/message-bus"
	messageBusTypes "github.com/jacobmcgowan/simple-scheduler/shared/message-bus/message-bus-types"
	"github.com/jacobmcgowan/simple-scheduler/shared/message-bus/rabbitmqMessageBus"
)

type MessageBusEnv struct {
	Type             string
	ConnectionString string
}

type MessageBusResources struct {
	Name       string
	MessageBus messageBus.MessageBus
}

func LoadMessageBusEnv() MessageBusEnv {
	return MessageBusEnv{
		Type:             os.Getenv("SIMPLE_SCHEDULER_MESSAGEBUS_TYPE"),
		ConnectionString: os.Getenv("SIMPLE_SCHEDULER_MESSAGEBUS_CONNECTION_STRING"),
	}
}

func RegisterMessageBus(env MessageBusEnv) (MessageBusResources, error) {
	conStrUrl, err := url.Parse(env.ConnectionString)
	if err != nil {
		msgBusResources := MessageBusResources{
			Name:       "",
			MessageBus: nil,
		}
		return msgBusResources, fmt.Errorf("connection string invalid: %s", err)
	}

	switch env.Type {
	case string(messageBusTypes.RabbitMq):
		msgBus := rabbitmqMessageBus.RabbitMessageBus{
			ConnectionString: env.ConnectionString,
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
		return msgBusResources, fmt.Errorf("message bus type %s not supported", env.Type)
	}
}
