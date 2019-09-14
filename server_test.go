package gorest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	ut "github.com/zdnscloud/cement/unittest"
	goresterr "github.com/zdnscloud/gorest/error"
	"github.com/zdnscloud/gorest/resource"
	"github.com/zdnscloud/gorest/resource/schema"
)

var (
	schemas = schema.NewSchemaManager()
	version = resource.APIVersion{
		Group:   "testing",
		Version: "v1",
	}
)

type dumbHandler struct{}

func (h *dumbHandler) Create(ctx *resource.Context) (interface{}, *goresterr.APIError) {
	return nil, nil
}

func (h *dumbHandler) List(ctx *resource.Context) interface{} {
	return nil
}

type Foo struct {
	resource.ResourceBase
}

var gnum int

var dumbHandler1 = func(ctx *resource.Context) *goresterr.APIError {
	ctx.Set("key", &gnum)
	return nil
}

var dumbHandler2 = func(ctx *resource.Context) *goresterr.APIError {
	val_, _ := ctx.Get("key")
	*(val_.(*int)) = 100
	return nil
}

func TestContextPassChain(t *testing.T) {
	handler, _ := resource.HandlerAdaptor(&dumbHandler{})
	schemas.Import(&version, Foo{}, handler)
	req, _ := http.NewRequest("GET", "/apis/testing/v1/foos", nil)
	req.Host = "127.0.0.1:1234"
	w := httptest.NewRecorder()
	s := NewAPIServer(schemas)
	s.Use(dumbHandler1)
	s.Use(dumbHandler2)

	ut.Equal(t, gnum, 0)
	s.ServeHTTP(w, req)
	ut.Equal(t, gnum, 100)

	s.ServeHTTP(w, req)
}
