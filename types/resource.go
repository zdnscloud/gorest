package types

import (
	"fmt"
	"path"
	"reflect"
	"strings"
	"time"
)

type Collection struct {
	Type         string            `json:"type,omitempty"`
	ResourceType string            `json:"resourceType,omitempty"`
	Links        map[string]string `json:"links,omitempty"`
	Data         interface{}       `json:"data"`
}

type APIVersion struct {
	Group   string `json:"group,omitempty"`
	Version string `json:"version,omitempty"`
}

func (v *APIVersion) GetVersionURL() string {
	return path.Join(GroupPrefix, v.Group, v.Version)
}

type RequestHandler func(request *Context) *APIError

func GetResourceType(obj interface{}) string {
	return strings.ToLower(reflect.TypeOf(obj).Name())
}

type ResourceType interface {
	GetParents() []string
	GetActions() []Action
	GetCollectionActions() []Action
}

type Resource struct {
	ID                string            `json:"id,omitempty"`
	Type              string            `json:"type,omitempty"`
	Links             map[string]string `json:"links,omitempty"`
	CreationTimestamp ISOTime           `json:"creationTimestamp,omitempty"`

	Parent Object  `json:"-"`
	Schema *Schema `json:"-"`
}

func (r Resource) GetActions() []Action {
	return nil
}

func (r Resource) GetCollectionActions() []Action {
	return nil
}

func (r Resource) GetParents() []string {
	return nil
}

func (r *Resource) GetID() string {
	return r.ID
}

func (r *Resource) SetID(id string) {
	r.ID = id
}

func (r *Resource) GetType() string {
	return r.Type
}

func (r *Resource) SetType(typ string) {
	r.Type = typ
}

func (r *Resource) GetLinks() map[string]string {
	return r.Links
}

func (r *Resource) SetLinks(links map[string]string) {
	r.Links = links
}

func (r *Resource) GetCreationTimestamp() time.Time {
	return time.Time(r.CreationTimestamp)
}

func (r *Resource) SetCreationTimestamp(timestamp time.Time) {
	r.CreationTimestamp = ISOTime(timestamp)
}

func (r *Resource) GetParent() Object {
	return r.Parent
}

func (r *Resource) SetParent(parent Object) {
	r.Parent = parent
}

func (r *Resource) GetSchema() *Schema {
	return r.Schema
}

func (r *Resource) SetSchema(schema *Schema) {
	r.Schema = schema
}

func GetAncestors(parent ObjectParent) []Object {
	var antiAncestors []Object
	for obj := parent.GetParent(); obj != nil; obj = obj.GetParent() {
		antiAncestors = append(antiAncestors, obj)
	}

	var ancestors []Object
	for i := len(antiAncestors) - 1; i >= 0; i-- {
		ancestors = append(ancestors, antiAncestors[i])
	}

	return ancestors
}

type ISOTime time.Time

func (t ISOTime) MarshalJSON() ([]byte, error) {
	if time.Time(t).IsZero() {
		return []byte("null"), nil
	}

	return []byte(fmt.Sprintf("\"%s\"", time.Time(t).Format(time.RFC3339))), nil
}

func (t *ISOTime) UnmarshalJSON(data []byte) (err error) {
	if len(data) == 4 && string(data) == "null" {
		*t = ISOTime(time.Time{})
		return
	}

	now, err := time.Parse(`"`+time.RFC3339+`"`, string(data))
	*t = ISOTime(now)
	return
}
