package types

import (
	"net/http"
)

type ResponseFormat string

const (
	ResponseYAML ResponseFormat = "yaml"
	ResponseJSON ResponseFormat = "json"
)

type Context struct {
	Schemas        *Schemas
	Request        *http.Request
	Response       http.ResponseWriter
	Object         Object
	Method         string
	Action         *Action
	ResponseFormat ResponseFormat
	params         map[string]interface{}
}

func NewContext(req *http.Request, resp http.ResponseWriter, schemas *Schemas) *Context {
	return &Context{
		Request:  req,
		Response: resp,
		Schemas:  schemas,
		Method:   req.Method,
		params:   make(map[string]interface{}),
	}
}

func (ctx *Context) Set(key string, value interface{}) {
	ctx.params[key] = value
}

func (ctx *Context) Get(key string) (interface{}, bool) {
	v, ok := ctx.params[key]
	return v, ok
}

func (ctx *Context) ParseRequestPath(url string) *APIError {
	obj, err := ctx.Schemas.CreateResourceFromUrl(url)
	if err == nil {
		ctx.Object = obj
	}
	return err
}
