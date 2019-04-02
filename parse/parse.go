package parse

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/zdnscloud/gorest/types"
)

var (
	multiSlashRegexp = regexp.MustCompile("//+")
	allowedFormats   = map[string]bool{
		"json": true,
		"yaml": true,
	}
)

func Parse(rw http.ResponseWriter, req *http.Request, schemas *types.Schemas) (*types.APIContext, *types.APIError) {
	result := types.NewAPIContext(req, rw, schemas)
	path := req.URL.EscapedPath()
	path = multiSlashRegexp.ReplaceAllString(path, "/")
	obj, err := parseVersionAndResource(schemas, path)
	if err != nil {
		return result, err
	}

	result.Obj = obj
	if err := ValidateMethod(result); err != nil {
		return result, err
	}

	return result, nil
}

func versionsForPath(schemas *types.Schemas, escapedPath string) *types.APIVersion {
	for _, version := range schemas.Versions() {
		if strings.HasPrefix(escapedPath, version.GetVersionURL()) {
			return &version
		}
	}
	return nil
}

func parseVersionAndResource(schemas *types.Schemas, escapedPath string) (types.Object, *types.APIError) {
	version := versionsForPath(schemas, escapedPath)
	if version == nil {
		return nil, types.NewAPIError(types.NotFound, "no found version with "+escapedPath)
	}

	if strings.HasSuffix(escapedPath, "/") {
		escapedPath = escapedPath[:len(escapedPath)-1]
	}

	versionURL := version.GetVersionURL()
	if len(escapedPath) <= len(versionURL) {
		return nil, types.NewAPIError(types.InvalidFormat, "no schema name in url "+escapedPath)
	}

	escapedPath = escapedPath[len(versionURL)+1:]
	pp := strings.Split(escapedPath, "/")
	var paths []string
	for _, p := range pp {
		part, err := url.PathUnescape(p)
		if err == nil {
			paths = append(paths, part)
		} else {
			paths = append(paths, p)
		}
	}

	if len(paths) == 0 {
		return nil, types.NewAPIError(types.NotFound, "no found schema with url "+escapedPath)
	}

	var obj *types.Resource
	var schema *types.Schema
	for i := 0; i < len(paths); i += 2 {
		schema = schemas.Schema(version, paths[i])
		if schema == nil {
			return nil, types.NewAPIError(types.NotFound, "no found schema for "+paths[i])
		}

		if i == 0 {
			obj = &types.Resource{
				ID:     safeIndex(paths, i+1),
				Type:   schema.ID,
				Schema: schema,
			}
			continue
		}

		if schema.Parent != obj.Type {
			return nil, types.NewAPIError(types.InvalidType,
				fmt.Sprintf("schema %v parent should not be %s", schema.ID, obj.Type))
		}

		obj = &types.Resource{
			ID:     safeIndex(paths, i+1),
			Type:   schema.ID,
			Parent: obj,
			Schema: schema,
		}
	}

	return obj, nil
}

func safeIndex(slice []string, index int) string {
	if index >= len(slice) {
		return ""
	}
	return slice[index]
}

func ParseResponseFormat(req *http.Request) string {
	format := req.URL.Query().Get("_format")

	if format != "" {
		format = strings.TrimSpace(strings.ToLower(format))
	}

	/* Format specified */
	if allowedFormats[format] {
		return format
	}

	if isYaml(req) {
		return "yaml"
	}

	return "json"
}

func isYaml(req *http.Request) bool {
	return strings.Contains(req.Header.Get("Accept"), "application/yaml")
}

func ParseMethod(req *http.Request) string {
	method := req.URL.Query().Get("_method")
	if method == "" {
		method = req.Method
	}
	return method
}

func ParseAction(url *url.URL) string {
	return url.Query().Get("action")
}
