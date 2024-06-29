package rabbitmqMessageBus

import (
	"fmt"
	"log"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	Queue           string
	Connection      *amqp.Connection
	MessageReceived func(body []byte) bool
	channel         *amqp.Channel
	messages        <-chan amqp.Delivery
	quit            chan struct{}
}

func (con *Consumer) Subscribe(wg *sync.WaitGroup) error {
	if con.quit != nil {
		return nil
	}

	log.Printf("Subscribing to queue %s...", con.Queue)
	ch, err := con.Connection.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %s", err)
	}

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
		ch.Close()
		return fmt.Errorf("failed to register consumer for queue %s: %s", con.Queue, err)
	}

	wg.Add(1)
	con.quit = make(chan struct{})
	con.channel = ch
	con.messages = msgs

	go con.consume(wg)

	log.Printf("Subscribed to queue %s", con.Queue)
	return nil
}

func (con Consumer) Unsubscribe() {
	log.Printf("Unsubscribing from queue %s...", con.Queue)
	con.quit <- struct{}{}
}

func (con *Consumer) consume(wg *sync.WaitGroup) {
	defer wg.Done()
	defer con.channel.Close()

	go func() {
		for msg := range con.messages {
			if con.MessageReceived(msg.Body) {
				con.channel.Ack(msg.DeliveryTag, false)
			} else {
				con.channel.Nack(msg.DeliveryTag, false, true)
			}
		}
	}()

	<-con.quit
	log.Printf("Unsubscribed from queue %s", con.Queue)
}
