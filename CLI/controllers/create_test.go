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
