package controllers_test

import (
	"cli/models"
	test_utils "cli/test"
	"maps"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateTemplateOfTypeGenericWorks(t *testing.T) {
	controller, mockAPI, _, _ := test_utils.NewControllerWithMocks(t)

	template := test_utils.GetEntity("genericTableTemplate", "generic-template", "", "")

	test_utils.MockCreateObject(mockAPI, "obj_template", template)

	err := controller.LoadTemplate(template)
	assert.Nil(t, err)
}

func TestApplyTemplateOfTypeDeviceWorks(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	device := maps.Clone(chassis)
	attributes := map[string]any{
		"template":    "device-template",
		"TDP":         "",
		"TDPmax":      "",
		"fbxModel":    "https://github.com/test.fbx",
		"height":      40.1,
		"heightUnit":  "mm",
		"model":       "TNF2LTX",
		"orientation": "front",
		"partNumber":  "0303XXXX",
		"size":        "[388.4, 205.9]",
		"sizeUnit":    "mm",
		"weightKg":    "1.81",
	}
	device["attributes"] = attributes
	template := test_utils.GetEntity("deviceChasisTemplate", "device-template", "", "")

	test_utils.MockGetObjTemplate(mockAPI, template)

	sizeU := int((float64(template["sizeWDHmm"].([]any)[2].(int)) / 1000) / models.RACKUNIT)
	err := controller.ApplyTemplate(attributes, device, models.DEVICE)
	assert.Nil(t, err)

	// we verify if the template was applied
	assert.Equal(t, 100, device["attributes"].(map[string]any)["height"])
	assert.Equal(t, template["attributes"].(map[string]any)["type"], device["attributes"].(map[string]any)["type"])
	assert.Equal(t, template["attributes"].(map[string]any)["vendor"], device["attributes"].(map[string]any)["vendor"])
	assert.Equal(t, []any{216, 659}, device["attributes"].(map[string]any)["size"])
	assert.Equal(t, sizeU, device["attributes"].(map[string]any)["sizeU"])
}

func TestApplyTemplateOfTypeDeviceError(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	device := test_utils.CopyMap(chassis)
	attributes := map[string]any{
		"template": "device-template",
	}
	device["attributes"] = attributes
	template := test_utils.GetEntity("deviceChasisTemplate", "device-template", "", "")
	template["sizeWDHmm"] = []any{216, 659}

	test_utils.MockGetObjTemplate(mockAPI, template)

	err := controller.ApplyTemplate(attributes, device, models.DEVICE)
	assert.NotNil(t, err)

	assert.Equal(t, "invalid size vector on given template", err.Error())
}

func TestApplyTemplateOfTypeDeviceConvertError(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	device := test_utils.CopyMap(chassis)
	attributes := map[string]any{
		"template": "device-template",
	}
	device["attributes"] = attributes
	template := test_utils.GetEntity("deviceChasisTemplate", "device-template", "", "")
	template["sizeWDHmm"] = []any{216, 659, "100"}

	test_utils.MockGetObjTemplate(mockAPI, template)

	err := controller.ApplyTemplate(attributes, device, models.DEVICE)
	assert.NotNil(t, err)

	assert.Equal(t, "cannot convert string to float64", err.Error())
}

func TestApplyTemplateOfTypeRoomWorks(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	room := test_utils.CopyMap(roomWithoutChildren)
	attributes := map[string]any{
		"template":   "room-template",
		"height":     "2.8",
		"heightUnit": "m",
		"rotation":   "-90",
		"posXY":      "[0, 0]",
		"posXYUnit":  "m",
	}
	room["attributes"] = attributes
	template := map[string]any{
		"slug":            "room-template",
		"category":        "room",
		"axisOrientation": "+x+y",
		"sizeWDHm":        []any{216, 659, 41},
		"floorUnit":       "m",
		"technicalArea":   []string{"front", "front", "front", "front"},
		"reservedArea":    []string{"front", "front", "front", "front"},
		"vertices":        []any{0, 0, 0},
		"tileAngle":       0,
		"separators": map[string]any{
			"sepname": map[string]any{
				"startPosXYm": []any{0, 0},
				"endPosXYm":   []any{0, 0},
				"type":        "wireframe|plain",
			},
		},
		"pillars": map[string]any{
			"pillarname": map[string]any{
				"centerXY": []any{0, 0},
				"sizeXY":   []any{0, 0},
				"rotation": 0,
			},
		},
		"tiles": []any{
			map[string]any{
				"location": "0/0",
				"name":     "my-tile",
				"label":    "my-tile",
				"texture":  "",
				"color":    "00ED00",
			},
		},
		"colors": []any{"my-color"},
		// "rows"            : [],
		// "center"          : [0,0],
	}

	test_utils.MockGetRoomTemplate(mockAPI, template)

	err := controller.ApplyTemplate(attributes, room, models.ROOM)
	assert.Nil(t, err)

	// we verify if the template was applied
	assert.Equal(t, 41, room["attributes"].(map[string]any)["height"])
	assert.Equal(t, []any{216, 659}, room["attributes"].(map[string]any)["size"])
	assert.Equal(t, template["axisOrientation"], room["attributes"].(map[string]any)["axisOrientation"])
	assert.Equal(t, template["floorUnit"], room["attributes"].(map[string]any)["floorUnit"])
	assert.Equal(t, []any{0, 0, 0}, room["attributes"].(map[string]any)["vertices"])
	assert.Equal(t, []any{"my-color"}, room["attributes"].(map[string]any)["colors"])
}

func TestLoadTemplateRoom(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	template := map[string]any{
		"slug":        "room-example",
		"description": "room example",
		"category":    "room",
		"sizeWDHm":    []any{216, 659, 41},
	}

	test_utils.MockCreateObject(mockAPI, "room_template", template)

	err := controller.LoadTemplate(template)
	assert.Nil(t, err)
}

func TestLoadTemplateBuilding(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	template := map[string]any{
		"slug":        "building-example",
		"description": "building example",
		"category":    "building",
		"sizeWDHm":    []any{216, 659, 41},
		"center":      []any{0, 0},
	}

	test_utils.MockCreateObject(mockAPI, "bldg_template", template)

	err := controller.LoadTemplate(template)
	assert.Nil(t, err)
}

func TestLoadTemplateInvalidCategory(t *testing.T) {
	controller, _, _ := layersSetup(t)

	template := map[string]any{
		"slug":        "invalid-example",
		"description": "invalid example",
		"category":    "invalid",
		"sizeWDHm":    []any{216, 659, 41},
		"center":      []any{0, 0},
	}

	err := controller.LoadTemplate(template)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "this template does not have a valid category. Please add a category attribute with a value of building, room, rack, device or generic")
}
