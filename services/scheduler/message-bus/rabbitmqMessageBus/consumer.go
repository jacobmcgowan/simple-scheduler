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
	isRunningLock   sync.Mutex `default:"sync.Mutex{}"`
	isRunning       bool
}

func (con *Consumer) Subscribe(wg *sync.WaitGroup) error {
	con.isRunningLock.Lock()
	defer con.isRunningLock.Unlock()

	if con.isRunning {
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

	con.quit = make(chan struct{})
	con.channel = ch
	con.messages = msgs

	go con.consume(wg)
	con.isRunning = true

	log.Printf("Subscribed to queue %s", con.Queue)
	return nil
}

func (con *Consumer) Unsubscribe() {
	con.isRunningLock.Lock()
	defer con.isRunningLock.Unlock()

	log.Printf("Unsubscribing from queue %s...", con.Queue)
	con.quit <- struct{}{}
	con.isRunning = false
}

func (con *Consumer) consume(wg *sync.WaitGroup) {
	wg.Add(1)
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
