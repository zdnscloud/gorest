package main

import (
	"fmt"
	"net/http"

	"github.com/zdnscloud/gorest/adaptor"
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
	ID   string `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
	Name string `json:"name,omitempty"`
}

func (foo *Foo) GetID() string {
	return foo.ID
}

func (foo *Foo) GetType() string {
	return foo.Type
}

func (foo *Foo) SetID(id string) {
	foo.ID = id
}

func (foo *Foo) SetType(typ string) {
	foo.Type = typ
}

type Handler struct{}

func (s *Handler) Create(obj types.Object) (interface{}, error) {
	fmt.Printf("create %s %s\n", obj.GetType(), obj.(*Foo).Name)
	return nil, nil
}

func (s *Handler) Delete(obj types.Object) error {
	fmt.Printf("delete %s %s\n", obj.GetType(), obj.GetID())
	return nil
}

func (s *Handler) BatchDelete(typ types.ObjectType) error {
	fmt.Printf("delete all %s\n", typ.GetType())
	return nil
}

func (s *Handler) Update(typ types.ObjectType, id types.ObjectID, obj types.Object) (interface{}, error) {
	fmt.Printf("update %s %s\n", typ.GetType(), id.GetID())
	return nil, nil
}

func (s *Handler) List(typ types.ObjectType) interface{} {
	fmt.Printf("get all %s\n", typ.GetType())
	return nil
}

func (s *Handler) Get(obj types.Object) interface{} {
	fmt.Printf("get %s %s\n", obj.GetType(), obj.GetID())
	return nil
}

func (s *Handler) Action(obj types.Object, action string, params map[string]interface{}) (interface{}, error) {
	fmt.Printf("do action %s with params %s for %s\n", action, params, obj.GetType())
	return nil, nil
}

func main() {
	server, err := adaptor.GetApiServer(&version, Foo{}, func(schema *types.Schema) {
		schema.Handler = &Handler{}
		schema.CollectionMethods = []string{"GET", "POST", "DELETE"}
		schema.ResourceMethods = []string{"GET", "PUT", "DELETE", "POST"}
		schema.CollectionActions = map[string]types.Action{
			"decrypt": types.Action{
				Input:  "cryptInput",
				Output: "file",
			}}
		schema.ResourceActions = map[string]types.Action{
			"encrypt": types.Action{
				Input:  "cryptInput",
				Output: "file",
			}}
	})

	if err != nil {
		panic(err.Error())
	}

	http.ListenAndServe("0.0.0.0:1234", server)
}
