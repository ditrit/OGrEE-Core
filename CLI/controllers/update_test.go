package controllers_test

import (
	"cli/models"
	test_utils "cli/test"
	"testing"

	"github.com/stretchr/testify/assert"
)

// region tags

func TestUpdateTagColor(t *testing.T) {
	controller, mockAPI, _, _ := test_utils.NewControllerWithMocks(t)

	oldSlug := "slug"
	path := models.TagsPath + oldSlug

	test_utils.MockGetObjectByEntity(mockAPI, "tags", map[string]any{
		"slug":        oldSlug,
		"description": "description",
		"color":       "aaaaaa",
	})

	dataUpdate := map[string]any{
		"color": "aaaaab",
	}
	dataUpdated := map[string]any{
		"slug":        oldSlug,
		"description": "description",
		"color":       "aaaaab",
	}

	test_utils.MockUpdateObject(mockAPI, dataUpdate, dataUpdated)
	_, err := controller.UpdateObj(path, dataUpdate, false)
	assert.Nil(t, err)
}

func TestUpdateTagSlug(t *testing.T) {
	controller, mockAPI, _, _ := test_utils.NewControllerWithMocks(t)

	oldSlug := "slug"
	newSlug := "new-slug"

	path := models.TagsPath + oldSlug

	test_utils.MockGetObjectByEntity(mockAPI, "tags", map[string]any{
		"slug":        oldSlug,
		"description": "description",
		"color":       "aaaaaa",
	})

	dataUpdate := map[string]any{
		"slug": newSlug,
	}
	dataUpdated := map[string]any{
		"slug":        newSlug,
		"description": "description",
		"color":       "aaaaaa",
	}

	test_utils.MockUpdateObject(mockAPI, dataUpdate, dataUpdated)
	_, err := controller.UpdateObj(path, dataUpdate, false)
	assert.Nil(t, err)
}

//endregion tags

// region device's sizeU

// Test an update of a device's sizeU with heightUnit == mm
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
			"sizeU":  float64(1),
			"height": 44.45,
		},
	}
	device["attributes"].(map[string]any)["sizeU"] = 1
	device["attributes"].(map[string]any)["height"] = 44.45

	test_utils.MockUpdateObject(mockAPI, mockDataUpdate, device)
	_, err := controller.UpdateObj(path, dataUpdate, false)
	assert.Nil(t, err)
}

// Test an update of a device's sizeU with heightUnit == cm
func TestUpdateDeviceSizeUcm(t *testing.T) {
	controller, mockAPI, _, _ := test_utils.NewControllerWithMocks(t)

	path := models.PhysicalIDToPath("BASIC.A.R1.A01.chU1")

	device := test_utils.GetEntity("device", "chU1", "BASIC.A.R1.A01", "test")
	device["attributes"].(map[string]any)["heightUnit"] = "cm"

	test_utils.MockGetObject(mockAPI, device)

	dataUpdate := map[string]any{
		"attributes": map[string]any{
			"sizeU": 1,
		},
	}
	mockDataUpdate := map[string]any{
		"attributes": map[string]any{
			"sizeU":  float64(1),
			"height": 4.445,
		},
	}
	device["attributes"].(map[string]any)["sizeU"] = 1
	device["attributes"].(map[string]any)["height"] = 4.445

	test_utils.MockUpdateObject(mockAPI, mockDataUpdate, device)
	_, err := controller.UpdateObj(path, dataUpdate, false)
	assert.Nil(t, err)
}

// endregion sizeU

// region device's height

// Test an update of a device's height with heightUnit == mm
func TestUpdateDeviceheightmm(t *testing.T) {
	controller, mockAPI, _, _ := test_utils.NewControllerWithMocks(t)

	path := models.PhysicalIDToPath("BASIC.A.R1.A01.chU1")

	device := test_utils.GetEntity("device", "chU1", "BASIC.A.R1.A01", "test")
	test_utils.MockGetObject(mockAPI, device)

	dataUpdate := map[string]any{
		"attributes": map[string]any{
			"height": 44.45,
		},
	}
	mockDataUpdate := map[string]any{
		"attributes": map[string]any{
			"sizeU":  float64(1),
			"height": 44.45,
		},
	}
	device["attributes"].(map[string]any)["sizeU"] = 1
	device["attributes"].(map[string]any)["height"] = 44.45

	test_utils.MockUpdateObject(mockAPI, mockDataUpdate, device)
	_, err := controller.UpdateObj(path, dataUpdate, false)
	assert.Nil(t, err)
}

// Test an update of a device's height with heightUnit == cm
func TestUpdateDeviceheightcm(t *testing.T) {
	controller, mockAPI, _, _ := test_utils.NewControllerWithMocks(t)

	path := models.PhysicalIDToPath("BASIC.A.R1.A01.chU1")

	device := test_utils.GetEntity("device", "chU1", "BASIC.A.R1.A01", "test")
	device["attributes"].(map[string]any)["heightUnit"] = "cm"
	test_utils.MockGetObject(mockAPI, device)

	dataUpdate := map[string]any{
		"attributes": map[string]any{
			"height": 4.445,
		},
	}
	mockDataUpdate := map[string]any{
		"attributes": map[string]any{
			"sizeU":  float64(1),
			"height": 4.445,
		},
	}
	device["attributes"].(map[string]any)["sizeU"] = 1
	device["attributes"].(map[string]any)["height"] = 4.445

	test_utils.MockUpdateObject(mockAPI, mockDataUpdate, device)
	_, err := controller.UpdateObj(path, dataUpdate, false)
	assert.Nil(t, err)
}

// endregion
