// Package helpers contains some tools that are not related with any specific domain but required
// to perform some basic computational operations
package helper

import "strings"

func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func ExtractVersion(image string) string {
	parts := strings.Split(image, ":")
	nparts := len(parts)
	if nparts > 1 {
		return parts[nparts-1]
	}
	return "unknown"
}

// IsSubSet returns whether the first argument contains all the keys and values of the second
// argument
func IsSubSet(set, subset map[string]string) bool {
	for k, v := range subset {
		if sv, ok := set[k]; !ok || v != sv {
			return false
		}
	}
	return true
}
