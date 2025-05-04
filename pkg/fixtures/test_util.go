package fixtures

import (
	"strings"

	"github.com/gofrs/uuid/v5"
)

func StrOfLen(n int) string {
	s := make([]string, 0)
	for range n {
		s = append(s, "a")
	}
	return strings.Join(s, "")
}

func PodEpGUID(s string) string {
	return uuid.NewV5(uuid.NamespaceOID, s).String()
}
