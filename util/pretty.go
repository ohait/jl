package util

import (
	"encoding/json"
	"regexp"
	"strings"
)

var reJson = regexp.MustCompile(`{.*}`)            // as much json as you can
var reLong = regexp.MustCompile(`[^\n]{80}\S*? +`) // a long line, with no /n

// expands any json
func Prettify(s string) []string {
	s = strings.ReplaceAll(s, "\n", "␍\n")
	s = reJson.ReplaceAllStringFunc(s, func(in string) string {
		var v interface{}
		err := json.Unmarshal([]byte(in), &v)
		if err != nil {
			// multiple json?
			//log.Printf("multiple json? %s", s)
			return in
		}
		j, _ := json.MarshalIndent(v, "", "  ")
		return string(j)
	})
	s = reLong.ReplaceAllStringFunc(s, func(in string) string {
		return strings.TrimRight(in, " ") + "\n"
	})
	s = strings.TrimRight(s, "\n ")
	s = strings.TrimLeft(s, "\n ")
	ss := strings.Split(s, "\n")
	for i := range ss {
		ss[i] = strings.ReplaceAll(ss[i], "␍", "\n")
	}
	return ss
}
