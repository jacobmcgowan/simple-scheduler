package messageBus

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	Queue           string
	Connection      *amqp.Connection
	MessageReceived func(body []byte) bool
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

	msgs, err := ch.Consume(
		con.Queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer for queue %s: %s", con.Queue, err)
	}

	con.Subscribed = make(chan bool)

	go func() {
		for {
			if <-con.Subscribed {
				return
			}

			for msg := range msgs {
				if con.MessageReceived(msg.Body) {
					ch.Ack(msg.DeliveryTag, false)
				} else {
					ch.Nack(msg.DeliveryTag, false, true)
				}
			}
		}
	}()

	return nil
}
