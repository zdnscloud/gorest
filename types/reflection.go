package types

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/zdnscloud/gorest/types/field"
)

func (s *Schemas) MustImport(version *APIVersion, obj ResourceType, objHandler interface{}) *Schemas {
	if reflect.ValueOf(obj).Kind() == reflect.Ptr {
		panic(fmt.Errorf("obj cannot be a pointer"))
	}

	objType := reflect.TypeOf(obj)
	if _, ok := reflect.New(objType).Interface().(Object); ok == false {
		panic("resource type doesn't implement object interface")
	}

	schema, err := s.importType(version, objType)
	if err != nil {
		panic(err)
	}

	handler, err := NewHandler(objHandler)
	if err != nil {
		panic(err)
	}

	schema.Handler = handler
	schema.ResourceMethods = GetResourceMethods(handler)
	schema.CollectionMethods = GetCollectionMethods(handler)
	schema.ResourceActions = obj.GetActions()
	schema.CollectionActions = obj.GetCollectionActions()
	schema.Parents = obj.GetParents()

	return s
}

func (s *Schemas) importType(version *APIVersion, t reflect.Type) (*Schema, error) {
	typeName := s.getTypeName(t)
	existing := s.Schema(version, typeName)
	if existing != nil {
		return existing, nil
	}

	schema, err := s.newSchemaFromType(version, t)
	if err != nil {
		return nil, err
	}

	if _, err := s.AddSchema(schema); err != nil {
		return nil, err
	}

	return s.Schema(&schema.Version, schema.GetType()), nil
}

func (s *Schemas) newSchemaFromType(version *APIVersion, t reflect.Type) (*Schema, error) {
	fields, err := field.NewBuilder().Build(t)
	if err != nil {
		return nil, err
	}

	return &Schema{
		Version:        *version,
		StructVal:      reflect.New(t).Elem(),
		ResourceFields: fields,
	}, nil
}

func (s *Schemas) getTypeName(t reflect.Type) string {
	if name, ok := s.typeNames[t]; ok {
		return name
	}

	name := strings.ToLower(t.Name())
	s.typeNames[t] = name
	return name
}
