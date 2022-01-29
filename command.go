package eventsource

import (
	"context"
	"net/http"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"github.com/slimloans/golly/errors"
	"github.com/slimloans/golly/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrorConflict = errors.Error{
		Key:    "ERROR.UPDATE_CONFLICT",
		Status: http.StatusConflict,
	}

	ErrorNewRecord = errors.Error{
		Key:    "ERROR.INVALID_RECORD",
		Status: http.StatusNotFound,
	}
)

type CommandRegistryType struct {
	Command   Command
	Aggregate interface{}
}

// CommandRegister holds the registery of commands
var CommandRegister = map[string]CommandRegistryType{}

var aggregateRegistry = map[string]Aggregate{}

// RegisterCommand register a name in commandregister
func RegisterCommand(aggregate Aggregate, cmds ...Command) {
	// Do this here for now
	aggregateRegistry[aggregate.Type()] = aggregate

	var ag interface{} = aggregate
	if ag != nil {
		aggregateVal := reflect.ValueOf(aggregate)
		if aggregateVal.Kind() == reflect.Ptr {
			ag = aggregateVal.Elem().Interface()
		}
	}

	for _, cmd := range cmds {
		tpe := utils.GetTypeWithPackage(cmd)
		CommandRegister[tpe] = CommandRegistryType{cmd, ag}
	}
}

type CommandInterfaces struct {
	Command   reflect.Type
	Aggregate reflect.Type
}

// FindCommand a command by name
func FindCommand(name string) (CommandInterfaces, bool) {
	if cmd, ok := CommandRegister[name]; ok {
		return CommandInterfaces{
			Command:   reflect.TypeOf(cmd.Command),
			Aggregate: reflect.TypeOf(cmd.Aggregate),
		}, true
	}
	return CommandInterfaces{}, false
}

type Command interface {
	Perform(context.Context, *gorm.DB, Aggregate) error
	Validate(Aggregate) error
}

// Call - call a command
// TODO we will want to
func Call(ctx context.Context, db *gorm.DB, command Command, aggregate Aggregate, metadata Metadata) (Aggregate, Event, error) {
	var event Event
	var changes []Event

	if err := command.Validate(aggregate); err != nil {
		return aggregate, Event{}, errors.WrapGeneric(err)
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		var newRecord = true
		if aggregate.GetID() != uuid.Nil {
			newRecord = false
			if err := aggregate.Load(tx.Clauses(clause.Locking{Strength: "UPDATE"})); err != nil {
				return errors.WrapGeneric(err)
			}
		}

		originalVersion := aggregate.GetVersion()

		var original interface{}

		fieldVal := reflect.ValueOf(aggregate)
		if fieldVal.Kind() == reflect.Ptr {
			original = fieldVal.Elem().Interface()
		} else {
			original = fieldVal.Interface()
		}

		if err := command.Perform(ctx, tx, aggregate); err != nil {
			return errors.WrapGeneric(err)
		}

		changes = aggregate.Uncommited()
		if len(changes) > 0 {
			if newRecord {
				if err := aggregate.Create(tx); err != nil {
					return errors.WrapGeneric(err)
				}
			} else {
				if err := aggregate.Save(tx, original); err != nil {
					return errors.WrapGeneric(err)
				}
			}

			for pos, change := range changes {
				change.Version = uint(int(originalVersion) + (pos + 1))
				change.AggregateID = aggregate.GetID()
				change.Metadata = mergeMetaData(change.Metadata, metadata)

				if eventDBToSave, err := change.Encode(); err != nil {
					return errors.WrapGeneric(err)
				} else if err = tx.Model(eventDBToSave).Create(&eventDBToSave).Error; err != nil {
					return errors.WrapGeneric(err)
				}

				changes[pos] = change
			}

		}
		return nil
	})

	for _, change := range changes {
		Dispatch(ctx, Topic(aggregate), change)
	}

	aggregate.ClearUncommited()

	return aggregate, event, errors.WrapGeneric(err)
}

func mergeMetaData(m1, m2 Metadata) Metadata {
	if m1 == nil {
		m1 = Metadata{}
	}

	if m2 == nil {
		m2 = Metadata{}
	}

	for k, v := range m2 {
		m1[k] = v
	}
	return m1
}

func Topic(aggregate Aggregate) string {
	return strings.ToLower("events." + aggregate.Type())
}
