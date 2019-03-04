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

type Node struct {
	Resource
	Name string `singlecloud:"notnullable"`
}

func TestReflection(t *testing.T) {
	schemas := NewSchemas()
	schemas.MustImportAndCustomize(&version, Node{}, nil, func(schema *Schema, handler Handler) {
		schema.Parent = "cluster"
		schema.CollectionMethods = []string{"GET", "POST"}
		schema.ResourceMethods = []string{"GET", "DELETE", "PUT"}
	})

	schema := schemas.Schema(&version, "node")
	ut.Equal(t, schema.ID, "node")
	ut.Equal(t, schema.PluralName, "nodes")
	ut.Equal(t, schema.Version.Group, "testing")
	ut.Equal(t, schema.Version.Path, "/v1")
	ut.Equal(t, schema.Parent, "cluster")
	ut.Equal(t, schema.CollectionMethods, []string{"GET", "POST"})
	ut.Equal(t, schema.ResourceMethods, []string{"GET", "DELETE", "PUT"})
	ut.Equal(t, len(schema.ResourceFields), 3)
	for _, field := range schema.ResourceFields {
		if field.CodeName == "Name" {
			ut.Equal(t, field.Nullable, false)
		}
	}
}
