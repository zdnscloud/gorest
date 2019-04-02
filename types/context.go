package types

import (
	"context"
	"net/http"
)

type APIContext struct {
	Schemas  *Schemas
	Request  *http.Request
	Response http.ResponseWriter
	Obj      Object
}

type apiContextKey struct{}

func NewAPIContext(req *http.Request, resp http.ResponseWriter, schemas *Schemas) *APIContext {
	apiCtx := &APIContext{
		Response: resp,
		Schemas:  schemas,
	}
	ctx := context.WithValue(req.Context(), apiContextKey{}, apiCtx)
	apiCtx.Request = req.WithContext(ctx)
	return apiCtx
}
