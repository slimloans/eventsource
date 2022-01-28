package eventsource

import (
	"context"

	"github.com/slimloans/golly"
)

type ConsumerFunc func(golly.Context, Event)

// Dispatch dispatch and event
func Dispatch(ctx context.Context, topic string, event Event) error {
	bus.Publish(topic, ctx, event)
	return nil
}

func Subscribe(topic string, fn ConsumerFunc) {
	bus.Subscribe(topic, fn)
}

func SubscribeAsync(topic string, fn ConsumerFunc) {
	bus.SubscribeAsync(topic, fn, true)
}

func Wait() {
	bus.WaitAsync()
}
