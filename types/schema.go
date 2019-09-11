package types

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/zdnscloud/gorest/types/resourcefield"
	"github.com/zdnscloud/gorest/util"
)

type Schema struct {
	Version           APIVersion                   `json:"version"`
	PluralName        string                       `json:"pluralName,omitempty"`
	ResourceMethods   []string                     `json:"resourceMethods,omitempty"`
	ResourceFields    *resourcefield.ResourceField `json:"resourceFields"`
	ResourceActions   []Action                     `json:"resourceActions,omitempty"`
	CollectionMethods []string                     `json:"collectionMethods,omitempty"`
	CollectionActions []Action                     `json:"collectionActions,omitempty"`

	StructVal reflect.Value `json:"-"`
	Handler   Handler       `json:"-"`
	Parents   []string      `json:"-"`
}

type Action struct {
	Name  string
	Input interface{} `json:"input,omitempty"`
}

func (s *Schema) GetType() string {
	return strings.ToLower(s.StructVal.Type().Name())
}

func newSchema(version *APIVersion, obj ResourceType, objHandler interface{}) (*Schema, error) {
	if reflect.ValueOf(obj).Kind() == reflect.Ptr {
		return nil, fmt.Errorf("obj cannot be a pointer")
	}

	objType := reflect.TypeOf(obj)
	if _, ok := reflect.New(objType).Interface().(Object); ok == false {
		return nil, fmt.Errorf("resource type doesn't implement object interface")
	}

	fields, err := resourcefield.NewBuilder().Build(objType)
	if err != nil {
		return nil, err
	}

	handler, err := NewHandler(objHandler)
	if err != nil {
		return nil, err
	}

	objectValue := reflect.New(objType).Elem()
	pluralName := util.GuessPluralName(strings.ToLower(objectValue.Type().Name()))

	return &Schema{
		Version:           *version,
		StructVal:         objectValue,
		ResourceFields:    fields,
		Handler:           handler,
		ResourceMethods:   GetResourceMethods(handler),
		CollectionMethods: GetCollectionMethods(handler),
		ResourceActions:   obj.GetActions(),
		CollectionActions: obj.GetCollectionActions(),
		Parents:           obj.GetParents(),
		PluralName:        pluralName,
	}, nil
}
