package controllers_test

import (
	"cli/controllers"
	"cli/models"
	test_utils "cli/test"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

	controllers.State.ObjsForUnity = controllers.SetObjsForUnity([]string{"all"})

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

	controllers.State.ObjsForUnity = controllers.SetObjsForUnity([]string{"all"})

	_, err := controller.UpdateObj(path, dataUpdate, false)
	assert.Nil(t, err)
}

func TestUpdateRoomTiles(t *testing.T) {
	tests := []struct {
		name         string
		attributeKey string
		oldValue     string
		newValue     string
	}{
		{"Color", "tilesColor", "aaaaaa", "aaaaab"},
		{"Name", "tilesName", "t1", "t2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller, mockAPI, mockOgree3D, _ := test_utils.NewControllerWithMocks(t)

			room := test_utils.CopyMap(roomWithoutChildren)
			room["attributes"] = map[string]any{
				tt.attributeKey: tt.oldValue,
			}
			updatedRoom := test_utils.CopyMap(room)
			updatedRoom["attributes"].(map[string]any)[tt.attributeKey] = tt.newValue
			dataUpdate := updatedRoom["attributes"].(map[string]any)
			entity := models.ROOM

			path := "/Physical/" + strings.Replace(room["id"].(string), ".", "/", -1)
			message := map[string]any{
				"type": "interact",
				"data": map[string]any{
					"id":    room["id"],
					"param": tt.attributeKey,
					"value": tt.newValue,
				},
			}

			mockOgree3D.On("InformOptional", "UpdateObj", entity, message).Return(nil)

			test_utils.MockGetObject(mockAPI, room)

			dataUpdated := test_utils.CopyMap(room)
			dataUpdated["attributes"].(map[string]any)[tt.attributeKey] = tt.newValue

			test_utils.MockUpdateObject(mockAPI, dataUpdate, dataUpdated)

			controllers.State.ObjsForUnity = controllers.SetObjsForUnity([]string{"all"})

			result, err := controller.UpdateObj(path, dataUpdate, false)
			assert.Nil(t, err)
			assert.Equal(t, result["data"].(map[string]any)["attributes"].(map[string]any)[tt.attributeKey], tt.newValue)
			mockOgree3D.AssertCalled(t, "InformOptional", "UpdateObj", entity, message)
		})
	}
}

func TestUpdateRackU(t *testing.T) {
	controller, mockAPI, mockOgree3D, _ := test_utils.NewControllerWithMocks(t)
	rack := test_utils.CopyMap(rack2)
	rack["attributes"] = map[string]any{
		"U": true,
	}
	updatedRack := test_utils.CopyMap(rack)
	updatedRack["attributes"].(map[string]any)["U"] = false
	dataUpdate := updatedRack["attributes"].(map[string]any)
	entity := models.RACK

	path := "/Physical/" + strings.Replace(rack["id"].(string), ".", "/", -1)
	message := map[string]any{
		"type": "interact",
		"data": map[string]any{
			"id":    rack["id"],
			"param": "U",
			"value": false,
		},
	}

	mockOgree3D.On("InformOptional", "UpdateObj", entity, message).Return(nil)

	test_utils.MockGetObject(mockAPI, rack)
	test_utils.MockUpdateObject(mockAPI, dataUpdate, updatedRack)

	controllers.State.ObjsForUnity = controllers.SetObjsForUnity([]string{"all"})

	result, err := controller.UpdateObj(path, dataUpdate, false)
	assert.Nil(t, err)
	assert.False(t, result["data"].(map[string]any)["attributes"].(map[string]any)["U"].(bool))
	mockOgree3D.AssertCalled(t, "InformOptional", "UpdateObj", entity, message)
}

func TestUpdateDeviceAlpha(t *testing.T) {
	controller, mockAPI, mockOgree3D, _ := test_utils.NewControllerWithMocks(t)
	device := test_utils.CopyMap(chassis)
	device["attributes"].(map[string]any)["alpha"] = true
	updatedDevice := test_utils.CopyMap(device)
	updatedDevice["attributes"].(map[string]any)["alpha"] = false
	dataUpdate := updatedDevice["attributes"].(map[string]any)
	entity := models.DEVICE

	path := "/Physical/" + strings.Replace(device["id"].(string), ".", "/", -1)
	message := map[string]any{
		"type": "interact",
		"data": map[string]any{
			"id":    device["id"],
			"param": "alpha",
			"value": false,
		},
	}

	mockOgree3D.On("InformOptional", "UpdateObj", entity, message).Return(nil)

	test_utils.MockGetObject(mockAPI, device)
	test_utils.MockUpdateObject(mockAPI, dataUpdate, updatedDevice)

	controllers.State.ObjsForUnity = controllers.SetObjsForUnity([]string{"all"})

	result, err := controller.UpdateObj(path, dataUpdate, false)
	assert.Nil(t, err)
	assert.False(t, result["data"].(map[string]any)["attributes"].(map[string]any)["alpha"].(bool))
	mockOgree3D.AssertCalled(t, "InformOptional", "UpdateObj", entity, message)
}

func TestUpdateGroupContent(t *testing.T) {
	controller, mockAPI, mockOgree3D, _ := test_utils.NewControllerWithMocks(t)
	group := test_utils.CopyMap(rackGroup)
	group["attributes"] = map[string]any{
		"content": "A,B",
	}
	newValue := "A,B,C"
	updatedGroup := test_utils.CopyMap(group)
	updatedGroup["attributes"].(map[string]any)["content"] = newValue
	dataUpdate := updatedGroup["attributes"].(map[string]any)
	entity := models.GROUP

	path := "/Physical/" + strings.Replace(group["id"].(string), ".", "/", -1)
	message := map[string]any{
		"type": "interact",
		"data": map[string]any{
			"id":    group["id"],
			"param": "content",
			"value": newValue,
		},
	}

	mockOgree3D.On("InformOptional", "UpdateObj", entity, message).Return(nil)

	test_utils.MockGetObject(mockAPI, group)
	test_utils.MockUpdateObject(mockAPI, dataUpdate, updatedGroup)

	controllers.State.ObjsForUnity = controllers.SetObjsForUnity([]string{"all"})

	result, err := controller.UpdateObj(path, dataUpdate, false)
	assert.Nil(t, err)
	assert.Equal(t, result["data"].(map[string]any)["attributes"].(map[string]any)["content"].(string), newValue)
	mockOgree3D.AssertCalled(t, "InformOptional", "UpdateObj", entity, message)
}
