package controllers_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateTemplateOfTypeGenericWorks(t *testing.T) {
	controller, mockAPI, _, _ := newControllerWithMocks(t)

	template := map[string]any{
		"slug":        "generic-example",
		"description": "a table",
		"category":    "generic",
		"sizeWDHmm":   []float64{447, 914.5, 263.3},
		"fbxModel":    "",
		"attributes": map[string]any{
			"type": "table",
		},
		"colors": []any{},
	}

	mockCreateObject(mockAPI, "obj-template", template)

	err := controller.LoadTemplate(template)
	assert.Nil(t, err)
}
