package authorization

import (
	"net/http"

	"github.com/zdnscloud/gorest/types"
	"github.com/zdnscloud/gorest/util/slice"
)

type AllAccess struct {
}

func (*AllAccess) CanCreate(apiContext *types.APIContext, schema *types.Schema) *types.APIError {
	if slice.ContainsString(schema.CollectionMethods, http.MethodPost) {
		return nil
	}
	return types.NewAPIError(types.PermissionDenied, "can not create "+schema.ID)
}

func (*AllAccess) CanGet(apiContext *types.APIContext, schema *types.Schema) *types.APIError {
	if slice.ContainsString(schema.ResourceMethods, http.MethodGet) {
		return nil
	}
	return types.NewAPIError(types.PermissionDenied, "can not get "+schema.ID)
}

func (*AllAccess) CanList(apiContext *types.APIContext, schema *types.Schema) *types.APIError {
	if slice.ContainsString(schema.CollectionMethods, http.MethodGet) {
		return nil
	}
	return types.NewAPIError(types.PermissionDenied, "can not list "+schema.ID)
}

func (*AllAccess) CanUpdate(apiContext *types.APIContext, schema *types.Schema) *types.APIError {
	if slice.ContainsString(schema.ResourceMethods, http.MethodPut) {
		return nil
	}
	return types.NewAPIError(types.PermissionDenied, "can not update "+schema.ID)
}

func (*AllAccess) CanDelete(apiContext *types.APIContext, schema *types.Schema) *types.APIError {
	if slice.ContainsString(schema.ResourceMethods, http.MethodDelete) {
		return nil
	}
	return types.NewAPIError(types.PermissionDenied, "can not delete "+schema.ID)
}
