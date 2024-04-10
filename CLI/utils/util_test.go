package utils_test

import (
	"cli/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFloat(t *testing.T) {
	number, err := utils.GetFloat(5)
	assert.Nil(t, err)
	assert.Equal(t, 5.0, number)

	number, err = utils.GetFloat("5")
	assert.NotNil(t, err)
	assert.Equal(t, 0.0, number)
	assert.ErrorContains(t, err, "cannot convert string to float64")
}

func TestValToFloat(t *testing.T) {
	number, err := utils.ValToFloat(5, "number")
	assert.Nil(t, err)
	assert.Equal(t, 5.0, number)

	number, err = utils.ValToFloat("5.5", "string number")
	assert.Nil(t, err)
	assert.Equal(t, 5.5, number)

	number, err = utils.ValToFloat("fifty", "string value")
	assert.NotNil(t, err)
	assert.Equal(t, 0.0, number)
	assert.ErrorContains(t, err, "string value should be a number")

	number, err = utils.ValToFloat([]int{5}, "list")
	assert.NotNil(t, err)
	assert.Equal(t, 0.0, number)
	assert.ErrorContains(t, err, "list should be a number")
}

func TestStringToNum(t *testing.T) {
	number, err := utils.StringToNum("5")
	assert.Nil(t, err)
	assert.Equal(t, 5, number) // returns an int

	number, err = utils.StringToNum("5.5")
	assert.Nil(t, err)
	assert.Equal(t, 5.5, number) // returns a float

	number, err = utils.StringToNum("fifty")
	assert.NotNil(t, err)
	assert.Nil(t, number)
	assert.ErrorContains(t, err, "the string is not a number")
}

func TestValToNum(t *testing.T) {
	number, err := utils.ValToNum("5", "string int")
	assert.Nil(t, err)
	assert.Equal(t, 5, number) // returns an int

	number, err = utils.ValToNum("5.5", "string float")
	assert.Nil(t, err)
	assert.Equal(t, 5.5, number) // returns a float

	number, err = utils.ValToNum("fifty", "string value")
	assert.NotNil(t, err)
	assert.Nil(t, number)
	assert.ErrorContains(t, err, "string value should be a number")

	number, err = utils.ValToNum(5, "int")
	assert.Nil(t, err)
	assert.Equal(t, 5, number)
}

func TestValToInt(t *testing.T) {
	number, err := utils.ValToInt("5", "string int")
	assert.Nil(t, err)
	assert.Equal(t, 5, number) // returns an int

	number, err = utils.ValToInt("5.5", "string float")
	assert.NotNil(t, err)
	assert.Equal(t, 0, number)
	assert.ErrorContains(t, err, "string float should be an integer")

	number, err = utils.ValToInt("fifty", "string value")
	assert.NotNil(t, err)
	assert.Equal(t, 0, number)
	assert.ErrorContains(t, err, "string value should be an integer")

	number, err = utils.ValToInt([]int{5}, "list")
	assert.NotNil(t, err)
	assert.Equal(t, 0, number)
	assert.ErrorContains(t, err, "list should be an integer")

	number, err = utils.ValToInt(5, "int")
	assert.Nil(t, err)
	assert.Equal(t, 5, number)
}

func TestValToBool(t *testing.T) {
	result, err := utils.ValToBool(true, "true boolean")
	assert.Nil(t, err)
	assert.True(t, result)

	result, err = utils.ValToBool(false, "false boolean")
	assert.Nil(t, err)
	assert.False(t, result)

	result, err = utils.ValToBool("true", "true string")
	assert.Nil(t, err)
	assert.True(t, result)

	result, err = utils.ValToBool("false", "false string")
	assert.Nil(t, err)
	assert.False(t, result)

	result, err = utils.ValToBool("fals", "error string")
	assert.NotNil(t, err)
	assert.False(t, result)
	assert.ErrorContains(t, err, "error string should be a boolean")

	result, err = utils.ValToBool(1, "int")
	assert.NotNil(t, err)
	assert.False(t, result)
	assert.ErrorContains(t, err, "int should be a boolean")

}

func TestValTo3dRotation(t *testing.T) {
	result, err := utils.ValTo3dRotation([]float64{1, 1})
	assert.Nil(t, err)
	assert.Equal(t, []float64{1, 1}, result)

	stringValues := map[string][]float64{
		"front":  []float64{0, 0, 180},
		"rear":   []float64{0, 0, 0},
		"left":   []float64{0, 90, 0},
		"right":  []float64{0, -90, 0},
		"top":    []float64{90, 0, 0},
		"bottom": []float64{-90, 0, 0},
	}

	for key, value := range stringValues {
		result, err = utils.ValTo3dRotation(key)
		assert.Nil(t, err)
		assert.Equal(t, value, result)
	}

	result, err = utils.ValTo3dRotation(false)
	assert.NotNil(t, err)
	assert.Nil(t, result)
	assert.ErrorContains(t, err,
		`rotation should be a vector3, or one of the following keywords :
		front, rear, left, right, top, bottom`)
}

func TestValToString(t *testing.T) {
	result, err := utils.ValToString(5, "int")
	assert.Nil(t, err)
	assert.Equal(t, "5", result)

	result, err = utils.ValToString("value", "string")
	assert.Nil(t, err)
	assert.Equal(t, "value", result)

	result, err = utils.ValToString(5.5, "float")
	assert.NotNil(t, err)
	assert.Equal(t, "", result)
	assert.ErrorContains(t, err, "float should be a string")
}

func TestValToVec(t *testing.T) {
	value := []float64{0, 1, 2}
	result, err := utils.ValToVec(value, len(value), "float vector")
	assert.Nil(t, err)
	assert.Equal(t, value, result)

	result, err = utils.ValToVec(value, len(value)-1, "float vector invalid size")
	assert.NotNil(t, err)
	assert.Nil(t, result)
	assert.ErrorContains(t, err, "float vector invalid size should be a vector2")

	result, err = utils.ValToVec("[0,0]", 2, "string")
	assert.NotNil(t, err)
	assert.Nil(t, result)
	assert.ErrorContains(t, err, "string should be a vector2")
}

func TestValToColor(t *testing.T) {
	color, ok := utils.ValToColor([]int{1})
	assert.False(t, ok)
	assert.Equal(t, "", color)

	// hex string of length != 6
	color, ok = utils.ValToColor("abcac")
	assert.False(t, ok)
	assert.Equal(t, "", color)

	// not hex string of length == 6
	color, ok = utils.ValToColor("zabaca")
	assert.False(t, ok)
	assert.Equal(t, "", color)

	// hex string of length == 6
	color, ok = utils.ValToColor("abcaca")
	assert.True(t, ok)
	assert.Equal(t, "abcaca", color)

	// int with 6 digits
	color, ok = utils.ValToColor(255255)
	assert.True(t, ok)
	assert.Equal(t, "255255", color)

	// int without 6 digits
	color, ok = utils.ValToColor(255)
	assert.False(t, ok)
	assert.Equal(t, "", color)

	// float without 6 digits
	color, ok = utils.ValToColor(255.0)
	assert.False(t, ok)
	assert.Equal(t, "", color)

	// float with 6 digits
	color, ok = utils.ValToColor(255255.0)
	assert.True(t, ok)
	assert.Equal(t, "255255", color)
}

func TestIsInfArr(t *testing.T) {
	ok := utils.IsInfArr([]any{1})
	assert.True(t, ok)

	ok = utils.IsInfArr([]any{1.0})
	assert.True(t, ok)

	ok = utils.IsInfArr("string")
	assert.False(t, ok)
}

func TestIsString(t *testing.T) {
	ok := utils.IsString(1)
	assert.False(t, ok)

	ok = utils.IsString(1.0)
	assert.False(t, ok)

	ok = utils.IsString("string")
	assert.True(t, ok)
}

func TestIsHexString(t *testing.T) {
	ok := utils.IsHexString("1.0")
	assert.False(t, ok)

	ok = utils.IsHexString("string")
	assert.False(t, ok)

	ok = utils.IsHexString("abc4")
	assert.True(t, ok)
}

func TestIsInt(t *testing.T) {
	ok := utils.IsInt(1.0)
	assert.False(t, ok)

	ok = utils.IsInt("string")
	assert.False(t, ok)

	ok = utils.IsInt(1)
	assert.True(t, ok)
}

func TestIsFloat(t *testing.T) {
	ok := utils.IsFloat(1.0)
	assert.True(t, ok)

	ok = utils.IsFloat("string")
	assert.False(t, ok)

	ok = utils.IsFloat(1)
	assert.False(t, ok)
}

func TestCompareVals(t *testing.T) {
	comparison, ok := utils.CompareVals(1.0, 2.0)
	assert.True(t, ok)
	assert.True(t, comparison)

	comparison, ok = utils.CompareVals(2.0, 1.0)
	assert.True(t, ok)
	assert.False(t, comparison)

	comparison, ok = utils.CompareVals("value1", "value2")
	assert.True(t, ok)
	assert.True(t, comparison)

	comparison, ok = utils.CompareVals("value2", "value1")
	assert.True(t, ok)
	assert.False(t, comparison)

	comparison, ok = utils.CompareVals(1.0, "abc")
	assert.False(t, ok)
	assert.False(t, comparison)
}

func TestNameOrSlug(t *testing.T) {
	result := utils.NameOrSlug(map[string]any{"slug": "my-slug"})
	assert.Equal(t, "my-slug", result)

	result = utils.NameOrSlug(map[string]any{"name": "my-name"})
	assert.Equal(t, "my-name", result)

	result = utils.NameOrSlug(map[string]any{"slug": "my-slug", "name": "my-name"})
	assert.Equal(t, "my-slug", result)
}

func TestObjectAttr(t *testing.T) {
	object := map[string]any{
		"name": "my-name",
	}
	value, ok := utils.ObjectAttr(object, "name")
	assert.True(t, ok)
	assert.Equal(t, object["name"], value)

	value, ok = utils.ObjectAttr(object, "color")
	assert.False(t, ok)
	assert.Nil(t, value)

	object["attributes"] = map[string]any{
		"color": "blue",
	}

	value, ok = utils.ObjectAttr(object, "color")
	assert.True(t, ok)
	assert.Equal(t, object["attributes"].(map[string]any)["color"], value)

	value, ok = utils.ObjectAttr(object, "other")
	assert.False(t, ok)
	assert.Nil(t, value)
}
