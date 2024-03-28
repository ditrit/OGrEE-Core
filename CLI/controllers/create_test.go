package controllers_test

import (
	"cli/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

var createRoom = map[string]any{
	"category": "room",
	"id":       "BASIC.A.R1",
	"name":     "R1",
	"parentId": "BASIC.A",
	"domain":   "test-domain",
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
			"height":     "1",
			"heightUnit": "cm",
			"rotation":   `[0, 0, 0]`,
			"posXYZ":     `[1, 1, 1]`,
			"posXYUnit":  "m",
			"size":       `[1, 1]`,
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
			"height":     "26.330000000000002",
			"heightUnit": "cm",
			"rotation":   `[0, 0, 0]`,
			"posXYZ":     `[1, 1, 1]`,
			"posXYUnit":  "m",
			"size":       `[44.7, 91.45]`,
			"sizeUnit":   "cm",
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

	mockGetObject(mockAPI, map[string]any{
		"category": "site",
		"children": []any{},
		"id":       "BASIC",
		"name":     "BASIC",
		"parentId": "",
		"domain":   "test-domain",
	})

	err := controller.CreateObject("/Physical/BASIC/A", models.BLDG, map[string]any{
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
			"size":       []float64{},
			"sizeUnit":   "m",
		},
	})
	// returns nil but the object is not created
	assert.Nil(t, err)
}

func TestCreateBuildingInvalidPosXY(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObject(mockAPI, map[string]any{
		"category": "site",
		"children": []any{},
		"id":       "BASIC",
		"name":     "BASIC",
		"parentId": "",
		"domain":   "test-domain",
	})

	err := controller.CreateObject("/Physical/BASIC/A", models.BLDG, map[string]any{
		"category": "building",
		"id":       "BASIC.A",
		"name":     "A",
		"parentId": "BASIC",
		"domain":   "test-domain",
		"attributes": map[string]any{
			"heightUnit": "m",
			"rotation":   30.5,
			"posXY":      []float64{},
			"posXYUnit":  "m",
			"size":       []float64{2, 2, 3},
			"sizeUnit":   "m",
		},
	})
	// returns nil but the object is not created
	assert.Nil(t, err)
}

func TestCreateBuilding(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObject(mockAPI, map[string]any{
		"category": "site",
		"children": []any{},
		"id":       "BASIC",
		"name":     "BASIC",
		"parentId": "",
		"domain":   "test-domain",
	})

	mockCreateObject(mockAPI, "building", map[string]any{
		"category":    "building",
		"id":          "BASIC.A",
		"name":        "A",
		"parentId":    "BASIC",
		"domain":      "test-domain",
		"description": "",
		"attributes": map[string]any{
			"height":     "5",
			"heightUnit": "m",
			"rotation":   `30.5`,
			"posXY":      `[4.6666666666667, -2]`,
			"posXYUnit":  "m",
			"size":       `[3, 3]`,
			"sizeUnit":   "m",
		},
	})

	err := controller.CreateObject("/Physical/BASIC/A", models.BLDG, map[string]any{
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
	})
	assert.Nil(t, err)
}

func TestCreateRoomInvalidSize(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObject(mockAPI, map[string]any{
		"category": "building",
		"children": []any{},
		"id":       "BASIC.A",
		"name":     "A",
		"parentId": "BASIC",
		"domain":   "test-domain",
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
			"size":            []float64{},
			"sizeUnit":        "m",
		},
	})
	assert.Nil(t, err)
}

func TestCreateRoomInvalidPosXY(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObject(mockAPI, map[string]any{
		"category": "building",
		"children": []any{},
		"id":       "BASIC.A",
		"name":     "A",
		"parentId": "BASIC",
		"domain":   "test-domain",
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
			"posXY":           []float64{},
			"posXYUnit":       "m",
			"size":            []float64{2, 3, 3},
			"sizeUnit":        "m",
		},
	})
	assert.Nil(t, err)
}

func TestCreateRoom(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObject(mockAPI, map[string]any{
		"category": "building",
		"children": []any{},
		"id":       "BASIC.A",
		"name":     "A",
		"parentId": "BASIC",
		"domain":   "test-domain",
	})

	mockCreateObject(mockAPI, "room", map[string]any{
		"category":    "room",
		"id":          "BASIC.A.R1",
		"name":        "R1",
		"parentId":    "BASIC.A",
		"domain":      "test-domain",
		"description": "",
		"attributes": map[string]any{
			"floorUnit":       "t",
			"height":          "5",
			"heightUnit":      "m",
			"axisOrientation": "+x+y",
			"rotation":        `30.5`,
			"posXY":           `[4.6666666666667, -2]`,
			"posXYUnit":       "m",
			"size":            `[3, 3]`,
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
			"height":     "47",
			"heightUnit": "U",
			"rotation":   `[45, 45, 45]`,
			"posXYZ":     `[4.6666666666667, -2, 0]`,
			"posXYUnit":  "m",
			"size":       `[1, 1]`,
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
			"height":      "47",
			"heightUnit":  "U",
			"orientation": "front",
			"size":        `[1,1]`,
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
			"height":      "47",
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
			"height":      "89",
			"sizeU":       "2",
			"heightUnit":  "U",
			"orientation": "front",
			"size":        `[1,1]`,
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
			"sizeU":       2,
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

	mockGetObject(mockAPI, map[string]any{
		"category": "site",
		"children": []any{},
		"id":       "BASIC",
		"name":     "BASIC",
		"parentId": "BASIC",
		"domain":   "test-domain",
	})

	mockCreateObject(mockAPI, "group", map[string]any{
		"attributes": map[string]any{
			"content": "R1,R2",
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
