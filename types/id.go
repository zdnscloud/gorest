package types

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/zdnscloud/gorest/types/rand"
)

var (
	lowerChars = regexp.MustCompile("[a-z]+")
)

func GenerateName(typeName string) string {
	base := typeName[0:1] + lowerChars.ReplaceAllString(typeName[1:], "")
	last := rand.String(5)
	return fmt.Sprintf("%s-%s", strings.ToLower(base), last)
}
