package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArithNodeExecute(t *testing.T) {
	left := 10
	right := 5
	results := map[string]any{
		"+": 15,
		"-": 5,
		"*": 50,
	}
	for op, result := range results {
		// apply operand with int values
		valNode := arithNode{op, &valueNode{left}, &valueNode{right}}
		value, err := valNode.execute()
		assert.Nil(t, err)
		assert.Equal(t, result, value)

		// apply operand with one float value
		valNode = arithNode{op, &valueNode{float64(left)}, &valueNode{right}}
		value, err = valNode.execute()
		assert.Nil(t, err)
		assert.Equal(t, float64(result.(int)), value)
	}

	valNode := arithNode{"/", &valueNode{10}, &valueNode{4}}
	value, err := valNode.execute()
	assert.Nil(t, err)
	assert.Equal(t, 2.5, value)

	valNode = arithNode{"/", &valueNode{10}, &valueNode{0}}
	_, err = valNode.execute()
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "cannot divide by 0")

	valNode = arithNode{"/", &valueNode{10.0}, &valueNode{4}}
	value, err = valNode.execute()
	assert.Nil(t, err)
	assert.Equal(t, 2.5, value)

	valNode = arithNode{"/", &valueNode{10.0}, &valueNode{0}}
	_, err = valNode.execute()
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "cannot divide by 0")

	// integer division
	valNode = arithNode{"\\", &valueNode{10}, &valueNode{4}}
	value, err = valNode.execute()
	assert.Nil(t, err)
	assert.Equal(t, 2, value)

	valNode = arithNode{"\\", &valueNode{10}, &valueNode{0}}
	_, err = valNode.execute()
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "cannot divide by 0")

	valNode = arithNode{"%", &valueNode{10}, &valueNode{4}}
	value, err = valNode.execute()
	assert.Nil(t, err)
	assert.Equal(t, 2, value)

	valNode = arithNode{"%", &valueNode{10.0}, &valueNode{4}}
	_, err = valNode.execute()
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "invalid operator for float operands")

	valNode = arithNode{"/%", &valueNode{10}, &valueNode{4}}
	_, err = valNode.execute()
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "invalid operator for integer operands")
}

func TestNegateNodeExecute(t *testing.T) {
	valNode := negateNode{&valueNode{10}}
	value, err := valNode.execute()
	assert.Nil(t, err)
	assert.Equal(t, -10, value)

	valNode = negateNode{&valueNode{10.0}}
	value, err = valNode.execute()
	assert.Nil(t, err)
	assert.Equal(t, -10.0, value)

	valNode = negateNode{&valueNode{true}}
	_, err = valNode.execute()
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "cannot negate non numeric value")
}
