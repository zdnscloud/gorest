package main

import (
	"fmt"
	"net/http"

	"github.com/zdnscloud/gorest/api"
	"github.com/zdnscloud/gorest/parse"
	"github.com/zdnscloud/gorest/types"
)

var (
	version = types.APIVersion{
		Version: "v1",
		Group:   "io.cattle.core.example",
		Path:    "/example/v1",
	}

	Schemas = types.NewSchemas()

	store = newStore()
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

	if kind, ok := data["kind"]; ok {
		foo.Kind = kind.(string)
	}

	if version, ok := data["apiVersion"]; ok {
		foo.ApiVersion = version.(string)
	}

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

func newStore() *Store {
	return &Store{resource: make(map[string]*Foo)}
}

func addFooToMem(data map[string]interface{}) error {
	foo := fooFromMap(data)
	id := foo.Metadata.Namespace + ":" + foo.Metadata.Name
	if _, ok := store.resource[id]; ok {
		return fmt.Errorf("duplicate foo: %s\n", id)
	} else {
		store.resource[id] = foo
	}
	return nil
}

func deleteFooFromMem(id string) error {
	delete(store.resource, id)
	return nil
}

func updateFooToMem(id string, data map[string]interface{}) error {
	if _, ok := store.resource[id]; ok {
		delete(store.resource, id)
		return addFooToMem(data)
	} else {
		return fmt.Errorf("no such foo: %s\n", id)
	}
}

func getAllFoosFromMem() []map[string]interface{} {
	var result []map[string]interface{}
	for _, foo := range store.resource {
		result = append(result, foo.fooToMap())
	}

	return result
}

func getFooFromMem(id string) map[string]interface{} {
	if foo, ok := store.resource[id]; ok {
		return foo.fooToMap()
	}

	return nil
}

func CreateHandler(apiContext *types.APIContext, next types.RequestHandler) error {
	data, err := parse.ReadBody(apiContext.Request)
	if err != nil {
		return err
	}

	data["kind"] = apiContext.Schema.ID
	data["apiVersion"] = apiContext.Schema.Version.Version

	if err := addFooToMem(data); err != nil {
		return err
	}

	apiContext.WriteResponse(http.StatusCreated, data)
	return nil
}

func DeleteHandler(apiContext *types.APIContext, next types.RequestHandler) error {
	if err := deleteFooFromMem(apiContext.ID); err != nil {
		return err
	}

	apiContext.WriteResponse(http.StatusCreated, nil)
	return nil
}

func UpdateHandler(apiContext *types.APIContext, next types.RequestHandler) error {
	data, err := parse.ReadBody(apiContext.Request)
	if err != nil {
		return err
	}

	data["kind"] = apiContext.Schema.ID
	data["apiVersion"] = apiContext.Schema.Version.Version

	if err := updateFooToMem(apiContext.ID, data); err != nil {
		return err
	}

	apiContext.WriteResponse(http.StatusCreated, data)
	return nil
}

func ListHandler(apiContext *types.APIContext, next types.RequestHandler) error {
	var result interface{}
	if apiContext.ID == "" {
		result = getAllFoosFromMem()
	} else {
		result = getFooFromMem(apiContext.ID)
	}

	apiContext.WriteResponse(http.StatusCreated, result)
	return nil
}

func main() {
	Schemas.MustImportAndCustomize(&version, Foo{}, func(schema *types.Schema) {
		schema.CreateHandler = CreateHandler
		schema.DeleteHandler = DeleteHandler
		schema.UpdateHandler = UpdateHandler
		schema.ListHandler = ListHandler
	})

	server := api.NewAPIServer()
	if err := server.AddSchemas(Schemas); err != nil {
		panic(err.Error())
	}

	fmt.Println("Listening on 0.0.0.0:1234")
	http.ListenAndServe("0.0.0.0:1234", server)
}
