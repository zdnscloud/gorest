package db

import (
	"fmt"

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
	n, err := reflector.StructName(r)
	if err != nil {
		panic(fmt.Sprintf("%v doesn't point to a struct implement resource:%s", r, err.Error()))
	}
	return ResourceType(stringtool.ToSnake(n))
}
