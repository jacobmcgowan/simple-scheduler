package repositories

import "context"

type DbContext interface {
	Connect(ctx context.Context) error
	Disconnect() error
}
