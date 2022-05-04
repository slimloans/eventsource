package eventsource

import (
	"time"

	"github.com/google/uuid"
	"github.com/slimloans/golly/utils"
)

var (
	eventrepo Repository
)

func SetEventRepository(repo Repository) {
	eventrepo = repo
}

type EventData interface{}

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

	Data EventData

	Metadata Metadata
}

func NewEvent(evtData EventData) Event {
	id, _ := uuid.NewUUID()
	return Event{
		ID:        id,
		Name:      utils.GetTypeWithPackage(evtData),
		Version:   0,
		Metadata:  Metadata{},
		Data:      evtData,
		CreatedAt: time.Now(),
	}
}
