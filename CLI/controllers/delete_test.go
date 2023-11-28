package controllers_test

import (
	"cli/controllers"
	"cli/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleteTagSendsDeleteTo3DWithSlug(t *testing.T) {
	controller, mockAPI, mockOgree3D, _ := newControllerWithMocks(t)

	slug := "slug"
	path := models.TagsPath + slug

	mockDeleteObjects(mockAPI, "namespace=logical.tag&slug=slug", []any{
		map[string]any{
			"slug":        slug,
			"description": "description",
			"color":       "aaaaaa",
		},
	})

	mockOgree3D.On("InformOptional", "DeleteObj", -1, map[string]any{
		"type": "delete-tag",
		"data": slug},
	).Return(nil)

	controllers.State.ObjsForUnity = controllers.SetObjsForUnity([]string{"all"})

	_, err := controller.DeleteObj(path)
	assert.Nil(t, err)
}
