package utils

import "strings"

func StartsWith(path string, prefix string, suffix *string) bool {
	if strings.HasPrefix(path, prefix) {
		if suffix != nil {
			*suffix = path[len(prefix):]
		}

		return true
	}

	return false
}
