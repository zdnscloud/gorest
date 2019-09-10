package field

import (
	"encoding/json"
	"fmt"
	"reflect"
)

var _ Field = &leafField{}
var _ Field = &compositeField{}

type leafField struct {
	name       string
	jsonName   string
	required   bool
	defaultVal interface{}
	validators []Validator
}

func newLeafField(name, jsonName string) *leafField {
	return &leafField{
		name:       name,
		jsonName:   jsonName,
		required:   false,
		defaultVal: nil,
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
		if _, ok := raw[jsonName]; ok == false {
			return fmt.Errorf("field %s is missing", jsonName)
		}
	}
	return nil
}

func (f *leafField) FillDefault(raw map[string]interface{}) {
	defaultVal := f.DefaultValue()
	jsonName := f.JsonName()
	if defaultVal != nil {
		if _, ok := raw[jsonName]; ok == false {
			raw[jsonName] = defaultVal
		}
	}
}

func (f *leafField) DefaultValue() interface{} {
	return f.defaultVal
}

func (f *leafField) SetDefault(val interface{}) {
	f.defaultVal = val
}

func (f *leafField) SetValidators(validators []Validator) {
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

type StructField struct {
	fields map[string]Field
}

func (f *StructField) SetFields(fields []Field) {
	m := make(map[string]Field)
	for _, field := range fields {
		m[field.Name()] = field
	}
	f.fields = m
}

func (f *StructField) DefaultValue() interface{} {
	def := make(map[string]interface{})
	for _, field := range f.fields {
		def[field.JsonName()] = field.DefaultValue()
	}
	return def
}

func (f *StructField) Validate(val interface{}) error {
	value := reflect.ValueOf(val)
	switch value.Kind() {
	case reflect.Ptr:
		if value.Elem().Kind() == reflect.Struct {
			return f.validateStruct(value.Elem())
		}
	case reflect.Struct:
		return f.validateStruct(value)
	}
	return fmt.Errorf("struct field doesn't support type %v", value.Kind())
}

func (f *StructField) validateStruct(value reflect.Value) error {
	st := value.Type()
	for i := 0; i < st.NumField(); i++ {
		sf := st.Field(i)
		if sf.PkgPath != "" {
			continue
		}

		if sf.Anonymous {
			if err := f.validateStruct(value.Field(i)); err != nil {
				return err
			}
			continue
		}

		if field, ok := f.fields[sf.Name]; ok {
			fieldValue := value.Field(i)
			switch sf.Type.Kind() {
			case reflect.Map:
				iter := fieldValue.MapRange()
				for iter.Next() {
					if err := field.Validate(iter.Value().Interface()); err != nil {
						return err
					}
				}
			case reflect.Slice:
				for i := 0; i < fieldValue.Cap(); i++ {
					if err := field.Validate(fieldValue.Index(i).Interface()); err != nil {
						return err
					}
				}
			default:
				if err := field.Validate(fieldValue.Interface()); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (f *StructField) FillDefault(raw map[string]interface{}) {
	for _, field := range f.fields {
		field.FillDefault(raw)
	}
}

func (f *StructField) CheckRequired(raw map[string]interface{}) error {
	for _, field := range f.fields {
		if err := field.CheckRequired(raw); err != nil {
			return err
		}
	}
	return nil
}

//filed like map with struct value type
//struct slice
//struct
type OwnerKind string

const (
	OwnerNone      OwnerKind = "none"
	OwnerSlice     OwnerKind = "slice"
	OwnerIntMap    OwnerKind = "int_map"
	OwnerStringMap OwnerKind = "string_map"
)

type compositeField struct {
	name      string
	jsonName  string
	required  bool
	ownerKind OwnerKind
	field     *StructField
}

func newCompositeField(name, jsonName string, inner *StructField) *compositeField {
	return &compositeField{
		name:      name,
		jsonName:  jsonName,
		required:  false,
		ownerKind: OwnerNone,
		field:     inner,
	}
}

func (f *compositeField) Name() string {
	return f.name
}

func (f *compositeField) JsonName() string {
	return f.jsonName
}

func (f *compositeField) IsRequired() bool {
	return f.required
}

func (f *compositeField) SetRequired(required bool) {
	f.required = required
}

func (f *compositeField) DefaultValue() interface{} {
	return nil
}

func (f *compositeField) SetOwner(kind OwnerKind) {
	f.ownerKind = kind
}

func (f *compositeField) SetDefault(_ interface{}) {
}

func (f *compositeField) Validate(val interface{}) error {
	return f.field.Validate(val)
}

func (f *compositeField) CheckRequired(json map[string]interface{}) error {
	jsonName := f.JsonName()
	if f.IsRequired() {
		if _, ok := json[jsonName]; ok == false {
			return fmt.Errorf("field %s is missing", jsonName)
		}
	}
	return nil
}

func (f *compositeField) FillDefault(raw map[string]interface{}) {
	jsonName := f.JsonName()
	if val, ok := raw[jsonName]; ok {
		nestRaw, _ := json.Marshal(val)
		switch f.ownerKind {
		case OwnerIntMap:
			nest := make(map[int64]json.RawMessage)
			err := json.Unmarshal(nestRaw, &nest)
			if err != nil {
				return
			}
			for k, v := range nest {
				structRaw := make(map[string]interface{})
				err := json.Unmarshal(v, &structRaw)
				if err != nil {
					return
				}
				f.field.FillDefault(structRaw)
				bytes, err := json.Marshal(structRaw)
				if err == nil {
					nest[k] = bytes
				}
			}
			raw[jsonName] = nest
		case OwnerStringMap:
			nest := make(map[string]json.RawMessage)
			err := json.Unmarshal(nestRaw, &nest)
			if err != nil {
				return
			}
			for k, v := range nest {
				structRaw := make(map[string]interface{})
				err := json.Unmarshal(v, &structRaw)
				if err != nil {
					return
				}
				f.field.FillDefault(structRaw)
				bytes, err := json.Marshal(structRaw)
				if err == nil {
					nest[k] = bytes
				}
			}
			raw[jsonName] = nest
		case OwnerSlice:
			nest := make([]json.RawMessage, 0)
			err := json.Unmarshal(nestRaw, &nest)
			if err != nil {
				return
			}
			for i, v := range nest {
				structRaw := make(map[string]interface{})
				err := json.Unmarshal(v, &structRaw)
				if err != nil {
					return
				}
				f.field.FillDefault(structRaw)
				bytes, err := json.Marshal(structRaw)
				if err == nil {
					nest[i] = bytes
				}
			}
			raw[jsonName] = nest
		case OwnerNone:
			structRaw := make(map[string]interface{})
			err := json.Unmarshal(nestRaw, &structRaw)
			if err != nil {
				return
			}
			f.field.FillDefault(structRaw)
			raw[jsonName] = structRaw
		}
	}
}
