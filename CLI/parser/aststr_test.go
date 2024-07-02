package parser

import (
	"cli/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathNode(t *testing.T) {
	path := models.PhysicalPath + "site/building/room/rack/../rack2/"
	valNode := pathNode{path: &valueNode{path}}
	value, err := valNode.Path()
	assert.Nil(t, err)
	assert.Equal(t, path, value)

	translatedPath, err := valNode.execute()
	assert.Nil(t, err)
	assert.Equal(t, models.PhysicalPath+"site/building/room/rack2", translatedPath)
}

func TestFormatStringNodeExecute(t *testing.T) {
	vals := []node{&valueNode{3}, &valueNode{4}, &valueNode{7}}
	valNode := formatStringNode{&valueNode{"%d + %d = %d"}, vals}
	value, err := valNode.execute()
	assert.Nil(t, err)
	assert.Equal(t, "3 + 4 = 7", value)

	vector := []float64{1, 2, 3}
	vals = []node{&valueNode{vector}}
	valNode = formatStringNode{&valueNode{"%v"}, vals}
	value, err = valNode.execute()
	assert.Nil(t, err)
	assert.Equal(t, vector, value)
}
