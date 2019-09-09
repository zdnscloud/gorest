package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	ut "github.com/zdnscloud/cement/unittest"
	"github.com/zdnscloud/gorest/types"
)

var (
	schemas = types.NewSchemas().MustImport(&version, Foo{}, &dumbHandler{})
	version = types.APIVersion{
		Group:   "testing",
		Version: "v1",
	}
)

type dumbHandler struct{}

func (h *dumbHandler) Create(ctx *types.Context, content []byte) (interface{}, *types.APIError) {
	return nil, nil
}

func (h *dumbHandler) List(ctx *types.Context) interface{} {
	return nil
}

type Foo struct {
	types.Resource
}

func (c Foo) GetParents() []string {
	return nil
}

func (c Foo) GetActions() []types.Action {
	return nil
}

func (c Foo) GetCollectionActions() []types.Action {
	return nil
}

var gnum int

var dumbHandler1 = func(ctx *types.Context) *types.APIError {
	ctx.Set("key", &gnum)
	return nil
}

var dumbHandler2 = func(ctx *types.Context) *types.APIError {
	val_, _ := ctx.Get("key")
	*(val_.(*int)) = 100
	return nil
}

func TestContextPassChain(t *testing.T) {
	req, _ := http.NewRequest("GET", "/apis/testing/v1/foos", nil)
	req.Host = "127.0.0.1:1234"
	w := httptest.NewRecorder()
	s := NewAPIServer()
	s.AddSchemas(schemas)
	s.Use(dumbHandler1)
	s.Use(dumbHandler2)

	ut.Equal(t, gnum, 0)
	s.ServeHTTP(w, req)
	ut.Equal(t, gnum, 100)

	s.Use(RestHandler)
	s.ServeHTTP(w, req)
}
