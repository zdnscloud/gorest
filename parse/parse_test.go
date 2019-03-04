package parse

import (
	"fmt"
	"net/http"
	"testing"

	ut "github.com/zdnscloud/cement/unittest"
	"github.com/zdnscloud/gorest/types"
)

var version = types.APIVersion{
	Version: "v1",
	Group:   "testing",
	Path:    "/v1",
}

type Cluster struct {
	types.Resource
}

type Node struct {
	types.Resource
}

func TestParse(t *testing.T) {
	schemas := types.NewSchemas()
	schemas.MustImportAndCustomize(&version, Cluster{}, nil, func(schema *types.Schema, handler types.Handler) {
		schema.CollectionMethods = []string{"GET", "POST"}
		schema.ResourceMethods = []string{"GET"}
	})
	schemas.MustImportAndCustomize(&version, Node{}, nil, func(schema *types.Schema, handler types.Handler) {
		schema.Parent = "cluster"
		schema.CollectionMethods = []string{"GET"}
		schema.ResourceMethods = []string{"GET"}
	})

	req, _ := http.NewRequest("GET", "/testing/v1/clusters/123321/nodes/345543", nil)
	ctx, _ := Parse(nil, req, schemas)
	ut.Equal(t, ctx.Schema.ID, "node")
	ut.Equal(t, ctx.ID, "345543")
	ut.Equal(t, ctx.Schema.Parent, "cluster")
	ut.Equal(t, ctx.Parent.ID, "123321")
	ut.Equal(t, ctx.Version.Group, "testing")
	ut.Equal(t, ctx.Version.Path, "/v1")

	req, _ = http.NewRequest("GET", "/testing/v1/clusters/123321", nil)
	ctx, _ = Parse(nil, req, schemas)
	ut.Equal(t, ctx.Schema.ID, "cluster")
	ut.Equal(t, ctx.ID, "123321")
	ut.Equal(t, ctx.Schema.Parent, "")
	ut.Equal(t, ctx.Version.Group, "testing")
	ut.Equal(t, ctx.Version.Path, "/v1")

	req, _ = http.NewRequest("DELETE", "/testing/v1/clusters/123321", nil)
	deleteNotAllowedErr := types.NewAPIError(types.MethodNotAllowed, fmt.Sprintf("Method %s not supported", req.Method))
	_, err := Parse(nil, req, schemas)
	ut.Equal(t, err, deleteNotAllowedErr)

	req, _ = http.NewRequest("POST", "/testing/v1/clusters", nil)
	var nilErr *types.APIError
	_, err = Parse(nil, req, schemas)
	ut.Equal(t, err, nilErr)

	req, _ = http.NewRequest("POST", "/testing/v1/clusters/123123/nodes", nil)
	postNotAllowedErr := types.NewAPIError(types.MethodNotAllowed, fmt.Sprintf("Method %s not supported", req.Method))
	_, err = Parse(nil, req, schemas)
	ut.Equal(t, err, postNotAllowedErr)

	req, _ = http.NewRequest("GET", "/testing/v1/noshemas", nil)
	schemaNoFoundErr := types.NewAPIError(types.NotFound, fmt.Sprintf("no found schema noshemas"))
	_, err = Parse(nil, req, schemas)
	ut.Equal(t, err, schemaNoFoundErr)

	req, _ = http.NewRequest("GET", "/testing/v2/clusters", nil)
	versionNoFoundErr := types.NewAPIError(types.NotFound, fmt.Sprintf("no found version with url /testing/v2/clusters"))
	_, err = Parse(nil, req, schemas)
	ut.Equal(t, err, versionNoFoundErr)
}
