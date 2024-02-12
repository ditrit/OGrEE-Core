package controllers_test

import (
	"cli/controllers"
	"cli/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleteTag(t *testing.T) {
	controller, mockAPI, _, _ := newControllerWithMocks(t)

	slug := "slug"
	path := models.TagsPath + slug

	mockDeleteObjects(mockAPI, "namespace=logical.tag&slug=slug", []any{
		map[string]any{
			"slug":        slug,
			"description": "description",
			"color":       "aaaaaa",
		},
	})

	controllers.State.ObjsForUnity = controllers.SetObjsForUnity([]string{"all"})

	_, err := controller.DeleteObj(path)
	assert.Nil(t, err)
}
