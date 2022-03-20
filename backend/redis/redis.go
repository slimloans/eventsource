package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/slimloans/eventsource/backend"
)

type Redis struct {
	c      backend.Config
	Client *redis.Client
}

type Config struct {
	backend.Config
	Client *redis.Client
}

func NewRedisBackend(c Config) Redis {
	if c.Client == nil {
		c.Client = redis.NewClient(&redis.Options{
			Addr:     c.Address,
			Password: c.Password,
			DB:       0,
		})
	}

	return Redis{c: c.Config, Client: c.Client}
}

func (r Redis) Subscribe(ctx context.Context, channels ...string) {
	r.Client.PSubscribe(ctx, channels...)
}

func (r Redis) Publish(ctx context.Context, channel string, data interface{}) {
	r.Client.Publish(ctx, channel, data)
}
