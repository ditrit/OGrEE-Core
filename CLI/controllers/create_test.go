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
