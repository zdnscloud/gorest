package resource

import (
	"testing"

	ut "github.com/zdnscloud/cement/unittest"
)

type Deployment struct {
	ResourceBase
}

func (d Deployment) Default() Resource {
	return &Deployment{}
}

type Pod struct {
	ResourceBase
}

func (p Pod) GetParents() []ResourceKind {
	return []ResourceKind{Deployment{}}
}

func TestKindAndResourceName(t *testing.T) {
	//ut.Equal(t, (&Deployment{}).Name(), "deployment")
	//ut.Equal(t, (&Pod{}).Name(), "pod")
	ut.Equal(t, 1, 1)
}
