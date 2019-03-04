package handler

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
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
	yamlContent := base64.StdEncoding.EncodeToString([]byte("testContent"))
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
	ut.Equal(t, string(content), "testContent")
	ut.Equal(t, obj.(*Foo).Type, "foo")
	ut.Equal(t, obj.(*Foo).Name, "bar")
}
