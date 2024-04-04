package controllers_test

import (
	"cli/controllers"
	"cli/models"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateTemplateOfTypeGenericWorks(t *testing.T) {
	controller, mockAPI, _, _ := newControllerWithMocks(t)

	template := map[string]any{
		"slug":        "generic-example",
		"description": "a table",
		"category":    "generic",
		"sizeWDHmm":   []float64{447, 914.5, 263.3},
		"fbxModel":    "",
		"attributes": map[string]any{
			"type": "table",
		},
		"colors": []any{},
	}

	mockCreateObject(mockAPI, "obj-template", template)

	err := controller.LoadTemplate(template)
	assert.Nil(t, err)
}

func TestApplyTemplateOfTypeDeviceWorks(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	device := copyMap(chassis)
	attributes := map[string]any{
		"template":    "device-template",
		"TDP":         "",
		"TDPmax":      "",
		"fbxModel":    "https://github.com/test.fbx",
		"height":      "40.1",
		"heightUnit":  "mm",
		"model":       "TNF2LTX",
		"orientation": "front",
		"partNumber":  "0303XXXX",
		"size":        "[388.4, 205.9]",
		"sizeUnit":    "mm",
		"weightKg":    "1.81",
	}
	device["attributes"] = attributes
	template := map[string]any{
		"slug":        "device-template",
		"description": "",
		"category":    "device",
		"sizeWDHmm":   []any{216, 659, 100},
		"fbxModel":    "",
		"attributes": map[string]any{
			"type":   "chassis",
			"vendor": "IBM",
		},
		"colors":     []any{},
		"components": []any{},
	}

	mockGetObjTemplate(mockAPI, template)

	sizeU := int((float64(template["sizeWDHmm"].([]any)[2].(int)) / 1000) / controllers.RACKUNIT)
	err := controller.ApplyTemplate(attributes, device, models.DEVICE)
	assert.Nil(t, err)

	// we verify if the template was applied
	assert.Equal(t, "100", device["attributes"].(map[string]any)["height"])
	assert.Equal(t, template["attributes"].(map[string]any)["type"], device["attributes"].(map[string]any)["type"])
	assert.Equal(t, template["attributes"].(map[string]any)["vendor"], device["attributes"].(map[string]any)["vendor"])
	assert.Equal(t, "[216, 659]", device["attributes"].(map[string]any)["size"])
	assert.Equal(t, strconv.Itoa(sizeU), device["attributes"].(map[string]any)["sizeU"])
}

func TestApplyTemplateOfTypeDeviceError(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	device := copyMap(chassis)
	attributes := map[string]any{
		"template": "device-template",
	}
	device["attributes"] = attributes
	template := map[string]any{
		"slug":        "device-template",
		"description": "",
		"category":    "device",
		"sizeWDHmm":   []any{216, 659, "100"},
		"fbxModel":    "",
		"attributes": map[string]any{
			"type":   "chassis",
			"vendor": "IBM",
		},
		"colors":     []any{},
		"components": []any{},
	}

	mockGetObjTemplate(mockAPI, template)

	err := controller.ApplyTemplate(attributes, device, models.DEVICE)
	assert.NotNil(t, err)

	assert.Equal(t, "invalid size vector on given template", err.Error())
}

func TestApplyTemplateOfTypeRoomWorks(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	room := copyMap(roomWithoutChildren)
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

	mockGetRoomTemplate(mockAPI, template)

	err := controller.ApplyTemplate(attributes, room, models.ROOM)
	assert.Nil(t, err)

	// we verify if the template was applied
	assert.Equal(t, "41", room["attributes"].(map[string]any)["height"])
	assert.Equal(t, "[216, 659]", room["attributes"].(map[string]any)["size"])
	assert.Equal(t, template["axisOrientation"], room["attributes"].(map[string]any)["axisOrientation"])
	assert.Equal(t, template["floorUnit"], room["attributes"].(map[string]any)["floorUnit"])
	assert.Equal(t, "[0,0,0]", room["attributes"].(map[string]any)["vertices"])
	assert.Equal(t, "[\"my-color\"]", room["attributes"].(map[string]any)["colors"])
}
