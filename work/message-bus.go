package work

type MessageBus interface {
	Connect(connStr string) error
	Close() error
	Publish(exchange string, key string, json string) error
	Subscribe(exchange string, key string, queue string, received func(json string) bool) error
}
