package schema

import (
	"github.com/zdnscloud/gorest/resource"
	"testing"

	ut "github.com/zdnscloud/cement/unittest"
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

	Name                  string
	Count                 uint32
	Annotations           map[string]string
	OtherInfo             OtherPodInfo
	OtherInfoSlice        []OtherPodInfo
	OtherInfoPointer      *OtherPodInfo
	OtherInfoPointerSlice []*OtherPodInfo
}

type OtherPodInfo struct {
	Name    string
	Numbers []uint32
}

func (c Pod) GetParents() []resource.ResourceKind {
	return []resource.ResourceKind{Deployment{}, DaemonSet{}, StatefulSet{}}
}

type Location struct {
	NodeName string `json:"nodeName"`
}

func (c Pod) CreateAction(name string) *resource.Action {
	if name == "move" {
		return &resource.Action{
			Name:  "move",
			Input: &Location{},
		}
	}
	return nil
}

func (c Pod) CreateDefaultResource() resource.Resource {
	return &Pod{
		Count: 20,
		OtherInfo: OtherPodInfo{
			Name:    "other",
			Numbers: []uint32{1, 2, 3},
		},
		OtherInfoSlice: []OtherPodInfo{
			OtherPodInfo{
				Name:    "other",
				Numbers: []uint32{1, 2, 3},
			},
		},
		OtherInfoPointer: &OtherPodInfo{
			Name:    "other",
			Numbers: []uint32{1, 2, 3},
		},
		OtherInfoPointerSlice: []*OtherPodInfo{
			&OtherPodInfo{
				Name:    "other",
				Numbers: []uint32{1, 2, 3},
			},
		},
	}
}

func PodEqual(t *testing.T, c, p *Pod) {
	ut.Equal(t, c.Name, p.Name)
	ut.Equal(t, c.Count, p.Count)
	ut.Equal(t, c.Annotations, p.Annotations)
	ut.Equal(t, c.OtherInfo, p.OtherInfo)
	ut.Equal(t, c.OtherInfoSlice, p.OtherInfoSlice)
	ut.Equal(t, c.OtherInfoPointer, p.OtherInfoPointer)
	ut.Equal(t, c.OtherInfoPointerSlice, p.OtherInfoPointerSlice)
}

func createSchemaManager() *SchemaManager {
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
		if err != nil {
			panic("test data isn't correct:" + err.Error())
		}
	}
	return mgr
}
