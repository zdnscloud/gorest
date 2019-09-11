package types

import (
	"testing"

	ut "github.com/zdnscloud/cement/unittest"
)

var version = APIVersion{
	Group:   "testing",
	Version: "v1",
}

type Cluster struct {
	Resource
	Name string
}

type Node struct {
	Resource
	Name string
}

func (c Node) GetParents() []string {
	return []string{GetResourceType(Cluster{})}
}

type NameSpace struct {
	Resource
	Name string
}

func (c NameSpace) GetParents() []string {
	return []string{GetResourceType(Cluster{})}
}

type Deployment struct {
	Resource
	Name string
}

func (c Deployment) GetParents() []string {
	return []string{GetResourceType(NameSpace{})}
}

type DaemonSet struct {
	Resource
	Name string
}

func (c DaemonSet) GetParents() []string {
	return []string{GetResourceType(NameSpace{})}
}

type StatefulSet struct {
	Resource
	Name string
}

func (c StatefulSet) GetParents() []string {
	return []string{GetResourceType(NameSpace{})}
}

type Pod struct {
	Resource
	Name string
}

func (c Pod) GetParents() []string {
	return []string{GetResourceType(Deployment{}), GetResourceType(DaemonSet{}), GetResourceType(StatefulSet{})}
}

type handleNoAction struct{}

func (h *handleNoAction) Create(ctx *Context, content []byte) (interface{}, *APIError) {
	return 10, nil
}

func (h *handleNoAction) Delete(ctx *Context) *APIError {
	return nil
}

func (h *handleNoAction) Update(ctx *Context) (interface{}, *APIError) {
	return 20, nil
}

func (h *handleNoAction) List(ctx *Context) interface{} {
	return []uint{1, 2, 3}
}

func (h *handleNoAction) Get(ctx *Context) interface{} {
	return 10
}

func TestReflection(t *testing.T) {
	schemas := NewSchemas()
	schemas.MustImport(&version, Cluster{}, &handleNoAction{})
	schemas.MustImport(&version, Node{}, &handleNoAction{})
	schemas.MustImport(&version, NameSpace{}, &handleNoAction{})
	schemas.MustImport(&version, Pod{}, &handleNoAction{})
	schemas.MustImport(&version, Deployment{}, &handleNoAction{})
	schemas.MustImport(&version, StatefulSet{}, &handleNoAction{})
	schemas.MustImport(&version, DaemonSet{}, &handleNoAction{})

	clusterChildren := schemas.GetChildren(GetResourceType(Cluster{}))
	ut.Equal(t, len(clusterChildren), 2)

	schema := schemas.Schema(&version, GetResourceType(Node{}))
	ut.Equal(t, schema.GetType(), GetResourceType(Node{}))
	ut.Equal(t, schema.PluralName, "nodes")
	ut.Equal(t, schema.Version.Group, "testing")
	ut.Equal(t, schema.Version.Version, "v1")
	ut.Equal(t, schema.Parents, []string{GetResourceType(Cluster{})})
	//ut.Equal(t, schema.CollectionMethods, []string{"GET", "POST"})
	//ut.Equal(t, schema.ResourceMethods, []string{"GET", "DELETE", "PUT"})
	//ut.Equal(t, len(schema.ResourceFields), 3)

	expectUrl := []string{
		"/apis/testing/v1/clusters",
		"/apis/testing/v1/clusters/:cluster_id",
		"/apis/testing/v1/clusters/:cluster_id/nodes",
		"/apis/testing/v1/clusters/:cluster_id/nodes/:node_id",
		"/apis/testing/v1/clusters/:cluster_id/namespaces",
		"/apis/testing/v1/clusters/:cluster_id/namespaces/:namespace_id",
		"/apis/testing/v1/clusters/:cluster_id/namespaces/:namespace_id/deployments",
		"/apis/testing/v1/clusters/:cluster_id/namespaces/:namespace_id/deployments/:deployment_id",
		"/apis/testing/v1/clusters/:cluster_id/namespaces/:namespace_id/daemonsets",
		"/apis/testing/v1/clusters/:cluster_id/namespaces/:namespace_id/daemonsets/:daemonset_id",
		"/apis/testing/v1/clusters/:cluster_id/namespaces/:namespace_id/statefulsets",
		"/apis/testing/v1/clusters/:cluster_id/namespaces/:namespace_id/statefulsets/:statefulset_id",
		"/apis/testing/v1/clusters/:cluster_id/namespaces/:namespace_id/deployments/:deployment_id/pods",
		"/apis/testing/v1/clusters/:cluster_id/namespaces/:namespace_id/deployments/:deployment_id/pods/:pod_id",
		"/apis/testing/v1/clusters/:cluster_id/namespaces/:namespace_id/daemonsets/:daemonset_id/pods",
		"/apis/testing/v1/clusters/:cluster_id/namespaces/:namespace_id/daemonsets/:daemonset_id/pods/:pod_id",
		"/apis/testing/v1/clusters/:cluster_id/namespaces/:namespace_id/statefulsets/:statefulset_id/pods",
		"/apis/testing/v1/clusters/:cluster_id/namespaces/:namespace_id/statefulsets/:statefulset_id/pods/:pod_id",
	}
	urlMethods := schemas.UrlMethods()
	ut.Equal(t, len(urlMethods), len(expectUrl))
	for _, url := range expectUrl {
		ut.Equal(t, len(urlMethods[url]) != 0, true)
	}
}
