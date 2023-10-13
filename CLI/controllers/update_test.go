package controllers_test

import (
	"cli/controllers"
	mocks "cli/mocks/controllers"
	"cli/models"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateTagColorSendsTagTo3DWithSameOldSlugAsSlug(t *testing.T) {
	mockAPI := mocks.NewAPIPort(t)
	mockOgree3D := mocks.NewOgree3DPort(t)
	mockGetController := mocks.NewGetController(t)
	updateController := controllers.UpdateControllerImpl{
		GetController: mockGetController,
		API:           mockAPI,
		Ogree3D:       mockOgree3D,
	}

	oldSlug := "slug"
	path := models.TagsPath + oldSlug

	mockGetController.On("GetObject", path).Return(map[string]any{
		"slug":  oldSlug,
		"name":  "name",
		"color": "aaaaaa",
	}, nil)

	dataUpdate := map[string]any{
		"color": "aaaaab",
	}
	dataUpdated := map[string]any{
		"slug":  oldSlug,
		"name":  "name",
		"color": "aaaaab",
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

	_, err := updateController.UpdateObj(path, dataUpdate)
	assert.Nil(t, err)
}

func TestUpdateTagSlugSendsTagTo3DWithNewSlug(t *testing.T) {
	mockAPI := mocks.NewAPIPort(t)
	mockOgree3D := mocks.NewOgree3DPort(t)
	mockGetController := mocks.NewGetController(t)
	updateController := controllers.UpdateControllerImpl{
		GetController: mockGetController,
		API:           mockAPI,
		Ogree3D:       mockOgree3D,
	}

	oldSlug := "slug"
	newSlug := "new-slug"

	path := models.TagsPath + oldSlug

	mockGetController.On("GetObject", path).Return(map[string]any{
		"slug":  oldSlug,
		"name":  "name",
		"color": "aaaaaa",
	}, nil)

	dataUpdate := map[string]any{
		"slug": newSlug,
	}
	dataUpdated := map[string]any{
		"slug":  newSlug,
		"name":  "name",
		"color": "aaaaaa",
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

	_, err := updateController.UpdateObj(path, dataUpdate)
	assert.Nil(t, err)
}
