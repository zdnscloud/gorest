package adaptor

import (
	"path"

	"github.com/gin-gonic/gin"
	"github.com/zdnscloud/gorest/api"
	"github.com/zdnscloud/gorest/types"
)

func RegisterHandler(router gin.IRoutes, server *api.Server) {
	handlerFunc := gin.WrapH(server)
	for _, schema := range server.Schemas.Schemas() {
		url := path.Join("/"+schema.Version.Group, schema.Version.Path, schema.ID)
		router.POST(url, handlerFunc)
		router.POST(path.Join(url, ":id"), handlerFunc)
		router.DELETE(path.Join(url, ":id"), handlerFunc)
		router.DELETE(url, handlerFunc)
		router.PUT(path.Join(url, ":id"), handlerFunc)
		router.GET(url, handlerFunc)
		router.GET(path.Join(url, ":id"), handlerFunc)
	}
}

func GetApiServer(version *types.APIVersion, obj interface{}, f func(*types.Schema)) (*api.Server, error) {
	server := api.NewAPIServer()
	if err := server.AddSchemas(types.NewSchemas().MustImportAndCustomize(version, obj, f)); err != nil {
		return nil, err
	}

	return server, nil
}
