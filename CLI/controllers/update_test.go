package controllers_test

import (
	"cli/models"
	test_utils "cli/test"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateDeviceSizeUmm(t *testing.T) {
	controller, mockAPI, _, _ := test_utils.NewControllerWithMocks(t)

	path := models.PhysicalIDToPath("BASIC.A.R1.A01.chU1")

	device := test_utils.GetEntity("device", "chU1", "BASIC.A.R1.A01", "test")
	test_utils.MockGetObject(mockAPI, device)

	dataUpdate := map[string]any{
		"attributes": map[string]any{
			"sizeU": 1,
		},
	}
	mockDataUpdate := map[string]any{
		"attributes": map[string]any{
			"sizeU":  float32(1),
			"height": float32(44.45),
		},
	}
	dataUpdated := map[string]any{
		"attributes": map[string]any{
			"height":       44.45,
			"heightUnit":   "mm",
			"invertOffset": false,
			"orientation":  "front",
			"posU":         1,
			"size":         []float64{60, 120},
			"sizeU":        1,
			"sizeUnit":     "mm",
			"type":         "chassis",
		},
		"category":    "device",
		"createdDate": "2024-07-04T15:33:57.941Z",
		"description": "poor chassis",
		"domain":      "test",
		"id":          "BASIC.A.R1.A01.chU1",
		"lastUpdated": "2024-07-04T16:33:12.22Z",
		"name":        "chU1",
		"parentId":    "BASIC.A.R1.A01",
		"tags":        []any{},
	}

	test_utils.MockUpdateObject(mockAPI, mockDataUpdate, dataUpdated)
	_, err := controller.UpdateObj(path, dataUpdate, false)
	assert.Nil(t, err)
}
