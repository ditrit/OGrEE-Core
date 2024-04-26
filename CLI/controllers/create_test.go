package controllers_test

import (
	"cli/controllers"
	"cli/models"
	"maps"
	"testing"

	"github.com/stretchr/testify/assert"
)

var baseSite = map[string]any{
	"category": "site",
	"children": []any{},
	"id":       "BASIC",
	"name":     "BASIC",
	"parentId": "",
	"domain":   "test-domain",
}

var baseBuilding = map[string]any{
	"category": "building",
	"id":       "BASIC.A",
	"name":     "A",
	"parentId": "BASIC",
	"domain":   "test-domain",
	"attributes": map[string]any{
		"heightUnit": "m",
		"rotation":   30.5,
		"posXY":      []float64{4.6666666666667, -2},
		"posXYUnit":  "m",
		"size":       []float64{3, 3, 5},
		"sizeUnit":   "m",
	},
}

var createRoom = map[string]any{
	"category": "room",
	"id":       "BASIC.A.R1",
	"name":     "R1",
	"parentId": "BASIC.A",
	"domain":   "test-domain",
}

func TestCreateObjectInvalidPath(t *testing.T) {
	controller, _, _ := layersSetup(t)

	err := controller.CreateObject("/.", models.RACK, map[string]any{})
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "Invalid path name provided for OCLI object creation")
}

func TestCreateObjectErrorParentNotFound(t *testing.T) {
	controller, _, _ := layersSetup(t)

	err := controller.CreateObject("/", models.RACK, map[string]any{})
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "parent not found")
}

func TestCreateObjectWithNotExistentTemplateReturnsError(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObject(mockAPI, createRoom)

	mockObjectNotFound(mockAPI, "/api/obj-templates/not-exists")

	err := controller.CreateObject("/Physical/BASIC/A/R1/A01", models.RACK, map[string]any{
		"attributes": map[string]any{
			"rotation": []float64{0, 0, 0},
			"template": "not-exists",
		},
	})
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "template not found")
}

func TestCreateObjectWithTemplateOfIncorrectCategoryReturnsError(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObject(mockAPI, createRoom)

	mockGetObjTemplate(mockAPI, map[string]any{
		"category": "device",
		"slug":     "device-template",
	})

	err := controller.CreateObject("/Physical/BASIC/A/R1/A01", models.RACK, map[string]any{
		"attributes": map[string]any{
			"rotation": []float64{0, 0, 0},
			"template": "device-template",
		},
	})
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "template of category device is not applicable to rack")
}

func TestCreateGenericWithoutTemplateWorks(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObject(mockAPI, createRoom)

	mockCreateObject(mockAPI, "generic", map[string]any{
		"name":        "A01",
		"category":    "generic",
		"description": "",
		"domain":      createRoom["domain"],
		"parentId":    createRoom["id"],
		"attributes": map[string]any{
			"height":     1.0,
			"heightUnit": "cm",
			"rotation":   []float64{0, 0, 0},
			"posXYZ":     []float64{1, 1, 1},
			"posXYUnit":  "m",
			"size":       []float64{1, 1},
			"sizeUnit":   "cm",
			"shape":      "cube",
			"type":       "box",
		},
	})

	err := controller.CreateObject("/Physical/BASIC/A/R1/A01", models.GENERIC, map[string]any{
		"attributes": map[string]any{
			"rotation":  []float64{0, 0, 0},
			"size":      []float64{1, 1, 1},
			"posXYZ":    []float64{1, 1, 1},
			"posXYUnit": "m",
			"shape":     "cube",
			"type":      "box",
		},
	})
	assert.Nil(t, err)
}

func TestCreateGenericWithTemplateWorks(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObject(mockAPI, createRoom)

	mockGetObjTemplate(mockAPI, map[string]any{
		"slug":        "generic-template",
		"description": "a table",
		"category":    "generic",
		"sizeWDHmm":   []any{447, 914.5, 263.3},
		"fbxModel":    "",
		"attributes": map[string]any{
			"type": "table",
		},
		"colors": []any{},
	})

	mockCreateObject(mockAPI, "generic", map[string]any{
		"name":        "A01",
		"category":    "generic",
		"description": "a table",
		"domain":      createRoom["domain"],
		"parentId":    createRoom["id"],
		"attributes": map[string]any{
			"height":     263.3,
			"heightUnit": "mm",
			"rotation":   []float64{0, 0, 0},
			"posXYZ":     []float64{1, 1, 1},
			"posXYUnit":  "m",
			"size":       []any{447, 914.5},
			"sizeUnit":   "mm",
			"template":   "generic-template",
			"fbxModel":   "",
			"type":       "table",
		},
	})

	err := controller.CreateObject("/Physical/BASIC/A/R1/A01", models.GENERIC, map[string]any{
		"attributes": map[string]any{
			"rotation":  []float64{0, 0, 0},
			"posXYZ":    []float64{1, 1, 1},
			"posXYUnit": "m",
			"template":  "generic-template",
		},
	})
	assert.Nil(t, err)
}

func TestCreateDomain(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	// domain with no parent
	mockCreateObject(mockAPI, "domain", map[string]any{
		"category":    "domain",
		"id":          "dom1",
		"name":        "dom1",
		"parentId":    "",
		"description": "",
		"attributes":  map[string]any{},
	})

	err := controller.CreateObject("/Organisation/Domain/dom1", models.DOMAIN, map[string]any{
		"category":    "domain",
		"id":          "dom1",
		"name":        "dom1",
		"description": "",
	})
	assert.Nil(t, err)

	// domain with parent
	mockGetObjectByEntity(mockAPI, "domains", map[string]any{
		"category": "domain",
		"id":       "domParent",
		"name":     "domParent",
		"parentId": "",
	})

	mockCreateObject(mockAPI, "domain", map[string]any{
		"category":    "domain",
		"id":          "domParent.dom2",
		"name":        "dom2",
		"parentId":    "domParent",
		"description": "",
		"attributes":  map[string]any{},
	})

	err = controller.CreateObject("/Organisation/Domain/domParent/dom2", models.DOMAIN, map[string]any{
		"category":    "domain",
		"id":          "domParent.dom2",
		"name":        "dom2",
		"parentId":    "domParent",
		"description": "",
	})
	assert.Nil(t, err)
}

func TestCreateBuildingInvalidSize(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)
	buildingInvalidSize := maps.Clone(baseBuilding)
	buildingInvalidSize["attributes"].(map[string]any)["size"] = "[1,2,3]"

	mockGetObject(mockAPI, baseSite)

	// with state.DebugLvl = 0
	err := controller.CreateObject("/Physical/BASIC/A", models.BLDG, buildingInvalidSize)
	// returns nil but the object is not created
	assert.Nil(t, err)

	// with state.DebugLvl > 0
	controllers.State.DebugLvl = 1
	mockGetObject(mockAPI, baseSite)
	err = controller.CreateObject("/Physical/BASIC/A", models.BLDG, buildingInvalidSize)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "Invalid size attribute provided."+
		" \nIt must be an array/list/vector with 3 elements."+
		" Please refer to the wiki or manual reference"+
		" for more details on how to create objects "+
		"using this syntax")
	controllers.State.DebugLvl = 0
	buildingInvalidSize["attributes"].(map[string]any)["size"] = []float64{3, 3, 5}
}

func TestCreateBuildingInvalidPosXY(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)
	buildingInvalidPosXY := maps.Clone(baseBuilding)
	buildingInvalidPosXY["attributes"].(map[string]any)["posXY"] = []float64{}

	// with state.DebugLvl = 0
	mockGetObject(mockAPI, baseSite)
	err := controller.CreateObject("/Physical/BASIC/A", models.BLDG, maps.Clone(buildingInvalidPosXY))
	// returns nil but the object is not created
	assert.Nil(t, err)

	// with state.DebugLvl > 0
	controllers.State.DebugLvl = 1
	mockGetObject(mockAPI, baseSite)
	err = controller.CreateObject("/Physical/BASIC/A", models.BLDG, buildingInvalidPosXY)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "Invalid posXY attribute provided."+
		" \nIt must be an array/list/vector with 2 elements."+
		" Please refer to the wiki or manual reference"+
		" for more details on how to create objects "+
		"using this syntax")
	controllers.State.DebugLvl = 0
	buildingInvalidPosXY["attributes"].(map[string]any)["posXY"] = []float64{4.6666666666667, -2}
}

func TestCreateBuilding(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObject(mockAPI, baseSite)
	mockCreateObject(mockAPI, "building", map[string]any{
		"category":    "building",
		"id":          "BASIC.A",
		"name":        "A",
		"parentId":    "BASIC",
		"domain":      "test-domain",
		"description": "",
		"attributes": map[string]any{
			"height":     float64(5),
			"heightUnit": "m",
			"rotation":   30.5,
			"posXY":      []float64{4.6666666666667, -2},
			"posXYUnit":  "m",
			"size":       []float64{3, 3},
			"sizeUnit":   "m",
		},
	})

	err := controller.CreateObject("/Physical/BASIC/A", models.BLDG, maps.Clone(baseBuilding))
	assert.Nil(t, err)
}

func TestCreateRoomInvalidSize(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)
	roomsBuilding := maps.Clone(baseBuilding)
	room := map[string]any{
		"category": "room",
		"id":       "BASIC.A.R1",
		"name":     "R1",
		"parentId": "BASIC.A",
		"domain":   "test-domain",
		"attributes": map[string]any{
			"floorUnit":       "t",
			"heightUnit":      "m",
			"rotation":        30.5,
			"axisOrientation": "+x+y",
			"posXY":           []float64{4.6666666666667, -2},
			"posXYUnit":       "m",
			"size":            []float64{},
			"sizeUnit":        "m",
		},
	}

	// with state.DebugLvl = 0
	mockGetObject(mockAPI, roomsBuilding)
	err := controller.CreateObject("/Physical/BASIC/A/R1", models.ROOM, room)
	assert.Nil(t, err)

	// with state.DebugLvl > 0
	controllers.State.DebugLvl = 1
	mockGetObject(mockAPI, roomsBuilding)
	err = controller.CreateObject("/Physical/BASIC/A/R1", models.ROOM, room)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "Invalid size attribute provided."+
		" \nIt must be an array/list/vector with 3 elements."+
		" Please refer to the wiki or manual reference"+
		" for more details on how to create objects "+
		"using this syntax")
	controllers.State.DebugLvl = 0
}

func TestCreateRoomInvalidPosXY(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)
	roomsBuilding := copyMap(baseBuilding)
	room := map[string]any{
		"category": "room",
		"id":       "BASIC.A.R1",
		"name":     "R1",
		"parentId": "BASIC.A",
		"domain":   "test-domain",
		"attributes": map[string]any{
			"floorUnit":       "t",
			"heightUnit":      "m",
			"rotation":        30.5,
			"axisOrientation": "+x+y",
			"posXY":           []float64{},
			"posXYUnit":       "m",
			"size":            []float64{2, 3, 3},
			"sizeUnit":        "m",
		},
	}

	// with state.DebugLvl = 0
	mockGetObject(mockAPI, roomsBuilding)

	err := controller.CreateObject("/Physical/BASIC/A/R1", models.ROOM, copyMap(room))
	assert.Nil(t, err)

	// with state.DebugLvl > 0
	controllers.State.DebugLvl = 1
	mockGetObject(mockAPI, roomsBuilding)
	err = controller.CreateObject("/Physical/BASIC/A/R1", models.ROOM, room)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "Invalid posXY attribute provided."+
		" \nIt must be an array/list/vector with 2 elements."+
		" Please refer to the wiki or manual reference"+
		" for more details on how to create objects "+
		"using this syntax")
	controllers.State.DebugLvl = 0
}

func TestCreateRoom(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObject(mockAPI, maps.Clone(baseBuilding))

	mockCreateObject(mockAPI, "room", map[string]any{
		"category":    "room",
		"id":          "BASIC.A.R1",
		"name":        "R1",
		"parentId":    "BASIC.A",
		"domain":      "test-domain",
		"description": "",
		"attributes": map[string]any{
			"floorUnit":       "t",
			"height":          float64(5),
			"heightUnit":      "m",
			"axisOrientation": "+x+y",
			"rotation":        30.5,
			"posXY":           []float64{4.6666666666667, -2},
			"posXYUnit":       "m",
			"size":            []float64{3, 3},
			"sizeUnit":        "m",
		},
	})

	err := controller.CreateObject("/Physical/BASIC/A/R1", models.ROOM, map[string]any{
		"category": "room",
		"id":       "BASIC.A.R1",
		"name":     "R1",
		"parentId": "BASIC.A",
		"domain":   "test-domain",
		"attributes": map[string]any{
			"floorUnit":       "t",
			"heightUnit":      "m",
			"rotation":        30.5,
			"axisOrientation": "+x+y",
			"posXY":           []float64{4.6666666666667, -2},
			"posXYUnit":       "m",
			"size":            []float64{3, 3, 5},
			"sizeUnit":        "m",
		},
	})
	assert.Nil(t, err)
}

func TestCreateRackInvalidSize(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)
	room := map[string]any{
		"category": "room",
		"children": []any{},
		"id":       "BASIC.A.R1",
		"name":     "R1",
		"parentId": "BASIC.A",
		"domain":   "test-domain",
	}
	rack := map[string]any{
		"category": "rack",
		"id":       "BASIC.A.R1.A01",
		"name":     "A01",
		"parentId": "BASIC.A.R1",
		"domain":   "test-domain",
		"attributes": map[string]any{
			"heightUnit": "U",
			"rotation":   []float64{45, 45, 45},
			"posXYZ":     []float64{4.6666666666667, -2, 0},
			"posXYUnit":  "m",
			"size":       []float64{},
			"sizeUnit":   "cm",
		},
	}

	// with state.DebugLvl = 0
	mockGetObject(mockAPI, room)
	err := controller.CreateObject("/Physical/BASIC/A/R1/A01", models.RACK, rack)
	assert.Nil(t, err)

	// with state.DebugLvl > 0
	controllers.State.DebugLvl = 1
	mockGetObject(mockAPI, room)
	err = controller.CreateObject("/Physical/BASIC/A/R1/A01", models.RACK, rack)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "Invalid size attribute/template provided."+
		" \nThe size must be an array/list/vector with "+
		"3 elements."+"\n\nIf you have provided a"+
		" template, please check that you are referring to "+
		"an existing template"+
		"\n\nFor more information "+
		"please refer to the wiki or manual reference"+
		" for more details on how to create objects "+
		"using this syntax")
	controllers.State.DebugLvl = 0
}

func TestCreateRack(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObject(mockAPI, map[string]any{
		"category": "room",
		"children": []any{},
		"id":       "BASIC.A.R1",
		"name":     "R1",
		"parentId": "BASIC.A",
		"domain":   "test-domain",
	})

	mockCreateObject(mockAPI, "rack", map[string]any{
		"category":    "rack",
		"id":          "BASIC.A.R1.A01",
		"name":        "A01",
		"parentId":    "BASIC.A.R1",
		"domain":      "test-domain",
		"description": "",
		"attributes": map[string]any{
			"height":     float64(47),
			"heightUnit": "U",
			"rotation":   []float64{45, 45, 45},
			"posXYZ":     []float64{4.6666666666667, -2, 0},
			"posXYUnit":  "m",
			"size":       []float64{1, 1},
			"sizeUnit":   "cm",
		},
	})

	err := controller.CreateObject("/Physical/BASIC/A/R1/A01", models.RACK, map[string]any{
		"category": "rack",
		"id":       "BASIC.A.R1.A01",
		"name":     "A01",
		"parentId": "BASIC.A.R1",
		"domain":   "test-domain",
		"attributes": map[string]any{
			"heightUnit": "U",
			"rotation":   []float64{45, 45, 45},
			"posXYZ":     []float64{4.6666666666667, -2, 0},
			"posXYUnit":  "m",
			"size":       []float64{1, 1, 47},
			"sizeUnit":   "cm",
		},
	})
	assert.Nil(t, err)
}

func TestCreateDevice(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObject(mockAPI, map[string]any{
		"category": "rack",
		"children": []any{},
		"id":       "BASIC.A.R1.A01",
		"name":     "A01",
		"parentId": "BASIC.A.R1",
		"domain":   "test-domain",
	})

	mockCreateObject(mockAPI, "device", map[string]any{
		"category":    "device",
		"id":          "BASIC.A.R1.A01.D1",
		"name":        "D1",
		"parentId":    "BASIC.A.R1.A01",
		"domain":      "test-domain",
		"description": "",
		"attributes": map[string]any{
			"height":      47,
			"heightUnit":  "U",
			"orientation": "front",
			"size":        []float64{1, 1},
			"sizeUnit":    "cm",
		},
	})

	err := controller.CreateObject("/Physical/BASIC/A/R1/A01/D1", models.DEVICE, map[string]any{
		"category": "device",
		"id":       "BASIC.A.R1.A01.D1",
		"name":     "D1",
		"parentId": "BASIC.A.R1.A01",
		"domain":   "test-domain",
		"attributes": map[string]any{
			"height":      47,
			"heightUnit":  "U",
			"orientation": "front",
			"size":        []float64{1, 1},
			"sizeUnit":    "cm",
		},
	})
	assert.Nil(t, err)
}

func TestCreateDeviceWithSizeU(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetResponse := map[string]any{
		"category": "rack",
		"children": []any{},
		"id":       "BASIC.A.R1.A01",
		"name":     "A01",
		"parentId": "BASIC.A.R1",
		"domain":   "test-domain",
	}

	mockCreateResponse := map[string]any{
		"category":    "device",
		"id":          "BASIC.A.R1.A01.D1",
		"name":        "D1",
		"parentId":    "BASIC.A.R1.A01",
		"domain":      "test-domain",
		"description": "",
		"attributes": map[string]any{
			"height":      float64(89),
			"sizeU":       float64(2),
			"heightUnit":  "U",
			"orientation": "front",
			"size":        []float64{1, 1},
			"sizeUnit":    "cm",
		},
	}

	// SizeU of int type
	mockGetObject(mockAPI, mockGetResponse)
	mockCreateObject(mockAPI, "device", mockCreateResponse)
	err := controller.CreateObject("/Physical/BASIC/A/R1/A01/D1", models.DEVICE, map[string]any{
		"category": "device",
		"id":       "BASIC.A.R1.A01.D1",
		"name":     "D1",
		"parentId": "BASIC.A.R1.A01",
		"domain":   "test-domain",
		"attributes": map[string]any{
			"sizeU":       float64(2),
			"heightUnit":  "U",
			"orientation": "front",
			"size":        []float64{1, 1},
			"sizeUnit":    "cm",
		},
	})
	assert.Nil(t, err)

	// SizeU of float type
	mockGetObject(mockAPI, mockGetResponse)
	mockCreateObject(mockAPI, "device", mockCreateResponse)
	err = controller.CreateObject("/Physical/BASIC/A/R1/A01/D1", models.DEVICE, map[string]any{
		"category": "device",
		"id":       "BASIC.A.R1.A01.D1",
		"name":     "D1",
		"parentId": "BASIC.A.R1.A01",
		"domain":   "test-domain",
		"attributes": map[string]any{
			"sizeU":       2.0,
			"heightUnit":  "U",
			"orientation": "front",
			"size":        []float64{1, 1},
			"sizeUnit":    "cm",
		},
	})
	assert.Nil(t, err)
}

func TestCreateGroup(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObject(mockAPI, baseSite)

	mockCreateObject(mockAPI, "group", map[string]any{
		"attributes": map[string]any{
			"content": []string{"R1", "R2"},
		},
		"category":    "group",
		"description": "",
		"domain":      "test-domain",
		"name":        "G1",
		"parentId":    "BASIC",
	})

	err := controller.CreateObject("/Physical/BASIC/G1", models.GROUP, map[string]any{
		"attributes": map[string]any{
			"content": []string{"R1", "R2"},
		},
		"category":    "group",
		"description": "",
		"domain":      "test-domain",
		"name":        "G1",
		"parentId":    "BASIC",
	})
	assert.Nil(t, err)
}

func TestCreateTag(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)
	color := "D0FF78"
	slug := "my-tag"

	mockCreateObject(mockAPI, "tag", map[string]any{
		"color":       color,
		"description": slug,
		"slug":        slug,
	})

	err := controller.CreateTag(slug, color)
	assert.Nil(t, err)
}
