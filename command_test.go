package eventsource

import (
	"context"
	"testing"

	"github.com/slimloans/golly"
	"github.com/slimloans/golly/orm"
	"gorm.io/gorm"
)

type TestAggregate struct {
	AggregateBase

	SomeValue int
}

func (t *TestAggregate) Load(db *gorm.DB) error   { return db.Find(&t, "id = ?", t.ID).Error }
func (t *TestAggregate) Create(db *gorm.DB) error { return db.Create(&t).Error }
func (t *TestAggregate) Save(db *gorm.DB, original interface{}) error {
	return db.Save(&t).Error
}

func (aggregate *TestAggregate) ApplyChange(ctx golly.Context, e Event) {
	switch event := e.Data.(type) {
	case TestEvent:
		aggregate.SomeValue = event.Value
	}
}

func (aggregate *TestAggregate) HandleCommand(ctx golly.Context, db *gorm.DB, c Command) error {
	return nil
}

type TestCommand struct{ Value int }
type TestEvent struct{ Value int }

func (TestCommand) Validate(Aggregate) error { return nil }
func (c TestCommand) Perform(ctx golly.Context, db *gorm.DB, aggregate Aggregate) error {
	aggregate.Apply(ctx, aggregate, TestEvent(c), true)
	return nil
}

func TestCall(t *testing.T) {
	db := orm.NewInMemoryConnection(TestAggregate{}, EventDB{})

	gctx := golly.NewContext(context.TODO())
	t.Run("it should call the test aggregate and save it", func(t *testing.T) {
		aggregate := TestAggregate{}

		Call(gctx, db, TestCommand{Value: 1}, &aggregate, Metadata{})
	})
}
