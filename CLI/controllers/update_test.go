package controllers_test

import (
	"cli/controllers"
	"cli/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateTagColorSendsTagTo3DWithSameOldSlugAsSlug(t *testing.T) {
	controller, mockAPI, mockOgree3D, _ := newControllerWithMocks(t)

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

	mockOgree3D.On("InformOptional", "UpdateObj", models.TAG, map[string]any{
		"type": "modify-tag",
		"data": map[string]any{
			"old-slug": oldSlug,
			"tag":      dataUpdated,
		},
	}).Return(nil)

	controllers.State.ObjsForUnity = controllers.SetObjsForUnity([]string{"all"})

	_, err := controller.UpdateObj(path, dataUpdate)
	assert.Nil(t, err)
}

func TestUpdateTagSlugSendsTagTo3DWithNewSlug(t *testing.T) {
	controller, mockAPI, mockOgree3D, _ := newControllerWithMocks(t)

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

	mockOgree3D.On("InformOptional", "UpdateObj", models.TAG, map[string]any{
		"type": "modify-tag",
		"data": map[string]any{
			"old-slug": oldSlug,
			"tag":      dataUpdated,
		},
	}).Return(nil)

	controllers.State.ObjsForUnity = controllers.SetObjsForUnity([]string{"all"})

	_, err := controller.UpdateObj(path, dataUpdate)
	assert.Nil(t, err)
}
