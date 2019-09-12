package types

import (
	"fmt"
	"net/http"
	"path"
	"reflect"
	"strings"

	"github.com/zdnscloud/gorest/types/resourcefield"
	"github.com/zdnscloud/gorest/util"
)

type Schema struct {
	Version           *APIVersion
	ResourceFields    *resourcefield.ResourceField
	ResourceActions   []Action
	CollectionActions []Action

	Handler        Handler
	resourceType   reflect.Type
	collectionName string
	children       []*Schema
}

type Action struct {
	Name  string
	Input interface{} `json:"input,omitempty"`
}

func (s *Schema) equals(other *Schema) bool {
	return s.resourceType == other.resourceType
}

func (s *Schema) SupportResourceMethod(method string) bool {
	return SupportResourceMethod(s.Handler, method)
}

func (s *Schema) SupportCollectionMethod(method string) bool {
	return SupportCollectionMethod(s.Handler, method)
}

func (s *Schema) GetCollectionName() string {
	return s.collectionName
}

func (s *Schema) GetType() string {
	return strings.ToLower(s.resourceType.Name())
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

	return &Schema{
		Version:           version,
		resourceType:      objType,
		ResourceFields:    fields,
		Handler:           handler,
		ResourceActions:   obj.GetActions(),
		CollectionActions: obj.GetCollectionActions(),
		collectionName:    util.GuessPluralName(strings.ToLower(objType.Name())),
	}, nil
}

func (s *Schema) CreateResourceFromUrlSegments(parent Object, segments []string) (Object, *APIError) {
	segmentCount := len(segments)
	if segmentCount == 0 {
		return parent, nil
	}

	if segments[0] != s.collectionName {
		return nil, nil
	}

	obj := reflect.New(s.resourceType).Interface().(Object)
	obj.SetType(s.GetType())
	obj.SetSchema(s)
	if parent != nil {
		obj.SetParent(parent)
	}
	if segmentCount > 1 {
		obj.SetID(segments[1])
	}

	if segmentCount > 2 {
		for _, child := range s.children {
			if obj, err := child.CreateResourceFromUrlSegments(obj, segments[2:]); err != nil {
				return nil, err
			} else if obj != nil {
				return obj, nil
			}
		}
		return nil, NewAPIError(NotFound, fmt.Sprintf("%s is not a child of %s", segments[2], s.collectionName))
	} else {
		return obj, nil
	}
}

func (s *Schema) AddChild(child *Schema) error {
	for _, c := range s.children {
		if c.equals(child) {
			return fmt.Errorf("duplicate import type %s", child.resourceType.Name())
		}
	}
	s.children = append(s.children, child)
	return nil
}

func (s *Schema) GetSchema(resource ResourceType) *Schema {
	if s.resourceType == reflect.TypeOf(resource) {
		return s
	}

	for _, c := range s.children {
		if target := c.GetSchema(resource); target != nil {
			return target
		}
	}

	return nil
}

func (s *Schema) GenUrls(parents []*Schema) map[string][]string {
	urls := s.genSelfUrls(parents)
	for _, child := range s.children {
		urls = mergeUrls(urls, child.GenUrls(append(parents, s)))
	}
	return urls
}

func (s *Schema) genSelfUrls(parents []*Schema) map[string][]string {
	segments := make([]string, 0, len(parents)*2+1)
	segments = append(segments, s.Version.GetUrl())
	for _, parent := range parents {
		segments = append(segments, parent.collectionName)
		segments = append(segments, ":"+strings.ToLower(parent.resourceType.Name())+"_id")
	}
	prefix := path.Join(segments...)

	urls := make(map[string][]string)
	for _, method := range GetResourceMethods(s.Handler) {
		urls[method] = append(urls[method], path.Join(prefix, s.collectionName, ":"+strings.ToLower(s.resourceType.Name())+"_id"))
	}
	for _, method := range GetCollectionMethods(s.Handler) {
		urls[method] = append(urls[method], path.Join(prefix, s.collectionName))
	}
	return urls
}

var supportedMethods = []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPost}

func mergeUrls(a, b map[string][]string) map[string][]string {
	for _, method := range supportedMethods {
		a[method] = append(a[method], b[method]...)
	}
	return a
}
