// Package helpers contains some tools that are not related with any specific domain but required
// to perform some basic computational operations
package helper

import (
	"sort"
	"strings"
)

// maximum length of a metadata label in Kubernetes
const maxLabelLength = 63

func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func RemoveAllStrings(slice []string, search string) []string {
	for i, v := range slice {
		if v == search {
			return RemoveAllStrings(append(slice[:i], slice[i+1:]...), search)
		}
	}
	return slice
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

// KeySorted returns the map key-value pairs sorted by Key
func KeySorted(set map[string]string) [][2]string {
	vals := make([][2]string, 0, len(set))
	for k, v := range set {
		vals = append(vals, [2]string{k, v})
	}
	sort.Slice(vals, func(i, j int) bool {
		return vals[i][0] < vals[j][0]
	})
	return vals
}

// MaxLabelLength cuts an input string it ifs length is largen than 63, the maximum length allowed
// by Kubernetes metadata
func MaxLabelLength(in string) string {
	if len(in) <= maxLabelLength {
		return in
	}
	return in[:maxLabelLength]
}
