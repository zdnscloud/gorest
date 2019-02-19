package crd

import (
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/zdnscloud/gorest/parse"
	"github.com/zdnscloud/gorest/types"
)

func AssignHandler(schema *types.Schema) {
	schema.CreateHandler = CreateHandler
	schema.DeleteHandler = DeleteHandler
	schema.UpdateHandler = UpdateHandler
	schema.ListHandler = ListHandler
}

func parseRequestBody(apiContext *types.APIContext) (interface{}, error) {
	val := apiContext.Schema.StructVal
	valPtr := reflect.New(val.Type())
	valPtr.Elem().Set(val)
	decode := parse.GetDecoder(apiContext.Request, io.LimitReader(apiContext.Request.Body, parse.MaxFormSize))
	if err := decode(valPtr.Interface()); err != nil {
		return nil, err
	}

	return valPtr.Interface(), nil
}

func CreateHandler(apiContext *types.APIContext, next types.RequestHandler) error {
	fmt.Printf("enter crd create handler\n")
	obj, err := parseRequestBody(apiContext)
	if err != nil {
		return err
	}

	object, ok := obj.(types.Object)
	if ok {
		if err := object.Create(); err != nil {
			return err
		}
	} else {
		fmt.Printf("obj to object failed\n")
	}

	apiContext.WriteResponse(http.StatusCreated, nil)
	return nil
}

func DeleteHandler(apiContext *types.APIContext, next types.RequestHandler) error {
	return nil
}

func UpdateHandler(apiContext *types.APIContext, next types.RequestHandler) error {
	return nil
}

func ListHandler(apiContext *types.APIContext, next types.RequestHandler) error {
	return nil
}
