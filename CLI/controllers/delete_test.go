package controllers_test

import (
	"cli/controllers"
	"cli/models"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleteTagSendsDeleteTo3DWithSlug(t *testing.T) {
	controller, mockAPI, mockOgree3D := newControllerWithMocks(t)

	slug := "slug"
	path := models.TagsPath + slug

	mockGetObjects(mockAPI, "namespace=logical.tag&slug=slug", []any{
		map[string]any{
			"slug":        slug,
			"description": "description",
			"color":       "aaaaaa",
		},
	})

	var nilMap map[string]any
	mockAPI.On("Request", http.MethodDelete, "/api/objects?namespace=logical.tag&slug=slug", nilMap, http.StatusOK).Return(
		nil, nil,
	)

	mockOgree3D.On("InformOptional", "DeleteObj", -1, map[string]any{
		"type": "delete-tag",
		"data": slug},
	).Return(nil)

	controllers.State.ObjsForUnity = controllers.SetObjsForUnity([]string{"all"})

	err := controller.DeleteObj(path)
	assert.Nil(t, err)
}
