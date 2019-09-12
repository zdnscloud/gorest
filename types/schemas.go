package types

import (
	"fmt"
)

const GroupPrefix = "/apis"

type Schemas struct {
	schemas []*VersionedSchemas
}

func NewSchemas() *Schemas {
	return &Schemas{}
}

func (s *Schemas) MustImport(v *APIVersion, obj ResourceType, objHandler interface{}) *Schemas {
	vs := s.getVersionedSchemas(v)
	if vs == nil {
		vs = NewVersionedSchemas(v)
		s.schemas = append(s.schemas, vs)
	}

	if err := vs.ImportResource(obj, objHandler); err != nil {
		panic(err)
	}

	return s
}

func (s *Schemas) getVersionedSchemas(v *APIVersion) *VersionedSchemas {
	for _, vs := range s.schemas {
		if vs.VersionEquals(v) {
			return vs
		}
	}
	return nil
}

func (s *Schemas) CreateResourceFromUrl(url string) (Object, *APIError) {
	for _, vs := range s.schemas {
		if obj, err := vs.CreateResourceFromUrl(url); err != nil {
			return nil, err
		} else if obj != nil {
			return obj, err
		}
	}
	return nil, NewAPIError(NotFound, fmt.Sprintf("%s has unknown api version", url))
}

func (s *Schemas) GetSchema(v *APIVersion, resource ResourceType) *Schema {
	if vs := s.getVersionedSchemas(v); vs != nil {
		return vs.GetSchema(resource)
	}
	return nil
}

func (s *Schemas) UrlMethods() map[string][]string {
	urls := make(map[string][]string)
	for _, vs := range s.schemas {
		urls = mergeUrls(urls, vs.GenUrls())
	}
	return urls
}

func (s *Schemas) GetChildren(parent string) []string {
	return nil

	/*
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
	*/
}
