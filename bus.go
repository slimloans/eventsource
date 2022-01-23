package eventsource

import (
	evbus "github.com/asaskevich/EventBus"
	"github.com/slimloans/golly"
)

var (
	bus = evbus.New()
)

func Server(a golly.Application) {
	evbus.NewServer(a.Config.GetString("eventbus_bind"), "/_server_bus_", bus)
}
