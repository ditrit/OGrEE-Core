package controllers_test

import (
	"cli/controllers"
	"cli/models"
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
