package handler

import (
	"net/http"

	"github.com/zdnscloud/gorest/httperror"
	"github.com/zdnscloud/gorest/types"
)

func UpdateHandler(apiContext *types.APIContext, next types.RequestHandler) error {
	data, err := ParseAndValidateBody(apiContext, false)
	if err != nil {
		return err
	}

	store := apiContext.Schema.Store
	if store == nil {
		return httperror.NewAPIError(httperror.NotFound, "no store found")
	}

	data, err = store.Update(apiContext, apiContext.Schema, data, apiContext.ID)
	if err != nil {
		return err
	}

	apiContext.WriteResponse(http.StatusOK, data)
	return nil
}
