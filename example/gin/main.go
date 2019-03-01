package main

import (
	"github.com/gin-gonic/gin"
	"github.com/zdnscloud/gorest/adaptor"
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

type Poo struct {
	ID     string       `json:"id,omitempty"`
	Type   string       `json:"type,omitempty"`
	Name   string       `json:"name,omitempty"`
	parent types.Parent `json:"-"`
}

func (poo *Poo) GetID() string {
	return poo.ID
}

func (poo *Poo) GetType() string {
	return poo.Type
}

func (poo *Poo) SetID(id string) {
	poo.ID = id
}

func (poo *Poo) SetType(typ string) {
	poo.Type = typ
}

func (poo *Poo) GetParent() types.Parent {
	return poo.parent
}

func (poo *Poo) SetParent(parent types.Parent) {
	poo.parent = parent
}

type Foo struct {
	ID     string       `json:"id,omitempty"`
	Type   string       `json:"type,omitempty"`
	Name   string       `json:"name,omitempty"`
	parent types.Parent `json:"-"`
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

func (foo *Foo) GetParent() types.Parent {
	return foo.parent
}

func (foo *Foo) SetParent(parent types.Parent) {
	foo.parent = parent
}

func getTestPoo() Poo {
	return Poo{
		ID:   "123321asd",
		Type: "poo",
		Name: "testPoo",
	}
}

func getTestFoo() Foo {
	return Foo{
		ID:   "123321asd",
		Type: "foo",
		Name: "testfoo",
	}
}

type Handler struct{}

func (s *Handler) Create(obj types.Object) (interface{}, error) {
	if obj.GetParent().Name != "" {
		return getTestFoo(), nil
	} else {
		return getTestPoo(), nil
	}
}

func (s *Handler) Delete(obj types.Object) error {
	return nil
}

func (s *Handler) BatchDelete(obj types.Object) error {
	return nil
}

func (s *Handler) Update(typ types.ObjectType, id types.ObjectID, obj types.Object) (interface{}, error) {
	if obj.GetParent().Name != "" {
		return getTestFoo(), nil
	} else {
		return getTestPoo(), nil
	}
}

func (s *Handler) List(obj types.Object) interface{} {
	if obj.GetParent().Name != "" {
		return []Foo{getTestFoo()}
	} else {
		return []Poo{getTestPoo()}
	}
}

func (s *Handler) Get(obj types.Object) interface{} {
	if obj.GetParent().Name != "" {
		return getTestFoo()
	} else {
		return getTestPoo()
	}
}

func (s *Handler) Action(obj types.Object, action string, params map[string]interface{}) (interface{}, error) {
	return params, nil
}

func main() {
	router := gin.Default()
	adaptor.RegisterHandler(router, getApiServer())
	router.Run("0.0.0.0:1234")
}

func getApiServer() *api.Server {
	server := api.NewAPIServer()
	schemas := types.NewSchemas()
	schemas.MustImportAndCustomize(&version, Poo{}, &Handler{}, func(schema *types.Schema, handler types.Handler) {
		schema.Handler = handler
		schema.CollectionMethods = []string{"GET", "POST", "DELETE"}
		schema.ResourceMethods = []string{"GET", "PUT", "DELETE", "POST"}
	})

	schemas.MustImportAndCustomize(&version, Foo{}, &Handler{}, func(schema *types.Schema, handler types.Handler) {
		schema.Parent = types.Parent{Name: "poo"}
		schema.Handler = handler
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

	if err := server.AddSchemas(schemas); err != nil {
		panic(err.Error())
	}

	return server
}
