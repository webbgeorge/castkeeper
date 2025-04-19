package fixures

import "strings"

func StrOfLen(n int) string {
	s := make([]string, 0)
	for range n {
		s = append(s, "a")
	}
	return strings.Join(s, "")
}
