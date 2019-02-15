package builtin

import (
	"github.com/zdnscloud/gorest/store/empty"
	"github.com/zdnscloud/gorest/types"
)

func APIRootFormatter(apiContext *types.APIContext, resource *types.RawResource) {
	path, _ := resource.Values["path"].(string)
	if path == "" {
		return
	}

	delete(resource.Values, "path")
}

type APIRootStore struct {
	empty.Store
	roots []string
}

func NewAPIRootStore(roots []string) types.Store {
	return &APIRootStore{roots: roots}
}

func (a *APIRootStore) ByID(apiContext *types.APIContext, schema *types.Schema, id string) (map[string]interface{}, error) {
	for _, version := range apiContext.Schemas.Versions() {
		if version.Path == id {
			return apiVersionToAPIRootMap(version), nil
		}
	}
	return nil, nil
}

func (a *APIRootStore) List(apiContext *types.APIContext, schema *types.Schema, opt *types.QueryOptions) ([]map[string]interface{}, error) {
	var roots []map[string]interface{}

	for _, version := range apiContext.Schemas.Versions() {
		roots = append(roots, apiVersionToAPIRootMap(version))
	}

	for _, root := range a.roots {
		roots = append(roots, map[string]interface{}{
			"path": root,
		})
	}

	return roots, nil
}

func apiVersionToAPIRootMap(version types.APIVersion) map[string]interface{} {
	return map[string]interface{}{
		"type": "/meta/schemas/apiRoot",
		"apiVersion": map[string]interface{}{
			"version": version.Version,
			"group":   version.Group,
			"path":    version.Path,
		},
		"path": version.Path,
	}
}
