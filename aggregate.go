package eventsource

import (
	"github.com/slimloans/golly"
)

type Aggregate interface {
	Repo(golly.Context) Repository

	Apply(golly.Context, Event)

	Type() string
	Topic() string

	IncrementVersion()

	Changes() []Event
	SetChanges([]Event)
	ClearChanges()

	GetVersion() uint

	GetID() interface{}
	SetID(interface{}) interface{}
}

// AggregateBase holds the base aggregate for the db
type AggregateBase struct {
	Version uint `json:"version"`

	changes []Event
}

func (ab *AggregateBase) IncrementVersion() {
	ab.Version++
}

// GetID return the aggregatebase id
func (ab *AggregateBase) GetVersion() uint {
	return ab.Version
}

func (ab *AggregateBase) Changes() []Event {
	return ab.changes
}

func (ab *AggregateBase) ClearChanges() {
	ab.changes = []Event{}
}

func (ab *AggregateBase) SetChanges(events []Event) {
	ab.changes = events
}

func Apply(ctx golly.Context, aggregate Aggregate, edata interface{}) {
	ApplyExt(ctx, aggregate, edata, nil, true)
}

func ApplyExt(ctx golly.Context, aggregate Aggregate, edata interface{}, meta Metadata, commit bool) {
	if edata == nil {
		return
	}

	event := NewEvent(edata)
	event.Metadata.Merge(meta)

	aggregate.IncrementVersion()
	aggregate.Apply(ctx, event)

	if !commit {
		return
	}

	event.Version = aggregate.GetVersion()

	aggregate.SetChanges(append(aggregate.Changes(), event))
}
