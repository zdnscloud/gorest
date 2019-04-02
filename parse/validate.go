package parse

import (
	"fmt"
	"net/http"

	"github.com/zdnscloud/gorest/types"
	"github.com/zdnscloud/gorest/util"
)

var (
	supportedMethods = map[string]bool{
		http.MethodPost:   true,
		http.MethodGet:    true,
		http.MethodPut:    true,
		http.MethodDelete: true,
	}
)

func ValidateMethod(apiContext *types.Context) *types.APIError {
	method := ParseMethod(apiContext.Request)
	if !supportedMethods[method] {
		return types.NewAPIError(types.MethodNotAllowed, fmt.Sprintf("Method %s not supported", method))
	}

	schema := apiContext.Object.GetSchema()
	if apiContext.Object.GetType() == "" || schema == nil {
		return types.NewAPIError(types.NotFound, "no found schema")
	}

	allowed := schema.ResourceMethods
	if apiContext.Object.GetID() == "" {
		allowed = schema.CollectionMethods
	}

	if util.ContainsString(allowed, method) {
		return nil
	}

	return types.NewAPIError(types.MethodNotAllowed, fmt.Sprintf("Method %s not supported", method))
}
