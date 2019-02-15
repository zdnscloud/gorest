gorest
========

An rest server by golang 

## Building

`make`

## Example

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/zdnscloud/gorest/api"
	"github.com/zdnscloud/gorest/types"
)

type Foo struct {
	types.Resource
	Name     string `json:"name"`
	Foo      string `json:"foo"`
	SubThing Baz    `json:"subThing"`
}

type Baz struct {
	Name string `json:"name"`
}

var (
	version = types.APIVersion{
		Version: "v1",
		Group:   "io.cattle.core.example",
		Path:    "/example/v1",
	}

	Schemas = types.NewSchemas()
)

func main() {
	Schemas.MustImport(&version, Foo{})

	server := api.NewAPIServer()
    if err := server.AddSchemas(Schemas); err != nil {
        panic(err.Error())
    }

	fmt.Println("Listening on 0.0.0.0:1234")
	http.ListenAndServe("0.0.0.0:1234", server)
}
```
