package backend

import (
	"context"
	"sync"
)

var (
	backend Interface
	lock    sync.RWMutex
)

type Config struct {
	Address  string
	Password string
	Username string
}

type Interface interface {
	Subscribe(context.Context, ...string)
	Publish(context.Context, string, interface{})
}

func RegisterBackend(be Interface) {
	defer lock.Unlock()
	lock.Lock()

	backend = be
}

func Dispatch(ctx context.Context, topic string, message interface{}) {
	if backend != nil {
		backend.Publish(ctx, topic, message)
	}
}

func Subscibe(ctx context.Context, topics ...string) {
	if backend != nil {
		backend.Subscribe(ctx, topics...)
	}
}
