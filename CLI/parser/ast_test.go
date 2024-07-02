package parser

import (
	"cli/controllers"
	"cli/models"
	test_utils "cli/test"
	"encoding/json"
	"strings"
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
				"data": []any{},
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
	assert.ErrorContains(t, err, "the U value must be positive")

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

func TestAddToMap(t *testing.T) {
	newMap, replaced := addToMap[int](map[string]any{"a": 3}, "b", 10)

	assert.Equal(t, map[string]any{"a": 3, "b": 10}, newMap)
	assert.False(t, replaced)

	newMap, replaced = addToMap[int](newMap, "b", 15)
	assert.Equal(t, map[string]any{"a": 3, "b": 15}, newMap)
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

func TestAddRoomSeparatorOrPillarError(t *testing.T) {
	tests := []struct {
		name         string
		addFunction  func(string, []any) (map[string]any, error)
		values       []any
		errorMessage string
	}{
		{"AddRoomSeparator", addRoomSeparator, []any{"mySeparator"}, "4 values (name, startPos, endPos, type) expected to add a separator"},
		{"AddRoomPillar", addRoomPillar, []any{"myPillar"}, "4 values (name, centerXY, sizeXY, rotation) expected to add a pillar"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj, err := tt.addFunction("/Physical/site/building/room", tt.values)

			assert.Nil(t, obj)
			assert.NotNil(t, err)
			assert.ErrorContains(t, err, tt.errorMessage)
		})
	}
}

func TestAddRoomSeparatorOrPillarWorks(t *testing.T) {
	tests := []struct {
		name          string
		addFunction   func(string, []any) (map[string]any, error)
		values        []any
		newAttributes map[string]any
	}{
		{"AddRoomSeparator", addRoomSeparator, []any{"mySeparator", []float64{1., 2.}, []float64{1., 2.}, "wireframe"}, map[string]interface{}{"separators": map[string]interface{}{"mySeparator": Separator{StartPos: []float64{1, 2}, EndPos: []float64{1, 2}, Type: "wireframe"}}}},
		{"AddRoomPillar", addRoomPillar, []any{"myPillar", []float64{1., 2.}, []float64{1., 2.}, 2.5}, map[string]interface{}{"pillars": map[string]interface{}{"myPillar": Pillar{CenterXY: []float64{1, 2}, SizeXY: []float64{1, 2}, Rotation: 2.5}}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, mockAPI, _, _ := test_utils.SetMainEnvironmentMock(t)
			room := test_utils.GetEntity("room", "room", "site.building", "domain")

			test_utils.MockGetObject(mockAPI, room)
			test_utils.MockGetObject(mockAPI, room)

			room["attributes"] = tt.newAttributes
			test_utils.MockUpdateObject(mockAPI, map[string]interface{}{"attributes": tt.newAttributes}, room)

			obj, err := tt.addFunction("/Physical/site/building/room", tt.values)
			assert.NotNil(t, obj)
			assert.Nil(t, err)
		})
	}
}

func TestDeleteRoomPillarOrSeparatorWithError(t *testing.T) {
	tests := []struct {
		name          string
		attributeName string
		separatorName string
		errorMessage  string
	}{
		{"InvalidArgument", "other", "separator", "other separator does not exist"},
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
	room["attributes"].(map[string]any)["separators"] = map[string]interface{}{"mySeparator": Separator{StartPos: []float64{1, 2}, EndPos: []float64{1, 2}, Type: "wireframe"}}

	updatedRoom := test_utils.GetEntity("room", "room", "site.building", "domain")
	updatedRoom["attributes"] = map[string]any{"separators": map[string]interface{}{}}

	test_utils.MockGetObject(mockAPI, room)
	test_utils.MockGetObject(mockAPI, room)

	test_utils.MockUpdateObject(mockAPI, map[string]interface{}{"attributes": map[string]interface{}{"separators": map[string]interface{}{}}}, updatedRoom)
	obj, err := deleteRoomPillarOrSeparator("/Physical/site/building/room", "separator", "mySeparator")

	assert.Nil(t, err)
	assert.NotNil(t, obj)
}

func TestDeleteRoomPillarOrSeparatorPillar(t *testing.T) {
	_, mockAPI, _, _ := test_utils.SetMainEnvironmentMock(t)

	room := test_utils.GetEntity("room", "room", "site.building", "domain")
	room["attributes"].(map[string]any)["pillars"] = map[string]interface{}{"myPillar": Pillar{CenterXY: []float64{1, 2}, SizeXY: []float64{1, 2}, Rotation: 2.5}}

	updatedRoom := maps.Clone(room)
	updatedRoom["attributes"] = map[string]any{"pillars": map[string]interface{}{}}

	test_utils.MockGetObject(mockAPI, room)
	test_utils.MockGetObject(mockAPI, room)
	test_utils.MockUpdateObject(mockAPI, map[string]interface{}{"attributes": map[string]interface{}{"pillars": map[string]interface{}{}}}, updatedRoom)
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

func TestTreeDrawAndUndraw(t *testing.T) {
	tests := []struct {
		name          string
		executionNode node
		isUndraw      bool
	}{
		{"TreeNodeExecution", &treeNode{path: &pathNode{path: &valueNode{"/Physical/site/building/room"}}}, false},
		{"DrawNodeExecution", &drawNode{path: &pathNode{path: &valueNode{"/Physical/site/building/room"}}}, false},
		{"UndrawNodeExecution", &undrawNode{path: &pathNode{path: &valueNode{"/Physical/site/building/room"}}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, mockAPI, mockOgree3D, _ := test_utils.SetMainEnvironmentMock(t)
			room := test_utils.GetEntity("room", "room", "site.building", "domain")
			test_utils.MockGetObject(mockAPI, room)

			if tt.isUndraw {
				mockOgree3D.On(
					"Inform", "Undraw", 0, map[string]interface{}{"type": "delete", "data": "site.building.room"},
				).Return(nil)
			}

			value, err := tt.executionNode.execute()
			assert.Nil(t, err)
			assert.Nil(t, value)
		})
	}
}

func TestLsogNodeExecution(t *testing.T) {
	tests := []struct {
		name             string
		clipboardContent []string
	}{
		{"EmptyClipboard", []string{}},
		{"OneElementClipboard", []string{"/Physical/site/building/room/rack"}},
		{"TwoElementClipboard", []string{"/Physical/site/building/room/rack", "/Physical/site/building/room2/rack2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test_utils.SetMainEnvironmentMock(t)

			controllers.State.ClipBoard = tt.clipboardContent
			array := selectNode{}
			value, err := array.execute()

			assert.Nil(t, err)
			assert.NotNil(t, value)
			assert.Len(t, value, len(tt.clipboardContent))
			assert.Equal(t, tt.clipboardContent, value)
		})
	}
}

func TestPwdNodeExecution(t *testing.T) {
	tests := []struct {
		name        string
		currentPath string
	}{
		{"SitePath", "/Physical/site"},
		{"RoomPath", "/Physical/site/room"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test_utils.SetMainEnvironmentMock(t)

			controllers.State.CurrPath = tt.currentPath
			array := pwdNode{}
			value, err := array.execute()

			assert.Nil(t, err)
			assert.NotNil(t, value)
			assert.Equal(t, controllers.State.CurrPath, value)
		})
	}
}

func TestSelectChildrenNodeExecution(t *testing.T) {
	rack := test_utils.GetEntity("rack", "rack", "site.building.room", "domain")
	secondRack := test_utils.GetEntity("rack", "rack2", "site.building.room2", "domain")
	tests := []struct {
		name     string
		entities []map[string]any
	}{
		{"EmptySelection", []map[string]any{}},
		{"OneSelection", []map[string]any{rack}},
		{"TwoSelections", []map[string]any{rack, secondRack}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, mockAPI, mockOgree3D, _ := test_utils.SetMainEnvironmentMock(t)

			paths := []node{}
			ids := []string{}

			for _, entity := range tt.entities {
				ids = append(ids, entity["id"].(string))
				path := models.PhysicalPath + strings.Replace(entity["id"].(string), ".", "/", -1)
				paths = append(paths, pathNode{path: &valueNode{path}})
				test_utils.MockGetObject(mockAPI, entity)
			}
			informOptionalData, _ := json.Marshal(ids)
			mockOgree3D.On(
				"InformOptional", "SetClipBoard",
				-1, map[string]interface{}{"data": string(informOptionalData), "type": "select"},
			).Return(nil)

			array := selectChildrenNode{
				paths: paths,
			}
			value, err := array.execute()

			assert.Nil(t, err)
			assert.NotNil(t, value)
			assert.Len(t, value, len(paths))
			assert.Equal(t, controllers.State.ClipBoard, value)
		})
	}
}

func TestUnsetFuncNodeExecution(t *testing.T) {
	test_utils.SetMainEnvironmentMock(t)
	functionBody := printNode{&formatStringNode{&valueNode{"%v"}, []node{&symbolReferenceNode{"i"}}}}
	functionName := "my_function"

	controllers.State.FuncTable[functionName] = functionBody

	unsetFunctionNode := unsetFuncNode{
		funcName: functionName,
	}

	value, err := unsetFunctionNode.execute()

	assert.Nil(t, err)
	assert.Nil(t, value)
	_, ok := controllers.State.FuncTable[functionName]
	assert.False(t, ok)

	// unset a non existent function does not generate an error
	value, err = unsetFunctionNode.execute()

	assert.Nil(t, err)
	assert.Nil(t, value)
}

func TestUnsetVarNodeExecution(t *testing.T) {
	test_utils.SetMainEnvironmentMock(t)
	varName := "i"

	controllers.State.DynamicSymbolTable[varName] = 5

	unsetVarNode := unsetVarNode{
		varName: varName,
	}

	value, err := unsetVarNode.execute()

	assert.Nil(t, err)
	assert.Nil(t, value)
	_, ok := controllers.State.FuncTable[varName]
	assert.False(t, ok)

	// unset a non existent variable does not generate an error
	value, err = unsetVarNode.execute()

	assert.Nil(t, err)
	assert.Nil(t, value)
}

func TestCreateDomainNodeExecution(t *testing.T) {
	_, mockAPI, _, _ := test_utils.SetMainEnvironmentMock(t)
	domain := test_utils.GetEntity("domain", "myDomain", "", "")
	domain["attributes"].(map[string]any)["color"] = "ffaaff"
	delete(domain, "id")

	test_utils.MockCreateObject(mockAPI, "domain", domain)

	executionNode := createDomainNode{
		path:  pathNode{path: &valueNode{"/Organisation/Domain/myDomain"}},
		color: &valueNode{"ffaaff"},
	}

	value, err := executionNode.execute()

	assert.Nil(t, err)
	assert.Nil(t, value)
}

func TestCreateSiteNodeExecution(t *testing.T) {
	_, mockAPI, _, _ := test_utils.SetMainEnvironmentMock(t)
	site := test_utils.GetEntity("site", "mySite", "", "")
	delete(site, "id")
	delete(site, "children")

	test_utils.MockCreateObject(mockAPI, "site", site)

	executionNode := createSiteNode{
		path: pathNode{path: &valueNode{"/Physical/mySite"}},
	}

	value, err := executionNode.execute()

	assert.Nil(t, err)
	assert.Nil(t, value)
}

func TestCreateBuildingNodeExecution(t *testing.T) {
	_, mockAPI, _, _ := test_utils.SetMainEnvironmentMock(t)
	site := test_utils.GetEntity("site", "mySite", "", "")
	building := test_utils.GetEntity("building", "myBuilding", "mySite", "")
	building["attributes"].(map[string]any)["posXY"] = []float64{0, 0}
	building["attributes"].(map[string]any)["rotation"] = 0.0
	building["attributes"].(map[string]any)["size"] = []float64{10, 10}
	building["attributes"].(map[string]any)["sizeUnit"] = "m"
	building["attributes"].(map[string]any)["height"] = 10.0
	building["attributes"].(map[string]any)["heightUnit"] = "m"
	building["attributes"].(map[string]any)["posXYUnit"] = "m"

	delete(building, "id")
	delete(building, "children")

	test_utils.MockGetObject(mockAPI, site)
	test_utils.MockCreateObject(mockAPI, "building", building)

	executionNode := createBuildingNode{
		path:           pathNode{path: &valueNode{"/Physical/mySite/myBuilding"}},
		posXY:          vec2(0, 0),
		rotation:       &valueNode{0.0},
		sizeOrTemplate: vec3(10, 10, 10),
	}

	value, err := executionNode.execute()

	assert.Nil(t, err)
	assert.Nil(t, value)
}

func TestCreateRoomNodeExecution(t *testing.T) {
	_, mockAPI, _, _ := test_utils.SetMainEnvironmentMock(t)
	building := test_utils.GetEntity("building", "myBuilding", "mySite", "")
	room := test_utils.GetEntity("room", "myRoom", "mySite.myBuilding", "")
	room["attributes"].(map[string]any)["posXY"] = []float64{0, 0}
	room["attributes"].(map[string]any)["rotation"] = 0.0
	room["attributes"].(map[string]any)["size"] = []float64{10, 10}
	room["attributes"].(map[string]any)["sizeUnit"] = "m"
	room["attributes"].(map[string]any)["height"] = 10.0
	room["attributes"].(map[string]any)["heightUnit"] = "m"
	room["attributes"].(map[string]any)["posXYUnit"] = "m"
	room["attributes"].(map[string]any)["axisOrientation"] = "+x+y"
	room["attributes"].(map[string]any)["floorUnit"] = "m"

	delete(room, "id")
	delete(room, "children")

	test_utils.MockGetObject(mockAPI, building)
	test_utils.MockCreateObject(mockAPI, "room", room)

	executionNode := createRoomNode{
		path:            pathNode{path: &valueNode{"/Physical/mySite/myBuilding/myRoom"}},
		posXY:           vec2(0, 0),
		rotation:        &valueNode{0.0},
		size:            vec3(10, 10, 10),
		axisOrientation: &valueNode{"+x+y"},
		floorUnit:       &valueNode{"m"},
	}

	value, err := executionNode.execute()

	assert.Nil(t, err)
	assert.Nil(t, value)
}

func TestCreateRackNodeExecution(t *testing.T) {
	_, mockAPI, _, _ := test_utils.SetMainEnvironmentMock(t)
	room := test_utils.GetEntity("room", "myRoom", "mySite.myBuilding", "")
	rack := test_utils.GetEntity("rack", "myRack", "mySite.myBuilding.myRoom", "")
	rack["attributes"].(map[string]any)["posXYZ"] = []float64{0, 0, 0}
	rack["attributes"].(map[string]any)["rotation"] = []float64{0, 0, 0}
	rack["attributes"].(map[string]any)["size"] = []float64{10, 10}
	rack["attributes"].(map[string]any)["sizeUnit"] = "cm"
	rack["attributes"].(map[string]any)["height"] = 10.0
	rack["attributes"].(map[string]any)["heightUnit"] = "U"
	rack["attributes"].(map[string]any)["posXYUnit"] = "m"

	delete(rack, "id")
	delete(rack, "children")

	test_utils.MockGetObject(mockAPI, room)
	test_utils.MockCreateObject(mockAPI, "rack", rack)

	executionNode := createRackNode{
		path:           pathNode{path: &valueNode{"/Physical/mySite/myBuilding/myRoom/myRack"}},
		pos:            vec3(0, 0, 0),
		rotation:       vec3(0, 0, 0),
		unit:           &valueNode{"m"},
		sizeOrTemplate: vec3(10, 10, 10),
	}

	value, err := executionNode.execute()

	assert.Nil(t, err)
	assert.Nil(t, value)
}

func TestCreateGenericNodeExecution(t *testing.T) {
	_, mockAPI, _, _ := test_utils.SetMainEnvironmentMock(t)
	room := test_utils.GetEntity("room", "myRoom", "mySite.myBuilding", "")
	generic := test_utils.GetEntity("generic", "myGeneric", "mySite.myBuilding.myRoom", "")
	generic["attributes"].(map[string]any)["posXYZ"] = []float64{0, 0, 0}
	generic["attributes"].(map[string]any)["rotation"] = []float64{0, 0, 0}
	generic["attributes"].(map[string]any)["size"] = []float64{10, 10}
	generic["attributes"].(map[string]any)["sizeUnit"] = "cm"
	generic["attributes"].(map[string]any)["height"] = 10.0
	generic["attributes"].(map[string]any)["heightUnit"] = "cm"
	generic["attributes"].(map[string]any)["posXYUnit"] = "m"
	generic["attributes"].(map[string]any)["shape"] = "cube"
	generic["attributes"].(map[string]any)["type"] = "box"

	delete(generic, "id")
	delete(generic, "children")

	test_utils.MockGetObject(mockAPI, room)
	test_utils.MockCreateObject(mockAPI, "generic", generic)

	executionNode := createGenericNode{
		path:           pathNode{path: &valueNode{"/Physical/mySite/myBuilding/myRoom/myGeneric"}},
		pos:            vec3(0, 0, 0),
		rotation:       vec3(0, 0, 0),
		unit:           &valueNode{"m"},
		sizeOrTemplate: vec3(10, 10, 10),
		shape:          &valueNode{"cube"},
		getype:         &valueNode{"box"},
	}

	value, err := executionNode.execute()

	assert.Nil(t, err)
	assert.Nil(t, value)
}

func TestCreateDeviceNodeExecution(t *testing.T) {
	_, mockAPI, _, _ := test_utils.SetMainEnvironmentMock(t)
	rack := test_utils.GetEntity("rack", "myRack", "mySite.myBuilding.myRoom", "")
	device := test_utils.GetEntity("device", "myDevice", "mySite.myBuilding.myRoom.myRack", "")
	device["attributes"].(map[string]any)["posU/slot"] = []string{}
	device["attributes"].(map[string]any)["sizeU"] = 10
	device["attributes"].(map[string]any)["sizeUnit"] = "mm"
	device["attributes"].(map[string]any)["height"] = 445.0
	device["attributes"].(map[string]any)["heightUnit"] = "mm"
	device["attributes"].(map[string]any)["orientation"] = "front"
	device["attributes"].(map[string]any)["invertOffset"] = false
	delete(device["attributes"].(map[string]any), "size")

	delete(device, "id")
	delete(device, "children")

	test_utils.MockGetObject(mockAPI, rack)
	test_utils.MockCreateObject(mockAPI, "device", device)

	executionNode := createDeviceNode{
		path:            pathNode{path: &valueNode{"/Physical/mySite/myBuilding/myRoom/myRack/myDevice"}},
		posUOrSlot:      []node{},
		invertOffset:    false,
		sizeUOrTemplate: &valueNode{10},
		side:            &valueNode{"front"},
	}

	value, err := executionNode.execute()

	assert.Nil(t, err)
	assert.Nil(t, value)
}

func TestCreateGroupNodeExecution(t *testing.T) {
	_, mockAPI, _, _ := test_utils.SetMainEnvironmentMock(t)
	site := test_utils.GetEntity("site", "mySite", "", "")
	group := map[string]any{
		"category":    "group",
		"description": "",
		"domain":      "",
		"name":        "myGroup",
		"parentId":    "mySite",
		"attributes": map[string]any{
			"content": []string{"myBuilding1", "myBuilding2"},
		},
	}

	test_utils.MockGetObject(mockAPI, site)
	test_utils.MockCreateObject(mockAPI, "group", group)

	executionNode := createGroupNode{
		path:  pathNode{path: &valueNode{"/Physical/mySite/myGroup"}},
		paths: []node{&valueNode{"/Physical/mySite/myBuilding1"}, &valueNode{"/Physical/mySite/myBuilding2"}},
	}

	value, err := executionNode.execute()

	assert.Nil(t, err)
	assert.Nil(t, value)
}

func TestCreateTagNodeExecution(t *testing.T) {
	_, mockAPI, _, _ := test_utils.SetMainEnvironmentMock(t)
	tag := map[string]any{
		"slug":        "myTag",
		"description": "myTag",
		"color":       "ffaaff",
	}

	test_utils.MockCreateObject(mockAPI, "tag", tag)

	executionNode := createTagNode{
		slug:  &valueNode{"myTag"},
		color: &valueNode{"ffaaff"},
	}

	value, err := executionNode.execute()

	assert.Nil(t, err)
	assert.Nil(t, value)
}

func TestCreateLayerNodeExecution(t *testing.T) {
	_, mockAPI, _, _ := test_utils.SetMainEnvironmentMock(t)
	layer := map[string]any{
		"slug":          "myLayer",
		"filter":        "category = rack",
		"applicability": "mySite.myBuilding.myRoom*",
	}

	test_utils.MockCreateObject(mockAPI, "layer", layer)

	executionNode := createLayerNode{
		slug:          &valueNode{"myLayer"},
		applicability: &valueNode{"/Physical/mySite/myBuilding/myRoom*"},
		filterValue:   &valueNode{"category = rack"},
	}

	value, err := executionNode.execute()

	assert.Nil(t, err)
	assert.Nil(t, value)
}

func TestCreateCorridorNodeExecution(t *testing.T) {
	_, mockAPI, _, _ := test_utils.SetMainEnvironmentMock(t)
	room := test_utils.GetEntity("room", "myRoom", "mySite.myBuilding", "myDomain")
	corridor := map[string]any{
		"category":    "corridor",
		"description": "",
		"name":        "myCorridor",
		"parentId":    "mySite.myBuilding.myRoom",
		"domain":      "myDomain",
		"attributes": map[string]any{
			"rotation":    []float64{0, 0, 0},
			"size":        []float64{10, 10},
			"temperature": "cold",
			"sizeUnit":    "cm",
			"height":      10.0,
			"heightUnit":  "cm",
			"posXYZ":      []float64{0, 0, 0},
			"posXYUnit":   "m",
		},
	}

	test_utils.MockGetObject(mockAPI, room)
	test_utils.MockCreateObject(mockAPI, "corridor", corridor)

	executionNode := createCorridorNode{
		path:     pathNode{path: &valueNode{"/Physical/mySite/myBuilding/myRoom/myCorridor"}},
		pos:      vec3(0, 0, 0),
		rotation: vec3(0, 0, 0),
		unit:     &valueNode{"m"},
		size:     vec3(10, 10, 10),
		temp:     &valueNode{"cold"},
	}

	value, err := executionNode.execute()

	assert.Nil(t, err)
	assert.Nil(t, value)
}

func TestCreateUserNodeExecution(t *testing.T) {
	_, mockAPI, _, _ := test_utils.SetMainEnvironmentMock(t)

	mockAPI.On(
		"Request",
		"POST",
		"/api/users",
		"mock.Anything", // It generates a random password
		201,
	).Return(
		&controllers.Response{
			Body: map[string]any{
				"account": map[string]any{
					"email": "user@user.com",
					"roles": map[string]any{
						"myDomain": "viewer",
					},
				},
			},
		}, nil,
	).Once()

	executionNode := createUserNode{
		email:  &valueNode{"user@user.com"},
		role:   &valueNode{"viewer"},
		domain: &valueNode{"myDomain"},
	}

	value, err := executionNode.execute()

	assert.Nil(t, err)
	assert.Nil(t, value)
}

func TestUiDelayNodeExecution(t *testing.T) {
	_, _, mockOgree3D, _ := test_utils.SetMainEnvironmentMock(t)

	mockOgree3D.On(
		"Inform",
		"HandleUI",
		-1,
		map[string]interface{}{"type": "ui", "data": map[string]interface{}{"command": "delay", "data": 10.0}},
	).Return(nil).Once()

	executionNode := uiDelayNode{
		time: 10,
	}

	value, err := executionNode.execute()

	assert.Nil(t, err)
	assert.Nil(t, value)
}

func TestUiToggleNodeExecution(t *testing.T) {
	_, _, mockOgree3D, _ := test_utils.SetMainEnvironmentMock(t)

	mockOgree3D.On(
		"Inform",
		"HandleUI",
		-1,
		map[string]interface{}{"type": "ui", "data": map[string]interface{}{"command": "myFeature", "data": false}},
	).Return(nil).Once()

	executionNode := uiToggleNode{
		feature: "myFeature",
		enable:  false,
	}

	value, err := executionNode.execute()

	assert.Nil(t, err)
	assert.Nil(t, value)
}
