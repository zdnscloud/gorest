package factory

import (
	"github.com/zdnscloud/gorest/types"
	"github.com/zdnscloud/gorest/types/mapper"
	//"k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ObjectMeta struct{}

func Schemas(version *types.APIVersion) *types.Schemas {
	s := types.NewSchemas()
	s.DefaultMappers = func() []types.Mapper {
		return []types.Mapper{
			mapper.NewObject(),
		}
	}
	s.DefaultPostMappers = func() []types.Mapper {
		return []types.Mapper{
			&mapper.RenameReference{},
		}
	}
	s.AddMapperForType(version, ObjectMeta{}, mapper.NewMetadataMapper())
	return s
}
