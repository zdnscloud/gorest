package schema

import (
	"net/http"
	"sort"
	"testing"

	ut "github.com/zdnscloud/cement/unittest"
	"github.com/zdnscloud/gorest/resource"
)

var version = resource.APIVersion{
	Group:   "testing",
	Version: "v1",
}

type Cluster struct {
	resource.ResourceBase
	Name string
}

type Node struct {
	resource.ResourceBase
	Name string
}

func (c Node) GetParents() []resource.ResourceKind {
	return []resource.ResourceKind{Cluster{}}
}

type NameSpace struct {
	resource.ResourceBase
	Name string
}

func (c NameSpace) GetParents() []resource.ResourceKind {
	return []resource.ResourceKind{Cluster{}}
}

type Deployment struct {
	resource.ResourceBase
	Name string
}

func (c Deployment) GetParents() []resource.ResourceKind {
	return []resource.ResourceKind{NameSpace{}}
}

type DaemonSet struct {
	resource.ResourceBase
	Name string
}

func (c DaemonSet) GetParents() []resource.ResourceKind {
	return []resource.ResourceKind{NameSpace{}}
}

type StatefulSet struct {
	resource.ResourceBase
	Name string
}

func (c StatefulSet) GetParents() []resource.ResourceKind {
	return []resource.ResourceKind{NameSpace{}}
}

type Pod struct {
	resource.ResourceBase
	Name string
}

func (c Pod) GetParents() []resource.ResourceKind {
	return []resource.ResourceKind{Deployment{}, DaemonSet{}, StatefulSet{}}
}

func createSchemaManager(t *testing.T) *SchemaManager {
	mgr := NewSchemaManager()
	handler, _ := resource.HandlerAdaptor(&resource.DumbHandler{})
	resourceKinds := []resource.ResourceKind{
		Cluster{},
		Node{},
		NameSpace{},
		Deployment{},
		StatefulSet{},
		DaemonSet{},
		Pod{},
	}
	for _, kind := range resourceKinds {
		err := mgr.Import(&version, kind, handler)
		ut.Assert(t, err == nil, "import get err %v", err)
	}
	return mgr
}

func TestGenerateResourceRoute(t *testing.T) {
	mgr := createSchemaManager(t)
	expectGetAndPostPaths := []string{
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

	expectDeleteAndPutPaths := []string{
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
	sort.StringSlice(expectGetAndPostPaths).Sort()
	sort.StringSlice(expectDeleteAndPutPaths).Sort()
	for method, urls := range mgr.GenerateResourceRoute() {
		sort.StringSlice(urls).Sort()
		if method == http.MethodGet || method == http.MethodPost {
			ut.Equal(t, urls, expectGetAndPostPaths)
		} else {
			ut.Equal(t, urls, expectDeleteAndPutPaths)
		}
	}
}

func TestCreateResourceFromRequest(t *testing.T) {
	mgr := createSchemaManager(t)

	invalidUrls := []string{
		"/apis/testings/v1/clusters/c1/namespaces/n1/deployments/d1/pods/p1",
		"/apis/testing/v2/clusters/c1/namespaces/n1/deployments/d1/pods/p1",
		"/apis/testing/v1/clusters/c1/namespaces/n1/deployments/d1/p1",
		"/apis/testing/v1/clusters/c1/namespacess/n1/deployments/d1/p1",
		"/apis/testing/v1/clusters/c1/deployments/d1/p1",
		"/apis/testing/v1/clusters/c1/namespacess/deployments/d1/p1",
	}
	for _, url := range invalidUrls {
		req, _ := http.NewRequest(http.MethodGet, url, nil)
		_, err := mgr.CreateResourceFromRequest(req)
		ut.Assert(t, err != nil, "")
	}

	validCases := []struct {
		url         string
		self        string
		parentIds   []string
		parentKinds []string
	}{
		{"/apis/testing/v1/clusters/c1/namespaces/n1/deployments/d1/pods/p1", "p1", []string{"d1", "n1", "c1"}, []string{"deployment", "namespace", "cluster"}},
		{"/apis/testing/v1/clusters/c1/namespaces/n1/statefulsets/s1/pods/p2", "p2", []string{"s1", "n1", "c1"}, []string{"statefulset", "namespace", "cluster"}},
		{"/apis/testing/v1/clusters/c1/namespaces/n2/daemonsets/d1/pods/p3", "p3", []string{"d1", "n2", "c1"}, []string{"daemonset", "namespace", "cluster"}},
		{"/apis/testing/v1/clusters/c1/namespaces/n2", "n2", []string{"c1"}, []string{"cluster"}},
		{"/apis/testing/v1/clusters/c1/namespaces", "", []string{"c1"}, []string{"cluster"}},
		{"/apis/testing/v1/clusters/c1", "c1", []string{}, nil},
	}
	for _, tc := range validCases {
		req, _ := http.NewRequest(http.MethodGet, tc.url, nil)
		r, err := mgr.CreateResourceFromRequest(req)
		ut.Equal(t, r.GetID(), tc.self)
		ut.Assert(t, err == nil, "")
		parent := r.GetParent()
		if len(tc.parentIds) == 0 {
			ut.Assert(t, parent == nil, "")
		} else {
			for i, parentId := range tc.parentIds {
				ut.Equal(t, parent.GetID(), parentId)
				ut.Equal(t, parent.GetSchema().(*Schema).ResourceKindName(), tc.parentKinds[i])
				parent = parent.GetParent()
			}
		}
	}
}

func TestGenerateLink(t *testing.T) {
	cases := []struct {
		url   string
		links map[resource.ResourceLinkType]resource.ResourceLink
	}{
		{
			"/apis/testing/v1/clusters/c1/namespaces/n1/deployments/d1/pods/p1",
			map[resource.ResourceLinkType]resource.ResourceLink{
				resource.SelfLink:       resource.ResourceLink("http:/127.0.0.1:5555/apis/testing/v1/clusters/c1/namespaces/n1/deployments/d1/pods/p1"),
				resource.UpdateLink:     resource.ResourceLink("http:/127.0.0.1:5555/apis/testing/v1/clusters/c1/namespaces/n1/deployments/d1/pods/p1"),
				resource.RemoveLink:     resource.ResourceLink("http:/127.0.0.1:5555/apis/testing/v1/clusters/c1/namespaces/n1/deployments/d1/pods/p1"),
				resource.CollectionLink: resource.ResourceLink("http:/127.0.0.1:5555/apis/testing/v1/clusters/c1/namespaces/n1/deployments/d1/pods"),
			},
		},

		{
			"/apis/testing/v1/clusters/c1/namespaces/n1",
			map[resource.ResourceLinkType]resource.ResourceLink{
				resource.SelfLink:                         resource.ResourceLink("http:/127.0.0.1:5555/apis/testing/v1/clusters/c1/namespaces/n1"),
				resource.UpdateLink:                       resource.ResourceLink("http:/127.0.0.1:5555/apis/testing/v1/clusters/c1/namespaces/n1"),
				resource.RemoveLink:                       resource.ResourceLink("http:/127.0.0.1:5555/apis/testing/v1/clusters/c1/namespaces/n1"),
				resource.CollectionLink:                   resource.ResourceLink("http:/127.0.0.1:5555/apis/testing/v1/clusters/c1/namespaces"),
				resource.ResourceLinkType("deployments"):  resource.ResourceLink("http:/127.0.0.1:5555/apis/testing/v1/clusters/c1/namespaces/n1/deployments"),
				resource.ResourceLinkType("daemonsets"):   resource.ResourceLink("http:/127.0.0.1:5555/apis/testing/v1/clusters/c1/namespaces/n1/daemonsets"),
				resource.ResourceLinkType("statefulsets"): resource.ResourceLink("http:/127.0.0.1:5555/apis/testing/v1/clusters/c1/namespaces/n1/statefulsets"),
			},
		},
		{
			"/apis/testing/v1/clusters/c1/namespaces",
			map[resource.ResourceLinkType]resource.ResourceLink{
				resource.SelfLink: resource.ResourceLink("http:/127.0.0.1:5555/apis/testing/v1/clusters/c1/namespaces"),
			},
		},
	}
	mgr := createSchemaManager(t)
	for _, tc := range cases {
		req, _ := http.NewRequest(http.MethodGet, tc.url, nil)
		r, _ := mgr.CreateResourceFromRequest(req)
		links, err := r.GetSchema().(*Schema).GenerateLinks(r, "http://127.0.0.1:5555")
		ut.Assert(t, err == nil, "")
		ut.Equal(t, links, tc.links)
	}
}
