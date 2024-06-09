package messageBus

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Producer struct {
	Exchange   string
	Connection *amqp.Connection
}

func (pro Producer) Publish(key string, body []byte) error {
	ch, err := pro.Connection.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %s", err)
	}
	defer ch.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(
		ctx,
		pro.Exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message %s to exchange %s: %s", body, pro.Exchange, err)
	}

	return nil
}
