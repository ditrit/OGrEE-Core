package unit

import (
	"testing"

	"github.com/elliotchance/pie/v2"
	"github.com/stretchr/testify/assert"
)

func ContainsObject(t *testing.T, objects []map[string]any, objectID string) {
	assert.NotEqual(
		t,
		-1,
		pie.FindFirstUsing(objects, func(object map[string]any) bool {
			return object["id"].(string) == objectID
		}),
		"%#v does not contain %#v", objects, objectID,
	)
}
