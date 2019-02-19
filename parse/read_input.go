package parse

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/zdnscloud/gorest/httperror"
	"github.com/zdnscloud/gorest/parse/yaml"
)

const reqMaxSize = (2 * 1 << 20) + 1

var bodyMethods = map[string]bool{
	http.MethodPut:  true,
	http.MethodPost: true,
}

type Decode func(interface{}) error

func ReadBody(req *http.Request) (map[string]interface{}, error) {
	if !bodyMethods[req.Method] {
		return nil, nil
	}

	decode := GetDecoder(req, io.LimitReader(req.Body, MaxFormSize))

	data := map[string]interface{}{}
	if err := decode(&data); err != nil {
		return nil, httperror.NewAPIError(httperror.InvalidBodyContent,
			fmt.Sprintf("Failed to parse body: %v", err))
	}

	return data, nil
}

func GetDecoder(req *http.Request, reader io.Reader) Decode {
	if req.Header.Get("Content-type") == "application/yaml" {
		return yaml.NewYAMLToJSONDecoder(reader).Decode
	}
	decoder := json.NewDecoder(reader)
	decoder.UseNumber()
	return decoder.Decode
}
