package types

import (
	"net/http"

	"github.com/zdnscloud/gorest/util/slice"
)

func (s *Schema) CanList(context *APIContext) *APIError {
	if context == nil {
		if slice.ContainsString(s.CollectionMethods, http.MethodGet) {
			return nil
		}
		return NewAPIError(PermissionDenied, "can not list "+s.ID)
	}
	return context.AccessControl.CanList(context, s)
}

func (s *Schema) CanGet(context *APIContext) *APIError {
	if context == nil {
		if slice.ContainsString(s.ResourceMethods, http.MethodGet) {
			return nil
		}
		return NewAPIError(PermissionDenied, "can not get "+s.ID)
	}
	return context.AccessControl.CanGet(context, s)
}

func (s *Schema) CanCreate(context *APIContext) *APIError {
	if context == nil {
		if slice.ContainsString(s.CollectionMethods, http.MethodPost) {
			return nil
		}
		return NewAPIError(PermissionDenied, "can not create "+s.ID)
	}
	return context.AccessControl.CanCreate(context, s)
}

func (s *Schema) CanUpdate(context *APIContext) *APIError {
	if context == nil {
		if slice.ContainsString(s.ResourceMethods, http.MethodPut) {
			return nil
		}
		return NewAPIError(PermissionDenied, "can not update "+s.ID)
	}
	return context.AccessControl.CanUpdate(context, s)
}

func (s *Schema) CanDelete(context *APIContext) *APIError {
	if context == nil {
		if slice.ContainsString(s.ResourceMethods, http.MethodDelete) {
			return nil
		}
		return NewAPIError(PermissionDenied, "can not delete "+s.ID)
	}
	return context.AccessControl.CanDelete(context, s)
}
