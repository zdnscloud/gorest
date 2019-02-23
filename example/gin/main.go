package main

import (
	"fmt"
	"path"

	"github.com/zdnscloud/gorest/api"
	"github.com/zdnscloud/gorest/types"

	"github.com/gin-gonic/gin"
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
	router := gin.Default()
	registerHandler(router, getApiServer())
	router.Run("0.0.0.0:1234")
}

func registerHandler(router gin.IRoutes, server *api.Server) {
	handlerFunc := gin.WrapH(server)
	for _, schema := range server.Schemas.Schemas() {
		url := path.Join("/"+schema.Version.Group, schema.Version.Path, schema.ID)
		router.POST(url, handlerFunc)
		router.POST(path.Join(url, ":id"), handlerFunc)
		router.DELETE(path.Join(url, ":id"), handlerFunc)
		router.DELETE(url, handlerFunc)
		router.PUT(path.Join(url, ":id"), handlerFunc)
		router.GET(url, handlerFunc)
		router.GET(path.Join(url, ":id"), handlerFunc)
	}
}

func getApiServer() *api.Server {
	schemas := types.NewSchemas().MustImportAndCustomize(&version, Foo{}, func(schema *types.Schema) {
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

	server := api.NewAPIServer()
	if err := server.AddSchemas(schemas); err != nil {
		panic(err.Error())
	}

	return server
}
