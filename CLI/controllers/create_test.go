package controllers_test

import (
	"cli/controllers"
	l "cli/logger"
	"cli/models"
	test_utils "cli/test"
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

func init() {
	l.InitLogs()
}

func TestCreateObjectPathErrors(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		errorMessage string
	}{
		{"InvalidPath", "/.", "invalid path name provided for OCLI object creation"},
		{"ParentNotFound", "/", "parent not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller, _, _ := layersSetup(t)

			err := controller.CreateObject(tt.path, models.RACK, map[string]any{})
			assert.NotNil(t, err)
			assert.ErrorContains(t, err, tt.errorMessage)
		})
	}
}

func TestCreateObjectWithTemplateErrors(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)
	rack := map[string]any{
		"attributes": map[string]any{
			"rotation": []float64{0, 0, 0},
			"template": "not-exists",
		},
	}

	// Template does not exist
	test_utils.MockGetObject(mockAPI, createRoom)
	test_utils.MockObjectNotFound(mockAPI, "/api/obj_templates/not-exists")

	err := controller.CreateObject("/Physical/BASIC/A/R1/A01", models.RACK, rack)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "template not found")

	// Template of incorrect category
	rack["attributes"].(map[string]any)["template"] = "device-template"
	test_utils.MockGetObject(mockAPI, createRoom)
	test_utils.MockGetObjTemplate(mockAPI, map[string]any{
		"category": "device",
		"slug":     "device-template",
	})

	err = controller.CreateObject("/Physical/BASIC/A/R1/A01", models.RACK, rack)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "template of category device is not applicable to rack")
}

func TestCreateGenericWithoutTemplateWorks(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	test_utils.MockGetObject(mockAPI, createRoom)

	genericObject := test_utils.GetEntity("generic", "A01", createRoom["id"].(string), createRoom["domain"].(string))
	test_utils.MockCreateObject(mockAPI, "generic", genericObject)

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

	test_utils.MockGetObject(mockAPI, createRoom)

	genericTableTemplate := test_utils.GetEntity("genericTableTemplate", "generic-template", "", "")
	test_utils.MockGetObjTemplate(mockAPI, genericTableTemplate)

	test_utils.MockCreateObject(mockAPI, "generic", map[string]any{
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
	}, false)
	assert.Nil(t, err)
}

func TestCreateDomain(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	// domain with no parent
	test_utils.MockCreateObject(mockAPI, "domain", test_utils.GetEntity("domain", "dom1", "", ""))

	err := controller.CreateObject("/Organisation/Domain/dom1", models.DOMAIN, map[string]any{
		"category":    "domain",
		"id":          "dom1",
		"name":        "dom1",
		"description": "",
	})
	assert.Nil(t, err)

	// domain with parent
	test_utils.MockGetObjectByEntity(mockAPI, "domains", map[string]any{
		"category": "domain",
		"id":       "domParent",
		"name":     "domParent",
		"parentId": "",
	})

	test_utils.MockCreateObject(mockAPI, "domain", test_utils.GetEntity("domain", "dom2", "domParent", ""))

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

	test_utils.MockGetObject(mockAPI, baseSite)
	err := controller.CreateObject("/Physical/BASIC/A", models.BLDG, buildingInvalidSize)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "invalid size attribute provided."+
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

	test_utils.MockGetObject(mockAPI, baseSite)
	err := controller.CreateObject("/Physical/BASIC/A", models.BLDG, buildingInvalidPosXY)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "invalid posXY attribute provided."+
		" \nIt must be an array/list/vector with 2 elements."+
		" Please refer to the wiki or manual reference"+
		" for more details on how to create objects "+
		"using this syntax")
	controllers.State.DebugLvl = 0
	buildingInvalidPosXY["attributes"].(map[string]any)["posXY"] = []float64{4.6666666666667, -2}
}

func TestCreateBuilding(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	test_utils.MockGetObject(mockAPI, baseSite)
	test_utils.MockCreateObject(mockAPI, "building", map[string]any{
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

	test_utils.MockGetObject(mockAPI, roomsBuilding)
	err := controller.CreateObject("/Physical/BASIC/A/R1", models.ROOM, room)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "invalid size attribute provided."+
		" \nIt must be an array/list/vector with 3 elements."+
		" Please refer to the wiki or manual reference"+
		" for more details on how to create objects "+
		"using this syntax")
	controllers.State.DebugLvl = 0
}

func TestCreateRoomInvalidPosXY(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)
	roomsBuilding := test_utils.CopyMap(baseBuilding)
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

	test_utils.MockGetObject(mockAPI, roomsBuilding)
	err := controller.CreateObject("/Physical/BASIC/A/R1", models.ROOM, room)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "invalid posXY attribute provided."+
		" \nIt must be an array/list/vector with 2 elements."+
		" Please refer to the wiki or manual reference"+
		" for more details on how to create objects "+
		"using this syntax")
	controllers.State.DebugLvl = 0
}

func TestCreateRoom(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	test_utils.MockGetObject(mockAPI, maps.Clone(baseBuilding))

	test_utils.MockCreateObject(mockAPI, "room", map[string]any{
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

	test_utils.MockGetObject(mockAPI, room)
	err := controller.CreateObject("/Physical/BASIC/A/R1/A01", models.RACK, rack)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "invalid size attribute provided.")
	controllers.State.DebugLvl = 0
}

func TestCreateRack(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	test_utils.MockGetObject(mockAPI, map[string]any{
		"category": "room",
		"children": []any{},
		"id":       "BASIC.A.R1",
		"name":     "R1",
		"parentId": "BASIC.A",
		"domain":   "test-domain",
	})

	test_utils.MockCreateObject(mockAPI, "rack", map[string]any{
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

	test_utils.MockGetObject(mockAPI, test_utils.GetEntity("rack", "A01", "BASIC.A.R1", "test-domain"))

	device := test_utils.GetEntity("device", "D1", "BASIC.A.R1.A01", "test-domain")

	test_utils.MockCreateObject(mockAPI, "device", device)

	err := controller.CreateObject("/Physical/BASIC/A/R1/A01/D1", models.DEVICE, device)
	assert.Nil(t, err)
}

func TestCreateDeviceWithSizeU(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetResponse := test_utils.GetEntity("rack", "A01", "BASIC.A.R1", "test-domain")
	sizeU := float64(2)
	height := sizeU * 44.5
	mockCreateResponse := map[string]any{
		"category":    "device",
		"id":          "BASIC.A.R1.A01.D1",
		"name":        "D1",
		"parentId":    "BASIC.A.R1.A01",
		"domain":      "test-domain",
		"description": "",
		"attributes": map[string]any{
			"height":      height,
			"sizeU":       sizeU,
			"heightUnit":  "U",
			"orientation": "front",
			"size":        []float64{1, 1},
			"sizeUnit":    "cm",
		},
	}

	test_utils.MockGetObject(mockAPI, mockGetResponse)
	test_utils.MockCreateObject(mockAPI, "device", mockCreateResponse)
	err := controller.CreateObject("/Physical/BASIC/A/R1/A01/D1", models.DEVICE, map[string]any{
		"category": "device",
		"id":       "BASIC.A.R1.A01.D1",
		"name":     "D1",
		"parentId": "BASIC.A.R1.A01",
		"domain":   "test-domain",
		"attributes": map[string]any{
			"sizeU":       sizeU,
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

	test_utils.MockGetObject(mockAPI, baseSite)

	object := map[string]any{
		"attributes": map[string]any{
			"content": []string{"R1", "R2"},
		},
		"category":    "group",
		"description": "",
		"domain":      "test-domain",
		"name":        "G1",
		"parentId":    "BASIC",
	}

	test_utils.MockCreateObject(mockAPI, "group", object)

	err := controller.CreateObject("/Physical/BASIC/G1", models.GROUP, object)
	assert.Nil(t, err)
}

func TestCreateTag(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)
	color := "D0FF78"
	slug := "my-tag"

	test_utils.MockCreateObject(mockAPI, "tag", map[string]any{
		"color":       color,
		"description": slug,
		"slug":        slug,
	})

	err := controller.CreateTag(slug, color)
	assert.Nil(t, err)
}

// Tests GetSlot
func TestGetSlotWithNoTemplate(t *testing.T) {
	rack := map[string]any{
		"attributes": map[string]any{},
	}
	result, err := controllers.C.GetSlot(rack, "")
	assert.Nil(t, err)
	assert.Nil(t, result)

	rack["attributes"].(map[string]any)["template"] = ""
	result, err = controllers.C.GetSlot(rack, "")
	assert.Nil(t, err)
	assert.Nil(t, result)
}

func TestGetSlotWithTemplateNonExistentSlot(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	template := map[string]any{
		"slug":        "rack-template",
		"description": "",
		"category":    "rack",
		"sizeWDHmm":   []any{605, 1200, 2003},
		"fbxModel":    "",
		"attributes": map[string]any{
			"vendor": "IBM",
			"model":  "9360-4PX",
		},
		"slots": []any{},
	}

	test_utils.MockGetObjTemplate(mockAPI, template)
	rack := map[string]any{
		"attributes": map[string]any{
			"template": "rack-template",
		},
	}
	_, err := controller.GetSlot(rack, "u02")
	assert.NotNil(t, err)
	assert.Equal(t, "the slot u02 does not exist", err.Error())
}

func TestGetSlotWithTemplateWorks(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)
	slot := map[string]any{
		"location":   "u01",
		"type":       "u",
		"elemOrient": []any{33.3, -44.4, 107},
		"elemPos":    []any{58, 51, 44.45},
		"elemSize":   []any{482.6, 1138, 44.45},
		"mandatory":  "no",
		"labelPos":   "frontrear",
	}

	template := map[string]any{
		"slug":        "rack-template",
		"description": "",
		"category":    "rack",
		"sizeWDHmm":   []any{605, 1200, 2003},
		"fbxModel":    "",
		"attributes": map[string]any{
			"vendor": "IBM",
			"model":  "9360-4PX",
		},
		"slots": []any{
			slot,
		},
	}

	test_utils.MockGetObjTemplate(mockAPI, template)
	rack := map[string]any{
		"attributes": map[string]any{
			"template": "rack-template",
		},
	}
	result, err := controller.GetSlot(rack, "u01")
	assert.Nil(t, err)
	assert.Equal(t, slot["location"], result["location"])
}

// Tests GetByAttr
func TestGetByAttrErrorWhenObjIsNotRack(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	test_utils.MockGetObjectHierarchy(mockAPI, chassis)

	err := controller.GetByAttr(models.PhysicalPath+"BASIC/A/R1/A01/chT", "colors")
	assert.NotNil(t, err)
	assert.Equal(t, "command may only be performed on rack objects", err.Error())
}

func TestGetByAttrErrorWhenObjIsRackWithSlotName(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	rack := test_utils.CopyMap(rack1)
	rack["attributes"] = map[string]any{
		"slot": []any{
			map[string]any{
				"location":   "u01",
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

	err := controller.GetByAttr(models.PhysicalPath+"BASIC/A/R1/A01", "u01")
	assert.Nil(t, err)
}

func TestGetByAttrErrorWhenObjIsRackWithHeight(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	rack := test_utils.CopyMap(rack1)
	rack["height"] = "47"
	test_utils.MockGetObjectHierarchy(mockAPI, rack)

	err := controller.GetByAttr(models.PhysicalPath+"BASIC/A/R1/A01", 47)
	assert.Nil(t, err)
}
