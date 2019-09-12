package types

import (
	//"fmt"
	"net/http"
	"sort"
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

func (c Node) GetParents() []ResourceType {
	return []ResourceType{Cluster{}}
}

type NameSpace struct {
	Resource
	Name string
}

func (c NameSpace) GetParents() []ResourceType {
	return []ResourceType{Cluster{}}
}

type Deployment struct {
	Resource
	Name string
}

func (c Deployment) GetParents() []ResourceType {
	return []ResourceType{NameSpace{}}
}

type DaemonSet struct {
	Resource
	Name string
}

func (c DaemonSet) GetParents() []ResourceType {
	return []ResourceType{NameSpace{}}
}

type StatefulSet struct {
	Resource
	Name string
}

func (c StatefulSet) GetParents() []ResourceType {
	return []ResourceType{NameSpace{}}
}

type Pod struct {
	Resource
	Name string
}

func (c Pod) GetParents() []ResourceType {
	return []ResourceType{Deployment{}, DaemonSet{}, StatefulSet{}}
}

func TestReflection(t *testing.T) {
	schemas := NewSchemas()
	schemas.MustImport(&version, Cluster{}, &dumbHandler{})
	schemas.MustImport(&version, Node{}, &dumbHandler{})
	schemas.MustImport(&version, NameSpace{}, &dumbHandler{})
	schemas.MustImport(&version, Deployment{}, &dumbHandler{})
	schemas.MustImport(&version, StatefulSet{}, &dumbHandler{})
	schemas.MustImport(&version, DaemonSet{}, &dumbHandler{})
	schemas.MustImport(&version, Pod{}, &dumbHandler{})

	expectGetAndPostUrls := []string{
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

	expectDeleteAndPutUrls := []string{
		"/apis/testing/v1/clusters/:cluster_id",
		"/apis/testing/v1/clusters/:cluster_id/nodes/:node_id",
		"/apis/testing/v1/clusters/:cluster_id/namespaces/:namespace_id",
		"/apis/testing/v1/clusters/:cluster_id/namespaces/:namespace_id/deployments/:deployment_id",
		"/apis/testing/v1/clusters/:cluster_id/namespaces/:namespace_id/daemonsets/:daemonset_id",
		"/apis/testing/v1/clusters/:cluster_id/namespaces/:namespace_id/statefulsets/:statefulset_id",
		"/apis/testing/v1/clusters/:cluster_id/namespaces/:namespace_id/deployments/:deployment_id/pods/:pod_id",
		"/apis/testing/v1/clusters/:cluster_id/namespaces/:namespace_id/statefulsets/:statefulset_id/pods/:pod_id",
		"/apis/testing/v1/clusters/:cluster_id/namespaces/:namespace_id/daemonsets/:daemonset_id/pods/:pod_id",
	}
	urlMethods := schemas.UrlMethods()
	sort.StringSlice(expectGetAndPostUrls).Sort()
	sort.StringSlice(expectDeleteAndPutUrls).Sort()
	for method, urls := range urlMethods {
		sort.StringSlice(urls).Sort()
		if method == http.MethodGet || method == http.MethodPost {
			ut.Equal(t, urls, expectGetAndPostUrls)
		} else {
			ut.Equal(t, urls, expectDeleteAndPutUrls)
		}
	}
}
