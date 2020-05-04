package db

import (
	"github.com/zdnscloud/cement/reflector"
	"github.com/zdnscloud/cement/stringtool"
	"github.com/zdnscloud/gorest/resource"
)

const (
	IDField         = "id"
	CreateTimeField = "create_time"
)

type ResourceType string

func ResourceDBType(r resource.Resource) ResourceType {
	n, _ := reflector.StructName(r)
	return ResourceType(stringtool.ToSnake(n))
}
