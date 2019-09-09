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

func (c Node) GetParents() []string {
	return []string{types.GetResourceType(Cluster{})}
}

type Namespace struct {
	types.Resource
}

func (c Namespace) GetParents() []string {
	return []string{types.GetResourceType(Cluster{})}
}

type Deployment struct {
	types.Resource
}

func (c Deployment) GetParents() []string {
	return []string{types.GetResourceType(Namespace{})}
}

type Pod struct {
	types.Resource
}

func (c Pod) GetParents() []string {
	return []string{types.GetResourceType(Deployment{})}
}

type Container struct {
	types.Resource
}

func (c Container) GetParents() []string {
	return []string{types.GetResourceType(Pod{})}
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
	schemas.MustImport(&version, Namespace{}, &getHandler{})
	schemas.MustImport(&version, Deployment{}, &getHandler{})
	schemas.MustImport(&version, Pod{}, &getHandler{})
	schemas.MustImport(&version, Container{}, &getHandler{})
	req, _ := http.NewRequest("GET", "/apis/testing/v1/clusters/123321/nodes/345543", nil)
	var noErr *types.APIError
	ctx, err := Parse(nil, req, schemas)
	ut.Equal(t, err, noErr)
	ut.Equal(t, ctx.Object.GetType(), "node")
	ut.Equal(t, ctx.Object.GetID(), "345543")
	ut.Equal(t, ctx.Object.GetSchema().Parents, []string{"cluster"})
	ut.Equal(t, ctx.Object.GetParent().GetID(), "123321")
	ut.Equal(t, ctx.Object.GetSchema().Version.Group, "testing")
	ut.Equal(t, ctx.Object.GetSchema().Version.Version, "v1")

	req, _ = http.NewRequest("GET", "/apis/testing/v1/clusters/clusters123/namespaces/namespaces123/deployments/deployments123/pods/pods123/containers/containers123", nil)
	ctx, err = Parse(nil, req, schemas)
	ut.Equal(t, err, noErr)
	ut.Equal(t, ctx.Object.GetType(), "container")
	ut.Equal(t, ctx.Object.GetID(), "containers123")
	ut.Equal(t, ctx.Object.GetSchema().Parents, []string{"pod"})
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
	ut.Equal(t, len(ctx.Object.GetSchema().Parents), 0)
	ut.Equal(t, ctx.Object.GetSchema().Version.Group, "testing")
	ut.Equal(t, ctx.Object.GetSchema().Version.Version, "v1")

	req, _ = http.NewRequest("DELETE", "/apis/testing/v1/clusters/123321", nil)
	deleteNotAllowedErr := types.NewAPIError(types.MethodNotAllowed, fmt.Sprintf("Method %s not supported", req.Method))
	_, err = Parse(nil, req, schemas)
	ut.Equal(t, err, deleteNotAllowedErr)

	req, _ = http.NewRequest("POST", "/apis/testing/v1/clusters", nil)
	var nilErr *types.APIError
	_, err = Parse(nil, req, schemas)
	ut.Equal(t, err, nilErr)

	req, _ = http.NewRequest("POST", "/apis/testing/v1/clusters/123123/nodes", nil)
	postNotAllowedErr := types.NewAPIError(types.MethodNotAllowed, fmt.Sprintf("Method %s not supported", req.Method))
	_, err = Parse(nil, req, schemas)
	ut.Equal(t, err, postNotAllowedErr)

	req, _ = http.NewRequest("GET", "/apis/testing/v1/noshemas", nil)
	schemaNoFoundErr := types.NewAPIError(types.NotFound, fmt.Sprintf("no found schema for noshemas"))
	_, err = Parse(nil, req, schemas)
	ut.Equal(t, err, schemaNoFoundErr)

	req, _ = http.NewRequest("GET", "/apis/testing/v2/clusters", nil)
	versionNoFoundErr := types.NewAPIError(types.NotFound, fmt.Sprintf("no found version with /apis/testing/v2/clusters"))
	_, err = Parse(nil, req, schemas)
	ut.Equal(t, err, versionNoFoundErr)

	req, _ = http.NewRequest("GET", "/apis/testing/v1", nil)
	noSchemaErr := types.NewAPIError(types.InvalidFormat, fmt.Sprintf("no schema name in url /apis/testing/v1"))
	_, err = Parse(nil, req, schemas)
	ut.Equal(t, err, noSchemaErr)
}
