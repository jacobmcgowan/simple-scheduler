package workers

type MessageBus interface {
	Connect(connStr string) error
	Close() error
	Register(exchange string, bindings map[string][]string) error
	Publish(exchange string, key string, body []byte) error
	Subscribe(exchange string, key string, queue string, received func(body []byte) bool) error
}
