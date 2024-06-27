package controllers_test

import (
	"cli/controllers"
	"cli/models"
	test_utils "cli/test"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
)

const testRackObjPath = "/api/hierarchy_objects/BASIC.A.R1.A01"

func TestDeleteTag(t *testing.T) {
	controller, mockAPI, _, _ := test_utils.NewControllerWithMocks(t)

	slug := "slug"
	path := models.TagsPath + slug

	test_utils.MockDeleteObjects(mockAPI, "namespace=logical.tag&slug=slug", []any{
		map[string]any{
			"slug":        slug,
			"description": "description",
			"color":       "aaaaaa",
		},
	})

	controllers.State.ObjsForUnity = controllers.SetObjsForUnity([]string{"all"})

	_, err := controller.DeleteObj(path)
	assert.Nil(t, err)
}

// Tests UnsetAttribute
func TestUnsetAttributeObjectNotFound(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	test_utils.MockObjectNotFound(mockAPI, testRackObjPath)

	err := controller.UnsetAttribute("/Physical/BASIC/A/R1/A01", "color")
	assert.NotNil(t, err)
	assert.Equal(t, "object not found", err.Error())
}

func TestUnsetAttributeWorks(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)
	rack := map[string]any{
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
			"color":      "00ED00",
		},
	}
	updatedRack := test_utils.CopyMap(rack)
	delete(updatedRack["attributes"].(map[string]any), "color")
	delete(updatedRack, "id")

	test_utils.MockGetObject(mockAPI, rack)
	test_utils.MockPutObject(mockAPI, updatedRack, updatedRack)

	err := controller.UnsetAttribute("/Physical/BASIC/A/R1/A01", "color")
	assert.Nil(t, err)
}

// Tests UnsetInObj
func TestUnsetInObjInvalidIndex(t *testing.T) {
	controller, _, _ := layersSetup(t)

	result, err := controller.UnsetInObj("/Physical/BASIC/A/R1/A01", "color", -1)
	assert.NotNil(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "Index out of bounds. Please provide an index greater than 0", err.Error())
}

func TestUnsetInObjObjectNotFound(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	test_utils.MockObjectNotFound(mockAPI, testRackObjPath)

	result, err := controller.UnsetInObj("/Physical/BASIC/A/R1/A01", "color", 0)
	assert.NotNil(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "object not found", err.Error())
}

func TestUnsetInObjAttributeNotFound(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	rack := test_utils.CopyMap(rack1)
	rack["attributes"] = map[string]any{}

	test_utils.MockGetObject(mockAPI, rack)

	result, err := controller.UnsetInObj("/Physical/BASIC/A/R1/A01", "color", 0)
	assert.NotNil(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "Attribute :color was not found", err.Error())
}

func TestUnsetInObjAttributeNotAnArray(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	rack := test_utils.CopyMap(rack1)
	rack["attributes"] = map[string]any{
		"color": "00ED00",
	}

	test_utils.MockGetObject(mockAPI, rack)

	result, err := controller.UnsetInObj("/Physical/BASIC/A/R1/A01", "color", 0)
	assert.NotNil(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "Attribute is not an array", err.Error())
}

func TestUnsetInObjEmptyArray(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	rack := test_utils.CopyMap(rack1)
	rack["attributes"] = map[string]any{
		"posXYZ": []any{},
	}

	test_utils.MockGetObject(mockAPI, rack)

	result, err := controller.UnsetInObj("/Physical/BASIC/A/R1/A01", "posXYZ", 0)
	assert.NotNil(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "Cannot delete anymore elements", err.Error())
}

func TestUnsetInObjWorksWithNestedAttribute(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	rack := test_utils.CopyMap(rack1)
	rack["attributes"] = map[string]any{
		"posXYZ": []any{1, 2, 3},
	}
	updatedRack := test_utils.CopyMap(rack1)
	updatedRack["attributes"] = map[string]any{
		"posXYZ": []any{1.0, 3.0},
	}
	delete(updatedRack, "children")

	test_utils.MockGetObject(mockAPI, rack)
	test_utils.MockPutObject(mockAPI, updatedRack, updatedRack)

	result, err := controller.UnsetInObj("/Physical/BASIC/A/R1/A01", "posXYZ", 1)
	assert.Nil(t, err)
	assert.Nil(t, result)
}

func TestUnsetInObjWorksWithAttribute(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	template := map[string]any{
		"slug":            "small-room",
		"category":        "room",
		"axisOrientation": "+x+y",
		"sizeWDHm":        []any{9.6, 22.8, 3.0},
		"floorUnit":       "t",
		"technicalArea":   []any{5.0, 0.0, 0.0, 0.0},
		"reservedArea":    []any{3.0, 1.0, 1.0, 3.0},
		"colors": []any{
			map[string]any{
				"name":  "my-color1",
				"value": "00ED00",
			},
			map[string]any{
				"name":  "my-color2",
				"value": "ffffff",
			},
		},
	}
	updatedTemplate := test_utils.CopyMap(template)
	updatedTemplate["colors"] = slices.Delete(updatedTemplate["colors"].([]any), 1, 2)
	test_utils.MockPutObject(mockAPI, updatedTemplate, updatedTemplate)
	test_utils.MockGetRoomTemplate(mockAPI, template)

	result, err := controller.UnsetInObj(models.RoomTemplatesPath+"small-room", "colors", 1)
	assert.Nil(t, err)
	assert.Nil(t, result)
}
