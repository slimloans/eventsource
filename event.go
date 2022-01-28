package eventsource

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/slimloans/go/utils"
	"github.com/slimloans/golly"
	"github.com/slimloans/golly/errors"
	"github.com/slimloans/golly/orm"
	"gorm.io/gorm"
)

var EventRegistry = map[string]reflect.Type{}

type Metadata = map[string]interface{}

type Event struct {
	orm.ModelUUID

	AggregateID   uuid.UUID `json:"aggregate_id"`
	AggregateType string    `json:"aggregate_type"`
	Version       uint      `json:"version"`

	Type string `json:"type"`

	Data interface{} `json:"data"`

	Metadata Metadata `json:"metadata"`
}

func (e Event) GetAggregateID() uuid.UUID {
	return e.AggregateID
}

func (e Event) GetAggregateType() string {
	return e.AggregateType
}

func (e Event) GetVersion() uint {
	return e.Version
}

type EventDB struct {
	orm.ModelUUID

	AggregateID   uuid.UUID `json:"aggregate_id"`
	AggregateType string    `json:"aggregate_type"`

	Version uint   `json:"version"`
	Type    string `json:"type"`

	RawData     postgres.Jsonb `json:"-" gorm:"type:jsonb;column:data"`
	RawMetadata postgres.Jsonb `json:"-" gorm:"type:jsonb;column:metadata"`
}

// TableName update the tablename
func (EventDB) TableName() string {
	return "events"
}

func RegisterEvents(events ...interface{}) {
	for _, event := range events {
		t := utils.GetRawType(event)
		name := utils.GetTypeWithPackage(event)

		EventRegistry[name] = t
	}
}

// var randssource = rand.NewSource(time.Now().UnixNano())
// var randssource = rand.New(randsource)

func newEvent(ctx golly.Context, aggregate Aggregate, data interface{}) Event {
	event := eventFromData(ctx, data)

	event.AggregateID = aggregate.GetID()
	event.AggregateType = utils.GetType(aggregate)

	event.Version = aggregate.GetVersion() + 1
	return event
}

func eventFromData(ctx golly.Context, data interface{}) Event {
	event := Event{}
	event.ID = uuid.New()
	event.CreatedAt = time.Now()

	event.Type = utils.GetTypeWithPackage(data)
	event.Data = data

	return event
}

// ReplayEvents allows you to reply specific events
// on an aggregate
func ReplayEvents(ctx golly.Context, aggregateID uuid.UUID, aggregateType string, events []Event) error {
	var aggregate Aggregate
	if ag, found := aggregateRegistry[aggregateType]; found {
		aggregate = reflect.New(reflect.TypeOf(ag)).Interface().(Aggregate)
	}

	for _, event := range events {
		aggregate.Apply(ctx, aggregate, event, false)
	}
	return nil
}

// Events returns **All** the persisted events for an aggregate
func Events(db *gorm.DB, aggregate Aggregate) ([]Event, error) {
	events := []EventDB{}
	ret := []Event{}

	db.Where(map[string]interface{}{
		"aggregate_type": utils.GetType(aggregate),
		"aggregate_id":   aggregate.GetID(),
	}).Order("created_at DESC").Find(&events)

	for _, event := range events {
		if ev, err := event.Decode(); err != nil {
			return []Event{}, err
		} else {
			ret = append(ret, ev)
		}
	}
	return ret, nil
}

// Decode return a deserialized event, ready to user
func (event EventDB) Decode() (Event, error) {
	evt, found := EventRegistry[event.Type]
	if !found {
		return Event{}, errors.WrapGeneric(fmt.Errorf("cannot find event %s", event.Type))
	}

	dataValue := reflect.New(evt)
	marshal := dataValue.Elem().Addr()

	// 	Elem()
	if err := json.Unmarshal(event.RawData.RawMessage, marshal.Interface()); err != nil {
		return Event{}, errors.WrapGeneric(fmt.Errorf("error when decoding %s %#v", event.Type, err))
	}

	ret := Event{
		ModelUUID: orm.ModelUUID{
			ID:        event.ID,
			CreatedAt: event.CreatedAt,
		},
		AggregateID:   event.AggregateID,
		AggregateType: event.AggregateType,
		Version:       event.Version,
		Type:          event.Type,
		Data:          marshal.Elem().Interface(),
	}

	if err := json.Unmarshal(event.RawMetadata.RawMessage, &ret.Metadata); err != nil {
		return Event{}, errors.WrapGeneric(err)
	}

	return ret, nil
}

// Encode returns a resiralized version of the event, ready to go to the Database
func (event Event) Encode() (EventDB, error) {
	var err error

	ret := EventDB{
		ModelUUID: orm.ModelUUID{
			ID:        event.ID,
			CreatedAt: event.CreatedAt,
		},
		AggregateID:   event.AggregateID,
		AggregateType: event.AggregateType,
		Type:          event.Type,
		Version:       event.Version,
	}

	ret.RawMetadata.RawMessage, err = json.Marshal(event.Metadata)
	if err != nil {
		return EventDB{}, err
	}

	ret.RawData.RawMessage, err = json.Marshal(event.Data)
	if err != nil {
		return EventDB{}, err
	}

	return ret, nil
}

func RemapEvent(event interface{}, dest interface{}) error {
	b, err := json.Marshal(event)
	if err != nil {
		return errors.WrapGeneric(err)
	}
	return json.Unmarshal(b, dest)
}
