package utils

import "github.com/elliotchance/pie/v2"

func SliceRemove[T comparable](ss []T, elem T) []T {
	filtered := pie.Filter(ss, func(oldElem T) bool {
		return oldElem != elem
	})
	if filtered == nil {
		return []T{}
	}

	return filtered
}
