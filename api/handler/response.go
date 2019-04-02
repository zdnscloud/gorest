package handler

import (
	"encoding/json"

	"github.com/zdnscloud/gorest/parse"
	"github.com/zdnscloud/gorest/types"
	yaml "gopkg.in/yaml.v2"
)

func WriteResponse(apiContext *types.Context, status int, result interface{}) {
	resp := apiContext.Response
	resp.WriteHeader(status)
	var body []byte
	switch parse.ParseResponseFormat(apiContext.Request) {
	case "json":
		resp.Header().Set("content-type", "application/json")
		body, _ = json.Marshal(result)
	case "yaml":
		resp.Header().Set("content-type", "application/yaml")
		body, _ = yaml.Marshal(result)
	}
	resp.Write(body)
}
