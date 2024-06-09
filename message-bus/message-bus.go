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

func (msgBus MessageBus) Register(exchange string, bindings map[string][]string) error {
	if msgBus.Connection == nil {
		return errors.New("a connection has not been established")
	}

	ch, err := msgBus.Connection.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %s", err)
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange %s: %s", exchange, err)
	}

	for queue, keys := range bindings {
		q, err := ch.QueueDeclare(
			queue,
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %s", queue, err)
		}

		for _, key := range keys {
			err = ch.QueueBind(
				q.Name,
				key,
				exchange,
				false,
				nil,
			)
			if err != nil {
				return fmt.Errorf("failed to bind queue %s to exchange %s with key %s: %s", q.Name, exchange, key, err)
			}
		}
	}

	return nil
}

func (msgBus MessageBus) Publish(exchange string, key string, body []byte) error {
	if msgBus.Connection == nil {
		return errors.New("a connection has not been established")
	}

	pro := Producer{
		Exchange:   exchange,
		Connection: msgBus.Connection,
	}
	err := pro.Publish(key, body)
	if err != nil {
		return fmt.Errorf("failed to publish message: %s", err)
	}

	return nil
}

func (msgBus *MessageBus) Subscribe(queue string, received func(body []byte) bool) error {
	if msgBus.Connection == nil {
		return errors.New("a connection has not been established")
	}

	_, hasConsumer := msgBus.Consumers[queue]
	if hasConsumer {
		return nil
	}

	con := Consumer{
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
