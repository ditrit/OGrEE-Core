package parser

import (
	"cli/models"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodeToFloat(t *testing.T) {
	valNode := valueNode{"3.5"}
	value, err := nodeToFloat(&valNode, "")

	assert.Nil(t, err)
	assert.Equal(t, 3.5, value)

	valNode = valueNode{"q3.5"}
	_, err = nodeToFloat(&valNode, "invalidFloatNode")

	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "invalidFloatNode should be a number")
}

func TestNodeToNum(t *testing.T) {
	valNode := valueNode{"3.5"}
	value, err := nodeToNum(&valNode, "")

	assert.Nil(t, err)
	assert.Equal(t, 3.5, value)

	valNode = valueNode{"3"}
	value, err = nodeToNum(&valNode, "")

	assert.Nil(t, err)
	assert.Equal(t, 3, value)

	valNode = valueNode{"q3"}
	_, err = nodeToNum(&valNode, "invalidNumberNode")

	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "invalidNumberNode should be a number")
}

func TestNodeToInt(t *testing.T) {
	valNode := valueNode{"3"}
	value, err := nodeToInt(&valNode, "")

	assert.Nil(t, err)
	assert.Equal(t, 3, value)

	valNode = valueNode{"3.5"}
	_, err = nodeToInt(&valNode, "invalidIntNode")

	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "invalidIntNode should be an integer")
}

func TestNodeToBool(t *testing.T) {
	valNode := valueNode{"false"}
	value, err := nodeToBool(&valNode, "")

	assert.Nil(t, err)
	assert.False(t, value)

	valNode = valueNode{"3.5"}
	_, err = nodeToBool(&valNode, "invalidBoolNode")

	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "invalidBoolNode should be a boolean")
}

func TestNodeTo3dRotation(t *testing.T) {
	valNode := valueNode{"front"}
	value, err := nodeTo3dRotation(&valNode)

	assert.Nil(t, err)
	assert.Equal(t, []float64{0, 0, 180}, value)

	valNode = valueNode{"3.5"}
	_, err = nodeTo3dRotation(&valNode)

	assert.NotNil(t, err)
	assert.ErrorContains(t, err,
		`rotation should be a vector3, or one of the following keywords :
		front, rear, left, right, top, bottom`)
}

func TestNodeToString(t *testing.T) {
	valNode := valueNode{3}
	value, err := nodeToString(&valNode, "int")

	assert.Nil(t, err)
	assert.Equal(t, "3", value)
}

func TestNodeToVec(t *testing.T) {
	valNode := valueNode{[]float64{1, 2, 3}}
	value, err := nodeToVec(&valNode, -1, "vector")

	assert.Nil(t, err)
	assert.Equal(t, []float64{1, 2, 3}, value)
}

func TestNodeToColorString(t *testing.T) {
	valNode := valueNode{"abcacc"}
	value, err := nodeToColorString(&valNode)

	assert.Nil(t, err)
	assert.Equal(t, "abcacc", value)

	valNode = valueNode{"3.5"}
	_, err = nodeToColorString(&valNode)

	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "Please provide a valid 6 digit Hex value for the color")
}

func TestFileToJson(t *testing.T) {
	basePath := t.TempDir() // temporary directory that will be deleted after the tests have finished
	fileContent := "{\"value\": [3,4,5]}\n"

	filename := "file_to_json_test_file.json"
	filePath := basePath + "/" + filename
	err := os.WriteFile(filePath, []byte(fileContent), 0644)

	if err != nil {
		t.Errorf("an error ocurred while creating the test file: %s", err)
	}

	json := fileToJSON(filePath)
	assert.Len(t, json, 1)
	assert.Equal(t, []any{3.0, 4.0, 5.0}, json["value"])

	json = fileToJSON(basePath + "invalidPath.json")
	assert.Nil(t, json)
}

func TestEvalNodeArr(t *testing.T) {
	nodes := []node{&valueNode{"abcacc"}, &valueNode{"abcddd"}}

	values, err := evalNodeArr(&nodes, []string{})
	assert.Nil(t, err)
	assert.Len(t, values, 2)
	assert.Equal(t, []string{"abcacc", "abcddd"}, values)

	nodes = []node{&valueNode{"abcacc"}, &valueNode{3}}

	_, err = evalNodeArr(&nodes, []string{})
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "Error unexpected element")
}

func TestErrorResponder(t *testing.T) {
	err := errorResponder("reserved", "4", false)
	assert.ErrorContains(t, err, "Invalid reserved attribute provided. It must be an array/list/vector with 4 elements. Please refer to the wiki or manual reference for more details on how to create objects using this syntax")

	err = errorResponder("reserved", "4", true)
	assert.ErrorContains(t, err, "Invalid reserved attributes provided. They must be arrays/lists/vectors with 4 elements. Please refer to the wiki or manual reference for more details on how to create objects using this syntax")
}

func TestFiltersToMapString(t *testing.T) {
	filters := map[string]node{
		"tag": &valueNode{"my-tag"},
	}
	mapFilters, err := filtersToMapString(filters)
	assert.Nil(t, err)
	assert.Len(t, mapFilters, 1)
	assert.Equal(t, "my-tag", mapFilters["tag"])
}

func TestRecursiveArgsToParams(t *testing.T) {
	args := recursiveArgs{false, "1", "2"}
	path := models.PhysicalPath + "site/building"
	recParams, err := args.toParams(path)
	assert.Nil(t, err)
	assert.Nil(t, recParams)

	args = recursiveArgs{true, "1", "2"}
	recParams, err = args.toParams(path)
	assert.Nil(t, err)
	assert.Equal(t, 1, recParams.MinDepth)
	assert.Equal(t, 2, recParams.MaxDepth)
	assert.Equal(t, path, recParams.PathEntered)
}

func TestStringToIntOr(t *testing.T) {
	value, err := stringToIntOr("3", 5)
	assert.Nil(t, err)
	assert.Equal(t, 3, value)

	value, err = stringToIntOr("", 5)
	assert.Nil(t, err)
	assert.Equal(t, 5, value)

	_, err = stringToIntOr("s", 5)
	assert.NotNil(t, err)
}

func TestAddSizeOrTemplate(t *testing.T) {
	valNode := valueNode{[]float64{1, 2, 3}}
	attributes := map[string]any{}
	err := addSizeOrTemplate(&valNode, attributes, models.ROOM)
	assert.Nil(t, err)
	assert.Contains(t, attributes, "size")

	valNode = valueNode{3.5}
	err = addSizeOrTemplate(&valNode, attributes, models.ROOM)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "vector3 (size) or string (template) expected")

	valNode = valueNode{"my-template"}
	err = addSizeOrTemplate(&valNode, attributes, models.ROOM)
	assert.Nil(t, err)
	assert.Contains(t, attributes, "template")
}

func TestNodeToSize(t *testing.T) {
	valNode := valueNode{[]float64{1, 2, 3}}
	size, err := nodeToSize(&valNode)
	assert.Nil(t, err)
	assert.Equal(t, []float64{1, 2, 3}, size)
}

func TestNodeToPosXYZ(t *testing.T) {
	valNode := valueNode{[]float64{1, 2, 3, 4}}
	_, err := nodeToPosXYZ(&valNode)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "position should be a vector2 or a vector3")

	valNode = valueNode{[]float64{1, 2, 3}}
	position, err := nodeToPosXYZ(&valNode)
	assert.Nil(t, err)
	assert.Equal(t, []float64{1, 2, 3}, position)
}
