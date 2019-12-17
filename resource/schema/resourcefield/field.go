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
	Validate(interface{}) error

	//work on json format string
	CheckRequired(json map[string]interface{}) error
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

func (f *leafField) CheckRequired(raw map[string]interface{}) error {
	jsonName := f.JsonName()
	if f.IsRequired() {
		val, ok := raw[jsonName]
		if ok == false {
			return fmt.Errorf("field %s is missing", jsonName)
		}

		v := reflect.ValueOf(val)
		if !v.IsValid() {
			return fmt.Errorf("field %s has invalid value", jsonName)
		}

		kind := v.Kind()
		if kind == reflect.String || kind == reflect.Map || kind == reflect.Slice {
			if v.Len() == 0 {
				return fmt.Errorf("field %s with empty slice or map", jsonName)
			}
		}
	}
	return nil
}

func (f *leafField) SetValidators(validators []validator.Validator) {
	f.validators = validators
}

func (f *leafField) Validate(val interface{}) error {
	for _, validator := range f.validators {
		if err := validator.Validate(val); err != nil {
			return err
		}
	}
	return nil
}

type sliceLeafField struct {
	Field
}

func newSliceLeafField(inner Field) *sliceLeafField {
	return &sliceLeafField{
		Field: inner,
	}
}

func (f *sliceLeafField) Validate(val interface{}) error {
	value := reflect.ValueOf(val)
	for i := 0; i < value.Len(); i++ {
		if err := f.Field.Validate(value.Index(i).Interface()); err != nil {
			return err
		}
	}
	return nil
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

func (f *sliceStructField) Validate(val interface{}) error {
	if f.inner == nil {
		return nil
	}

	value := reflect.ValueOf(val)
	for i := 0; i < value.Len(); i++ {
		if err := f.inner.Validate(value.Index(i).Interface()); err != nil {
			return err
		}
	}
	return nil
}

func (f *sliceStructField) CheckRequired(raw map[string]interface{}) error {
	if f.Field != nil {
		if err := f.Field.CheckRequired(raw); err != nil {
			return err
		}
	}

	if f.inner == nil {
		return nil
	}

	jsonName := f.Field.JsonName()
	v, ok := raw[jsonName]
	if !ok || v == nil {
		return nil
	}

	value := reflect.ValueOf(v)
	if value.Kind() != reflect.Slice {
		return fmt.Errorf("elem of field %s is not a slice but %v", jsonName, value.Kind())
	}

	for i := 0; i < value.Len(); i++ {
		d, _ := json.Marshal(value.Index(i).Interface())
		m := make(map[string]interface{})
		if err := json.Unmarshal(d, &m); err != nil {
			return fmt.Errorf("elem of field %s is not a struct", jsonName)
		}

		if err := f.inner.CheckRequired(m); err != nil {
			return err
		}
	}

	return nil
}

type mapLeafField struct {
	Field
}

func newMapLeafField(inner Field) *mapLeafField {
	return &mapLeafField{
		Field: inner,
	}
}

func (f *mapLeafField) Validate(val interface{}) error {
	iter := reflect.ValueOf(val).MapRange()
	for iter.Next() {
		if err := f.Field.Validate(iter.Value().Interface()); err != nil {
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

func (f *mapStructField) Validate(val interface{}) error {
	if f.inner == nil {
		return nil
	}

	iter := reflect.ValueOf(val).MapRange()
	for iter.Next() {
		if err := f.inner.Validate(iter.Value().Interface()); err != nil {
			return err
		}
	}
	return nil
}

func (f *mapStructField) CheckRequired(raw map[string]interface{}) error {
	if f.Field != nil {
		if err := f.Field.CheckRequired(raw); err != nil {
			return err
		}
	}

	if f.inner == nil {
		return nil
	}

	jsonName := f.Field.JsonName()
	v, ok := raw[jsonName]
	if !ok || v == nil {
		return nil
	}

	value := reflect.ValueOf(v)
	if value.Kind() != reflect.Map {
		return fmt.Errorf("field %s isn't a map but %v", jsonName, value.Kind())
	}

	iter := value.MapRange()
	for iter.Next() {
		d, _ := json.Marshal(iter.Value().Interface())
		m := make(map[string]interface{})
		if err := json.Unmarshal(d, &m); err != nil {
			return fmt.Errorf("value of field %s is not a struct", jsonName)
		}
		if err := f.inner.CheckRequired(m); err != nil {
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

func (f *structField) Validate(val interface{}) error {
	kind := reflect.TypeOf(val).Kind()
	if kind == reflect.Ptr {
		value := reflect.ValueOf(val)
		if !value.IsNil() {
			return f.Validate(value.Elem().Interface())
		} else {
			return nil
		}
	}

	if kind != reflect.Struct {
		return fmt.Errorf("struct field doesn't support type %v", kind)
	}

	value := reflect.ValueOf(val)
	typ := value.Type()
	for i := 0; i < typ.NumField(); i++ {
		ft := typ.Field(i)
		if ft.PkgPath != "" {
			continue
		}

		if ft.Anonymous {
			if err := f.Validate(value.Field(i).Interface()); err != nil {
				return err
			}
			continue
		}

		if field, ok := f.fields[ft.Name]; ok {
			if err := field.Validate(value.Field(i).Interface()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (f *structField) CheckRequired(jd map[string]interface{}) error {
	if f.Field != nil {
		if err := f.Field.CheckRequired(jd); err != nil {
			return err
		}

		jsonName := f.Field.JsonName()
		d, _ := json.Marshal(jd[jsonName])
		m := make(map[string]interface{})
		if err := json.Unmarshal(d, &m); err != nil {
			return fmt.Errorf("elem of field %s is not a struct", jsonName)
		}
		jd = m
	}

	for _, field := range f.fields {
		if err := field.CheckRequired(jd); err != nil {
			return err
		}
	}
	return nil
}
