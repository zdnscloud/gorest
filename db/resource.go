package db

import (
	"github.com/zdnscloud/cement/reflector"
	"github.com/zdnscloud/cement/stringtool"
)

const (
	IDField         = "id"
	CreateTimeField = "create_time"
)

type Resource interface {
}

type ResourceType string

func ResourceDBType(r Resource) ResourceType {
	n, _ := reflector.StructName(r)
	return ResourceType(stringtool.ToSnake(n))
}
