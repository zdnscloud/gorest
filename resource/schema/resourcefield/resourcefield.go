package resourcefield

import (
	"reflect"
)

type ResourceField interface {
	Validate(interface{}) error
	CheckRequired(raw map[string]interface{}) error
}

func New(typ reflect.Type) (ResourceField, error) {
	builder := NewBuilder()
	if f, err := builder.Build(typ); err != nil {
		return nil, err
	} else if f == nil {
		return nil, nil
	} else {
		return newResourceField(f), nil
	}
}

type resourceField struct {
	field Field
}

func newResourceField(field Field) *resourceField {
	return &resourceField{
		field: field,
	}
}

//validate the resource go struct
func (f *resourceField) Validate(value interface{}) error {
	return f.field.Validate(value)
}

//check the json string whether the required field is specified
func (f *resourceField) CheckRequired(raw map[string]interface{}) error {
	return f.field.CheckRequired(raw)
}
