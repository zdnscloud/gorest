package handler

import (
	"net/url"

	"github.com/zdnscloud/gorest/httperror"
	"github.com/zdnscloud/gorest/types"
)

func ErrorHandler(request *types.APIContext, err error) {
	var error *httperror.APIError
	if apiError, ok := err.(*httperror.APIError); ok {
		if apiError.Cause != nil {
			url, _ := url.PathUnescape(request.Request.URL.String())
			if url == "" {
				url = request.Request.URL.String()
			}
		}
		error = apiError
	} else {
		error = &httperror.APIError{
			Code:    httperror.ServerError,
			Message: err.Error(),
		}
	}

	data := toError(error)
	request.WriteResponse(error.Code.Status, data)
}

func toError(apiError *httperror.APIError) map[string]interface{} {
	e := map[string]interface{}{
		"type":    "/meta/schemas/error",
		"status":  apiError.Code.Status,
		"code":    apiError.Code.Code,
		"message": apiError.Message,
	}
	if apiError.FieldName != "" {
		e["fieldName"] = apiError.FieldName
	}

	return e
}
