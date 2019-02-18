package main

import (
	"fmt"
	"net/http"

	"github.com/zdnscloud/gorest/api"
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
	Name string `json:"name"`
	Foo  string `json:"foo"`
}

func fooFromData(data map[string]interface{}) *Foo {
	foo := &Foo{}
	if name, ok := data["name"]; ok {
		foo.Name = name.(string)
	}

	if fo, ok := data["foo"]; ok {
		foo.Foo = fo.(string)
	}

	return foo
}

type Store struct {
	resource map[string][]*Foo
}

func NewStore() *Store {
	return &Store{resource: make(map[string][]*Foo)}
}

func (s *Store) Context() types.StorageContext {
	return types.DefaultStorageContext
}

func (s *Store) ByID(apiContext *types.APIContext, schema *types.Schema, id string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for _, foo := range s.resource[schema.ID] {
		if foo.Name == id {
			result["name"] = foo.Name
			result["foo"] = foo.Foo
			break
		}
	}

	if schema.Mapper != nil {
		schema.Mapper.FromInternal(result)
	}
	return result, nil
}

func (s *Store) List(apiContext *types.APIContext, schema *types.Schema, opt *types.QueryOptions) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	for _, foos := range s.resource {
		for _, foo := range foos {
			result := map[string]interface{}{"name": foo.Name, "foo": foo.Foo}
			if schema.Mapper != nil {
				schema.Mapper.FromInternal(result)
			}
			results = append(results, result)
		}
	}

	return results, nil
}

func (s *Store) Create(apiContext *types.APIContext, schema *types.Schema, data map[string]interface{}) (map[string]interface{}, error) {
	if schema.Mapper != nil {
		schema.Mapper.ToInternal(data)
	}

	s.resource[schema.ID] = append(s.resource[schema.ID], fooFromData(data))

	if schema.Mapper != nil {
		schema.Mapper.FromInternal(data)
	}

	return data, nil
}

func (s *Store) Update(apiContext *types.APIContext, schema *types.Schema, data map[string]interface{}, id string) (map[string]interface{}, error) {
	if schema.Mapper != nil {
		schema.Mapper.ToInternal(data)
	}

	for i, foo := range s.resource[schema.ID] {
		if foo.Name == id {
			s.resource[schema.ID][i] = fooFromData(data)
			break
		}
	}

	if schema.Mapper != nil {
		schema.Mapper.FromInternal(data)
	}
	return data, nil
}

func (s *Store) Delete(apiContext *types.APIContext, schema *types.Schema, id string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for i, foo := range s.resource[schema.ID] {
		if foo.Name == id {
			s.resource[schema.ID] = append(s.resource[schema.ID][:i], s.resource[schema.ID][i+1:]...)
			break
		}
	}

	if schema.Mapper != nil {
		schema.Mapper.FromInternal(result)
	}
	return result, nil
}

func (s *Store) Watch(apiContext *types.APIContext, schema *types.Schema, opt *types.QueryOptions) (chan map[string]interface{}, error) {
	return nil, nil
}

func main() {
	Schemas.MustImportAndCustomize(&version, Foo{}, func(schema *types.Schema) {
		schema.Store = NewStore()
	})

	server := api.NewAPIServer()
	if err := server.AddSchemas(Schemas); err != nil {
		panic(err.Error())
	}

	fmt.Println("Listening on 0.0.0.0:1234")
	http.ListenAndServe("0.0.0.0:1234", server)
}
