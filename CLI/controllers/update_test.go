package controllers_test

import (
	"cli/controllers"
	"cli/models"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateTagColorSendsTagTo3DWithSameOldSlugAsSlug(t *testing.T) {
	controller, mockAPI, mockOgree3D := newControllerWithMocks(t)

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

	mockAPI.On("Request", http.MethodPatch, mock.Anything, dataUpdate, http.StatusOK).Return(
		&controllers.Response{
			Body: map[string]any{
				"data": dataUpdated,
			},
		}, nil,
	)

	mockOgree3D.On("InformOptional", "UpdateObj", controllers.TAG, map[string]any{
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
	controller, mockAPI, mockOgree3D := newControllerWithMocks(t)

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

	mockAPI.On("Request", http.MethodPatch, mock.Anything, dataUpdate, http.StatusOK).Return(
		&controllers.Response{
			Body: map[string]any{
				"data": dataUpdated,
			},
		}, nil,
	)

	mockOgree3D.On("InformOptional", "UpdateObj", controllers.TAG, map[string]any{
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
