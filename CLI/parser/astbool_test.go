package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEqualityNodeExecute(t *testing.T) {
	valNode := equalityNode{"==", &valueNode{5}, &valueNode{6}}
	value, err := valNode.execute()

	assert.Nil(t, err)
	assert.False(t, value.(bool))

	valNode = equalityNode{"==", &valueNode{5}, &valueNode{5}}
	value, err = valNode.execute()

	assert.Nil(t, err)
	assert.True(t, value.(bool))

	valNode = equalityNode{"!=", &valueNode{5}, &valueNode{6}}
	value, err = valNode.execute()

	assert.Nil(t, err)
	assert.True(t, value.(bool))

	valNode = equalityNode{"=", &valueNode{5}, &valueNode{5}}
	_, err = valNode.execute()

	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "Invalid equality node operator : =")
}

func TestComparatorNodeExecute(t *testing.T) {
	valNode := comparatorNode{"<", &valueNode{5}, &valueNode{5}}
	value, err := valNode.execute()
	assert.Nil(t, err)
	assert.False(t, value.(bool))

	valNode = comparatorNode{"<=", &valueNode{5}, &valueNode{5}}
	value, err = valNode.execute()
	assert.Nil(t, err)
	assert.True(t, value.(bool))

	valNode = comparatorNode{">", &valueNode{6}, &valueNode{5}}
	value, err = valNode.execute()
	assert.Nil(t, err)
	assert.True(t, value.(bool))

	valNode = comparatorNode{">=", &valueNode{6}, &valueNode{5}}
	value, err = valNode.execute()
	assert.Nil(t, err)
	assert.True(t, value.(bool))

	valNode = comparatorNode{">!", &valueNode{6}, &valueNode{5}}
	_, err = valNode.execute()
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "Invalid comparison operator : >!")
}

func TestLogicalNodeExecute(t *testing.T) {
	valNode := logicalNode{"||", &valueNode{false}, &valueNode{true}}
	value, err := valNode.execute()
	assert.Nil(t, err)
	assert.True(t, value.(bool))

	valNode = logicalNode{"&&", &valueNode{false}, &valueNode{true}}
	value, err = valNode.execute()
	assert.Nil(t, err)
	assert.False(t, value.(bool))

	valNode = logicalNode{"&|", &valueNode{false}, &valueNode{true}}
	_, err = valNode.execute()
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "Invalid logical operator : &|")
}

func TestNegateBoolNodeExecute(t *testing.T) {
	valNode := negateBoolNode{&valueNode{"true"}}
	value, err := valNode.execute()
	assert.Nil(t, err)
	assert.False(t, value.(bool))

	valNode = negateBoolNode{&valueNode{"false"}}
	value, err = valNode.execute()
	assert.Nil(t, err)
	assert.True(t, value.(bool))

	valNode = negateBoolNode{&valueNode{"3.5"}}
	_, err = valNode.execute()
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "expression should be a boolean")
}
