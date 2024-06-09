package messageBus

import (
	"errors"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type MessageBus struct {
	Connection *amqp.Connection
	Consumers  map[string]Consumer
}

func (msgBus *MessageBus) Connect(connStr string) error {
	if msgBus.Connection != nil {
		return nil
	}

	conn, err := amqp.Dial(connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %s", err)
	}
	msgBus.Connection = conn
	msgBus.Consumers = make(map[string]Consumer)

	return nil
}

func (msgBus *MessageBus) Close() error {
	if msgBus.Connection == nil {
		return nil
	}

	for _, con := range msgBus.Consumers {
		con.Subscribed <- false
	}

	clear(msgBus.Consumers)

	err := msgBus.Connection.Close()
	if err != nil {
		return fmt.Errorf("failed to close connection: %s", err)
	}

	msgBus.Connection = nil

	return nil
}

func (msgBus MessageBus) Publish(exchange string, key string, json string) error {
	if msgBus.Connection == nil {
		return errors.New("a connection has not been established")
	}

	pro := Producer{
		Exchange:   exchange,
		Connection: msgBus.Connection,
	}
	err := pro.Publish(key, json)
	if err != nil {
		return fmt.Errorf("failed to publish message: %s", err)
	}

	return nil
}

func (msgBus *MessageBus) Subscribe(exchange string, key string, queue string, received func(json string) bool) error {
	if msgBus.Connection == nil {
		return errors.New("a connection has not been established")
	}

	_, hasConsumer := msgBus.Consumers[queue]
	if hasConsumer {
		return nil
	}

	con := Consumer{
		Exchange:        exchange,
		Key:             key,
		Queue:           queue,
		Connection:      msgBus.Connection,
		MessageReceived: received,
	}
	err := con.Subscribe()
	if err != nil {
		return fmt.Errorf("failed to subscribe: %s", err)
	}

	msgBus.Consumers[queue] = con

	return nil
}

func (msgBus *MessageBus) Unsubscribe(queue string) {
	con, hasConsumer := msgBus.Consumers[queue]
	if hasConsumer {
		con.Subscribed <- false
		delete(msgBus.Consumers, queue)
	}
}
