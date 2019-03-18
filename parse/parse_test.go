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

type Namespace struct {
	types.Resource
}

type Deployment struct {
	types.Resource
}

type Pod struct {
	types.Resource
}

type Container struct {
	types.Resource
}

func TestParse(t *testing.T) {
	schemas := types.NewSchemas()
	schemas.MustImportAndCustomize(&version, Cluster{}, nil, func(schema *types.Schema, handler types.Handler) {
		schema.CollectionMethods = []string{"GET", "POST"}
		schema.ResourceMethods = []string{"GET"}
	})
	schemas.MustImportAndCustomize(&version, Node{}, nil, func(schema *types.Schema, handler types.Handler) {
		schema.Parent = types.GetResourceType(Cluster{})
		schema.CollectionMethods = []string{"GET"}
		schema.ResourceMethods = []string{"GET"}
	})

	schemas.MustImportAndCustomize(&version, Namespace{}, nil, func(schema *types.Schema, handler types.Handler) {
		schema.Parent = types.GetResourceType(Cluster{})
		schema.CollectionMethods = []string{"GET"}
		schema.ResourceMethods = []string{"GET"}
	})

	schemas.MustImportAndCustomize(&version, Deployment{}, nil, func(schema *types.Schema, handler types.Handler) {
		schema.Parent = types.GetResourceType(Namespace{})
		schema.CollectionMethods = []string{"GET"}
		schema.ResourceMethods = []string{"GET"}
	})

	schemas.MustImportAndCustomize(&version, Pod{}, nil, func(schema *types.Schema, handler types.Handler) {
		schema.Parent = types.GetResourceType(Deployment{})
		schema.CollectionMethods = []string{"GET"}
		schema.ResourceMethods = []string{"GET"}
	})

	schemas.MustImportAndCustomize(&version, Container{}, nil, func(schema *types.Schema, handler types.Handler) {
		schema.Parent = types.GetResourceType(Pod{})
		schema.CollectionMethods = []string{"GET"}
		schema.ResourceMethods = []string{"GET"}
	})

	req, _ := http.NewRequest("GET", "/apis/testing/v1/clusters/123321/nodes/345543", nil)
	var noErr *types.APIError
	ctx, err := Parse(nil, req, schemas)
	ut.Equal(t, err, noErr)
	ut.Equal(t, ctx.Schema.ID, "node")
	ut.Equal(t, ctx.ID, "345543")
	ut.Equal(t, ctx.Schema.Parent, "cluster")
	ut.Equal(t, ctx.Parent.GetID(), "123321")
	ut.Equal(t, ctx.Version.Group, "testing")
	ut.Equal(t, ctx.Version.Version, "v1")

	req, _ = http.NewRequest("GET", "/apis/testing/v1/clusters/clusters123/namespaces/namespaces123/deployments/deployments123/pods/pods123/containers/containers123", nil)
	ctx, err = Parse(nil, req, schemas)
	ut.Equal(t, err, noErr)
	ut.Equal(t, ctx.Schema.ID, "container")
	ut.Equal(t, ctx.ID, "containers123")
	ut.Equal(t, ctx.Schema.Parent, "pod")
	ut.Equal(t, ctx.Parent.GetID(), "pods123")
	ut.Equal(t, ctx.Parent.GetType(), "pod")
	ut.Equal(t, ctx.Version.Group, "testing")
	ut.Equal(t, ctx.Version.Version, "v1")
	objs := types.GetAncestors(ctx.Parent)
	ut.Equal(t, len(objs), 3)
	ut.Equal(t, objs[0].GetType(), "cluster")
	ut.Equal(t, objs[1].GetType(), "namespace")
	ut.Equal(t, objs[2].GetType(), "deployment")

	req, _ = http.NewRequest("GET", "/apis/testing/v1/clusters/123321", nil)
	ctx, err = Parse(nil, req, schemas)
	ut.Equal(t, err, noErr)
	ut.Equal(t, ctx.Schema.ID, "cluster")
	ut.Equal(t, ctx.ID, "123321")
	ut.Equal(t, ctx.Schema.Parent, "")
	ut.Equal(t, ctx.Version.Group, "testing")
	ut.Equal(t, ctx.Version.Version, "v1")

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
