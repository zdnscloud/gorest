package handler

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	ut "github.com/zdnscloud/cement/unittest"
	"github.com/zdnscloud/gorest/types"
)

var (
	schemas = types.NewSchemas().MustImportAndCustomize(&version, Foo{}, &Handler{}, func(schema *types.Schema, handler types.Handler) {
		schema.Handler = handler
	})

	schema = schemas.Schema(&version, types.GetResourceType(Foo{}))

	ctx = &types.APIContext{
		Schemas: schemas,
		Obj: &types.Resource{
			Type:   schema.ID,
			Schema: schema,
		},
	}
)

type Foo struct {
	types.Resource
	Name string `json:"name"singlecloud:"required=true"`
	Role string `json:"role"singlecloud:"required=true"`
}

type testServer struct {
	ctx *types.APIContext
}

func (t *testServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var err *types.APIError
	switch req.Method {
	case http.MethodPost:
		err = CreateHandler(t.ctx)
	case http.MethodPut:
		err = UpdateHandler(t.ctx)
	case http.MethodDelete:
		err = DeleteHandler(t.ctx)
	case http.MethodGet:
		err = ListHandler(t.ctx)
	default:
		panic("unspport method " + req.Method)
	}

	if err != nil {
		WriteResponse(t.ctx, err.Status, err)
	}
}

func TestCreateHandler(t *testing.T) {
	yamlContent := "testContent"
	expectBody := "{\"id\":\"12138\",\"type\":\"foo\",\"links\":{\"self\":\"http://127.0.0.1:1234/test/v1/foos/12138\"},\"creationTimestamp\":null,\"name\":\"bar\",\"role\":\"master\"}"
	req, _ := http.NewRequest("POST", "/test/v1/foos", bytes.NewBufferString(fmt.Sprintf("{\"name\":\"bar\", \"yaml_\":\"%s\",\"role\":\"master\"}", yamlContent)))
	req.Host = "127.0.0.1:1234"
	w := httptest.NewRecorder()
	ctx.Request = req
	ctx.Response = w
	server := &testServer{}
	server.ctx = ctx
	server.ServeHTTP(w, req)
	ut.Equal(t, w.Code, 201)
	ut.Equal(t, w.Body.String(), expectBody)
}

func TestDeleteHandler(t *testing.T) {
	req, _ := http.NewRequest("DELETE", "/test/v1/foos/12138", nil)
	req.Host = "127.0.0.1:1234"
	w := httptest.NewRecorder()
	ctx.Request = req
	ctx.Response = w
	server := &testServer{}
	server.ctx = ctx
	server.ServeHTTP(w, req)
	ut.Equal(t, w.Code, 204)
}

func TestUpdateHandler(t *testing.T) {
	yamlContent := "testContent"
	expectBody := "{\"id\":\"12138\",\"type\":\"foo\",\"links\":{\"self\":\"http://127.0.0.1:1234/test/v1/foos/12138\"},\"creationTimestamp\":null,\"name\":\"bar\",\"role\":\"worker\"}"
	req, _ := http.NewRequest("PUT", "/test/v1/foos/12138", bytes.NewBufferString(fmt.Sprintf("{\"name\":\"bar\", \"yaml_\":\"%s\",\"role\": \"worker\"}", yamlContent)))
	req.Host = "127.0.0.1:1234"
	w := httptest.NewRecorder()
	ctx.Request = req
	ctx.Response = w
	server := &testServer{}
	server.ctx = ctx
	server.ServeHTTP(w, req)
	ut.Equal(t, w.Code, 200)
	ut.Equal(t, w.Body.String(), expectBody)
}

func TestListHandler(t *testing.T) {
	expectCollection := "{\"type\":\"collection\",\"resourceType\":\"foo\",\"links\":{\"self\":\"http://127.0.0.1:1234/test/v1/foos\"},\"data\":[]}"
	req, _ := http.NewRequest("GET", "/test/v1/foos", nil)
	req.Host = "127.0.0.1:1234"
	w := httptest.NewRecorder()
	ctx.Request = req
	ctx.Response = w
	server := &testServer{}
	server.ctx = ctx
	server.ServeHTTP(w, req)
	ut.Equal(t, w.Code, 200)
	ut.Equal(t, w.Body.String(), expectCollection)
}

func TestListHandlerForGetOne(t *testing.T) {
	expectResult := "{\"id\":\"12138\",\"type\":\"foo\",\"links\":{\"self\":\"http://127.0.0.1:1234/test/12138\"},\"creationTimestamp\":null,\"name\":\"bar\",\"role\":\"worker\"}"
	req, _ := http.NewRequest("GET", "/test/12138", nil)
	req.Host = "127.0.0.1:1234"
	w := httptest.NewRecorder()
	ctx.Request = req
	ctx.Response = w
	ctx.Obj.SetID("12138")
	server := &testServer{}
	server.ctx = ctx
	server.ServeHTTP(w, req)
	ut.Equal(t, w.Code, 200)
	ut.Equal(t, w.Body.String(), expectResult)
}

func TestGetFail(t *testing.T) {
	expectResult := "{\"code\":\"NotFound\",\"status\":404,\"type\":\"error\",\"message\":\"no found foo with id 23456\"}"
	req, _ := http.NewRequest("GET", "/test/12138", nil)
	req.Host = "127.0.0.1:1234"
	w := httptest.NewRecorder()
	ctx.Request = req
	ctx.Response = w
	ctx.Obj.SetID("23456")
	server := &testServer{}
	server.ctx = ctx
	server.ServeHTTP(w, req)
	ut.Equal(t, w.Code, 404)
	ut.Equal(t, w.Body.String(), expectResult)
}

type Handler struct{}

func (h *Handler) Create(obj types.Object, content []byte) (interface{}, *types.APIError) {
	obj.SetID("12138")
	return obj, nil
}

func (h *Handler) Delete(obj types.Object) *types.APIError {
	return nil
}

func (h *Handler) Update(obj types.Object) (interface{}, *types.APIError) {
	obj.SetID("12138")
	return obj, nil
}

func (h *Handler) List(obj types.Object) interface{} {
	return []types.Object{}
}

func (h *Handler) Get(obj types.Object) interface{} {
	if obj.GetID() == "12138" {
		foo := obj.(*Foo)
		foo.Name = "bar"
		foo.Role = "worker"
		return foo
	}
	return nil
}

func (h *Handler) Action(obj types.Object, action string, params map[string]interface{}) (interface{}, *types.APIError) {
	return params, nil
}
