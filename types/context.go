package types

import (
	"context"
	"net/http"
)

type Context struct {
	Schemas  *Schemas
	Request  *http.Request
	Response http.ResponseWriter
	Object   Object
}

type apiContextKey struct{}

func NewContext(req *http.Request, resp http.ResponseWriter, schemas *Schemas) *Context {
	apiCtx := &Context{
		Response: resp,
		Schemas:  schemas,
	}
	ctx := context.WithValue(req.Context(), apiContextKey{}, apiCtx)
	apiCtx.Request = req.WithContext(ctx)
	return apiCtx
}
