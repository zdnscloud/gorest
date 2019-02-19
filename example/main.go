package main

import (
	"fmt"
	"net/http"

	"github.com/zdnscloud/gorest/api"
	"github.com/zdnscloud/gorest/api/crd"
	"github.com/zdnscloud/gorest/types"
)

var (
	version = types.APIVersion{
		Version: "v1",
		Group:   "io.cattle.core.example",
		Path:    "/example/v1",
	}

	Schemas = types.NewSchemas()
)

type Foo struct {
	types.Resource
	types.TypeMeta   `json:",inline"`
	types.ObjectMeta `json:"metadata,omitempty"`
	types.SpecMeta   `json:"spec,omitempty"`
}

func (foo *Foo) TypeMetaData() *types.TypeMeta {
	return &types.TypeMeta{
		Kind:       foo.Kind,
		APIVersion: foo.APIVersion,
	}
}

func (foo *Foo) ObjectMetaData() *types.ObjectMeta {
	return &types.ObjectMeta{
		Name:      foo.Name,
		Namespace: foo.Namespace,
		Labels:    foo.Labels,
	}
}

func (foo *Foo) Create() error {
	fmt.Printf("create foo %s in namespace %s\n", foo.Name, foo.Namespace)
	return nil
}

func (foo *Foo) Delete(name string, namespace string) error {
	fmt.Printf("delete foo %s in namespace %s\n", name, namespace)
	return nil
}

func (foo *Foo) Update(name string, namespace string) error {
	fmt.Printf("update foo %s in namespace %s\n", name, namespace)
	return nil
}

func (foo *Foo) List() []map[string]interface{} {
	fmt.Printf("get all foos\n")
	return nil
}

func (foo *Foo) Get(name string, namespace string) map[string]interface{} {
	fmt.Printf("get foo %s in namespace %s\n", name, namespace)
	return nil
}

func (foo *Foo) Action(action string) error {
	fmt.Printf("do action: %s\n", action)
	return nil
}

func main() {
	Schemas.MustImportAndCustomize(&version, Foo{}, func(schema *types.Schema) {
		crd.AssignHandler(schema)
	})

	server := api.NewAPIServer()
	if err := server.AddSchemas(Schemas); err != nil {
		panic(err.Error())
	}

	fmt.Println("Listening on 0.0.0.0:1234")
	http.ListenAndServe("0.0.0.0:1234", server)
}
