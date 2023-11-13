package utils

import (
	"testing"

	"github.com/elliotchance/pie/v2"
	"github.com/stretchr/testify/assert"
)

func ContainsObjectNamed(t *testing.T, objects []map[string]any, name string) {
	assert.NotEqual(t, -1, pie.FindFirstUsing(objects, func(object map[string]any) bool {
		return object["name"].(string) == name
	}))
}
