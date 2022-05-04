package eventsource

import (
	"reflect"

	"github.com/slimloans/golly/utils"
)

var (
	registry = map[reflect.Type]RegistryItem{}
)

type RegistryOptions struct {
	Aggregate Aggregate

	Commands []Command
	Events   []EventData
	Topics   []string
}

type RegistryItem struct {
	Name string

	RegistryOptions
}

func FindAggregateByName(name string) *RegistryItem {
	for _, reg := range registry {
		if reg.Name == name {
			return &reg
		}
	}
	return nil
}

func FindRegistryItem(ag Aggregate) *RegistryItem {
	if ri, found := registry[reflect.TypeOf(ag)]; found {
		return &ri
	}
	return nil
}

func DefineAggregate(opts RegistryOptions) {
	registry[reflect.TypeOf(opts.Aggregate)] = RegistryItem{
		Name:            utils.GetTypeWithPackage(opts.Aggregate),
		RegistryOptions: opts,
	}
}

//  Register(RegistryOptions{
// 		Aggregate: Aggregate,
// 		Events: []Events{
// 			AggregateCreated{},
// 			AggregateUpdated{},
// 			AggregateDeleted{
// 		},
// 		Commands: []Commands{CreateAggregate{}},
// 		Topics: []string{}
// 	})
