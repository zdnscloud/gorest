package builtin

import (
	"github.com/zdnscloud/gorest/types"
)

var (
	Version = types.APIVersion{
		Group:   "zcloud.cn",
		Version: "v1",
		Path:    "/meta",
	}

	Schema = types.Schema{
		ID:         "schema",
		PluralName: "schemas",
		Version:    Version,
	}

	Error = types.Schema{
		ID:      "error",
		Version: Version,
		ResourceFields: map[string]types.Field{
			"code":      {Type: "string"},
			"message":   {Type: "string"},
			"fieldName": {Type: "string"},
			"status":    {Type: "int"},
		},
	}
)
