package parse

import (
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
		schema.CollectionMethods = []string{"GET"}
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

	req, _ = http.NewRequest("GET", "/testing/v1/clusters/123321", nil)
	ctx, _ = Parse(nil, req, schemas)
	ut.Equal(t, ctx.Schema.ID, "cluster")
	ut.Equal(t, ctx.ID, "123321")
	ut.Equal(t, ctx.Schema.Parent, "")
}
