package controllers_test

import (
	cmd "cli/controllers"
	mocks "cli/mocks/controllers"
	"cli/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func interactTestSetup(t *testing.T) (cmd.Controller, *mocks.APIPort, *mocks.Ogree3DPort) {
	controller, mockAPI, mockOgree3D := layersSetup(t)

	mockGetObject(mockAPI, createRoom)

	mockCreateObject(mockAPI, "generic", map[string]any{
		"name":        "A01",
		"category":    "generic",
		"description": []any{},
		"domain":      createRoom["domain"],
		"parentId":    createRoom["id"],
		"attributes": map[string]any{
			"height":       "1",
			"heightUnit":   "cm",
			"rotation":     `{"x":0, "y":0, "z":0}`,
			"posXYZ":       `{"x":1 ,"y":1 ,"z":1 }`,
			"posXYUnit":    "m",
			"size":         `{"x":1 ,"y":1 }`,
			"sizeUnit":     "cm",
			"shape":        "cube",
			"type":         "box",
			"template":     "",
			"diameterUnit": "cm",
		},
	})

	err := controller.CreateObject("/Physical/BASIC/A/R1/A01", models.GENERIC, map[string]any{
		"attributes": map[string]any{
			"rotation":     []float64{0, 0, 0},
			"size":         []float64{1, 1, 1},
			"posXYZ":       []float64{1, 1, 1},
			"posXYUnit":    "m",
			"shape":        "cube",
			"type":         "box",
			"diameterUnit": "cm",
		},
	})
	assert.Nil(t, err)

	return controller, mockAPI, mockOgree3D
}

func interactLabelTestSetup(t *testing.T) (cmd.Controller, *mocks.APIPort, *mocks.Ogree3DPort) {
	controller, mockAPI, mockOgree3D := interactTestSetup(t)

	mockGetObject(mockAPI, map[string]any{
		"category": "rack",
		"children": []any{chassis, rackGroup, pdu},
		"id":       "BASIC.A.R1.A01",
		"name":     "A01",
		"parentId": "BASIC.A.R1",
		"attributes": map[string]any{
			"category": "rack",
			"children": []any{chassis, rackGroup, pdu},
			"id":       "BASIC.A.R1.A01",
			"name":     "A01",
			"parentId": "BASIC.A.R1",
		},
	})

	return controller, mockAPI, mockOgree3D
}

func TestLabelNotStringReturnsError(t *testing.T) {
	err := cmd.C.InteractObject("/Physical/BASIC/A/R1", "label", 1, false)
	assert.NotNil(t, err)
	assert.Errorf(t, err, "The label value must be a string")
}

func TestNonExistingAttrReturnsError(t *testing.T) {
	controller, _, _ := interactLabelTestSetup(t)
	err := controller.InteractObject("/Physical/BASIC/A/R1/A01", "label", "abc", true)
	assert.NotNil(t, err)
	assert.Errorf(t, err, "The specified attribute 'abc' does not exist in the object. \nPlease view the object (ie. $> get) and try again")
}

func TestLabelStringOk(t *testing.T) {
	controller, _, _ := interactLabelTestSetup(t)
	err := controller.InteractObject("/Physical/BASIC/A/R1/A01", "label", "string", false)
	assert.Nil(t, err)
}

func TestLabelSingleAttrOk(t *testing.T) {
	controller, _, _ := interactLabelTestSetup(t)
	err := controller.InteractObject("/Physical/BASIC/A/R1/A01", "label", "name", true)
	assert.Nil(t, err)
}

func TestLabelStringWithOneAttrOk(t *testing.T) {
	controller, _, _ := interactLabelTestSetup(t)
	err := controller.InteractObject("/Physical/BASIC/A/R1/A01", "label", "My name is #name", false)
	assert.Nil(t, err)
}

func TestLabelStringWithMultipleAttrOk(t *testing.T) {
	controller, _, _ := interactLabelTestSetup(t)
	err := controller.InteractObject("/Physical/BASIC/A/R1/A01", "label", "My name is #name and I am a #category", false)
	assert.Nil(t, err)
}

func TestLabelSingleAttrAndStringOk(t *testing.T) {
	controller, _, _ := interactLabelTestSetup(t)
	err := controller.InteractObject("/Physical/BASIC/A/R1/A01", "label", "name is my name", true)
	assert.Nil(t, err)
}

func TestLabelSingleAttrAndStringWithAttrOk(t *testing.T) {
	controller, _, _ := interactLabelTestSetup(t)
	err := controller.InteractObject("/Physical/BASIC/A/R1/A01", "label", "name\n#id", true)
	assert.Nil(t, err)
}

func TestLabelFontItalicOk(t *testing.T) {
	controller, _, _ := interactLabelTestSetup(t)

	err := controller.InteractObject("/Physical/BASIC/A/R1/A01", "labelFont", "italic", false)
	assert.Nil(t, err)
}

func TestLabelFontBoldOk(t *testing.T) {
	controller, _, _ := interactLabelTestSetup(t)

	err := controller.InteractObject("/Physical/BASIC/A/R1/A01", "labelFont", "bold", false)
	assert.Nil(t, err)
}

func TestLabelColorOk(t *testing.T) {
	controller, _, _ := interactLabelTestSetup(t)

	err := controller.InteractObject("/Physical/BASIC/A/R1/A01", "labelFont", "color@C0FFEE", false)
	assert.Nil(t, err)
}

func TestLabelBackgroundOk(t *testing.T) {
	controller, _, _ := interactLabelTestSetup(t)

	err := controller.InteractObject("/Physical/BASIC/A/R1/A01", "labelBackground", "C0FFEE", false)
	assert.Nil(t, err)
}

func TestContentOk(t *testing.T) {
	controller, _, _ := interactLabelTestSetup(t)
	err := controller.InteractObject("/Physical/BASIC/A/R1/A01", "content", true, false)
	assert.Nil(t, err)
}

func TestAlphaOk(t *testing.T) {
	controller, _, _ := interactLabelTestSetup(t)
	err := controller.InteractObject("/Physical/BASIC/A/R1/A01", "alpha", true, false)
	assert.Nil(t, err)
}

func TestTilesNameOk(t *testing.T) {
	controller, mockAPI, _ := interactTestSetup(t)

	mockGetObject(mockAPI, createRoom)
	err := controller.InteractObject("/Physical/BASIC/A/R1", "tilesName", true, false)
	assert.Nil(t, err)
}

func TestTilesColorOk(t *testing.T) {
	controller, mockAPI, _ := interactTestSetup(t)

	mockGetObject(mockAPI, createRoom)
	err := controller.InteractObject("/Physical/BASIC/A/R1", "tilesColor", true, false)
	assert.Nil(t, err)
}

func TestUOk(t *testing.T) {
	controller, mockAPI, _ := interactTestSetup(t)

	mockGetObject(mockAPI, createRoom)
	err := controller.InteractObject("/Physical/BASIC/A/R1", "U", true, false)
	assert.Nil(t, err)
}

func TestSlotsOk(t *testing.T) {
	controller, mockAPI, _ := interactTestSetup(t)

	mockGetObject(mockAPI, createRoom)
	err := controller.InteractObject("/Physical/BASIC/A/R1", "slots", true, false)
	assert.Nil(t, err)
}

func TestLocalCSOk(t *testing.T) {
	controller, mockAPI, _ := interactTestSetup(t)

	mockGetObject(mockAPI, createRoom)
	err := controller.InteractObject("/Physical/BASIC/A/R1", "localCS", true, false)
	assert.Nil(t, err)
}
