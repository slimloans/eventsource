package eventsource

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"github.com/slimloans/golly"
	"github.com/slimloans/golly/errors"
)

type CommandDTO struct {
	Name        string      `json:"name"`
	AggregateID uuid.UUID   `json:"aggregate_id"`
	Data        interface{} `json:"data"`
	Metadata    Metadata    `json:"metadata"`
}

type commandController struct{}

func (c commandController) Routes(r *golly.Route) {
	r.Post("/", c.IssueCommand)
	r.Get("/", c.ListCommands)
}

func (commandController) ListCommands(ctx golly.WebContext) {
	cmds := []interface{}{}

	for x, val := range CommandRegister {
		attribs := map[string]string{}
		dataValue := reflect.ValueOf(val.Command)

		n := dataValue.NumField()

		for i := 0; i < n; i++ {
			field := dataValue.Type().Field(i)
			name := field.Tag.Get("json")
			if name == "" {
				name = field.Name
			}
			fType := field.Type.String()
			if pos := strings.Index(fType, "."); pos >= 0 {
				fType = fType[pos+1:]
			}

			attribs[name] = strings.ToLower(fType)
		}

		ret := map[string]interface{}{"name": x, "attributes": attribs}

		cmds = append(cmds, ret)
	}

	ctx.RenderJSON(cmds)
}

func (c commandController) IssueCommand(ctx golly.WebContext) {
	dto := CommandDTO{}

	if err := ctx.Params(&dto); err != nil {
		golly.Render(ctx, err)
		return
	}

	_, agg, err := IssueCommandDTO(ctx.Context, dto)
	if err != nil {
		golly.Render(ctx, err)
		return
	}

	golly.Render(ctx, agg)
}

// IssueCommandDTO issues a command via the DTO, this is use for decoupling
// so you can build an app in the same place
func IssueCommandDTO(ctx golly.Context, dto CommandDTO) (Command, Aggregate, error) {
	cmd, aggregate, err := NewCommand(dto)
	if err != nil {
		return cmd, aggregate, err
	}

	agg, _, err := Call(ctx, cmd, aggregate, dto.Metadata)
	return cmd, agg, err
}

// NewCommand create a new command and aggregate based on the looked up DTO
func NewCommand(dto CommandDTO) (Command, Aggregate, error) {
	interfaces, found := FindCommand(dto.Name)

	if !found {
		return nil, nil, errors.WrapUnprocessable(fmt.Errorf("no such command"))
	}

	b, _ := json.Marshal(dto.Data)

	dataValue := reflect.New(interfaces.Command)

	marshal := dataValue.Elem().Addr()

	json.Unmarshal(b, marshal.Interface())

	aggregate := reflect.
		New(interfaces.Aggregate).
		Interface().(Aggregate)

	if dto.AggregateID != uuid.Nil {
		aggregate.SetID(dto.AggregateID)
	}

	return marshal.Elem().Interface().(Command), aggregate, nil
}
