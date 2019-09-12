package types

import (
	"fmt"
	"net/url"
	"strings"
)

type VersionedSchemas struct {
	version *APIVersion
	//instead generate from version very time
	//to optimize search performance
	versionUrl      string
	toplevelSchemas []*Schema
}

func NewVersionedSchemas(version *APIVersion) *VersionedSchemas {
	return &VersionedSchemas{
		version:    version,
		versionUrl: version.GetUrl(),
	}
}

func (s *VersionedSchemas) VersionEquals(v *APIVersion) bool {
	return s.version.Equals(v)
}

func (s *VersionedSchemas) ImportResource(obj ResourceType, objHandler interface{}) error {
	schema, err := newSchema(s.version, obj, objHandler)
	if err != nil {
		return err
	}

	parents := obj.GetParents()
	for _, parent := range parents {
		parentSchema := s.GetSchema(parent)
		if parentSchema == nil {
			return fmt.Errorf("%s who is parent of %s hasn't been imported", resourceTypeDebugName(parent), resourceTypeDebugName(obj))
		} else {
			if err := parentSchema.AddChild(schema); err != nil {
				return err
			}
		}
	}

	if len(parents) == 0 {
		return s.addTopleveSchema(schema)
	}

	return nil
}

func (s *VersionedSchemas) CreateResourceFromUrl(path string) (Object, *APIError) {
	if strings.HasPrefix(path, s.versionUrl) == false {
		return nil, nil
	}

	path = strings.TrimPrefix(path, s.versionUrl)

	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1] //get rid of last '/'
	}

	if len(path) == 0 {
		return nil, NewAPIError(InvalidFormat, "no schema name in url")
	} else {
		path = path[1:] //get rid of first '/'
	}

	segments := strings.Split(path, "/")
	for i, segment := range segments {
		if seg, err := url.PathUnescape(segment); err == nil {
			segments[i] = seg
		}
	}

	segmentCount := len(segments)
	if segmentCount == 0 {
		return nil, NewAPIError(NotFound, "only api version without any resource")
	}

	for _, schema := range s.toplevelSchemas {
		if obj, err := schema.CreateResourceFromUrlSegments(nil, segments); err != nil {
			return nil, err
		} else if obj != nil {
			return obj, nil
		}
	}
	return nil, NewAPIError(NotFound, fmt.Sprintf("no resource with collection name %s", segments[0]))
}

func (s *VersionedSchemas) addTopleveSchema(schema *Schema) error {
	for _, old := range s.toplevelSchemas {
		if old.equals(schema) {
			return fmt.Errorf("duplicate import type %s", schema.resourceType.Name())
		}
	}
	s.toplevelSchemas = append(s.toplevelSchemas, schema)
	return nil
}

func (s *VersionedSchemas) GetSchema(resource ResourceType) *Schema {
	for _, schema := range s.toplevelSchemas {
		if target := schema.GetSchema(resource); target != nil {
			return target
		}
	}
	return nil
}

func (s *VersionedSchemas) GenUrls() map[string][]string {
	urls := make(map[string][]string)
	for _, schema := range s.toplevelSchemas {
		urls = mergeUrls(urls, schema.GenUrls(nil))
	}
	return urls
}
