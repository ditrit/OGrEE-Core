package controllers_test

import (
	"cli/controllers"
	"cli/models"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateTagColor(t *testing.T) {
	controller, mockAPI, _, _ := newControllerWithMocks(t)

	oldSlug := "slug"
	path := models.TagsPath + oldSlug

	mockGetObjectByEntity(mockAPI, "tags", map[string]any{
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

	mockUpdateObject(mockAPI, dataUpdate, dataUpdated)

	controllers.State.ObjsForUnity = controllers.SetObjsForUnity([]string{"all"})

	_, err := controller.UpdateObj(path, dataUpdate, false)
	assert.Nil(t, err)
}

func TestUpdateTagSlug(t *testing.T) {
	controller, mockAPI, _, _ := newControllerWithMocks(t)

	oldSlug := "slug"
	newSlug := "new-slug"

	path := models.TagsPath + oldSlug

	mockGetObjectByEntity(mockAPI, "tags", map[string]any{
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

	mockUpdateObject(mockAPI, dataUpdate, dataUpdated)

	controllers.State.ObjsForUnity = controllers.SetObjsForUnity([]string{"all"})

	_, err := controller.UpdateObj(path, dataUpdate, false)
	assert.Nil(t, err)
}

func TestUpdateRoomTilesColor(t *testing.T) {
	controller, mockAPI, mockOgree3D, _ := newControllerWithMocks(t)

	room := copyMap(roomWithoutChildren)
	room["attributes"] = map[string]any{
		"tilesColor": "aaaaaa",
	}
	newColor := "aaaaab"
	updatedRoom := copyMap(room)
	updatedRoom["attributes"].(map[string]any)["tilesColor"] = newColor
	dataUpdate := updatedRoom["attributes"].(map[string]any)
	entity := models.ROOM

	path := "/Physical/" + strings.Replace(room["id"].(string), ".", "/", -1)
	message := map[string]any{
		"type": "interact",
		"data": map[string]any{
			"id":    room["id"],
			"param": "tilesColor",
			"value": newColor,
		},
	}

	mockOgree3D.On("InformOptional", "UpdateObj", entity, message).Return(nil)

	mockGetObject(mockAPI, room)

	dataUpdated := copyMap(room)
	dataUpdated["attributes"].(map[string]any)["tilesColor"] = newColor

	mockUpdateObject(mockAPI, dataUpdate, dataUpdated)

	controllers.State.ObjsForUnity = controllers.SetObjsForUnity([]string{"all"})

	result, err := controller.UpdateObj(path, dataUpdate, false)
	assert.Nil(t, err)
	assert.Equal(t, result["data"].(map[string]any)["attributes"].(map[string]any)["tilesColor"], newColor)
	mockOgree3D.AssertCalled(t, "InformOptional", "UpdateObj", entity, message)
}

func TestUpdateRoomTilesName(t *testing.T) {
	controller, mockAPI, mockOgree3D, _ := newControllerWithMocks(t)

	room := copyMap(roomWithoutChildren)
	room["attributes"] = map[string]any{
		"tilesName": "t1",
	}
	newName := "t2"
	updatedRoom := copyMap(room)
	updatedRoom["attributes"].(map[string]any)["tilesName"] = newName
	dataUpdate := updatedRoom["attributes"].(map[string]any)
	entity := models.ROOM

	path := "/Physical/" + strings.Replace(room["id"].(string), ".", "/", -1)
	message := map[string]any{
		"type": "interact",
		"data": map[string]any{
			"id":    room["id"],
			"param": "tilesName",
			"value": newName,
		},
	}

	mockOgree3D.On("InformOptional", "UpdateObj", entity, message).Return(nil)

	mockGetObject(mockAPI, room)

	dataUpdated := copyMap(room)
	dataUpdated["attributes"].(map[string]any)["tilesName"] = newName

	mockUpdateObject(mockAPI, dataUpdate, dataUpdated)

	controllers.State.ObjsForUnity = controllers.SetObjsForUnity([]string{"all"})

	result, err := controller.UpdateObj(path, dataUpdate, false)
	assert.Nil(t, err)
	assert.Equal(t, result["data"].(map[string]any)["attributes"].(map[string]any)["tilesName"], newName)
	mockOgree3D.AssertCalled(t, "InformOptional", "UpdateObj", entity, message)
}

func TestUpdateRackU(t *testing.T) {
	controller, mockAPI, mockOgree3D, _ := newControllerWithMocks(t)
	rack := copyMap(rack2)
	rack["attributes"] = map[string]any{
		"U": true,
	}
	updatedRack := copyMap(rack)
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

	mockGetObject(mockAPI, rack)
	mockUpdateObject(mockAPI, dataUpdate, updatedRack)

	controllers.State.ObjsForUnity = controllers.SetObjsForUnity([]string{"all"})

	result, err := controller.UpdateObj(path, dataUpdate, false)
	assert.Nil(t, err)
	assert.False(t, result["data"].(map[string]any)["attributes"].(map[string]any)["U"].(bool))
	mockOgree3D.AssertCalled(t, "InformOptional", "UpdateObj", entity, message)
}

func TestUpdateDeviceAlpha(t *testing.T) {
	controller, mockAPI, mockOgree3D, _ := newControllerWithMocks(t)
	device := copyMap(chassis)
	device["attributes"].(map[string]any)["alpha"] = true
	updatedDevice := copyMap(device)
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

	mockGetObject(mockAPI, device)
	mockUpdateObject(mockAPI, dataUpdate, updatedDevice)

	controllers.State.ObjsForUnity = controllers.SetObjsForUnity([]string{"all"})

	result, err := controller.UpdateObj(path, dataUpdate, false)
	assert.Nil(t, err)
	assert.False(t, result["data"].(map[string]any)["attributes"].(map[string]any)["alpha"].(bool))
	mockOgree3D.AssertCalled(t, "InformOptional", "UpdateObj", entity, message)
}

func TestUpdateGroupContent(t *testing.T) {
	controller, mockAPI, mockOgree3D, _ := newControllerWithMocks(t)
	group := copyMap(rackGroup)
	group["attributes"] = map[string]any{
		"content": "A,B",
	}
	newValue := "A,B,C"
	updatedGroup := copyMap(group)
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

	mockGetObject(mockAPI, group)
	mockUpdateObject(mockAPI, dataUpdate, updatedGroup)

	controllers.State.ObjsForUnity = controllers.SetObjsForUnity([]string{"all"})

	result, err := controller.UpdateObj(path, dataUpdate, false)
	assert.Nil(t, err)
	assert.Equal(t, result["data"].(map[string]any)["attributes"].(map[string]any)["content"].(string), newValue)
	mockOgree3D.AssertCalled(t, "InformOptional", "UpdateObj", entity, message)
}
