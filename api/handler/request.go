package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/zdnscloud/gorest/types"
)

func CreateHandler(apiContext *types.APIContext) *types.APIError {
	handler := apiContext.Schema.Handler
	if handler == nil {
		return types.NewAPIError(types.NotFound, "no found schema handler")
	}

	object, err := parseRequestBody(apiContext)
	if err != nil {
		return err
	}

	result, err := handler.Create(object)
	if err != nil {
		return err
	}

	WriteResponse(apiContext, http.StatusCreated, result)
	return nil
}

func DeleteHandler(apiContext *types.APIContext) *types.APIError {
	handler := apiContext.Schema.Handler
	if handler == nil {
		return types.NewAPIError(types.NotFound, "no found schema handler")
	}

	obj, err := getSchemaObject(apiContext)
	if err != nil {
		return err
	}

	obj.SetID(apiContext.ID)
	if err = handler.Delete(obj); err != nil {
		return err
	}

	WriteResponse(apiContext, http.StatusOK, nil)
	return nil
}

func UpdateHandler(apiContext *types.APIContext) *types.APIError {
	handler := apiContext.Schema.Handler
	if handler == nil {
		return types.NewAPIError(types.NotFound, "no found schema handler")
	}

	object, err := parseRequestBody(apiContext)
	if err != nil {
		return err
	}

	object.SetID(apiContext.ID)
	result, err := handler.Update(object)
	if err != nil {
		return err
	}

	WriteResponse(apiContext, http.StatusOK, result)
	return nil
}

func ListHandler(apiContext *types.APIContext) *types.APIError {
	handler := apiContext.Schema.Handler
	if handler == nil {
		return types.NewAPIError(types.NotFound, "no found schema handler")
	}

	var result interface{}
	obj, err := getSchemaObject(apiContext)
	if err != nil {
		return err
	}

	if apiContext.ID == "" {
		result = types.Collection{
			Type:         "collection",
			ResourceType: apiContext.Schema.ID,
			Data:         handler.List(obj),
		}
	} else {
		obj.SetID(apiContext.ID)
		result = handler.Get(obj)
	}

	WriteResponse(apiContext, http.StatusOK, result)
	return nil
}

func ActionHandler(apiContext *types.APIContext, action *types.Action) *types.APIError {
	handler := apiContext.Schema.Handler
	if handler == nil {
		return types.NewAPIError(types.NotFound, "no found schema handler")
	}

	var params map[string]interface{}
	if err := decodeBody(apiContext.Request, &params); err != nil {
		return err
	}

	obj, err := getSchemaObject(apiContext)
	if err != nil {
		return err
	}

	obj.SetID(apiContext.ID)
	result, err := handler.Action(obj, apiContext.Action, params)
	if err != nil {
		return err
	}

	WriteResponse(apiContext, http.StatusOK, result)
	return nil
}

func getSchemaObject(apiContext *types.APIContext) (types.Object, *types.APIError) {
	obj, ok := getSchemaStructVal(apiContext).(types.Object)
	if ok == false {
		return nil, types.NewAPIError(types.NotFound, "no found resource schema")
	}

	obj.SetType(apiContext.Schema.ID)
	obj.SetParent(apiContext.Parent)
	return obj, nil
}

func getSchemaStructVal(apiContext *types.APIContext) interface{} {
	val := apiContext.Schema.StructVal
	valPtr := reflect.New(val.Type())
	valPtr.Elem().Set(val)
	return valPtr.Interface()
}

func parseRequestBody(apiContext *types.APIContext) (types.Object, *types.APIError) {
	val := getSchemaStructVal(apiContext)
	if err := decodeBody(apiContext.Request, val); err != nil {
		return nil, err
	}

	if obj, ok := val.(types.Object); ok {
		obj.SetType(apiContext.Schema.ID)
		obj.SetParent(apiContext.Parent)
		return obj, nil
	} else {
		return nil, types.NewAPIError(types.InvalidBodyContent, fmt.Sprintf("Request Body mismatch resource schema"))
	}
}

func decodeBody(req *http.Request, params interface{}) *types.APIError {
	reqBody, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		return types.NewAPIError(types.InvalidBodyContent,
			fmt.Sprintf("Failed to read request body: %v", err.Error()))
	}

	err = json.Unmarshal(reqBody, params)
	if err != nil {
		return types.NewAPIError(types.InvalidBodyContent,
			fmt.Sprintf("Failed to parse request body: %v", err.Error()))
	}

	return nil
}
