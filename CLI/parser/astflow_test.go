package parser

import (
	"cli/controllers"
	test_utils "cli/test"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIfNodeExecute(t *testing.T) {
	// we execute the if condition and returns nil
	valNode := ifNode{&valueNode{true}, &valueNode{5}, &valueNode{3}}
	value, err := valNode.execute()
	assert.Nil(t, err)
	assert.Nil(t, value)

	// we execute the if condition which is invalid, so it returns error
	valNode = ifNode{&valueNode{true}, &negateBoolNode{&valueNode{"3.5"}}, &valueNode{3}}
	_, err = valNode.execute()
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "expression should be a boolean")

	// we execute the else condition which is invalid, so it returns error
	valNode = ifNode{&valueNode{false}, &valueNode{3}, &negateBoolNode{&valueNode{"3.5"}}}
	_, err = valNode.execute()
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "expression should be a boolean")
}

func TestWhileNodeExecute(t *testing.T) {
	test_utils.SetMainEnvironmentMock(t)

	variable := &assignNode{"i", &valueNode{"0"}}
	variable.execute()
	// "while $i<6 {var: i = eval $i+1}"
	valNode := whileNode{
		condition: &comparatorNode{"<", &symbolReferenceNode{"i"}, &valueNode{6}},
		body:      &assignNode{"i", &arithNode{"+", &symbolReferenceNode{"i"}, &valueNode{1}}},
	}
	value, err := valNode.execute()
	assert.Nil(t, err)
	assert.Nil(t, value)
	assert.Contains(t, controllers.State.DynamicSymbolTable, "i")
	assert.Equal(t, 6, controllers.State.DynamicSymbolTable["i"])
}

func TestForNodeExecute(t *testing.T) {
	test_utils.SetMainEnvironmentMock(t)

	// "for i=0;i<10;i=i+1 { print $i }"
	valNode := forNode{
		init:        &assignNode{"i", &valueNode{"0"}},
		condition:   &comparatorNode{"<", &symbolReferenceNode{"i"}, &valueNode{10}},
		incrementor: &assignNode{"i", &arithNode{"+", &symbolReferenceNode{"i"}, &valueNode{1}}},
		body:        &printNode{&formatStringNode{&valueNode{"%v"}, []node{&symbolReferenceNode{"i"}}}},
	}
	value, err := valNode.execute()
	assert.Nil(t, err)
	assert.Nil(t, value)

	assert.Contains(t, controllers.State.DynamicSymbolTable, "i")
	assert.Equal(t, 10, controllers.State.DynamicSymbolTable["i"])
}

// ToDo: enable this test once the forArrayNode.execute is fixed
// func TestForArrayNodeExecute(t *testing.T) {
// 	oldValue := controllers.State.DynamicSymbolTable
// 	controllers.State.DynamicSymbolTable = map[string]any{}

// 	// "for i in [1,3,5] { print $i }"
// 	array := arrNode{[]node{&valueNode{"a"}, &valueNode{"b"}, &valueNode{"c"}}}

// 	valNode := forArrayNode{
// 		variable: "i",
// 		arr:      &array,
// 		body:     &printNode{&formatStringNode{&valueNode{"%v"}, []node{&symbolReferenceNode{"i"}}}},
// 	}
// 	value, err := valNode.execute()
// 	assert.Nil(t, err)
// 	assert.Nil(t, value)

// 	assert.Contains(t, controllers.State.DynamicSymbolTable, "i")
// 	assert.Equal(t, "c", controllers.State.DynamicSymbolTable["i"])
// 	controllers.State.DynamicSymbolTable = oldValue
// }

func TestForArrayNodeExecuteError(t *testing.T) {
	test_utils.SetMainEnvironmentMock(t)

	// "for i in 2 { print $i }"
	valNode := forArrayNode{
		variable: "i",
		arr:      &valueNode{2},
		body:     &printNode{&formatStringNode{&valueNode{"%v"}, []node{&symbolReferenceNode{"i"}}}},
	}
	value, err := valNode.execute()
	assert.Nil(t, value)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "only an array can be iterated")
}

func TestForRangeNodeExecute(t *testing.T) {
	test_utils.SetMainEnvironmentMock(t)

	// "for i in 0..3 { print $i }"
	valNode := forRangeNode{
		variable: "i",
		start:    &valueNode{0},
		end:      &valueNode{3},
		body:     &printNode{&formatStringNode{&valueNode{"%v"}, []node{&symbolReferenceNode{"i"}}}},
	}
	value, err := valNode.execute()
	assert.Nil(t, err)
	assert.Nil(t, value)

	assert.Contains(t, controllers.State.DynamicSymbolTable, "i")
	assert.Equal(t, 3, controllers.State.DynamicSymbolTable["i"])
}

func TestForRangeNodeExecuteError(t *testing.T) {
	test_utils.SetMainEnvironmentMock(t)

	// Start value higher than end value
	// "for i in 3..0 { print $i }"
	valNode := forRangeNode{
		variable: "i",
		start:    &valueNode{3},
		end:      &valueNode{0},
		body:     &printNode{&formatStringNode{&valueNode{"%v"}, []node{&symbolReferenceNode{"i"}}}},
	}
	value, err := valNode.execute()
	assert.Nil(t, value)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "start index should be lower than end index")

	// Body is nil
	valNode = forRangeNode{
		variable: "i",
		start:    &valueNode{0},
		end:      &valueNode{3},
		body:     nil,
	}
	value, err = valNode.execute()
	assert.Nil(t, value)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "loop body should not be empty")
}
