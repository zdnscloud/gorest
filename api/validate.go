package api

import (
	"fmt"
	"net/http"

	"github.com/zdnscloud/gorest/parse"
	"github.com/zdnscloud/gorest/types"
)

const (
	csrfCookie = "CSRF"
	csrfHeader = "X-API-CSRF"
)

func ValidateAction(apiContext *types.APIContext, method string) (*types.Action, *types.APIError) {
	urlAction := parse.ParseAction(apiContext.Request.URL)
	if urlAction == "" || method != http.MethodPost {
		return nil, nil
	}

	actions := apiContext.Obj.GetSchema().CollectionActions
	if apiContext.Obj.GetID() != "" {
		actions = apiContext.Obj.GetSchema().ResourceActions
	}

	action, ok := actions[urlAction]
	if !ok {
		return nil, types.NewAPIError(types.InvalidAction, fmt.Sprintf("Invalid action: %s", urlAction))
	}

	return &action, nil
}
