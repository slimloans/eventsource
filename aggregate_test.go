package eventsource

import "github.com/slimloans/golly"

type testAggregate struct {
	AggregateBase

	repo Repository
}

func (a *testAggregate) Repo(golly.Context) Repository   { return a.repo }
func (*testAggregate) Type() string                      { return "test-aggregate" }
func (*testAggregate) Topic() string                     { return "test/topic" }
func (*testAggregate) Apply(golly.Context, Event)        {}
func (*testAggregate) GetID() interface{}                { return nil }
func (*testAggregate) SetID(obj interface{}) interface{} { return obj }

type testRepostoryBase struct {
	loadCalled       int
	saveCalled       int
	trasactionCalled int
}

func (r *testRepostoryBase) Load(golly.Context, interface{}) error {
	r.loadCalled++
	return nil
}

func (r *testRepostoryBase) Save(golly.Context, interface{}) error {
	r.saveCalled++
	return nil
}

func (r *testRepostoryBase) Transaction(handler func(Repository) error) error {
	r.trasactionCalled++

	return handler(r)
}

func (r *testRepostoryBase) IsNewRecord(interface{}) bool {
	return true
}
