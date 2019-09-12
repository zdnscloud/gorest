package handler

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/zdnscloud/gorest/types"
)

type Collection struct {
	Type         string            `json:"type,omitempty"`
	ResourceType string            `json:"resourceType,omitempty"`
	Links        map[string]string `json:"links,omitempty"`
	Data         interface{}       `json:"data"`
}

func addLinks(ctx *types.Context, obj types.Object) {
	urlPrefix := getUrlPrefix(ctx.Request)

	links := make(map[string]string)
	ctx.Object.SetID(obj.GetID())
	self := types.GenerateResourceUrl(urlPrefix, ctx.Object)
	schema := ctx.Object.GetSchema()
	if schema.SupportResourceMethod(http.MethodGet) {
		links["self"] = self
	}

	if schema.SupportResourceMethod(http.MethodPut) {
		links["update"] = self
	}

	if schema.SupportResourceMethod(http.MethodDelete) {
		links["remove"] = self
	}

	if schema.SupportCollectionMethod(http.MethodGet) {
		links["collection"] = types.GenerateResourceCollectionUrl(urlPrefix, ctx.Object)
	}

	for childCollection, url := range types.GenerateChildrenUrl(urlPrefix, ctx.Object) {
		links[childCollection] = url
	}

	obj.SetLinks(links)
}

func addResourceLinks(ctx *types.Context, obj interface{}) {
	if object, ok := obj.(types.Object); ok {
		addLinks(ctx, object)
	}
}

func addCollectionLinks(ctx *types.Context, collection *Collection) {
	urlPrefix := getUrlPrefix(ctx.Request)

	collection.Links = map[string]string{
		"self": types.GenerateResourceCollectionUrl(urlPrefix, ctx.Object),
	}

	sliceData := reflect.ValueOf(collection.Data)
	if sliceData.Kind() == reflect.Slice {
		for i := 0; i < sliceData.Len(); i++ {
			addResourceLinks(ctx, sliceData.Index(i).Interface())
		}
	}
}

func getUrlPrefix(req *http.Request) string {
	scheme := "http"
	if req.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, req.Host)
}
