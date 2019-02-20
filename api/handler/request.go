package handler

import (
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/zdnscloud/gorest/httperror"
	"github.com/zdnscloud/gorest/parse"
	"github.com/zdnscloud/gorest/types"
)

func CreateHandler(apiContext *types.APIContext, next types.RequestHandler) error {
	handler := apiContext.Schema.Handler
	if handler == nil {
		return httperror.NewAPIError(httperror.NotFound, "no handler found")
	}

	object, err := parseRequestBody(apiContext)
	if err != nil {
		return err
	}

	if err := handler.Create(object); err != nil {
		return err
	}

	apiContext.WriteResponse(http.StatusCreated, nil)
	return nil
}

func DeleteHandler(apiContext *types.APIContext, next types.RequestHandler) error {
	handler := apiContext.Schema.Handler
	if handler == nil {
		return httperror.NewAPIError(httperror.NotFound, "no handler found")
	}

	namespace, name := getNamespaceAndName(apiContext.ID)
	if err := handler.Delete(types.TypeMeta{Kind: apiContext.Schema.ID},
		types.ObjectMeta{Namespace: namespace, Name: name}); err != nil {
		return err
	}

	apiContext.WriteResponse(http.StatusCreated, nil)
	return nil
}

func UpdateHandler(apiContext *types.APIContext, next types.RequestHandler) error {
	handler := apiContext.Schema.Handler
	if handler == nil {
		return httperror.NewAPIError(httperror.NotFound, "no handler found")
	}

	object, err := parseRequestBody(apiContext)
	if err != nil {
		return err
	}

	namespace, name := getNamespaceAndName(apiContext.ID)
	if err := handler.Update(types.TypeMeta{Kind: apiContext.Schema.ID},
		types.ObjectMeta{Namespace: namespace, Name: name}, object); err != nil {
		return err
	}

	apiContext.WriteResponse(http.StatusCreated, nil)
	return nil
}

func ListHandler(apiContext *types.APIContext, next types.RequestHandler) error {
	handler := apiContext.Schema.Handler
	if handler == nil {
		return httperror.NewAPIError(httperror.NotFound, "no handler found")
	}

	var result interface{}
	if apiContext.ID == "" {
		result = handler.List()
	} else {
		namespace, name := getNamespaceAndName(apiContext.ID)
		result = handler.Get(types.TypeMeta{Kind: apiContext.Schema.ID},
			types.ObjectMeta{Namespace: namespace, Name: name})
	}

	apiContext.WriteResponse(http.StatusCreated, result)
	return nil
}

func ActionHandler(actionName string, action *types.Action, apiContext *types.APIContext) error {
	handler := apiContext.Schema.Handler
	if handler == nil {
		return httperror.NewAPIError(httperror.NotFound, "no handler found")
	}

	object, err := parseRequestBody(apiContext)
	if err != nil {
		return err
	}

	if err := handler.Action(apiContext.Action, nil, object); err != nil {
		return err
	}

	apiContext.WriteResponse(http.StatusCreated, nil)
	return nil
}

func getNamespaceAndName(id string) (string, string) {
	namespace := ""
	name := id
	namespaceAndName := strings.SplitN(id, ":", 2)
	if len(namespaceAndName) == 2 {
		namespace = namespaceAndName[0]
		name = namespaceAndName[1]
	}

	return namespace, name
}

func parseRequestBody(apiContext *types.APIContext) (types.Object, error) {
	val := apiContext.Schema.StructVal
	valPtr := reflect.New(val.Type())
	valPtr.Elem().Set(val)
	decode := parse.GetDecoder(apiContext.Request, io.LimitReader(apiContext.Request.Body, parse.MaxFormSize))
	if err := decode(valPtr.Interface()); err != nil {
		return nil, httperror.NewAPIError(httperror.InvalidBodyContent,
			fmt.Sprintf("Failed to parse body: %v", err))
	}

	if object, ok := valPtr.Interface().(types.Object); ok {
		object.SetTypeMeta(types.TypeMeta{
			Kind:       apiContext.Schema.ID,
			APIVersion: getAPIVersion(apiContext.Schema.Version)})
		return object, nil
	}

	return nil, httperror.NewAPIError(httperror.InvalidBodyContent,
		fmt.Sprintf("Failed trans object interface when parse request body"))
}

func getAPIVersion(version types.APIVersion) string {
	if version.Group != "" {
		return version.Group + "/" + version.Version
	} else {
		return version.Version
	}
}
