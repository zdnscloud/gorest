package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/zdnscloud/gorest/api"
	"github.com/zdnscloud/gorest/httperror"
	"github.com/zdnscloud/gorest/types"
)

var (
	version = types.APIVersion{
		Version: "v1",
		Group:   "io.cattle.core.example",
		Path:    "/example/v1",
	}

	Schemas = types.NewSchemas()
)

type Foo struct {
	types.Resource
	ApiVersion string    `json:"apiVersion"`
	Kind       string    `json:"kind"`
	Metadata   *Metadata `json:"metadata"`
}

type Metadata struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Label     string `json:"label"`
}

func fooFromMap(data map[string]interface{}) *Foo {
	foo := &Foo{Metadata: &Metadata{}}
	if metaif, ok := data["metadata"]; ok {
		if meta, ok := metaif.(map[string]interface{}); ok {
			if name, ok := meta["name"]; ok {
				foo.Metadata.Name = name.(string)
			}

			if namespace, ok := meta["namespace"]; ok {
				foo.Metadata.Namespace = namespace.(string)
			}

			if label, ok := meta["label"]; ok {
				foo.Metadata.Label = label.(string)
			}
		}
	}
	return foo
}

func (foo *Foo) fooToMap() map[string]interface{} {
	return map[string]interface{}{
		"apiVersion": foo.ApiVersion,
		"kind":       foo.Kind,
		"metadata": map[string]interface{}{
			"name":      foo.Metadata.Name,
			"namespace": foo.Metadata.Namespace,
			"label":     foo.Metadata.Label,
		},
	}
}

type Store struct {
	resource map[string]*Foo
}

func NewStore(schema *types.Schema) *Store {
	return &Store{
		resource: make(map[string]*Foo),
	}
}

func (s *Store) Context() types.StorageContext {
	return types.DefaultStorageContext
}

func (s *Store) ByID(apiContext *types.APIContext, schema *types.Schema, id string) (map[string]interface{}, error) {
	namespaceAndName := strings.SplitN(id, ":", 2)
	if len(namespaceAndName) == 2 {
		if foo, ok := s.resource[namespaceAndName[0]+":"+schema.ID+":"+namespaceAndName[1]]; ok {
			return foo.fooToMap(), nil
		}
	}

	return nil, nil
}

func (s *Store) List(apiContext *types.APIContext, schema *types.Schema, opt *types.QueryOptions) ([]map[string]interface{}, error) {
	var resources []map[string]interface{}
	for _, foo := range s.resource {
		resources = append(resources, foo.fooToMap())
	}
	return resources, nil
}

func (s *Store) Create(apiContext *types.APIContext, schema *types.Schema, data map[string]interface{}) (map[string]interface{}, error) {
	return s.create(schema, data)
}

func (s *Store) create(schema *types.Schema, data map[string]interface{}) (map[string]interface{}, error) {
	foo := fooFromMap(data)
	foo.Kind = schema.ID
	foo.ApiVersion = schema.Version.Version
	id := foo.Metadata.Namespace + ":" + foo.Kind + ":" + foo.Metadata.Name
	if _, ok := s.resource[id]; ok {
		return nil, httperror.NewAPIError(httperror.Conflict, "duplicate resource: "+id)
	} else {
		s.resource[id] = foo
	}
	return data, nil
}

func (s *Store) Update(apiContext *types.APIContext, schema *types.Schema, data map[string]interface{}, id string) (map[string]interface{}, error) {
	namespaceAndName := strings.SplitN(id, ":", 2)
	if len(namespaceAndName) == 2 {
		id := namespaceAndName[0] + ":" + schema.ID + ":" + namespaceAndName[1]
		if _, ok := s.resource[id]; ok {
			delete(s.resource, id)
			return s.create(schema, data)
		} else {
			return nil, httperror.NewAPIError(httperror.NotFound, "non-exist resource: "+id)
		}
	}

	return nil, nil
}

func (s *Store) Delete(apiContext *types.APIContext, schema *types.Schema, id string) (map[string]interface{}, error) {
	namespaceAndName := strings.SplitN(id, ":", 2)
	if len(namespaceAndName) == 2 {
		delete(s.resource, namespaceAndName[0]+":"+schema.ID+":"+namespaceAndName[1])
	}

	return nil, nil
}

func (s *Store) Watch(apiContext *types.APIContext, schema *types.Schema, opt *types.QueryOptions) (chan map[string]interface{}, error) {
	return nil, nil
}

func main() {
	Schemas.MustImportAndCustomize(&version, Foo{}, func(schema *types.Schema) {
		schema.Store = NewStore(schema)
	})

	server := api.NewAPIServer()
	if err := server.AddSchemas(Schemas); err != nil {
		panic(err.Error())
	}

	fmt.Println("Listening on 0.0.0.0:1234")
	http.ListenAndServe("0.0.0.0:1234", server)
}
