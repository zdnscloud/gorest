package empty

import (
	"github.com/zdnscloud/gorest/types"
)

type Handler struct {
}

func (s *Handler) Create(types.Object) error {
	return nil
}

func (s *Handler) Delete(types.TypeMeta, types.ObjectMeta) error {
	return nil
}

func (s *Handler) Update(types.TypeMeta, types.ObjectMeta, types.Object) error {
	return nil
}

func (s *Handler) List() interface{} {
	return nil
}

func (s *Handler) Get(types.TypeMeta, types.ObjectMeta) interface{} {
	return nil
}

func (s *Handler) Action(string, map[string]interface{}, types.Object) error {
	return nil
}

func (s *Handler) Watch(string, string, string) (chan map[string]interface{}, error) {
	return nil, nil
}
