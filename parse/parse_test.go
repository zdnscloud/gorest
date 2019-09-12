package parse

import (
	"fmt"
	"net/http"
	"testing"

	ut "github.com/zdnscloud/cement/unittest"
	"github.com/zdnscloud/gorest/types"
)

var version = types.APIVersion{
	Group:   "testing",
	Version: "v1",
}

type Cluster struct {
	types.Resource
}

type Node struct {
	types.Resource
}

func (c Node) GetParents() []types.ResourceType {
	return []types.ResourceType{Cluster{}}
}

type NameSpace struct {
	types.Resource
}

func (c NameSpace) GetParents() []types.ResourceType {
	return []types.ResourceType{Cluster{}}
}

type Deployment struct {
	types.Resource
}

func (c Deployment) GetParents() []types.ResourceType {
	return []types.ResourceType{NameSpace{}}
}

type Pod struct {
	types.Resource
}

func (c Pod) GetParents() []types.ResourceType {
	return []types.ResourceType{Deployment{}}
}

type Container struct {
	types.Resource
}

func (c Container) GetParents() []types.ResourceType {
	return []types.ResourceType{Pod{}}
}

type dumbHandler struct{}

func (h *dumbHandler) Create(ctx *types.Context, content []byte) (interface{}, *types.APIError) {
	return nil, nil
}

func (h *dumbHandler) List(ctx *types.Context) interface{} {
	return nil
}

func (h *dumbHandler) Get(ctx *types.Context) interface{} {
	return nil
}

type getHandler struct{}

func (h *getHandler) List(ctx *types.Context) interface{} {
	return nil
}

func (h *getHandler) Get(ctx *types.Context) interface{} {
	return nil
}

func TestParse(t *testing.T) {
	schemas := types.NewSchemas()
	schemas.MustImport(&version, Cluster{}, &dumbHandler{})
	schemas.MustImport(&version, Node{}, &getHandler{})
	schemas.MustImport(&version, NameSpace{}, &getHandler{})
	schemas.MustImport(&version, Deployment{}, &getHandler{})
	schemas.MustImport(&version, Pod{}, &getHandler{})
	schemas.MustImport(&version, Container{}, &getHandler{})
	req, _ := http.NewRequest("GET", "/apis/testing/v1/clusters/123321/nodes/345543", nil)
	var noErr *types.APIError
	ctx, err := Parse(nil, req, schemas)
	ut.Equal(t, err, noErr)
	ut.Equal(t, ctx.Object.GetType(), "node")
	ut.Equal(t, ctx.Object.GetID(), "345543")
	ut.Equal(t, ctx.Object.GetParent().GetID(), "123321")
	ut.Equal(t, ctx.Object.GetSchema().Version.Group, "testing")
	ut.Equal(t, ctx.Object.GetSchema().Version.Version, "v1")

	req, _ = http.NewRequest("GET", "/apis/testing/v1/clusters/clusters123/namespaces/namespaces123/deployments/deployments123/pods/pods123/containers/containers123", nil)
	ctx, err = Parse(nil, req, schemas)
	ut.Equal(t, err, noErr)
	ut.Equal(t, ctx.Object.GetType(), "container")
	ut.Equal(t, ctx.Object.GetID(), "containers123")
	ut.Equal(t, ctx.Object.GetParent().GetID(), "pods123")
	ut.Equal(t, ctx.Object.GetParent().GetType(), "pod")
	ut.Equal(t, ctx.Object.GetSchema().Version.Group, "testing")
	ut.Equal(t, ctx.Object.GetSchema().Version.Version, "v1")
	objs := types.GetAncestors(ctx.Object.GetParent())
	ut.Equal(t, len(objs), 3)
	ut.Equal(t, objs[0].GetType(), "cluster")
	ut.Equal(t, objs[1].GetType(), "namespace")
	ut.Equal(t, objs[2].GetType(), "deployment")

	req, _ = http.NewRequest("GET", "/apis/testing/v1/clusters/123321", nil)
	ctx, err = Parse(nil, req, schemas)
	ut.Equal(t, err, noErr)
	ut.Equal(t, ctx.Object.GetType(), "cluster")
	ut.Equal(t, ctx.Object.GetID(), "123321")
	ut.Equal(t, ctx.Object.GetSchema().Version.Group, "testing")
	ut.Equal(t, ctx.Object.GetSchema().Version.Version, "v1")

	req, _ = http.NewRequest("POST", "/apis/testing/v1/clusters", nil)
	var nilErr *types.APIError
	_, err = Parse(nil, req, schemas)
	ut.Equal(t, err, nilErr)

	req, _ = http.NewRequest("GET", "/apis/testing/v1/noshemas", nil)
	schemaNoFoundErr := types.NewAPIError(types.NotFound, fmt.Sprintf("no resource with collection name noshemas"))
	_, err = Parse(nil, req, schemas)
	ut.Equal(t, err, schemaNoFoundErr)

	req, _ = http.NewRequest("GET", "/apis/testing/v2/clusters", nil)
	versionNoFoundErr := types.NewAPIError(types.NotFound, fmt.Sprintf("/apis/testing/v2/clusters has unknown api version"))
	_, err = Parse(nil, req, schemas)
	ut.Equal(t, err, versionNoFoundErr)

	req, _ = http.NewRequest("GET", "/apis/testing/v1", nil)
	noSchemaErr := types.NewAPIError(types.InvalidFormat, fmt.Sprintf("no schema name in url"))
	_, err = Parse(nil, req, schemas)
	ut.Equal(t, err, noSchemaErr)
}
