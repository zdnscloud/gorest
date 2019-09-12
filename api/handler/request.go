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
	handler := ctx.Object.GetSchema().Handler.GetCreateHandler()
	if handler == nil {
		return types.NewAPIError(types.NotFound, "no handler for create")
	}

	content, err := parseCreateBody(ctx)
	if err != nil {
		return err
	}

	result, err := handler(ctx, content)
	if err != nil {
		return err
	}

	addResourceLinks(ctx, result)
	WriteResponse(ctx, http.StatusCreated, result)
	return nil
}

func DeleteHandler(ctx *types.Context) *types.APIError {
	handler := ctx.Object.GetSchema().Handler.GetDeleteHandler()
	if handler == nil {
		return types.NewAPIError(types.NotFound, "no handler for delete")
	}

	if err := handler(ctx); err != nil {
		return err
	}

	WriteResponse(ctx, http.StatusNoContent, nil)
	return nil
}

func UpdateHandler(ctx *types.Context) *types.APIError {
	handler := ctx.Object.GetSchema().Handler.GetUpdateHandler()
	if handler == nil {
		return types.NewAPIError(types.NotFound, "no handler for update")
	}

	if err := decodeBody(ctx.Request, ctx.Object); err != nil {
		return err
	}

	result, err := handler(ctx)
	if err != nil {
		return err
	}

	addResourceLinks(ctx, result)
	WriteResponse(ctx, http.StatusOK, result)
	return nil
}

func ListHandler(ctx *types.Context) *types.APIError {
	var result interface{}
	if ctx.Object.GetID() == "" {
		handler := ctx.Object.GetSchema().Handler.GetListHandler()
		if handler == nil {
			return types.NewAPIError(types.NotFound, "no found for list")
		}
		data := handler(ctx)
		if data == nil {
			data = make([]types.Object, 0)
		} else if reflect.ValueOf(data).Kind() != reflect.Slice {
			return types.NewAPIError(types.ServerError,
				fmt.Sprintf("list handler doesn't return slice but %v", reflect.ValueOf(data).Kind()))
		}
		collection := &Collection{
			Type:         "collection",
			ResourceType: ctx.Object.GetType(),
			Data:         data,
		}
		addCollectionLinks(ctx, collection)
		result = collection
	} else {
		handler := ctx.Object.GetSchema().Handler.GetGetHandler()
		if handler == nil {
			return types.NewAPIError(types.NotFound, "no found for list")
		}
		result = handler(ctx)
		if result == nil {
			return types.NewAPIError(types.NotFound,
				fmt.Sprintf("%s resource with id %s doesn't exist", ctx.Object.GetType(), ctx.Object.GetID()))
		} else if reflect.ValueOf(result).Kind() != reflect.Ptr {
			return types.NewAPIError(types.ServerError,
				fmt.Sprintf("get handler doesn't return pointer but %v", reflect.ValueOf(result).Kind()))
		}
		addResourceLinks(ctx, result)
	}

	WriteResponse(ctx, http.StatusOK, result)
	return nil
}

func ActionHandler(ctx *types.Context) *types.APIError {
	handler := ctx.Object.GetSchema().Handler.GetActionHandler()
	if handler == nil {
		return types.NewAPIError(types.NotFound, "no handler for action")
	}

	if ctx.Action.Input != nil {
		val := createRuntimeActionInput(ctx)
		if err := decodeBody(ctx.Request, val); err != nil {
			return err
		}
		ctx.Action.Input = val
	}
	result, err := handler(ctx)
	if err != nil {
		return err
	}

	WriteResponse(ctx, http.StatusOK, result)
	return nil
}

func createRuntimeActionInput(ctx *types.Context) interface{} {
	val := reflect.ValueOf(ctx.Action.Input)
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

func parseCreateBody(ctx *types.Context) ([]byte, *types.APIError) {
	var params struct {
		Yaml string `json:"yaml_"`
	}

	reqBody, err := ioutil.ReadAll(ctx.Request.Body)
	defer ctx.Request.Body.Close()
	if err != nil {
		return nil, types.NewAPIError(types.InvalidBodyContent,
			fmt.Sprintf("Failed to read request body: %v", err.Error()))
	}

	if err := json.Unmarshal(reqBody, &params); err != nil {
		return nil, types.NewAPIError(types.InvalidBodyContent,
			fmt.Sprintf("Failed to parse request body: %v", err.Error()))
	}

	raw := make(map[string]interface{})
	if err := json.Unmarshal(reqBody, &raw); err != nil {
		return nil, types.NewAPIError(types.InvalidBodyContent,
			fmt.Sprintf("Failed to parse request body: %v as a map", err.Error()))
	}
	schema := ctx.Object.GetSchema()
	if err := schema.ResourceFields.CheckRequired(raw); err != nil {
		return nil, types.NewAPIError(types.InvalidBodyContent, err.Error())
	}
	schema.ResourceFields.FillDefault(raw)
	reqBody, _ = json.Marshal(raw)

	if err := json.Unmarshal(reqBody, ctx.Object); err != nil {
		return nil, types.NewAPIError(types.InvalidBodyContent,
			fmt.Sprintf("Failed to parse request body: %v", err.Error()))
	}

	if err := schema.ResourceFields.Validate(ctx.Object); err != nil {
		return nil, types.NewAPIError(types.InvalidBodyContent, err.Error())
	}

	return []byte(params.Yaml), nil
}
