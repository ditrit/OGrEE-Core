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

func TestDeleteTagSendsDeleteTo3DWithSlug(t *testing.T) {
	mockAPI := mocks.NewAPIPort(t)
	mockOgree3D := mocks.NewOgree3DPort(t)
	mockGetController := mocks.NewGetController(t)
	deleteController := controllers.DeleteControllerImpl{
		GetController: mockGetController,
		API:           mockAPI,
		Ogree3D:       mockOgree3D,
	}

	slug := "slug"
	path := models.TagsPath + slug

	mockGetController.On("GetObject", path).Return(map[string]any{
		"slug":  slug,
		"color": "aaaaaa",
	}, nil)

	var nilMap map[string]any
	mockAPI.On("Request", http.MethodDelete, mock.Anything, nilMap, http.StatusNoContent).Return(
		nil, nil,
	)

	mockOgree3D.On("InformOptional", "DeleteObj", -1, map[string]any{
		"type": "delete-tag",
		"data": slug},
	).Return(nil)

	controllers.State.ObjsForUnity = controllers.SetObjsForUnity([]string{"all"})

	err := deleteController.DeleteObj(path)
	assert.Nil(t, err)
}
