package utils

import (
	"net/url"
	"strings"
)

func IsMappingEqual[T any](a, b map[string]T, eq func(a, b T) bool) bool {
	if len(a) != len(b) {
		return false
	}

	for k, av := range a {
		bv, ok := b[k]
		if !ok {
			return false
		}

		if !eq(av, bv) {
			return false
		}
	}

	return true
}

func ParseAddr(raw string) (scheme, service, method string, err error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", "", "", err
	}

	scheme = u.Scheme                        // consul / http / grpc ...
	service = u.Host                         // Greeter
	method = strings.TrimPrefix(u.Path, "/") // Echo

	return
}
