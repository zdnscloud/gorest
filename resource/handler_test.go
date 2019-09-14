package resource

import (
	"net/http"
	"testing"

	ut "github.com/zdnscloud/cement/unittest"
	err "github.com/zdnscloud/gorest/error"
)

type dumbHandlerTwo struct{}

func (h *dumbHandlerTwo) Create(ctx *Context) (interface{}, *err.APIError) {
	return 10, nil
}

type emptyHandler struct{}

func TestHandlerGen(t *testing.T) {
	handler, _ := HandlerAdaptor(&DumbHandler{})
	resourceMethods := GetResourceMethods(handler)
	collectionMethods := GetCollectionMethods(handler)
	ut.Equal(t, resourceMethods, []HttpMethod{http.MethodGet, http.MethodDelete, http.MethodPut, http.MethodPost})
	ut.Equal(t, collectionMethods, []HttpMethod{http.MethodGet, http.MethodPost})

	createResult, err := handler.GetCreateHandler()(nil)
	ut.Assert(t, err == nil, "")
	ut.Equal(t, createResult.(int), 10)

	updateResult, err := handler.GetUpdateHandler()(nil)
	ut.Assert(t, err == nil, "")
	ut.Equal(t, updateResult.(int), 20)

	err = handler.GetDeleteHandler()(nil)
	ut.Assert(t, err == nil, "")

	listResult := handler.GetListHandler()(nil)
	ut.Equal(t, listResult.([]uint), []uint{1, 2, 3})

	getResult := handler.GetGetHandler()(nil)
	ut.Equal(t, getResult.(int), 10)

	actionResult, err := handler.GetActionHandler()(nil)
	ut.Equal(t, actionResult.(int), 10)
	ut.Assert(t, err == nil, "")

	handler, _ = HandlerAdaptor(&dumbHandlerTwo{})
	resourceMethods = GetResourceMethods(handler)
	collectionMethods = GetCollectionMethods(handler)
	ut.Equal(t, len(resourceMethods), 0)
	ut.Equal(t, collectionMethods, []HttpMethod{http.MethodPost})

	_, err_ := HandlerAdaptor(&emptyHandler{})
	ut.Assert(t, err_ != nil, "")
}
