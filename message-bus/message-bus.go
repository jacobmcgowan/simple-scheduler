package messageBus

import "sync"

type MessageBus interface {
	Connect(connStr string) error
	Close() error
	Register(exchange string, bindings map[string][]string) error
	Publish(exchange string, key string, body []byte) error
	Subscribe(wg *sync.WaitGroup, string, key string, queue string, received func(body []byte) bool) error
}
