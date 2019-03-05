package handler

import (
	"net/http"
	"testing"

	ut "github.com/zdnscloud/cement/unittest"
	"github.com/zdnscloud/gorest/types"
)

type Testresourceobject struct {
	types.Resource
}

type Testnoresourceobject struct {
	ID    string
	Type  string
	Links map[string]string
}

func TestAddResourceLink(t *testing.T) {
	expectLink := "http://127.0.0.1:1234/test/testresourceobjects/1de5f1bb403524c280c220f3a366b538"
	req, _ := http.NewRequest("POST", "/test/testresourceobjects", nil)
	req.Host = "127.0.0.1:1234"
	apiContext := &types.APIContext{
		Request: req,
		Schema:  &types.Schema{},
	}

	obj := &Testresourceobject{
		types.Resource{
			ID:   "1de5f1bb403524c280c220f3a366b538",
			Type: "testResourceObject",
		},
	}
	addResourceLinks(apiContext, obj)
	ut.Equal(t, len(obj.Links), 1)
	ut.Equal(t, obj.Links["self"], expectLink)

	req, _ = http.NewRequest("PUT", "/test/testresourceobjects/1de5f1bb403524c280c220f3a366b538", nil)
	req.Host = "127.0.0.1:1234"
	apiContext.Request = req
	addResourceLinks(apiContext, obj)
	ut.Equal(t, len(obj.Links), 1)
	ut.Equal(t, obj.Links["self"], expectLink)

	expectLink = "http://127.0.0.1:1234/test/resourceobjectparent/d6db994a406ab41c80dc6e4e31ecf890/testresourceobjects/1de5f1bb403524c280c220f3a366b538"
	req, _ = http.NewRequest("POST", "/test/resourceobjectparent/d6db994a406ab41c80dc6e4e31ecf890/testresourceobjects", nil)
	req.Host = "127.0.0.1:1234"
	apiContext.Request = req
	addResourceLinks(apiContext, obj)
	ut.Equal(t, len(obj.Links), 1)
	ut.Equal(t, obj.Links["self"], expectLink)
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
	req, _ := http.NewRequest("GET", "/test/testresourceobjects", nil)
	req.Host = "127.0.0.1:1234"
	apiContext := &types.APIContext{
		Request: req,
		Schema:  &types.Schema{},
	}

	collection := &types.Collection{
		Type:         "collection",
		ResourceType: "testresourceobject",
		Data: []*Testresourceobject{
			&Testresourceobject{
				types.Resource{
					ID:   "1de5f1bb403524c280c220f3a366b538",
					Type: "testResourceObject",
				},
			},
			&Testresourceobject{
				types.Resource{
					ID:   "0ad4bcfd408086438084f774097712d5",
					Type: "testResourceObject",
				},
			},
		},
	}
	expectCollectionLink := "http://127.0.0.1:1234/test/testresourceobjects"
	expectResourceLink1 := "http://127.0.0.1:1234/test/testresourceobjects/1de5f1bb403524c280c220f3a366b538"
	expectResourceLink2 := "http://127.0.0.1:1234/test/testresourceobjects/0ad4bcfd408086438084f774097712d5"

	addCollectionLinks(apiContext, collection)
	ut.Equal(t, len(collection.Links), 1)
	ut.Equal(t, collection.Links["self"], expectCollectionLink)
	ut.Equal(t, len(collection.Data.([]*Testresourceobject)), 2)
	ut.Equal(t, collection.Data.([]*Testresourceobject)[0].Links["self"], expectResourceLink1)
	ut.Equal(t, collection.Data.([]*Testresourceobject)[1].Links["self"], expectResourceLink2)
}
