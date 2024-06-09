package messageBus

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	Exchange        string
	Key             string
	Queue           string
	Connection      *amqp.Connection
	MessageReceived func(json string) bool
	Subscribed      chan bool
}

func (con *Consumer) Subscribe() error {
	if con.Subscribed != nil {
		return nil
	}

	ch, err := con.Connection.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %s", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		con.Queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue %s: %s", con.Queue, err)
	}

	err = ch.QueueBind(
		q.Name,
		con.Key,
		con.Exchange,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue %s to exchange %s: %s", con.Queue, con.Exchange, err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer for queue %s: %s", q.Name, err)
	}

	con.Subscribed = make(chan bool)

	go func() {
		for {
			if <-con.Subscribed {
				return
			}

			for msg := range msgs {
				if con.MessageReceived(string(msg.Body[:])) {
					ch.Ack(msg.DeliveryTag, false)
				} else {
					ch.Nack(msg.DeliveryTag, false, true)
				}
			}
		}
	}()

	return nil
}
