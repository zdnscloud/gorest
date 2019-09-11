package resourcefield

import (
	"fmt"
	"reflect"
)

type ResourceField struct {
	fields map[string]Field
}

//validate the resource go struct
func (f *ResourceField) Validate(value interface{}) error {
	fieldValue := reflect.ValueOf(value)
	switch fieldValue.Kind() {
	case reflect.Ptr:
		if fieldValue.IsNil() {
			return nil
		}

		if fieldValue.Elem().Kind() == reflect.Struct {
			return f.validateStruct(fieldValue.Elem())
		}
	case reflect.Struct:
		return f.validateStruct(fieldValue)
	}
	return fmt.Errorf("struct field doesn't support type %v", fieldValue.Kind())
}

func (f *ResourceField) validateStruct(value reflect.Value) error {
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
			if err := field.Validate(value.Field(i).Interface()); err != nil {
				return err
			}
		}
	}
	return nil
}

//fill the default value to json string before unmarshall to resource go object
func (f *ResourceField) FillDefault(raw map[string]interface{}) {
	for _, field := range f.fields {
		field.FillDefault(raw)
	}
}

//check the json string whether the required field is specified
func (f *ResourceField) CheckRequired(raw map[string]interface{}) error {
	for _, field := range f.fields {
		if err := field.CheckRequired(raw); err != nil {
			return err
		}
	}
	return nil
}

func newResourceField(fields []Field) *ResourceField {
	fields_ := make(map[string]Field)
	for _, field := range fields {
		fields_[field.Name()] = field
	}

	return &ResourceField{
		fields: fields_,
	}
}
