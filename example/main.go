package main

import (
	"encoding/base64"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zdnscloud/cement/uuid"
	"github.com/zdnscloud/gorest"
	"github.com/zdnscloud/gorest/adaptor"
	goresterr "github.com/zdnscloud/gorest/error"
	"github.com/zdnscloud/gorest/resource"
	"github.com/zdnscloud/gorest/resource/schema"
)

var (
	version = resource.APIVersion{
		Group:   "zdns.cloud.example",
		Version: "example/v1",
	}
	clusterKind = resource.DefaultKindName(Cluster{})
	nodeKind    = resource.DefaultKindName(Node{})
)

type Cluster struct {
	resource.ResourceBase `json:",inline"`
	Name                  string `json:"name,omitempty"`
}

func (c Cluster) CreateActions(name string) *resource.Action {
	if name == "encode" {
		return &resource.Action{
			Name:  "encode",
			Input: &Input{},
		}
	} else if name == "decode" {
		return &resource.Action{
			Name:  "decode",
			Input: &Input{},
		}
	} else {
		return nil
	}
}

type Node struct {
	resource.ResourceBase `json:",inline"`
	Name                  string `json:"name,omitempty"`
}

func (n Node) GetParents() []resource.ResourceKind {
	return []resource.ResourceKind{Cluster{}}
}

type Handler struct {
	objects map[string]resource.Resource
}

func newHandler() *Handler {
	return &Handler{
		objects: make(map[string]resource.Resource),
	}
}

func (h *Handler) Create(ctx *resource.Context) (interface{}, *goresterr.APIError) {
	id, _ := uuid.Gen()
	switch ctx.Resource.GetType() {
	case clusterKind:
		cluster := ctx.Resource.(*Cluster)
		for _, object := range h.objects {
			if object.GetType() == clusterKind && object.(*Cluster).Name == cluster.Name {
				return nil, goresterr.NewAPIError(goresterr.DuplicateResource, "cluster "+cluster.Name+" already exists")
			}
		}

		cluster.SetID(id)
		cluster.SetCreationTimestamp(time.Now())
		h.objects[id] = cluster
		return cluster, nil
	case nodeKind:
		if parent := ctx.Resource.GetParent(); parent != nil {
			if h.hasID(parent.GetID()) == false {
				return nil, goresterr.NewAPIError(goresterr.NotFound, "cluster "+parent.GetID()+" is non-exists")
			}
		}

		node := ctx.Resource.(*Node)
		for _, object := range h.objects {
			if object.GetType() == nodeKind && object.(*Node).Name == node.Name {
				return nil, goresterr.NewAPIError(goresterr.DuplicateResource, "node "+node.Name+" already exists")
			}
		}

		node.SetID(id)
		node.SetCreationTimestamp(time.Now())
		h.objects[id] = node
		return node, nil
	default:
		return nil, goresterr.NewAPIError(goresterr.NotFound, "no found resource type "+ctx.Resource.GetType())
	}
}

func (h *Handler) hasObject(obj resource.Resource) *goresterr.APIError {
	if parent := obj.GetParent(); parent != nil {
		if h.hasID(parent.GetID()) == false {
			return goresterr.NewAPIError(goresterr.NotFound, "cluster "+parent.GetID()+" is non-exists")
		}
	}

	if h.hasID(obj.GetID()) == false {
		return goresterr.NewAPIError(goresterr.NotFound, "no found resource "+obj.GetType()+" with id "+obj.GetID())
	}

	return nil
}

func (h *Handler) hasID(id string) bool {
	_, ok := h.objects[id]
	return ok
}

func (h *Handler) hasChild(id string) bool {
	for _, obj := range h.objects {
		if parent := obj.GetParent(); parent != nil && parent.GetID() == id {
			return true
		}
	}

	return false
}

func (h *Handler) Delete(ctx *resource.Context) *goresterr.APIError {
	if err := h.hasObject(ctx.Resource); err != nil {
		return err
	}

	if h.hasChild(ctx.Resource.GetID()) {
		return goresterr.NewAPIError(goresterr.DeleteParent, "resource has child resource")
	}

	delete(h.objects, ctx.Resource.GetID())
	return nil
}

func (h *Handler) Update(ctx *resource.Context) (interface{}, *goresterr.APIError) {
	if err := h.hasObject(ctx.Resource); err != nil {
		return nil, err
	}

	h.objects[ctx.Resource.GetID()] = ctx.Resource
	return ctx.Resource, nil
}

func (h *Handler) List(ctx *resource.Context) interface{} {
	var result []resource.Resource
	for _, object := range h.objects {
		if object.GetType() == ctx.Resource.GetType() {
			result = append(result, object)
		}
	}
	return result
}

func (h *Handler) Get(ctx *resource.Context) interface{} {
	if parent := ctx.Resource.GetParent(); parent != nil && h.hasID(parent.GetID()) == false {
		return nil
	}

	return h.objects[ctx.Resource.GetID()]
}

func (h *Handler) Action(ctx *resource.Context) (interface{}, *goresterr.APIError) {
	err := h.hasObject(ctx.Resource)
	if err != nil {
		return nil, err
	}

	r := ctx.Resource
	input, ok := r.GetAction().Input.(*Input)
	if ok == false {
		return nil, goresterr.NewAPIError(goresterr.InvalidFormat, "action input type invalid")
	}

	switch r.GetAction().Name {
	case "encode":
		return base64.StdEncoding.EncodeToString([]byte(input.Data)), nil
	case "decode":
		if data, e := base64.StdEncoding.DecodeString(input.Data); e != nil {
			err = goresterr.NewAPIError(goresterr.InvalidFormat, "decode failed: "+e.Error())
		} else {
			return string(data), nil
		}
	default:
		err = goresterr.NewAPIError(goresterr.NotFound, "not found action "+r.GetAction().Name)
	}

	return nil, err
}

type Input struct {
	Data string `json:"data,omitempty"`
}

func main() {
	router := gin.Default()
	apiServer := getApiServer()
	adaptor.RegisterHandler(router, apiServer, apiServer.Schemas.GenerateResourceRoute())
	router.Run("0.0.0.0:1234")
}

func getApiServer() *gorest.Server {
	schemas := schema.NewSchemaManager()
	handler, _ := resource.HandlerAdaptor(newHandler())
	schemas.Import(&version, Cluster{}, handler)
	schemas.Import(&version, Node{}, handler)
	server := gorest.NewAPIServer(schemas)
	return server
}
