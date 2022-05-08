package eventsource

import (
	"time"

	"github.com/google/uuid"
	"github.com/slimloans/golly"
	"github.com/slimloans/golly/utils"
)

var (
	eventBackend EventBackend
)

type EventBackend interface {
	Repository

	Publish(golly.Context, string, Event)
}

func SetEventRepository(backend EventBackend) {
	eventBackend = backend
}

type Metadata map[string]interface{}

func (m1 Metadata) Merge(m2 Metadata) {
	if m2 == nil || len(m2) == 0 {
		return
	}

	if m1 == nil {
		m1 = Metadata{}
	}

	for k, v := range m2 {
		m1[k] = v
	}
}

type Event struct {
	ID        uuid.UUID
	CreatedAt time.Time

	Name    string
	Version uint

	AggregateID   interface{}
	AggregateType string

	Data interface{}

	Metadata Metadata
}

func NewEvent(evtData interface{}) Event {
	id, _ := uuid.NewUUID()

	return Event{
		ID:        id,
		Name:      utils.GetTypeWithPackage(evtData),
		Metadata:  Metadata{},
		Data:      evtData,
		CreatedAt: time.Now(),
	}
}
