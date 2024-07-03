package controllers_test

import (
	"cli/controllers"
	"cli/models"
	test_utils "cli/test"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Tests IsEntityDrawable
func TestIsEntityDrawableObjectNotFound(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	test_utils.MockObjectNotFound(mockAPI, testRackObjPath)

	isDrawable, err := controller.IsEntityDrawable(models.PhysicalPath + "BASIC/A/R1/A01")
	assert.False(t, isDrawable)
	assert.NotNil(t, err)
	assert.Equal(t, "object not found", err.Error())
}

func TestIsEntityDrawable(t *testing.T) {
	tests := []struct {
		name               string
		drawableObjects    []int
		expectedIsDrawable bool
	}{
		{"CategoryIsNotDrawable", []int{models.EntityStrToInt("device")}, false},
		{"CategoryIsDrawable", []int{models.EntityStrToInt("rack")}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller, mockAPI, _ := layersSetup(t)
			controllers.State.DrawableObjs = tt.drawableObjects

			test_utils.MockGetObject(mockAPI, rack1)

			isDrawable, err := controller.IsEntityDrawable(models.PhysicalPath + "BASIC/A/R1/A01")
			assert.Equal(t, tt.expectedIsDrawable, isDrawable)
			assert.Nil(t, err)
		})
	}
}

// Tests IsAttrDrawable (and IsCategoryAttrDrawable)
func TestIsAttrDrawableObjectNotFound(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)
	path := testRackObjPath

	test_utils.MockObjectNotFound(mockAPI, path)

	isAttrDrawable, err := controller.IsAttrDrawable(models.PhysicalPath+"BASIC/A/R1/A01", "color")
	assert.False(t, isAttrDrawable)
	assert.NotNil(t, err)
	assert.Equal(t, "object not found", err.Error())
}

func TestIsAttrDrawableTemplateJsonIsNil(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)
	controllers.State.DrawableObjs = []int{models.EntityStrToInt("rack")}

	controllers.State.DrawableJsons = map[string]map[string]any{
		"rack": nil,
	}

	test_utils.MockGetObject(mockAPI, rack1)

	isAttrDrawable, err := controller.IsAttrDrawable(models.PhysicalPath+"BASIC/A/R1/A01", "color")
	assert.True(t, isAttrDrawable)
	assert.Nil(t, err)
}

func TestIsAttrDrawable(t *testing.T) {
	tests := []struct {
		name                 string
		attributeDrawable    string
		attributeNonDrawable string
	}{
		{"SpecialAttribute", "name", "description"},
		{"SpecialAttribute", "color", "height"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller, mockAPI, _ := layersSetup(t)
			controllers.State.DrawableObjs = []int{models.EntityStrToInt("rack")}

			controllers.State.DrawableJsons = test_utils.GetTestDrawableJson()

			test_utils.MockGetObject(mockAPI, rack1)
			isAttrDrawable, err := controller.IsAttrDrawable(models.PhysicalPath+"BASIC/A/R1/A01", tt.attributeDrawable)
			assert.True(t, isAttrDrawable)
			assert.Nil(t, err)

			test_utils.MockGetObject(mockAPI, rack1)
			isAttrDrawable, err = controller.IsAttrDrawable(models.PhysicalPath+"BASIC/A/R1/A01", tt.attributeNonDrawable)
			assert.False(t, isAttrDrawable)
			assert.Nil(t, err)
		})
	}
}
