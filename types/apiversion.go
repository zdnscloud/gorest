package types

import (
	"path"
)

type APIVersion struct {
	Group   string `json:"group,omitempty"`
	Version string `json:"version,omitempty"`
}

func (v *APIVersion) GetVersionURL() string {
	return path.Join(GroupPrefix, v.Group, v.Version)
}

func (v *APIVersion) Equals(other *APIVersion) bool {
	return v.Group == other.Group && v.Version == other.Version
}
