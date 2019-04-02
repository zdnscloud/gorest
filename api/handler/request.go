package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/zdnscloud/gorest/types"
)

func CreateHandler(ctx *types.Context) *types.APIError {
	handler := ctx.Object.GetSchema().Handler
	if handler == nil {
		return types.NewAPIError(types.NotFound, "no found schema handler")
	}

	content, object, err := parseCreateBody(ctx)
	if err != nil {
		return err
	}

	result, err := handler.Create(object, content)
	if err != nil {
		return err
	}

	addResourceLinks(ctx, result)
	WriteResponse(ctx, http.StatusCreated, result)
	return nil
}

func DeleteHandler(ctx *types.Context) *types.APIError {
	handler := ctx.Object.GetSchema().Handler
	if handler == nil {
		return types.NewAPIError(types.NotFound, "no found schema handler")
	}

	obj, err := getObject(ctx, getSchemaStructVal(ctx))
	if err != nil {
		return err
	}

	obj.SetID(ctx.Object.GetID())
	if err = handler.Delete(obj); err != nil {
		return err
	}

	WriteResponse(ctx, http.StatusNoContent, nil)
	return nil
}

func UpdateHandler(ctx *types.Context) *types.APIError {
	handler := ctx.Object.GetSchema().Handler
	if handler == nil {
		return types.NewAPIError(types.NotFound, "no found schema handler")
	}

	val := getSchemaStructVal(ctx)
	if err := decodeBody(ctx.Request, val); err != nil {
		return err
	}

	object, err := getObject(ctx, val)
	if err != nil {
		return err
	}

	object.SetID(ctx.Object.GetID())
	result, err := handler.Update(object)
	if err != nil {
		return err
	}

	addResourceLinks(ctx, result)
	WriteResponse(ctx, http.StatusOK, result)
	return nil
}

func ListHandler(ctx *types.Context) *types.APIError {
	handler := ctx.Object.GetSchema().Handler
	if handler == nil {
		return types.NewAPIError(types.NotFound, "no found schema handler")
	}

	var result interface{}
	obj, err := getObject(ctx, getSchemaStructVal(ctx))
	if err != nil {
		return err
	}

	if ctx.Object.GetID() == "" {
		data := handler.List(obj)
		if data == nil || reflect.ValueOf(data).IsNil() {
			data = make([]types.Object, 0)
		}

		collection := &types.Collection{
			Type:         "collection",
			ResourceType: ctx.Object.GetType(),
			Data:         data,
		}
		addCollectionLinks(ctx, collection)
		result = collection
	} else {
		obj.SetID(ctx.Object.GetID())
		result = handler.Get(obj)
		if result == nil || reflect.ValueOf(result).IsNil() {
			return types.NewAPIError(types.NotFound,
				fmt.Sprintf("no found %v with id %v", obj.GetType(), ctx.Object.GetID()))
		}
		addResourceLinks(ctx, result)
	}

	WriteResponse(ctx, http.StatusOK, result)
	return nil
}

func ActionHandler(ctx *types.Context) *types.APIError {
	handler := ctx.Object.GetSchema().Handler
	if handler == nil {
		return types.NewAPIError(types.NotFound, "no found schema handler")
	}

	var params map[string]interface{}
	if err := decodeBody(ctx.Request, &params); err != nil {
		return err
	}

	obj, err := getObject(ctx, getSchemaStructVal(ctx))
	if err != nil {
		return err
	}

	obj.SetID(ctx.Object.GetID())
	result, err := handler.Action(obj, ctx.Action.Name, params)
	if err != nil {
		return err
	}

	WriteResponse(ctx, http.StatusAccepted, result)
	return nil
}

func getSchemaStructVal(ctx *types.Context) interface{} {
	val := ctx.Object.GetSchema().StructVal
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

func getObject(ctx *types.Context, val interface{}) (types.Object, *types.APIError) {
	if obj, ok := val.(types.Object); ok {
		obj.SetType(ctx.Object.GetType())
		obj.SetParent(ctx.Object.GetParent())
		return obj, nil
	} else {
		return nil, types.NewAPIError(types.NotFound, fmt.Sprintf("no found resource schema"))
	}
}

func parseCreateBody(ctx *types.Context) ([]byte, types.Object, *types.APIError) {
	var params struct {
		Yaml string `json:"yaml_"`
	}

	reqBody, err := ioutil.ReadAll(ctx.Request.Body)
	defer ctx.Request.Body.Close()
	if err != nil {
		return nil, nil, types.NewAPIError(types.InvalidBodyContent,
			fmt.Sprintf("Failed to read request body: %v", err.Error()))
	}

	if err := json.Unmarshal(reqBody, &params); err != nil {
		return nil, nil, types.NewAPIError(types.InvalidBodyContent,
			fmt.Sprintf("Failed to parse request body: %v", err.Error()))
	}

	val := getSchemaStructVal(ctx)
	if err := json.Unmarshal(reqBody, val); err != nil {
		return nil, nil, types.NewAPIError(types.InvalidBodyContent,
			fmt.Sprintf("Failed to parse request body: %v", err.Error()))
	}

	obj, apiErr := getObject(ctx, val)
	if apiErr != nil {
		return nil, nil, apiErr
	}

	return []byte(params.Yaml), obj, CheckObjectFields(ctx, obj)
}
