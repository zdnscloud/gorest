package handler

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/zdnscloud/gorest/types"
	"github.com/zdnscloud/gorest/util/slice"
)

func addLinks(apiContext *types.APIContext, obj types.Object) {
	links := make(map[string]string)
	self := genResourceLink(apiContext.Request, obj.GetID())
	links["self"] = self
	if slice.ContainsString(apiContext.Schema.ResourceMethods, http.MethodPut) {
		links["update"] = self
	}

	if slice.ContainsString(apiContext.Schema.ResourceMethods, http.MethodDelete) {
		links["remove"] = self
	}

	obj.SetLinks(links)
}

func addResourceLinks(apiContext *types.APIContext, obj interface{}) {
	if object, ok := obj.(types.Object); ok {
		addLinks(apiContext, object)
	}
}

func addCollectionLinks(apiContext *types.APIContext, collection *types.Collection) {
	collection.Links = map[string]string{
		"self": getRequestURL(apiContext.Request),
	}

	slice := reflect.ValueOf(collection.Data)
	if slice.Kind() == reflect.Slice {
		for i := 0; i < slice.Len(); i++ {
			addResourceLinks(apiContext, slice.Index(i).Interface())
		}
	}
}

func genResourceLink(req *http.Request, id string) string {
	if id == "" {
		return ""
	}

	requestURL := getRequestURL(req)
	if strings.HasSuffix(requestURL, "/"+id) {
		return requestURL
	} else {
		return requestURL + "/" + id
	}
}

func getRequestURL(req *http.Request) string {
	scheme := "http"
	if req.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s%s", scheme, req.Host, req.URL.Path)
}
