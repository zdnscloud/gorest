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

	object, err := parseBody(apiContext)
	if err != nil {
		return err
	}

	result, err := handler.Create(object)
	if err != nil {
		return err
	}

	addResourceLinks(apiContext, result)
	WriteResponse(apiContext, http.StatusCreated, result)
	return nil
}

func DeleteHandler(apiContext *types.APIContext) *types.APIError {
	handler := apiContext.Schema.Handler
	if handler == nil {
		return types.NewAPIError(types.NotFound, "no found schema handler")
	}

	obj, err := getObject(apiContext, getSchemaStructVal(apiContext))
	if err != nil {
		return err
	}

	obj.SetID(apiContext.ID)
	if err = handler.Delete(obj); err != nil {
		return err
	}

	WriteResponse(apiContext, http.StatusNoContent, nil)
	return nil
}

func UpdateHandler(apiContext *types.APIContext) *types.APIError {
	handler := apiContext.Schema.Handler
	if handler == nil {
		return types.NewAPIError(types.NotFound, "no found schema handler")
	}

	object, err := parseBody(apiContext)
	if err != nil {
		return err
	}

	object.SetID(apiContext.ID)
	result, err := handler.Update(object)
	if err != nil {
		return err
	}

	addResourceLinks(apiContext, result)
	WriteResponse(apiContext, http.StatusOK, result)
	return nil
}

func ListHandler(apiContext *types.APIContext) *types.APIError {
	handler := apiContext.Schema.Handler
	if handler == nil {
		return types.NewAPIError(types.NotFound, "no found schema handler")
	}

	var result interface{}
	obj, err := getObject(apiContext, getSchemaStructVal(apiContext))
	if err != nil {
		return err
	}

	if apiContext.ID == "" {
		data := handler.List(obj)
		if data == nil || reflect.ValueOf(data).IsNil() {
			data = make([]types.Object, 0)
		}

		collection := &types.Collection{
			Type:         "collection",
			ResourceType: apiContext.Schema.ID,
			Data:         data,
		}
		addCollectionLinks(apiContext, collection)
		result = collection
	} else {
		obj.SetID(apiContext.ID)
		result = handler.Get(obj)
		if result == nil || reflect.ValueOf(result).IsNil() {
			return types.NewAPIError(types.NotFound,
				fmt.Sprintf("no found %v with id %v", obj.GetType(), apiContext.ID))
		}
		addResourceLinks(apiContext, result)
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

	obj, err := getObject(apiContext, getSchemaStructVal(apiContext))
	if err != nil {
		return err
	}

	obj.SetID(apiContext.ID)
	result, err := handler.Action(obj, apiContext.Action, params)
	if err != nil {
		return err
	}

	WriteResponse(apiContext, http.StatusAccepted, result)
	return nil
}

func getSchemaStructVal(apiContext *types.APIContext) interface{} {
	val := apiContext.Schema.StructVal
	valPtr := reflect.New(val.Type())
	valPtr.Elem().Set(val)
	return valPtr.Interface()
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

func parseBody(apiContext *types.APIContext) (types.Object, *types.APIError) {
	val := getSchemaStructVal(apiContext)
	if err := decodeBody(apiContext.Request, val); err != nil {
		return nil, err
	}

	return getObject(apiContext, val)
}

func getObject(apiContext *types.APIContext, val interface{}) (types.Object, *types.APIError) {
	if obj, ok := val.(types.Object); ok {
		obj.SetType(apiContext.Schema.ID)
		obj.SetParent(apiContext.Parent)
		return obj, nil
	} else {
		return nil, types.NewAPIError(types.NotFound, fmt.Sprintf("no found resource schema"))
	}
}
