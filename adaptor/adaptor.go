package adaptor

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterHandler(router gin.IRoutes, handler http.Handler, urlMethods map[string][]string) {
	handlerFunc := gin.WrapH(handler)
	for method, urls := range urlMethods {
		switch method {
		case http.MethodPost:
			for _, url := range urls {
				router.POST(url, handlerFunc)
			}
		case http.MethodDelete:
			for _, url := range urls {
				router.DELETE(url, handlerFunc)
			}
		case http.MethodPut:
			for _, url := range urls {
				router.PUT(url, handlerFunc)
			}
		case http.MethodGet:
			for _, url := range urls {
				router.GET(url, handlerFunc)
			}
		}
	}
}
