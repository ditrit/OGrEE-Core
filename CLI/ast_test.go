package main

import (
	"cli/controllers"
	mocks "cli/mocks/controllers"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func setMainEnvironmentMock(t *testing.T) (*mocks.APIPort, *mocks.ClockPort, func()) {
	oldDynamicSymbolTable := controllers.State.DynamicSymbolTable
	oldFuncTable := controllers.State.FuncTable
	controllers.State.DynamicSymbolTable = map[string]any{}
	controllers.State.FuncTable = map[string]any{}

	mockAPI := mocks.NewAPIPort(t)
	mockOgree3D := mocks.NewOgree3DPort(t)
	mockClock := mocks.NewClockPort(t)
	controller := controllers.Controller{
		API:     mockAPI,
		Ogree3D: mockOgree3D,
		Clock:   mockClock,
	}
	oldControllerValue := controllers.C
	controllers.C = controller
	oldHierarchy := controllers.State.Hierarchy
	controllers.State.Hierarchy = controllers.BuildBaseTree(controller)

	deferFunction := func() {
		controllers.State.DynamicSymbolTable = oldDynamicSymbolTable
		controllers.State.FuncTable = oldFuncTable
		controllers.C = oldControllerValue
		controllers.State.Hierarchy = oldHierarchy
	}
	return mockAPI, mockClock, deferFunction
}

func TestValueNodeExecute(t *testing.T) {
	valNode := valueNode{5}
	value, err := valNode.execute()

	assert.Nil(t, err)
	assert.Equal(t, 5, value)
}

func TestAstExecute(t *testing.T) {
	_, _, deferFunction := setMainEnvironmentMock(t)
	defer deferFunction()

	commands := ast{
		statements: []node{
			&assignNode{"i", &valueNode{5}},
			&assignNode{"j", &valueNode{10}},
		},
	}
	value, err := commands.execute()

	assert.Nil(t, err)
	assert.Nil(t, value)

	assert.Contains(t, controllers.State.DynamicSymbolTable, "i")
	assert.Contains(t, controllers.State.DynamicSymbolTable, "j")
	assert.Equal(t, 5, controllers.State.DynamicSymbolTable["i"])
	assert.Equal(t, 10, controllers.State.DynamicSymbolTable["j"])
}

func TestFuncDefNodeExecute(t *testing.T) {
	_, _, deferFunction := setMainEnvironmentMock(t)
	defer deferFunction()

	// alias my_function { print $i }
	functionBody := printNode{&formatStringNode{&valueNode{"%v"}, []node{&symbolReferenceNode{"i"}}}}
	funcNode := funcDefNode{
		name: "my_function",
		body: &functionBody,
	}
	value, err := funcNode.execute()

	assert.Nil(t, err)
	assert.Nil(t, value)

	assert.Contains(t, controllers.State.FuncTable, "my_function")
	assert.Equal(t, &functionBody, controllers.State.FuncTable["my_function"])
}

func TestFuncCallNodeExecute(t *testing.T) {
	_, _, deferFunction := setMainEnvironmentMock(t)
	defer deferFunction()

	// we define the function
	// alias my_function { .var: i = 5 }
	functionName := "my_function"
	functionBody := assignNode{"i", &valueNode{5}}
	funcNode := funcDefNode{
		name: functionName,
		body: &functionBody,
	}
	value, err := funcNode.execute()

	assert.Nil(t, err)
	assert.Nil(t, value)

	callNode := funcCallNode{functionName}
	value, err = callNode.execute()
	assert.Nil(t, err)
	assert.Nil(t, value)

	assert.Contains(t, controllers.State.DynamicSymbolTable, "i")
	assert.Equal(t, 5, controllers.State.DynamicSymbolTable["i"])
}

func TestFuncCallNodeExecuteUndefinedFunction(t *testing.T) {
	_, _, deferFunction := setMainEnvironmentMock(t)
	defer deferFunction()

	functionName := "my_function"
	callNode := funcCallNode{functionName}
	value, err := callNode.execute()

	assert.Nil(t, value)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "undefined function "+functionName)
}

func TestArrNodeExecute(t *testing.T) {
	array := arrNode{[]node{&valueNode{5}, &valueNode{6}}}
	value, err := array.execute()

	assert.Nil(t, err)
	assert.Equal(t, []float64{5, 6}, value) // it only returns an array of floats
}

func TestLenNodeExecute(t *testing.T) {
	_, _, deferFunction := setMainEnvironmentMock(t)
	defer deferFunction()

	controllers.State.DynamicSymbolTable["myArray"] = []float64{1, 2, 3, 4}
	array := lenNode{"myArray"}
	value, err := array.execute()

	assert.Nil(t, err)
	assert.Equal(t, 4, value)

	array = lenNode{"myArray2"}
	_, err = array.execute()

	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "Undefined variable myArray2")
}

func TestCdNodeExecute(t *testing.T) {
	mockAPI, _, deferFunction := setMainEnvironmentMock(t)
	defer deferFunction()

	mockAPI.On(
		"Request", "GET",
		"/api/hierarchy-objects/site",
		"mock.Anything", 200,
	).Return(
		&controllers.Response{
			Body: map[string]any{
				"data": map[string]any{
					"category": "site",
					"id":       "site",
					"name":     "site",
					"parentId": "",
				},
			},
		}, nil,
	).Once()

	array := cdNode{&pathNode{path: &valueNode{"/Physical/site"}}}
	value, err := array.execute()

	assert.Nil(t, err)
	assert.Nil(t, value)
}

func TestLsNodeExecute(t *testing.T) {
	mockAPI, mockClock, deferFunction := setMainEnvironmentMock(t)
	defer deferFunction()

	mockAPI.On(
		"Request", "GET",
		"/api/hierarchy-objects/site/all?limit=1",
		"mock.Anything", 200,
	).Return(
		&controllers.Response{
			Body: map[string]any{
				"data": map[string]any{
					"category": "site",
					"id":       "site",
					"name":     "site",
					"parentId": "",
				},
			},
		}, nil,
	).Once()
	mockAPI.On(
		"Request", "GET",
		"/api/layers",
		"mock.Anything", 200,
	).Return(
		&controllers.Response{
			Body: map[string]any{
				"data": map[string]any{
					"objects": []any{},
				},
			},
		}, nil,
	).Once()
	mockClock.On("Now").Return(time.Now()).Once()

	ls := lsNode{
		path: &pathNode{path: &valueNode{"/Physical/site"}},
	}
	value, err := ls.execute()

	assert.Nil(t, err)
	assert.Nil(t, value)
}

func TestGetUNodeExecute(t *testing.T) {
	mockAPI, _, deferFunction := setMainEnvironmentMock(t)
	defer deferFunction()

	mockAPI.On(
		"Request", "GET",
		"/api/hierarchy-objects/site.building.room.rack/all?limit=1",
		"mock.Anything", 200,
	).Return(
		&controllers.Response{
			Body: map[string]any{
				"data": map[string]any{
					"category": "rack",
					"children": []any{},
					"id":       "site.building.room.rack",
					"name":     "rack",
					"parentId": "site.building.room",
				},
			},
		}, nil,
	).Once()

	uNode := getUNode{
		path: &pathNode{path: &valueNode{"/Physical/site/building/room/rack"}},
		u:    &valueNode{-42},
	}
	value, err := uNode.execute()

	assert.Nil(t, value)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "The U value must be positive")

	uNode = getUNode{
		path: &pathNode{path: &valueNode{"/Physical/site/building/room/rack"}},
		u:    &valueNode{42},
	}
	value, err = uNode.execute()

	assert.Nil(t, value)
	assert.Nil(t, err)
}

func TestGetSlotNodeExecute(t *testing.T) {
	mockAPI, _, deferFunction := setMainEnvironmentMock(t)
	defer deferFunction()

	mockAPI.On(
		"Request", "GET",
		"/api/hierarchy-objects/site.building.room.rack/all?limit=1",
		"mock.Anything", 200,
	).Return(
		&controllers.Response{
			Body: map[string]any{
				"data": map[string]any{
					"category": "rack",
					"children": []any{map[string]any{
						"category": "device",
						"attributes": map[string]any{
							"type": "chassis",
							"slot": "slot",
						},
						"children": []any{},
						"id":       "BASIC.A.R1.A01.chT",
						"name":     "chT",
						"parentId": "BASIC.A.R1.A01",
					}},
					"id":       "site.building.room.rack",
					"name":     "rack",
					"parentId": "site.building.room",
					"attributes": map[string]any{
						"slot": []any{
							map[string]any{
								"location":   "slot",
								"type":       "u",
								"elemOrient": []any{33.3, -44.4, 107},
								"elemPos":    []any{58, 51, 44.45},
								"elemSize":   []any{482.6, 1138, 44.45},
								"mandatory":  "no",
								"labelPos":   "frontrear",
								"color":      "@color1",
							},
						},
					},
				},
			},
		}, nil,
	).Once()

	slotNode := getSlotNode{
		path: &pathNode{path: &valueNode{"/Physical/site/building/room/rack"}},
		slot: &valueNode{"slot"},
	}
	value, err := slotNode.execute()

	assert.Nil(t, value)
	assert.Nil(t, err)
}

func TestPrintNodeExecute(t *testing.T) {
	executable := printNode{&formatStringNode{&valueNode{"%v"}, []node{&valueNode{5}}}}
	value, err := executable.execute()

	assert.Nil(t, value)
	assert.Nil(t, err)
}
