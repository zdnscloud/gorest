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
		Group:   "zdns.cloud.example",
		Path:    "/example/v1",
	}
)

type Foo struct {
	id   string `json:"id,omitempty"`
	typ  string `json:"type,omitempty"`
	Name string `json:"name,omitempty"`
}

func (foo *Foo) ID() string {
	return foo.id
}

func (foo *Foo) Type() string {
	return foo.typ
}

func (foo *Foo) SetID(id string) {
	foo.id = id
}

func (foo *Foo) SetType(typ string) {
	foo.typ = typ
}

type Handler struct{}

func (s *Handler) Create(obj types.Object) (interface{}, error) {
	fmt.Printf("create %s %s\n", obj.Type(), obj.(*Foo).Name)
	return nil, nil
}

func (s *Handler) Delete(obj types.Object) error {
	fmt.Printf("delete %s %s\n", obj.Type(), obj.ID())
	return nil
}

func (s *Handler) BatchDelete(typ types.ObjectType) error {
	fmt.Printf("delete all %s\n", typ.Type())
	return nil
}

func (s *Handler) Update(typ types.ObjectType, id types.ObjectID, obj types.Object) (interface{}, error) {
	fmt.Printf("update %s %s\n", typ.Type(), id.ID())
	return nil, nil
}

func (s *Handler) List(typ types.ObjectType) interface{} {
	fmt.Printf("get all %s\n", typ.Type())
	return nil
}

func (s *Handler) Get(obj types.Object) interface{} {
	fmt.Printf("get %s %s\n", obj.Type(), obj.ID())
	return nil
}

func (s *Handler) Action(obj types.Object, action string, params map[string]interface{}) (interface{}, error) {
	fmt.Printf("do action %s with params %s for %s\n", action, params, obj.Type())
	return nil, nil
}

func main() {
	schemas := types.NewSchemas().MustImportAndCustomize(&version, Foo{}, func(schema *types.Schema) {
		schema.Handler = &Handler{}
		schema.CollectionMethods = []string{"GET", "POST"}
		schema.ResourceMethods = []string{"GET", "PUT", "DELETE"}
	})

	server := api.NewAPIServer()
	if err := server.AddSchemas(schemas); err != nil {
		panic(err.Error())
	}

	http.ListenAndServe("0.0.0.0:1234", server)
}
