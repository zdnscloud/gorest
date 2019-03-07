package types

import (
	"testing"

	ut "github.com/zdnscloud/cement/unittest"
)

var version = APIVersion{
	Version: "v1",
	Group:   "testing",
	Path:    "/v1",
}

type Cluster struct {
	Resource
	Name string `singlecloud:"notnullable"`
}

type Node struct {
	Resource
	Name string `singlecloud:"notnullable"`
}

type Namespace struct {
	Resource
	Name string `singlecloud:"notnullable"`
}

type Pod struct {
	Resource
	Name string `singlecloud:"notnullable"`
}

type Container struct {
	Resource
	Name string `singlecloud:"notnullable"`
}

func TestReflection(t *testing.T) {
	schemas := NewSchemas()
	schemas.MustImportAndCustomize(&version, Cluster{}, nil, func(schema *Schema, handler Handler) {
		schema.CollectionMethods = []string{"GET", "POST"}
		schema.ResourceMethods = []string{"GET", "DELETE", "PUT"}
	})

	schemas.MustImportAndCustomize(&version, Node{}, nil, func(schema *Schema, handler Handler) {
		schema.Parent = GetResourceType(Cluster{})
		schema.CollectionMethods = []string{"GET", "POST"}
		schema.ResourceMethods = []string{"GET", "DELETE", "PUT"}
	})
	schemas.MustImportAndCustomize(&version, Namespace{}, nil, func(schema *Schema, handler Handler) {
		schema.Parent = GetResourceType(Cluster{})
		schema.CollectionMethods = []string{"GET", "POST"}
		schema.ResourceMethods = []string{"GET", "DELETE", "PUT"}
	})
	schemas.MustImportAndCustomize(&version, Pod{}, nil, func(schema *Schema, handler Handler) {
		schema.Parent = GetResourceType(Namespace{})
		schema.CollectionMethods = []string{"GET", "POST"}
		schema.ResourceMethods = []string{"GET", "DELETE", "PUT"}
	})
	schemas.MustImportAndCustomize(&version, Container{}, nil, func(schema *Schema, handler Handler) {
		schema.Parent = GetResourceType(Pod{})
		schema.CollectionMethods = []string{"GET", "POST"}
		schema.ResourceMethods = []string{"GET", "DELETE", "PUT"}
	})

	schema := schemas.Schema(&version, GetResourceType(Node{}))
	ut.Equal(t, schema.ID, GetResourceType(Node{}))
	ut.Equal(t, schema.PluralName, "nodes")
	ut.Equal(t, schema.Version.Group, "testing")
	ut.Equal(t, schema.Version.Path, "/v1")
	ut.Equal(t, schema.Parent, GetResourceType(Cluster{}))
	ut.Equal(t, schema.CollectionMethods, []string{"GET", "POST"})
	ut.Equal(t, schema.ResourceMethods, []string{"GET", "DELETE", "PUT"})
	ut.Equal(t, len(schema.ResourceFields), 5)
	for _, field := range schema.ResourceFields {
		if field.CodeName == "Name" {
			ut.Equal(t, field.Nullable, false)
		}
	}

	expectUrl := []string{
		"/testing/v1/clusters",
		"/testing/v1/clusters/:cluster_id",
		"/testing/v1/clusters/:cluster_id/nodes",
		"/testing/v1/clusters/:cluster_id/nodes/:node_id",
		"/testing/v1/clusters/:cluster_id/namespaces",
		"/testing/v1/clusters/:cluster_id/namespaces/:namespace_id",
		"/testing/v1/clusters/:cluster_id/namespaces/:namespace_id/pods",
		"/testing/v1/clusters/:cluster_id/namespaces/:namespace_id/pods/:pod_id",
		"/testing/v1/clusters/:cluster_id/namespaces/:namespace_id/pods/:pod_id/containers",
		"/testing/v1/clusters/:cluster_id/namespaces/:namespace_id/pods/:pod_id/containers/:container_id",
	}
	urlMethods := schemas.UrlMethods()
	ut.Equal(t, len(urlMethods), len(expectUrl))
	for _, url := range expectUrl {
		ut.Equal(t, len(urlMethods[url]) != 0, true)
	}
}
