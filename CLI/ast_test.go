package main

import (
	"cli/controllers"
	"cli/models"
	test_utils "cli/test"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/maps"
)

func TestValueNodeExecute(t *testing.T) {
	valNode := valueNode{5}
	value, err := valNode.execute()

	assert.Nil(t, err)
	assert.Equal(t, 5, value)
}

func TestAstExecute(t *testing.T) {
	test_utils.SetMainEnvironmentMock(t)

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
	test_utils.SetMainEnvironmentMock(t)

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
	test_utils.SetMainEnvironmentMock(t)

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
	test_utils.SetMainEnvironmentMock(t)

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
	test_utils.SetMainEnvironmentMock(t)

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
	_, mockAPI, _, _ := test_utils.SetMainEnvironmentMock(t)

	site := test_utils.GetEntity("site", "site", "", "domain")
	test_utils.MockGetObject(mockAPI, site)

	array := cdNode{&pathNode{path: &valueNode{"/Physical/site"}}}
	value, err := array.execute()

	assert.Nil(t, err)
	assert.Nil(t, value)
}

func TestLsNodeExecute(t *testing.T) {
	_, mockAPI, _, mockClock := test_utils.SetMainEnvironmentMock(t)

	site := test_utils.GetEntity("site", "site", "", "domain")
	test_utils.MockGetObjectHierarchy(mockAPI, site)
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
	_, mockAPI, _, _ := test_utils.SetMainEnvironmentMock(t)

	rack := test_utils.GetEntity("rack", "rack", "site.building.room", "domain")
	test_utils.MockGetObjectHierarchy(mockAPI, rack)

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
	_, mockAPI, _, _ := test_utils.SetMainEnvironmentMock(t)

	rack := test_utils.GetEntity("rack", "rack", "site.building.room", "domain")
	rack["children"] = []any{map[string]any{
		"category": "device",
		"attributes": map[string]any{
			"type": "chassis",
			"slot": "slot",
		},
		"children": []any{},
		"id":       "BASIC.A.R1.A01.chT",
		"name":     "chT",
		"parentId": "BASIC.A.R1.A01",
	}}
	rack["attributes"] = map[string]any{
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
	}
	test_utils.MockGetObjectHierarchy(mockAPI, rack)

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

func TestDeleteObjNodeExecute(t *testing.T) {
	_, mockAPI, _, _ := test_utils.SetMainEnvironmentMock(t)
	rack := test_utils.GetEntity("rack", "rack", "site.building.room", "domain")
	test_utils.MockDeleteObjects(mockAPI, "id=site.building.room.rack&namespace=physical.hierarchy", []any{rack})

	executable := deleteObjNode{&pathNode{path: &valueNode{"/Physical/site/building/room/rack"}}}
	value, err := executable.execute()

	assert.Nil(t, value)
	assert.Nil(t, err)
}

func TestDeleteSelectionNodeExecute(t *testing.T) {
	_, mockAPI, _, _ := test_utils.SetMainEnvironmentMock(t)
	controllers.State.ClipBoard = []string{"/Physical/site/building/room/rack", "/Physical/site/building/room2/rack2"}

	rack := test_utils.GetEntity("rack", "rack", "site.building.room", "domain")
	test_utils.MockDeleteObjects(mockAPI, "id=site.building.room.rack&namespace=physical.hierarchy", []any{rack})

	secondRack := test_utils.GetEntity("rack", "rack2", "site.building.room2", "domain")
	test_utils.MockDeleteObjects(mockAPI, "id=site.building.room2.rack2&namespace=physical.hierarchy", []any{secondRack})

	executable := deleteSelectionNode{}
	value, err := executable.execute()

	assert.Nil(t, value)
	assert.Nil(t, err)
}

func TestIsEntityDrawableNodeExecute(t *testing.T) {
	_, mockAPI, _, _ := test_utils.SetMainEnvironmentMock(t)
	rack := test_utils.GetEntity("rack", "rack", "site.building.room", "domain")

	test_utils.MockGetObject(mockAPI, rack)

	executable := isEntityDrawableNode{&pathNode{path: &valueNode{"/Physical/site/building/room/rack"}}}
	value, err := executable.execute()

	assert.False(t, value.(bool))
	assert.Nil(t, err)

	// We add the Rack to the drawable objects list
	controllers.State.DrawableObjs = []int{models.RACK}
	test_utils.MockGetObject(mockAPI, rack)

	value, err = executable.execute()

	assert.True(t, value.(bool))
	assert.Nil(t, err)
}

func TestIsAttrDrawableNodeExecute(t *testing.T) {
	_, mockAPI, _, _ := test_utils.SetMainEnvironmentMock(t)
	rack := test_utils.GetEntity("rack", "rack", "site.building.room", "domain")

	test_utils.MockGetObject(mockAPI, rack)

	executable := isAttrDrawableNode{&pathNode{path: &valueNode{"/Physical/site/building/room/rack"}}, "sdsdasd"}
	value, err := executable.execute()

	assert.Nil(t, err)
	assert.True(t, value.(bool))
}

func TestGetObjectNodeExecute(t *testing.T) {
	_, mockAPI, _, _ := test_utils.SetMainEnvironmentMock(t)
	rack := test_utils.GetEntity("rack", "rack", "site.building.room", "domain")

	test_utils.MockGetObjectsWithComplexFilters(mockAPI, "id=%2A%2A.site.building.room&namespace=physical.hierarchy", map[string]interface{}{"filter": "(category=rack) & (name=rack)"}, []any{rack})

	executable := getObjectNode{
		path:      &pathNode{path: &valueNode{"/Physical/site/building/room"}},
		filters:   map[string]node{"filter": &valueNode{"(category=rack) & (name=rack)"}},
		recursive: recursiveArgs{isRecursive: true},
	}
	value, err := executable.execute()

	assert.Nil(t, err)
	assert.Len(t, value, 1)
	assert.Equal(t, rack["id"], value.([]map[string]any)[0]["id"])
}

func TestSelectObjectNodeExecuteOnePath(t *testing.T) {
	_, mockAPI, mockOgree3D, _ := test_utils.SetMainEnvironmentMock(t)
	rack := test_utils.GetEntity("rack", "rack", "site.building.room", "domain")

	test_utils.MockGetObject(mockAPI, rack)
	test_utils.MockGetObject(mockAPI, rack)
	mockOgree3D.On(
		"InformOptional", "SetClipBoard",
		-1, map[string]interface{}{"data": "[\"site.building.room.rack\"]", "type": "select"},
	).Return(nil)

	executable := selectObjectNode{&pathNode{path: &valueNode{"/Physical/site/building/room/rack"}}}
	value, err := executable.execute()

	assert.Nil(t, err)
	assert.Nil(t, value)
	assert.Len(t, controllers.State.ClipBoard, 1)
	assert.Equal(t, []string{"/Physical/site/building/room/rack"}, controllers.State.ClipBoard)
}

func TestSelectObjectNodeExecuteReset(t *testing.T) {
	_, _, mockOgree3D, _ := test_utils.SetMainEnvironmentMock(t)
	controllers.State.ClipBoard = []string{"/Physical/site/building/room/rack"}
	mockOgree3D.On(
		"InformOptional", "SetClipBoard",
		-1, map[string]interface{}{"data": "[]", "type": "select"},
	).Return(nil)

	executable := selectObjectNode{&valueNode{""}}
	value, err := executable.execute()

	assert.Nil(t, err)
	assert.Nil(t, value)
	assert.Len(t, controllers.State.ClipBoard, 0)
}

func TestSetRoomAreas(t *testing.T) {
	_, mockAPI, _, _ := test_utils.SetMainEnvironmentMock(t)

	room := test_utils.GetEntity("room", "room", "site.building", "domain")

	roomResponse := test_utils.GetEntity("room", "room", "site.building", "domain")
	test_utils.MockGetObject(mockAPI, room)

	roomResponse["attributes"] = map[string]any{
		"reserved":  []float64{1, 2, 3, 4},
		"technical": []float64{1, 2, 3, 4},
	}
	test_utils.MockUpdateObject(mockAPI, map[string]interface{}{"attributes": map[string]interface{}{"reserved": []float64{1, 2, 3, 4}, "technical": []float64{1, 2, 3, 4}}}, roomResponse)

	reservedArea := []float64{1, 2, 3, 4}
	technicalArea := []float64{1, 2, 3, 4}
	value, err := setRoomAreas("/Physical/site/building/room", []any{reservedArea, technicalArea})

	assert.Nil(t, err)
	assert.NotNil(t, value)
}

func TestSetLabel(t *testing.T) {
	_, mockAPI, _, _ := test_utils.SetMainEnvironmentMock(t)

	room := test_utils.GetEntity("rack", "rack", "site.building.room", "domain")
	test_utils.MockGetObject(mockAPI, room)
	value, err := setLabel("/Physical/site/building/room/rack", []any{"myLabel"}, false)

	assert.Nil(t, err)
	assert.Nil(t, value)
}

func TestAddToStringMap(t *testing.T) {
	newMap, replaced := addToStringMap[int]("{\"a\":3}", "b", 10)

	assert.Equal(t, "{\"a\":3,\"b\":10}", newMap)
	assert.False(t, replaced)

	newMap, replaced = addToStringMap[int](newMap, "b", 15)
	assert.Equal(t, "{\"a\":3,\"b\":15}", newMap)
	assert.True(t, replaced)
}

func TestRemoveFromStringMap(t *testing.T) {
	newMap, deleted := removeFromStringMap[int]("{\"a\":3,\"b\":10}", "b")

	assert.Equal(t, "{\"a\":3}", newMap)
	assert.True(t, deleted)

	newMap, deleted = removeFromStringMap[int](newMap, "b")
	assert.Equal(t, "{\"a\":3}", newMap)
	assert.False(t, deleted)
}

func TestAddRoomSeparatorError(t *testing.T) {
	obj, err := addRoomSeparator("/Physical/site/building/room", []any{"mySeparator"})

	assert.Nil(t, obj)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "4 values (name, startPos, endPos, type) expected to add a separator")
}

func TestAddRoomSeparator(t *testing.T) {
	_, mockAPI, _, _ := test_utils.SetMainEnvironmentMock(t)

	room := test_utils.GetEntity("room", "room", "site.building", "domain")
	test_utils.MockGetObject(mockAPI, room)
	test_utils.MockGetObject(mockAPI, room)

	newAttributes := map[string]interface{}{
		"separators": "{\"mySeparator\":{\"startPosXYm\":[1,2],\"endPosXYm\":[1,2],\"type\":\"wireframe\"}}",
	}
	room["attributes"] = newAttributes
	test_utils.MockUpdateObject(mockAPI, map[string]interface{}{"attributes": newAttributes}, room)

	obj, err := addRoomSeparator("/Physical/site/building/room", []any{"mySeparator", []float64{1., 2.}, []float64{1., 2.}, "wireframe"})

	assert.Nil(t, err)
	assert.NotNil(t, obj)

}

func TestAddRoomPillarError(t *testing.T) {
	obj, err := addRoomPillar("/Physical/site/building/room", []any{"myPillar"})

	assert.Nil(t, obj)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "4 values (name, centerXY, sizeXY, rotation) expected to add a pillar")
}

func TestAddRoomPillar(t *testing.T) {
	_, mockAPI, _, _ := test_utils.SetMainEnvironmentMock(t)

	room := test_utils.GetEntity("room", "room", "site.building", "domain")
	test_utils.MockGetObject(mockAPI, room)
	test_utils.MockGetObject(mockAPI, room)

	newAttributes := map[string]interface{}{
		"pillars": "{\"myPillar\":{\"centerXY\":[1,2],\"sizeXY\":[1,2],\"rotation\":2.5}}",
	}
	room["attributes"] = newAttributes
	test_utils.MockUpdateObject(mockAPI, map[string]interface{}{"attributes": newAttributes}, room)

	obj, err := addRoomPillar("/Physical/site/building/room", []any{"myPillar", []float64{1., 2.}, []float64{1., 2.}, 2.5})

	assert.Nil(t, err)
	assert.NotNil(t, obj)
}

func TestDeleteRoomPillarOrSeparatorWithError(t *testing.T) {
	tests := []struct {
		name          string
		attributeName string
		separatorName string
		errorMessage  string
	}{
		{"InvalidArgument", "other", "separator", "\"separator\" or \"pillar\" expected"},
		{"SeparatorDoesNotExist", "separator", "mySeparator", "separator mySeparator does not exist"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, mockAPI, _, _ := test_utils.SetMainEnvironmentMock(t)

			room := test_utils.GetEntity("room", "room", "site.building", "domain")
			test_utils.MockGetObject(mockAPI, room)
			obj, err := deleteRoomPillarOrSeparator("/Physical/site/building/room", tt.attributeName, tt.separatorName)

			assert.Nil(t, obj)
			assert.NotNil(t, err)
			assert.ErrorContains(t, err, tt.errorMessage)
		})
	}
}

func TestDeleteRoomPillarOrSeparatorSeparator(t *testing.T) {
	_, mockAPI, _, _ := test_utils.SetMainEnvironmentMock(t)

	room := test_utils.GetEntity("room", "room", "site.building", "domain")
	room["attributes"].(map[string]any)["separators"] = "{\"mySeparator\":{\"startPosXYm\":[1,2],\"endPosXYm\":[1,2],\"type\":\"wireframe\"}}"

	updatedRoom := test_utils.GetEntity("room", "room", "site.building", "domain")
	updatedRoom["attributes"] = map[string]any{"separators": "{}"}

	test_utils.MockGetObject(mockAPI, room)
	test_utils.MockGetObject(mockAPI, room)

	test_utils.MockUpdateObject(mockAPI, map[string]interface{}{"attributes": map[string]interface{}{"separators": "{}"}}, updatedRoom)
	obj, err := deleteRoomPillarOrSeparator("/Physical/site/building/room", "separator", "mySeparator")

	assert.Nil(t, err)
	assert.NotNil(t, obj)
}

func TestDeleteRoomPillarOrSeparatorPillar(t *testing.T) {
	_, mockAPI, _, _ := test_utils.SetMainEnvironmentMock(t)

	room := test_utils.GetEntity("room", "room", "site.building", "domain")
	room["attributes"].(map[string]any)["pillars"] = "{\"myPillar\":{\"centerXY\":[1,2],\"sizeXY\":[1,2],\"rotation\":\"2.5\"}}"

	updatedRoom := maps.Clone(room)
	updatedRoom["attributes"] = map[string]any{"pillars": "{}"}

	test_utils.MockGetObject(mockAPI, room)
	test_utils.MockGetObject(mockAPI, room)
	test_utils.MockUpdateObject(mockAPI, map[string]interface{}{"attributes": map[string]interface{}{"pillars": "{}"}}, updatedRoom)
	obj, err := deleteRoomPillarOrSeparator("/Physical/site/building/room", "pillar", "myPillar")

	assert.Nil(t, err)
	assert.NotNil(t, obj)
}

func TestUpdateObjNodeExecuteUpdateDescription(t *testing.T) {
	_, mockAPI, _, _ := test_utils.SetMainEnvironmentMock(t)

	room := test_utils.GetEntity("room", "room", "site.building", "domain")

	test_utils.MockGetObject(mockAPI, room)
	room["description"] = "newDescription"
	test_utils.MockUpdateObject(mockAPI, map[string]interface{}{"description": "newDescription"}, room)

	array := updateObjNode{
		path:      &pathNode{path: &valueNode{"/Physical/site/building/room"}},
		attr:      "description",
		values:    []node{&valueNode{"newDescription"}},
		hasSharpe: false,
	}
	value, err := array.execute()

	assert.Nil(t, err)
	assert.Nil(t, value)
}
