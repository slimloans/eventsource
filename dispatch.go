package eventsource

import (
	"github.com/slimloans/go/slim"
	"github.com/slimloans/golly"
	"github.com/slimloans/golly/orm"
	"github.com/slimloans/golly/utils"
)

type ConsumerFunc func(slim.Context, Event)

func DispatchData(ctx golly.Context, data interface{}) {
	event := eventFromData(ctx, data)

	Dispatch(ctx, event)
}

// Dispatch dispatch and event
func Dispatch(ctx golly.Context, event Event) error {
	sctx := *(&ctx)
	orm.SetDBOnContext(sctx, orm.DB(ctx))

	ctx.Logger().Infof("Dispatching: %#v", topicFromEvent(event.Data))

	bus.Publish(topicFromEvent(event.Data), sctx, event)

	return nil
}

func Subscribe(topic string, fn ConsumerFunc) {

	bus.Subscribe(topic, fn)
}

func SubscribeAsync(topic string, fn ConsumerFunc) {
	bus.SubscribeAsync(topic, fn, true)
}

func topicFromEvent(event interface{}) string {
	return utils.GetTypeWithPackage(event)
}

func Wait() {
	bus.WaitAsync()
}
