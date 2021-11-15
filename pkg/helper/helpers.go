// Package helpers contains some tools that are not related with any specific domain but required
// to perform some basic computational operations
package helper

func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
