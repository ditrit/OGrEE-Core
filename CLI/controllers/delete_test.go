package controllers_test

import (
	"cli/controllers"
	"cli/models"
	test_utils "cli/test"
	"testing"

	"github.com/stretchr/testify/assert"
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
