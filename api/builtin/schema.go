package builtin

import (
	"github.com/zdnscloud/gorest/store/schema"
	"github.com/zdnscloud/gorest/types"
)

var (
	Version = types.APIVersion{
		Group:   "meta.cattle.io",
		Version: "v1",
		Path:    "/meta",
	}

	Schema = types.Schema{
		ID:                "schema",
		PluralName:        "schemas",
		Version:           Version,
		CollectionMethods: []string{"GET"},
		ResourceMethods:   []string{"GET"},
		ResourceFields: map[string]types.Field{
			"collectionActions": {Type: "map[json]"},
			"collectionFields":  {Type: "map[json]"},
			"collectionFilters": {Type: "map[json]"},
			"collectionMethods": {Type: "array[string]"},
			"pluralName":        {Type: "string"},
			"resourceActions":   {Type: "map[json]"},
			"resourceFields":    {Type: "map[json]"},
			"resourceMethods":   {Type: "array[string]"},
			"version":           {Type: "map[json]"},
		},
		Store: schema.NewSchemaStore(),
	}

	Error = types.Schema{
		ID:                "error",
		Version:           Version,
		ResourceMethods:   []string{},
		CollectionMethods: []string{},
		ResourceFields: map[string]types.Field{
			"code":      {Type: "string"},
			"detail":    {Type: "string", Nullable: true},
			"message":   {Type: "string", Nullable: true},
			"fieldName": {Type: "string", Nullable: true},
			"status":    {Type: "int"},
		},
	}

	Collection = types.Schema{
		ID:                "collection",
		Version:           Version,
		ResourceMethods:   []string{},
		CollectionMethods: []string{},
		ResourceFields: map[string]types.Field{
			"data":       {Type: "array[json]"},
			"pagination": {Type: "map[json]"},
			"sort":       {Type: "map[json]"},
			"filters":    {Type: "map[json]"},
		},
	}

	APIRoot = types.Schema{
		ID:                "apiRoot",
		Version:           Version,
		CollectionMethods: []string{"GET"},
		ResourceMethods:   []string{"GET"},
		ResourceFields: map[string]types.Field{
			"apiVersion": {Type: "map[json]"},
			"path":       {Type: "string"},
		},
		Formatter: APIRootFormatter,
		Store:     NewAPIRootStore(nil),
	}

	Schemas = types.NewSchemas().
		AddSchema(Schema).
		AddSchema(Error).
		AddSchema(Collection).
		AddSchema(APIRoot)
)

func contains(list []string, needle string) bool {
	for _, v := range list {
		if v == needle {
			return true
		}
	}
	return false
}
