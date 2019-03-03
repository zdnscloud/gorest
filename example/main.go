package main

import (
	"github.com/gin-gonic/gin"
	"github.com/zdnscloud/cement/uuid"
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
	types.Resource `json:",inline"`
	Name           string `json:"name,omitempty"`
}

type Foo struct {
	types.Resource `json:",inline"`
	Name           string `json:"name,omitempty"`
}

type Handler struct {
	objects map[string]types.Object
}

func newHandler() *Handler {
	return &Handler{
		objects: make(map[string]types.Object),
	}
}

func (h *Handler) Create(obj types.Object) (interface{}, *types.APIError) {
	id, _ := uuid.Gen()
	switch obj.GetType() {
	case "poo":
		poo := obj.(*Poo)
		for _, object := range h.objects {
			if object.GetType() == "poo" && object.(*Poo).Name == poo.Name {
				return nil, types.NewAPIError(types.DuplicateResource, "poo "+poo.Name+" already exists")
			}
		}

		poo.SetID(id)
		h.objects[id] = poo
		return poo, nil
	case "foo":
		if h.hasID(obj.GetParent().ID) == false {
			return nil, types.NewAPIError(types.NotFound, "poo "+obj.GetParent().ID+" is non-exists")
		}

		foo := obj.(*Foo)
		for _, object := range h.objects {
			if object.GetType() == "foo" && object.(*Foo).Name == foo.Name {
				return nil, types.NewAPIError(types.DuplicateResource, "poo "+foo.Name+" already exists")
			}
		}

		foo.SetID(id)
		h.objects[id] = foo
		return foo, nil
	default:
		return nil, types.NewAPIError(types.NotFound, "no found resource type "+obj.GetType())
	}
}

func (h *Handler) hasObject(obj types.Object) *types.APIError {
	parentID := obj.GetParent().ID
	if parentID != "" && h.hasID(parentID) == false {
		return types.NewAPIError(types.NotFound, "no found resource "+obj.GetParent().Name+" with id "+parentID)
	}

	if h.hasID(obj.GetID()) == false {
		return types.NewAPIError(types.NotFound, "no found resource "+obj.GetType()+" with id "+obj.GetID())
	}

	return nil
}

func (h *Handler) hasID(id string) bool {
	_, ok := h.objects[id]
	return ok
}

func (h *Handler) hasChild(id string) bool {
	for _, obj := range h.objects {
		if obj.GetParent().ID == id {
			return true
		}
	}

	return false
}

func (h *Handler) Delete(obj types.Object) *types.APIError {
	if err := h.hasObject(obj); err != nil {
		return err
	}

	if h.hasChild(obj.GetID()) {
		return types.NewAPIError(types.DeleteParent, "resource has child resource")
	}

	delete(h.objects, obj.GetID())
	return nil
}

func (h *Handler) Update(obj types.Object) (interface{}, *types.APIError) {
	if err := h.hasObject(obj); err != nil {
		return nil, err
	}

	h.objects[obj.GetID()] = obj
	return obj, nil
}

func (h *Handler) List(obj types.Object) interface{} {
	var result []types.Object
	for _, object := range h.objects {
		if object.GetType() == obj.GetType() {
			result = append(result, object)
		}
	}
	return result
}

func (h *Handler) Get(obj types.Object) interface{} {
	parentID := obj.GetParent().ID
	if parentID != "" && h.hasID(parentID) == false {
		return nil
	}

	return h.objects[obj.GetID()]
}

func (h *Handler) Action(obj types.Object, action string, params map[string]interface{}) (interface{}, *types.APIError) {
	if err := h.hasObject(obj); err != nil {
		return nil, err
	}

	return params, nil
}

func main() {
	router := gin.Default()
	apiServer := getApiServer()
	adaptor.RegisterHandler(router, gin.WrapH(apiServer), apiServer.Schemas.UrlMethods())
	router.Run("0.0.0.0:1234")
}

func getApiServer() *api.Server {
	server := api.NewAPIServer()
	schemas := types.NewSchemas()
	handler := newHandler()
	schemas.MustImportAndCustomize(&version, Poo{}, handler, func(schema *types.Schema, handler types.Handler) {
		schema.Handler = handler
		schema.CollectionMethods = []string{"GET", "POST"}
		schema.ResourceMethods = []string{"GET", "PUT", "DELETE", "POST"}
	})

	schemas.MustImportAndCustomize(&version, Foo{}, handler, func(schema *types.Schema, handler types.Handler) {
		schema.Parent = "poo"
		schema.Handler = handler
		schema.CollectionMethods = []string{"GET", "POST"}
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
