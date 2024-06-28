package rabbitmqMessageBus

import (
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	Queue           string
	Connection      *amqp.Connection
	MessageReceived func(body []byte) bool
	quit            chan struct{}
}

func (con *Consumer) Subscribe(wg *sync.WaitGroup) error {
	if con.quit != nil {
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

	wg.Add(1)
	con.quit = make(chan struct{})

	go func() {
		defer wg.Done()

		for {
			select {
			case <-con.quit:
				return
			default:
				for msg := range msgs {
					if con.MessageReceived(msg.Body) {
						ch.Ack(msg.DeliveryTag, false)
					} else {
						ch.Nack(msg.DeliveryTag, false, true)
					}
				}
			}
		}
	}()

	return nil
}

func (con *Consumer) Unsubscribe() {
	con.quit <- struct{}{}
}
