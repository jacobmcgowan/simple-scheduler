package messageBus

import "sync"

type MessageBus interface {
	Connect() error
	Close() error
	Register(exchange string, bindings map[string][]string) error
	Publish(exchange string, key string, body []byte) error
	Subscribe(wg *sync.WaitGroup, queue string, received func(body []byte) bool) error
}
