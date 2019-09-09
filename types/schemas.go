package types

import (
	"bytes"
	"fmt"
	"path"
	"reflect"
	"strings"

	"github.com/zdnscloud/cement/slice"
	"github.com/zdnscloud/gorest/util"
)

const GroupPrefix = "/apis"

type Schema struct {
	Version           APIVersion       `json:"version"`
	PluralName        string           `json:"pluralName,omitempty"`
	ResourceMethods   []string         `json:"resourceMethods,omitempty"`
	ResourceFields    map[string]Field `json:"resourceFields"`
	ResourceActions   []Action         `json:"resourceActions,omitempty"`
	CollectionMethods []string         `json:"collectionMethods,omitempty"`
	CollectionActions []Action         `json:"collectionActions,omitempty"`

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

type Schemas struct {
	typeNames        map[reflect.Type]string
	schemasByVersion map[string]map[string]*Schema
	versions         []APIVersion
	schemas          []*Schema
}

func NewSchemas() *Schemas {
	return &Schemas{
		typeNames:        map[reflect.Type]string{},
		schemasByVersion: map[string]map[string]*Schema{},
	}
}

func (s *Schema) Complete() error {
	if s.GetType() == "" {
		return fmt.Errorf("get type from schema failed: %v", s)
	}

	if s.Version.Version == "" {
		return fmt.Errorf("version is not set on schema: %s", s.GetType())
	}

	if s.PluralName == "" {
		s.PluralName = util.GuessPluralName(s.GetType())
	}
	return nil
}

func (s *Schemas) AddSchema(schema *Schema) (*Schemas, error) {
	if err := schema.Complete(); err != nil {
		return nil, err
	}

	schemas, ok := s.schemasByVersion[schema.Version.Version]
	if !ok {
		schemas = map[string]*Schema{}
		s.schemasByVersion[schema.Version.Version] = schemas
		s.versions = append(s.versions, schema.Version)
	}

	if _, ok := schemas[schema.PluralName]; !ok {
		schemas[schema.PluralName] = schema
		s.schemas = append(s.schemas, schema)
	}

	return s, nil
}

func (s *Schemas) Versions() []APIVersion {
	return s.versions
}

func (s *Schemas) Schemas() []*Schema {
	return s.schemas
}

func (s *Schemas) Schema(version *APIVersion, name string) *Schema {
	schemas, ok := s.schemasByVersion[version.Version]
	if !ok {
		return nil
	}

	schema := schemas[name]
	if schema != nil {
		return schema
	}

	for _, check := range schemas {
		if strings.EqualFold(check.GetType(), name) || strings.EqualFold(check.PluralName, name) {
			return check
		}
	}

	return nil
}

func (s *Schemas) UrlMethods() map[string][]string {
	urlMethods := make(map[string][]string)
	for _, schema := range s.schemas {
		urls := map[int][]string{0: []string{schema.GetType()}}
		childrenUrls := urls
		children := urls[0]
		for parents := schema.Parents; len(parents) != 0; {
			var grandparents []string
			index := 0
			for i, child := range children {
				childSchema := s.Schema(&schema.Version, util.GuessPluralName(child))
				if childSchema == nil {
					panic(fmt.Sprintf("no found schema %s", child))
				}

				childUrl := childrenUrls[i]
				for _, parent := range childSchema.Parents {
					if parentSchema := s.Schema(&schema.Version, util.GuessPluralName(parent)); parentSchema != nil {
						urls[index] = append(childUrl, parent)
						grandparents = append(grandparents, parentSchema.Parents...)
						index += 1
					} else {
						panic(fmt.Sprintf("no found schema %s", parent))
					}
				}
			}

			children = parents
			parents = grandparents
			childrenUrls = urls
		}

		for _, parents := range urls {
			buffer := bytes.Buffer{}
			for i := len(parents) - 1; i > 0; i-- {
				buffer.WriteString("/")
				buffer.WriteString(util.GuessPluralName(parents[i]))
				buffer.WriteString("/:")
				buffer.WriteString(parents[i])
				buffer.WriteString("_id")
			}

			parentUrl := buffer.String()
			url := path.Join(schema.Version.GetVersionURL(), parentUrl, schema.PluralName)
			if len(schema.CollectionMethods) != 0 {
				urlMethods[url] = schema.CollectionMethods
			}

			if len(schema.ResourceMethods) != 0 {
				urlMethods[path.Join(url, ":"+schema.GetType()+"_id")] = schema.ResourceMethods
			}
		}
	}

	return urlMethods
}

func (s *Schemas) GetChildren(parent string) []string {
	if parent == "" {
		return nil
	}

	var children []string
	for _, schema := range s.schemas {
		if slice.SliceIndex(schema.Parents, parent) != -1 {
			children = append(children, schema.PluralName)
		}
	}

	return children
}
