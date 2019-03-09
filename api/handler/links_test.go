package handler

import (
	"net/http"
	"testing"

	ut "github.com/zdnscloud/cement/unittest"
	"github.com/zdnscloud/gorest/types"
)

var version = types.APIVersion{
	Version: "v1",
	Group:   "testing",
	Path:    "/v1",
}

type Testresourceobjectparent struct {
	types.Resource
}

type Testresourceobject struct {
	types.Resource
}

type Testnoresourceobject struct {
	ID    string
	Type  string
	Links map[string]string
}

func TestAddResourceLink(t *testing.T) {
	expectSelfLink := "http://127.0.0.1:1234/test/testresourceobjects/1de5f1bb403524c280c220f3a366b538"
	expectCollectionLink := "http://127.0.0.1:1234/test/testresourceobjects"
	req, _ := http.NewRequest("POST", "/test/testresourceobjects", nil)
	req.Host = "127.0.0.1:1234"
	schemas := types.NewSchemas()
	schemas.MustImportAndCustomize(&version, Testresourceobjectparent{}, nil, func(schema *types.Schema, handler types.Handler) {
		schema.CollectionMethods = []string{"GET", "POST"}
		schema.ResourceMethods = []string{"GET", "DELETE", "PUT"}
	})

	schemas.MustImportAndCustomize(&version, Testresourceobject{}, nil, func(schema *types.Schema, handler types.Handler) {
		schema.Parent = types.GetResourceType(Testresourceobjectparent{})
		schema.CollectionMethods = []string{"GET", "POST"}
		schema.ResourceMethods = []string{"GET", "DELETE", "PUT"}
	})

	apiContext := &types.APIContext{
		Request: req,
		Schemas: schemas,
		Schema:  schemas.Schema(&version, types.GetResourceType(Testresourceobject{})),
	}

	obj := &Testresourceobject{
		types.Resource{
			ID:   "1de5f1bb403524c280c220f3a366b538",
			Type: "testResourceObject",
		},
	}
	addResourceLinks(apiContext, obj)
	ut.Equal(t, len(obj.Links), 4)
	ut.Equal(t, obj.Links["self"], expectSelfLink)
	ut.Equal(t, obj.Links["collection"], expectCollectionLink)
	ut.Equal(t, obj.Links["remove"], expectSelfLink)
	ut.Equal(t, obj.Links["update"], expectSelfLink)

	req, _ = http.NewRequest("PUT", "/test/testresourceobjects/1de5f1bb403524c280c220f3a366b538", nil)
	req.Host = "127.0.0.1:1234"
	apiContext.Request = req
	addResourceLinks(apiContext, obj)
	ut.Equal(t, len(obj.Links), 4)
	ut.Equal(t, obj.Links["self"], expectSelfLink)
	ut.Equal(t, obj.Links["collection"], expectCollectionLink)
	ut.Equal(t, obj.Links["remove"], expectSelfLink)
	ut.Equal(t, obj.Links["update"], expectSelfLink)

	expectSelfLink = "http://127.0.0.1:1234/test/resourceobjectparents/d6db994a406ab41c80dc6e4e31ecf890"
	expectCollectionLink = "http://127.0.0.1:1234/test/resourceobjectparents"
	expectTestObjectLink := "http://127.0.0.1:1234/test/resourceobjectparents/d6db994a406ab41c80dc6e4e31ecf890/testresourceobjects"

	req, _ = http.NewRequest("POST", "/test/resourceobjectparents", nil)
	req.Host = "127.0.0.1:1234"
	apiContext.Schema = schemas.Schema(&version, types.GetResourceType(Testresourceobjectparent{}))
	apiContext.Request = req
	objParent := &Testresourceobjectparent{
		types.Resource{
			ID:   "d6db994a406ab41c80dc6e4e31ecf890",
			Type: "testResourceObjectParent",
		},
	}

	addResourceLinks(apiContext, objParent)
	ut.Equal(t, len(objParent.Links), 5)
	ut.Equal(t, objParent.Links["self"], expectSelfLink)
	ut.Equal(t, objParent.Links["collection"], expectCollectionLink)
	ut.Equal(t, objParent.Links["testresourceobject"], expectTestObjectLink)
}

func TestAddLinkFail(t *testing.T) {
	req, _ := http.NewRequest("POST", "/test/testresourceobjects", nil)
	req.Host = "127.0.0.1:1234"
	apiContext := &types.APIContext{
		Request: req,
		Schema:  &types.Schema{},
	}

	obj := &Testnoresourceobject{
		ID:   "1de5f1bb403524c280c220f3a366b538",
		Type: "testnoresoureobject",
	}

	addResourceLinks(apiContext, obj)
	ut.Equal(t, len(obj.Links), 0)
	ut.Equal(t, obj.Links["self"], "")
}

func TestAddCollectionLinks(t *testing.T) {
	schemas := types.NewSchemas()
	schemas.MustImportAndCustomize(&version, Testresourceobjectparent{}, nil, func(schema *types.Schema, handler types.Handler) {
		schema.CollectionMethods = []string{"GET", "POST"}
		schema.ResourceMethods = []string{"GET", "DELETE", "PUT"}
	})

	schemas.MustImportAndCustomize(&version, Testresourceobject{}, nil, func(schema *types.Schema, handler types.Handler) {
		schema.Parent = types.GetResourceType(Testresourceobjectparent{})
		schema.CollectionMethods = []string{"GET", "POST"}
		schema.ResourceMethods = []string{"GET", "DELETE", "PUT"}
	})

	req, _ := http.NewRequest("GET", "/test/testresourceobjectparents", nil)
	req.Host = "127.0.0.1:1234"
	apiContext := &types.APIContext{
		Request: req,
		Schemas: schemas,
		Schema:  schemas.Schema(&version, types.GetResourceType(Testresourceobjectparent{})),
	}

	collection := &types.Collection{
		Type:         "collection",
		ResourceType: "testresourceobject",
		Data: []*Testresourceobjectparent{
			&Testresourceobjectparent{
				types.Resource{
					ID:   "1de5f1bb403524c280c220f3a366b538",
					Type: "testResourceObject",
				},
			},
			&Testresourceobjectparent{
				types.Resource{
					ID:   "0ad4bcfd408086438084f774097712d5",
					Type: "testResourceObject",
				},
			},
		},
	}
	expectCollectionLink := "http://127.0.0.1:1234/test/testresourceobjectparents"
	expectResourceLink1 := "http://127.0.0.1:1234/test/testresourceobjectparents/1de5f1bb403524c280c220f3a366b538"
	expectChildLink1 := "http://127.0.0.1:1234/test/testresourceobjectparents/1de5f1bb403524c280c220f3a366b538/testresourceobjects"
	expectResourceLink2 := "http://127.0.0.1:1234/test/testresourceobjectparents/0ad4bcfd408086438084f774097712d5"
	expectChildLink2 := "http://127.0.0.1:1234/test/testresourceobjectparents/0ad4bcfd408086438084f774097712d5/testresourceobjects"

	addCollectionLinks(apiContext, collection)
	ut.Equal(t, len(collection.Links), 1)
	ut.Equal(t, collection.Links["self"], expectCollectionLink)
	ut.Equal(t, len(collection.Data.([]*Testresourceobjectparent)), 2)
	ut.Equal(t, collection.Data.([]*Testresourceobjectparent)[0].Links["self"], expectResourceLink1)
	ut.Equal(t, collection.Data.([]*Testresourceobjectparent)[0].Links["collection"], expectCollectionLink)
	ut.Equal(t, collection.Data.([]*Testresourceobjectparent)[0].Links["testresourceobject"], expectChildLink1)
	ut.Equal(t, collection.Data.([]*Testresourceobjectparent)[1].Links["self"], expectResourceLink2)
	ut.Equal(t, collection.Data.([]*Testresourceobjectparent)[1].Links["collection"], expectCollectionLink)
	ut.Equal(t, collection.Data.([]*Testresourceobjectparent)[1].Links["testresourceobject"], expectChildLink2)
}
