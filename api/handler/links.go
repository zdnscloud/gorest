package handler

import (
	"fmt"
	"net/http"
	"path"

	"github.com/zdnscloud/gorest/types"
	"github.com/zdnscloud/gorest/util/slice"
)

func addLinks(apiContext *types.APIContext, obj types.Object) {
	links := make(map[string]string)
	self := genSelfURL(apiContext.Request, obj.GetID())
	links["self"] = self
	if slice.ContainsString(apiContext.Schema.ResourceMethods, http.MethodPut) {
		links["update"] = self
	}

	if slice.ContainsString(apiContext.Schema.ResourceMethods, http.MethodDelete) {
		links["remove"] = self
	}

	obj.SetLinks(links)
}

func addCollectionLinks(apiContext *types.APIContext, objs interface{}) {
	for _, obj := range objs.([]types.Object) {
		addLinks(apiContext, obj)
	}
}

func genSelfURL(req *http.Request, id string) string {
	if id == "" {
		return ""
	}

	return path.Join(getRequestURL(req), id)
}

func getRequestURL(req *http.Request) string {
	scheme := "http"
	if req.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s%s", scheme, req.Host, req.URL.Path)
}
