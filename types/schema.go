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
	ResourceFields    *resourcefield.ResourceField `json:"resourceFields"`
	ResourceActions   []Action                     `json:"resourceActions,omitempty"`
	CollectionActions []Action                     `json:"collectionActions,omitempty"`

	resourceType reflect.Type `json:"-"`
	Handler      Handler      `json:"-"`
	Parents      []string     `json:"-"`
}

type Action struct {
	Name  string
	Input interface{} `json:"input,omitempty"`
}

func (s *Schema) GetType() string {
	return strings.ToLower(s.resourceType.Name())
}

func (s *Schema) resourceMethods() []string {
	return GetResourceMethods(s.Handler)
}

func (s *Schema) SupportResourceMethod(method string) bool {
	return SupportResourceMethod(s.Handler, method)
}

func (s *Schema) collectionMethods() []string {
	return GetCollectionMethods(s.Handler)
}

func (s *Schema) SupportCollectionMethod(method string) bool {
	return SupportCollectionMethod(s.Handler, method)
}

func (s *Schema) NewResource() interface{} {
	return reflect.New(s.resourceType).Interface()
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

	pluralName := util.GuessPluralName(strings.ToLower(objType.Name()))
	return &Schema{
		Version:           *version,
		resourceType:      objType,
		ResourceFields:    fields,
		Handler:           handler,
		ResourceActions:   obj.GetActions(),
		CollectionActions: obj.GetCollectionActions(),
		Parents:           obj.GetParents(),
		PluralName:        pluralName,
	}, nil
}
