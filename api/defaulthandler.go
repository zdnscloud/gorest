package api

import (
	"github.com/zdnscloud/gorest/types"
)

var _ types.Handler = &DefaultHandler{}

type DefaultHandler struct {
	CreateHandler types.CreateHandler
	DeleteHandler types.DeleteHandler
	UpdateHandler types.UpdateHandler
	ListHandler   types.ListHandler
	GetHandler    types.GetHandler
	ActionHandler types.ActionHandler
}

func (h *DefaultHandler) GetCreateHandler() types.CreateHandler {
	return h.CreateHandler
}

func (h *DefaultHandler) GetDeleteHandler() types.DeleteHandler {
	return h.DeleteHandler
}

func (h *DefaultHandler) GetUpdateHandler() types.UpdateHandler {
	return h.UpdateHandler
}

func (h *DefaultHandler) GetListHandler() types.ListHandler {
	return h.ListHandler
}

func (h *DefaultHandler) GetGetHandler() types.GetHandler {
	return h.GetHandler
}

func (h *DefaultHandler) GetActionHandler() types.ActionHandler {
	return h.ActionHandler
}
