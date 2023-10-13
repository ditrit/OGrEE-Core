package controllers_test

import (
	"cli/controllers"
	mocks "cli/mocks/controllers"
	"cli/models"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
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

	mockGetController.On("GetObjectsWildcard", path).Return(
		[]map[string]any{
			{
				"slug":  slug,
				"color": "aaaaaa",
			},
		},
		nil,
		nil,
	)

	var nilMap map[string]any
	mockAPI.On("Request", http.MethodDelete, "/api/objects?namespace=logical.tag&slug=slug", nilMap, http.StatusOK).Return(
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
