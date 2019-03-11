package handler

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	ut "github.com/zdnscloud/cement/unittest"
	"github.com/zdnscloud/gorest/types"
)

type Foo struct {
	types.Resource
	Name string `json:"name"`
}

func TestParseCreateBody(t *testing.T) {
	var noerr *types.APIError
	yamlContent := "testContent"
	req, _ := http.NewRequest("POST", "/", bytes.NewBufferString(fmt.Sprintf("{\"name\":\"bar\", \"yaml_\":\"%s\"}", yamlContent)))
	apiContext := &types.APIContext{
		Request: req,
		Schema: &types.Schema{
			ID:        "foo",
			StructVal: reflect.ValueOf(Foo{}),
		},
	}

	content, obj, err := parseCreateBody(apiContext)
	ut.Equal(t, err, noerr)
	ut.Equal(t, obj.(*Foo).Type, "foo")
	ut.Equal(t, obj.(*Foo).Name, "bar")
	ut.Equal(t, string(content), "testContent")
}

type Handler struct{}

func (h *Handler) Create(obj types.Object, content []byte) (interface{}, *types.APIError) {
	return nil, nil
}

func (h *Handler) Delete(obj types.Object) *types.APIError {
	return nil
}

func (h *Handler) Update(obj types.Object) (interface{}, *types.APIError) {
	return nil, nil
}
func (h *Handler) List(obj types.Object) interface{} {
	return nil
}

func (h *Handler) Get(obj types.Object) interface{} {
	return nil
}

func (h *Handler) Action(obj types.Object, action string, params map[string]interface{}) (interface{}, *types.APIError) {
	return nil, nil
}

func TestList(t *testing.T) {
	var noerr *types.APIError
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Host = "127.0.0.1:1234"
	apiContext := &types.APIContext{
		Request:  req,
		Response: httptest.NewRecorder(),
		Schema: &types.Schema{
			ID:        "foo",
			StructVal: reflect.ValueOf(Foo{}),
			Handler:   &Handler{},
		},
		ResponseFormat: "json",
	}

	expectCollection := "{\"type\":\"collection\",\"resourceType\":\"foo\",\"links\":{\"self\":\"http://127.0.0.1:1234/test\"},\"data\":[]}"
	err := ListHandler(apiContext)
	ut.Equal(t, err, noerr)
	ut.Equal(t, apiContext.Response.(*httptest.ResponseRecorder).Code, http.StatusOK)
	body := apiContext.Response.(*httptest.ResponseRecorder).Body
	ut.Equal(t, body.String(), expectCollection)
}

func TestGet(t *testing.T) {
	req, _ := http.NewRequest("GET", "/test/123", nil)
	req.Host = "127.0.0.1:1234"
	apiContext := &types.APIContext{
		ID:       "123",
		Request:  req,
		Response: httptest.NewRecorder(),
		Schema: &types.Schema{
			ID:        "foo",
			StructVal: reflect.ValueOf(Foo{}),
			Handler:   &Handler{},
		},
		ResponseFormat: "json",
	}

	expectResut := types.NewAPIError(types.NotFound, "no found foo with id 123")
	err := ListHandler(apiContext)
	ut.Equal(t, err, expectResut)
}
