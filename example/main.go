package main

import (
	"fmt"
	"net/http"

	"github.com/zdnscloud/gorest/api"
	"github.com/zdnscloud/gorest/types"
)

var (
	version = types.APIVersion{
		Version: "v1",
		Group:   "zdns.cloud.example",
		Path:    "/example/v1",
	}
)

type Foo struct {
	types.TypeMeta   `json:",inline"`
	types.ObjectMeta `json:"metadata,omitempty"`
	Spec             `json:"spec,omitempty"`
}

type Spec struct{}

func (foo *Foo) TypeMetaData() types.TypeMeta {
	return foo.TypeMeta
}

func (foo *Foo) ObjectMetaData() types.ObjectMeta {
	return foo.ObjectMeta
}

func (foo *Foo) SetTypeMeta(meta types.TypeMeta) {
	foo.TypeMeta = meta
}

type Handler struct{}

func (s *Handler) Create(obj types.Object) error {
	fmt.Printf("create %s %s in namespace %s with apiVersion %s\n",
		obj.(*Foo).Kind, obj.(*Foo).Name, obj.(*Foo).Namespace, obj.(*Foo).APIVersion)
	return nil
}

func (s *Handler) Delete(typeMeta types.TypeMeta, objMeta types.ObjectMeta) error {
	fmt.Printf("delete %s %s in namespace %s \n", typeMeta.Kind, objMeta.Name, objMeta.Namespace)
	return nil
}

func (s *Handler) Update(typeMeta types.TypeMeta, objMeta types.ObjectMeta, obj types.Object) error {
	fmt.Printf("update %s %s in namespace %s \n", typeMeta.Kind, objMeta.Name, objMeta.Namespace)
	return nil
}

func (s *Handler) List() interface{} {
	fmt.Printf("get all foos\n")
	return nil
}

func (s *Handler) Get(typeMeta types.TypeMeta, objMeta types.ObjectMeta) interface{} {
	fmt.Printf("get %s %s in namespace %s \n", typeMeta.Kind, objMeta.Name, objMeta.Namespace)
	return nil
}

func (s *Handler) Action(action string, params map[string]interface{}, obj types.Object) error {
	fmt.Printf("do action: %s\n", action)
	return nil
}

func main() {
	schemas := types.NewSchemas().MustImportAndCustomize(&version, Foo{}, func(schema *types.Schema) {
		schema.Handler = &Handler{}
		schema.CollectionMethods = []string{"GET", "POST"}
		schema.ResourceMethods = []string{"GET", "PUT", "DELETE"}
	})

	server := api.NewAPIServer()
	if err := server.AddSchemas(schemas); err != nil {
		panic(err.Error())
	}

	http.ListenAndServe("0.0.0.0:1234", server)
}
