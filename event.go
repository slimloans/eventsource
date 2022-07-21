package eventsource

import (
	"time"

	"github.com/google/uuid"
	"github.com/slimloans/golly"
	"github.com/slimloans/golly/utils"
)

const (
	PUBLISH_TYPE_EVENT   = "event"
	PUBLISH_TYPE_COMMAND = "command"
)

var (
	eventBackend EventBackend
)

type EventBackend interface {
	Repository

	Publish(golly.Context, string, DTO)
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
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"event_at"`

	Event   string `json:"event"`
	Version uint   `json:"version"`

	AggregateID   string `json:"arggregate_id"`
	AggregateType string `json:"aggregate_type"`

	Data     interface{} `json:"data" gorm:"-"`
	Metadata Metadata    `json:"metadata" gorm:"-"`

	commit bool
}

type Events []Event

func (evts Events) HasCommited() bool {
	for _, event := range evts {
		if event.commit {
			return true
		}
	}
	return false
}

func (event Event) DTO() DTO {
	return DTO{Type: PUBLISH_TYPE_EVENT, Event: &event}
}

func NewEvent(evtData interface{}) Event {
	id, _ := uuid.NewUUID()

	return Event{
		ID: id,

		Event:     utils.GetTypeWithPackage(evtData),
		Metadata:  Metadata{},
		Data:      evtData,
		CreatedAt: time.Now(),
	}
}
