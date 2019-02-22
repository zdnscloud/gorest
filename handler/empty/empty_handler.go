package empty

import (
	"github.com/zdnscloud/gorest/types"
)

type Handler struct {
}

func (s *Handler) Create(types.Object) (interface{}, error) {
	return nil, nil
}

func (s *Handler) BatchCreate([]types.Object) (interface{}, error) {
	return nil, nil
}

func (s *Handler) Delete(types.Object) error {
	return nil
}

func (s *Handler) BatchDelete(types.ObjectType) error {
	return nil
}

func (s *Handler) Update(types.ObjectType, types.ObjectID, types.Object) (interface{}, error) {
	return nil, nil
}

func (s *Handler) BatchUpdate([]types.Object) (interface{}, error) {
	return nil, nil
}

func (s *Handler) List(types.ObjectType) interface{} {
	return nil
}

func (s *Handler) Get(types.Object) interface{} {
	return nil
}

func (s *Handler) Action(types.Object, string, map[string]interface{}) (interface{}, error) {
	return nil, nil
}
