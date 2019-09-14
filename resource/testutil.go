package resource

import (
	"github.com/zdnscloud/gorest/error"
)

type DumbHandler struct{}

func (h *DumbHandler) Create(ctx *Context) (interface{}, *error.APIError) {
	return 10, nil
}

func (h *DumbHandler) Delete(ctx *Context) *error.APIError {
	return nil
}

func (h *DumbHandler) Update(ctx *Context) (interface{}, *error.APIError) {
	return 20, nil
}

func (h *DumbHandler) List(ctx *Context) interface{} {
	return []uint{1, 2, 3}
}

func (h *DumbHandler) Get(ctx *Context) interface{} {
	return 10
}

func (h *DumbHandler) Action(ctx *Context) (interface{}, *error.APIError) {
	return 10, nil
}
