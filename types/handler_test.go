package types

import (
	"testing"

	ut "github.com/zdnscloud/cement/unittest"
)

type dumbHandler struct{}

func (h *dumbHandler) Create(ctx *Context, content []byte) (interface{}, *APIError) {
	return 10, nil
}

func (h *dumbHandler) Delete(ctx *Context) *APIError {
	return nil
}

func (h *dumbHandler) Update(ctx *Context) (interface{}, *APIError) {
	return 20, nil
}

func (h *dumbHandler) List(ctx *Context) interface{} {
	return []uint{1, 2, 3}
}

func (h *dumbHandler) Get(ctx *Context) interface{} {
	return 10
}

func (h *dumbHandler) Action(ctx *Context) (interface{}, *APIError) {
	return 10, nil
}

type dumbHandlerTwo struct{}

func (h *dumbHandlerTwo) Create(ctx *Context, content []byte) (interface{}, *APIError) {
	return 10, nil
}

type emptyHandler struct{}

func TestHandlerGen(t *testing.T) {
	handler, _ := NewHandler(&dumbHandler{})
	resourceMethods := GetResourceMethods(handler)
	collectionMethods := GetCollectionMethods(handler)
	ut.Equal(t, resourceMethods, []string{"GET", "DELETE", "PUT", "POST"})
	ut.Equal(t, collectionMethods, []string{"GET", "POST"})

	createResult, err := handler.GetCreateHandler()(nil, nil)
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

	handler, _ = NewHandler(&dumbHandlerTwo{})
	resourceMethods = GetResourceMethods(handler)
	collectionMethods = GetCollectionMethods(handler)
	ut.Equal(t, len(resourceMethods), 0)
	ut.Equal(t, collectionMethods, []string{"POST"})

	_, err_ := NewHandler(&emptyHandler{})
	ut.Assert(t, err_ != nil, "")
}
