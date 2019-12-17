package resourcefield

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/zdnscloud/gorest/resource/schema/resourcefield/validator"
)

type Field interface {
	JsonName() string

	Name() string

	IsRequired() bool
	SetRequired(bool)

	//validate fields of go struct
	Validate(interface{}, map[string]interface{}) error
}

var _ Field = &leafField{}
var _ Field = &structField{}

type leafField struct {
	name       string
	jsonName   string
	required   bool
	validators []validator.Validator
}

func newLeafField(name, jsonName string) *leafField {
	return &leafField{
		name:     name,
		jsonName: jsonName,
		required: false,
	}
}

func (f *leafField) Name() string {
	return f.name
}

func (f *leafField) JsonName() string {
	return f.jsonName
}

func (f *leafField) IsRequired() bool {
	return f.required
}

func (f *leafField) SetRequired(required bool) {
	f.required = required
}

func (f *leafField) SetValidators(validators []validator.Validator) {
	f.validators = validators
}

func (f *leafField) Validate(val interface{}, raw map[string]interface{}) error {
	jsonName := f.JsonName()
	jsonVal, specified := raw[jsonName]
	if f.IsRequired() {
		if !specified {
			return fmt.Errorf("field %s is missing", jsonName)
		}

		if !reflect.ValueOf(jsonVal).IsValid() {
			return fmt.Errorf("field %s has invalid value", jsonName)
		}
	}

	if !specified {
		return nil
	}

	return f.doValidate(val)
}

func (f *leafField) doValidate(val interface{}) error {
	for _, validator := range f.validators {
		if err := validator.Validate(val); err != nil {
			return err
		}
	}
	return nil
}

type sliceLeafField struct {
	*leafField
}

func newSliceLeafField(inner *leafField) *sliceLeafField {
	return &sliceLeafField{
		leafField: inner,
	}
}

func (f *sliceLeafField) Validate(val interface{}, raw map[string]interface{}) error {
	specified, err := fieldIsSpecifiedWithKind(f.leafField, raw, reflect.Slice)
	if err != nil {
		return err
	}

	value := reflect.ValueOf(val)
	if specified {
		for i := 0; i < value.Len(); i++ {
			if err := f.leafField.doValidate(value.Index(i).Interface()); err != nil {
				return err
			}
		}
	}
	return nil
}

func fieldIsSpecifiedWithKind(f Field, raw map[string]interface{}, kind reflect.Kind) (bool, error) {
	jsonName := f.JsonName()
	jsonVal, specified := raw[jsonName]
	if jsonVal == nil {
		specified = false
	}

	if f.IsRequired() {
		if !specified {
			return specified, fmt.Errorf("field %s is missing", jsonName)
		}

		v := reflect.ValueOf(jsonVal)
		if !v.IsValid() {
			return specified, fmt.Errorf("field %s has invalid value", jsonName)
		}

		if v.Kind() != kind {
			return specified, fmt.Errorf("field %s isn't %v", jsonName, kind)
		}
		if v.Len() == 0 {
			return specified, fmt.Errorf("field %s with empty slice ", jsonName)
		}
	}
	return specified, nil
}

type sliceStructField struct {
	Field
	inner Field
}

func newSliceStructField(self, inner Field) *sliceStructField {
	return &sliceStructField{
		Field: self,
		inner: inner,
	}
}

func (f *sliceStructField) Validate(val interface{}, raw map[string]interface{}) error {
	specified, err := fieldIsSpecifiedWithKind(f.Field, raw, reflect.Slice)
	if err != nil {
		return err
	}

	if !specified || f.inner == nil {
		return nil
	}

	jsonName := f.Field.JsonName()
	value := reflect.ValueOf(val)
	jsonValue := reflect.ValueOf(raw[jsonName])
	if value.Len() != jsonValue.Len() {
		panic("json and unmarshalled value isn't related")
	}

	for i := 0; i < value.Len(); i++ {
		d, _ := json.Marshal(jsonValue.Index(i).Interface())
		m := make(map[string]interface{})
		if err := json.Unmarshal(d, &m); err != nil {
			return fmt.Errorf("elem of field %s is not a struct", jsonName)
		}

		if err := f.inner.Validate(value.Index(i).Interface(), m); err != nil {
			return err
		}
	}

	return nil
}

type mapLeafField struct {
	*leafField
}

func newMapLeafField(inner *leafField) *mapLeafField {
	return &mapLeafField{
		leafField: inner,
	}
}

func (f *mapLeafField) Validate(val interface{}, raw map[string]interface{}) error {
	specified, err := fieldIsSpecifiedWithKind(f.leafField, raw, reflect.Map)
	if err != nil {
		return err
	}
	if !specified {
		return nil
	}

	iter := reflect.ValueOf(val).MapRange()
	for iter.Next() {
		if err := f.leafField.doValidate(iter.Value().Interface()); err != nil {
			return err
		}
	}
	return nil
}

type mapStructField struct {
	Field
	inner Field
}

func newMapStructField(self, inner Field) *mapStructField {
	return &mapStructField{
		Field: self,
		inner: inner,
	}
}

func (f *mapStructField) Validate(val interface{}, raw map[string]interface{}) error {
	specified, err := fieldIsSpecifiedWithKind(f.Field, raw, reflect.Map)
	if err != nil {
		return err
	}

	if !specified || f.inner == nil {
		return nil
	}

	jsonName := f.Field.JsonName()
	jsonValue := reflect.ValueOf(raw[jsonName])
	value := reflect.ValueOf(val)
	if jsonValue.Len() != value.Len() {
		panic("json and unmarshalled value isn't related")
	}

	ji := jsonValue.MapRange()
	vi := value.MapRange()
	for ji.Next() && vi.Next() {
		d, _ := json.Marshal(ji.Value().Interface())
		m := make(map[string]interface{})
		if err := json.Unmarshal(d, &m); err != nil {
			return fmt.Errorf("value of field %s is not a struct", jsonName)
		}
		if err := f.inner.Validate(vi.Value().Interface(), m); err != nil {
			return err
		}
	}
	return nil
}

type structField struct {
	Field
	fields map[string]Field
}

func newStructField(self Field, fields map[string]Field) *structField {
	return &structField{
		Field:  self,
		fields: fields,
	}
}

func (f *structField) Validate(val interface{}, raw map[string]interface{}) error {
	value := reflect.ValueOf(val)
	kind := reflect.TypeOf(val).Kind()
	if kind == reflect.Ptr {
		value = value.Elem()
	}

	if value.Kind() != reflect.Struct {
		return fmt.Errorf("struct field doesn't support type %v", kind)
	}

	if f.Field != nil {
		jsonName := f.Field.JsonName()
		jsonVal, ok := raw[jsonName]
		if f.Field.IsRequired() && !ok {
			return fmt.Errorf("struct field %s is missing", jsonName)
		}

		d, _ := json.Marshal(jsonVal)
		m := make(map[string]interface{})
		if err := json.Unmarshal(d, &m); err != nil {
			return fmt.Errorf("value of field %s is not a struct", jsonName)
		}
		raw = m
	}

	typ := value.Type()
	for i := 0; i < typ.NumField(); i++ {
		ft := typ.Field(i)
		if ft.PkgPath != "" {
			continue
		}

		if ft.Anonymous {
			if err := f.Validate(value.Field(i).Interface(), raw); err != nil {
				return err
			}
			continue
		}

		if field, ok := f.fields[ft.Name]; ok {
			if err := field.Validate(value.Field(i).Interface(), raw); err != nil {
				return err
			}
		}
	}
	return nil
}
