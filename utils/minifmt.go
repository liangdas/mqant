package mqant_tools

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	regFields = regexp.MustCompile(`\{(\w+)\}`)
	regField  = regexp.MustCompile(`[\{\}]`)
)

func Sprintf(format string, extra map[string]string) string {
	fields := regFields.FindAllString(format, -1)
	ret := format
	for _, fieldName := range fields {
		field := regField.ReplaceAllString(fieldName, "")
		if v, ok := extra[field]; !ok {

		} else {
			ret = strings.Replace(ret, fieldName, fmt.Sprintf("%v", v), 1)
		}
	}
	return ret
}
