package eventsource

import (
	"github.com/google/uuid"

	"github.com/slimloans/golly"
	"github.com/slimloans/golly/orm"
	"github.com/slimloans/golly/utils"
	"gorm.io/gorm"
)

// AggregateBase holds the base aggregate for the db
type AggregateBase struct {
	orm.ModelUUID

	Version uint `json:"version"`

	changes []Event
}

func (ab *AggregateBase) BeforeCreate(db *gorm.DB) error {
	return ab.ModelUUID.BeforeCreate(db)
}

func (ab *AggregateBase) IncrementVersion() {
	ab.Version++
}

// GetID return the aggregatebase id
func (ab *AggregateBase) GetID() uuid.UUID {
	return ab.ID
}

func (ab *AggregateBase) SetID(id uuid.UUID) {
	ab.ID = id
}

// GetID return the aggregatebase id
func (ab *AggregateBase) GetVersion() uint {
	return ab.Version
}

func (ab *AggregateBase) Uncommited() []Event {
	return ab.changes
}

func (ab *AggregateBase) ClearUncommited() {
	ab.changes = []Event{}
}

func (ab *AggregateBase) NewRecord() bool {
	return ab.GetID() == uuid.Nil
}

// ApplyChangeHelper increments the version of an aggregate and apply the change itself
func (b *AggregateBase) ApplyChangeHelper(ctx golly.Context, aggregate Aggregate, data interface{}, commit bool) {
	event := newEvent(ctx, aggregate, data)

	// increments the version in event and aggregate
	b.IncrementVersion()

	// apply the event itself
	aggregate.ApplyChange(ctx, event)
	if commit {
		event.Version = b.Version
		if event.Data != nil {
			event.Type = utils.GetTypeWithPackage(event.Data)
		}

		if event.Metadata == nil {
			event.Metadata = Metadata{}
		}

		b.changes = append(b.changes, event)
	}
}

type DiffType map[string]interface{}

type Aggregate interface {
	// HandleCommand(slim.Context, Command) error
	ApplyChangeHelper(golly.Context, Aggregate, interface{}, bool)
	ApplyChange(golly.Context, Event)
	HandleCommand(golly.Context, Command) error

	IncrementVersion()

	Uncommited() []Event
	ClearUncommited()

	GetID() uuid.UUID
	SetID(uuid.UUID)

	GetVersion() uint
	GetType() string

	NewRecord() bool

	Save(db *gorm.DB, original interface{}) error
	Create(db *gorm.DB) error
	Load(db *gorm.DB) error
}
