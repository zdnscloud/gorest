package writer

import (
	"github.com/zdnscloud/gorest/types"
)

var (
	start = `
<!DOCTYPE html>
<!-- If you are reading this, there is a good chance you would prefer sending an
"Accept: application/json" header and receiving actual JSON responses. -->
<script>
var data =
`
	end = []byte(`</script>`)
)

type StringGetter func() string

type HTMLResponseWriter struct {
	EncodingResponseWriter
	CSSURL       StringGetter
	JSURL        StringGetter
	APIUIVersion StringGetter
}

func (h *HTMLResponseWriter) start(apiContext *types.APIContext, code int, obj interface{}) {
	AddCommonResponseHeader(apiContext)
	apiContext.Response.Header().Set("content-type", "text/html")
	apiContext.Response.WriteHeader(code)
}

func (h *HTMLResponseWriter) Write(apiContext *types.APIContext, code int, obj interface{}) {
	h.start(apiContext, code, obj)
	apiContext.Response.Write([]byte(start))
	h.Body(apiContext, apiContext.Response, obj)
	apiContext.Response.Write(end)
}
