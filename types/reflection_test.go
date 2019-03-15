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
	Name string
}

type Node struct {
	Resource
	Name string
}

type NameSpace struct {
	Resource
	Name string
}

type Pod struct {
	Resource
	Name string
}

type Container struct {
	Resource
	Name string
}

type ConfigMap struct {
	Resource
	Name string
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
	schemas.MustImportAndCustomize(&version, NameSpace{}, nil, func(schema *Schema, handler Handler) {
		schema.Parent = GetResourceType(Cluster{})
		schema.CollectionMethods = []string{"GET", "POST"}
		schema.ResourceMethods = []string{"GET", "DELETE", "PUT"}
	})
	schemas.MustImportAndCustomize(&version, Pod{}, nil, func(schema *Schema, handler Handler) {
		schema.Parent = GetResourceType(NameSpace{})
		schema.CollectionMethods = []string{"GET", "POST"}
		schema.ResourceMethods = []string{"GET", "DELETE", "PUT"}
	})
	schemas.MustImportAndCustomize(&version, Container{}, nil, func(schema *Schema, handler Handler) {
		schema.Parent = GetResourceType(Pod{})
		schema.CollectionMethods = []string{"GET", "POST"}
		schema.ResourceMethods = []string{"GET", "DELETE", "PUT"}
	})
	schemas.MustImportAndCustomize(&version, ConfigMap{}, nil, func(schema *Schema, handler Handler) {
		schema.Parent = GetResourceType(NameSpace{})
		schema.CollectionMethods = []string{"GET", "POST"}
		schema.ResourceMethods = []string{"GET", "DELETE", "PUT"}
	})

	clusterChildren := schemas.GetChildren(GetResourceType(Cluster{}))
	ut.Equal(t, clusterChildren[GetResourceType(Node{})], "nodes")
	ut.Equal(t, clusterChildren[GetResourceType(NameSpace{})], "namespaces")

	schema := schemas.Schema(&version, GetResourceType(Node{}))
	ut.Equal(t, schema.ID, GetResourceType(Node{}))
	ut.Equal(t, schema.PluralName, "nodes")
	ut.Equal(t, schema.Version.Group, "testing")
	ut.Equal(t, schema.Version.Path, "/v1")
	ut.Equal(t, schema.Parent, GetResourceType(Cluster{}))
	ut.Equal(t, schema.CollectionMethods, []string{"GET", "POST"})
	ut.Equal(t, schema.ResourceMethods, []string{"GET", "DELETE", "PUT"})
	ut.Equal(t, len(schema.ResourceFields), 3)

	expectUrl := []string{
		"/apis/testing/v1/clusters",
		"/apis/testing/v1/clusters/:cluster_id",
		"/apis/testing/v1/clusters/:cluster_id/nodes",
		"/apis/testing/v1/clusters/:cluster_id/nodes/:node_id",
		"/apis/testing/v1/clusters/:cluster_id/namespaces",
		"/apis/testing/v1/clusters/:cluster_id/namespaces/:namespace_id",
		"/apis/testing/v1/clusters/:cluster_id/namespaces/:namespace_id/pods",
		"/apis/testing/v1/clusters/:cluster_id/namespaces/:namespace_id/pods/:pod_id",
		"/apis/testing/v1/clusters/:cluster_id/namespaces/:namespace_id/configmaps",
		"/apis/testing/v1/clusters/:cluster_id/namespaces/:namespace_id/configmaps/:configmap_id",
		"/apis/testing/v1/clusters/:cluster_id/namespaces/:namespace_id/pods/:pod_id/containers",
		"/apis/testing/v1/clusters/:cluster_id/namespaces/:namespace_id/pods/:pod_id/containers/:container_id",
	}
	urlMethods := schemas.UrlMethods()
	ut.Equal(t, len(urlMethods), len(expectUrl))
	for _, url := range expectUrl {
		ut.Equal(t, len(urlMethods[url]) != 0, true)
	}
}
