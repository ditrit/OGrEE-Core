package controllers_test

import (
	cmd "cli/controllers"
	mocks "cli/mocks/controllers"
	"cli/models"
	test_utils "cli/test"
	"testing"

	"github.com/stretchr/testify/assert"
)

func interactTestSetup(t *testing.T) (cmd.Controller, *mocks.APIPort, *mocks.Ogree3DPort) {
	controller, mockAPI, mockOgree3D := layersSetup(t)

	test_utils.MockGetObject(mockAPI, createRoom)

	test_utils.MockCreateObject(mockAPI, "generic", map[string]any{
		"name":        "A01",
		"category":    "generic",
		"description": "",
		"domain":      createRoom["domain"],
		"parentId":    createRoom["id"],
		"attributes": map[string]any{
			"height":     float64(1),
			"heightUnit": "cm",
			"rotation":   []float64{0, 0, 0},
			"posXYZ":     []float64{1, 1, 1},
			"posXYUnit":  "m",
			"size":       []float64{1, 1},
			"sizeUnit":   "cm",
			"shape":      "cube",
			"type":       "box",
		},
	})

	err := controller.CreateObject("/Physical/BASIC/A/R1/A01", models.GENERIC, map[string]any{
		"attributes": map[string]any{
			"rotation":  []float64{0, 0, 0},
			"size":      []float64{1, 1, 1},
			"posXYZ":    []float64{1, 1, 1},
			"posXYUnit": "m",
			"shape":     "cube",
			"type":      "box",
		},
	})
	assert.Nil(t, err)

	return controller, mockAPI, mockOgree3D
}

func interactLabelTestSetup(t *testing.T) (cmd.Controller, *mocks.APIPort, *mocks.Ogree3DPort) {
	controller, mockAPI, mockOgree3D := interactTestSetup(t)

	test_utils.MockGetObject(mockAPI, map[string]any{
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

func TestInteractObject(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		keyword  string
		value    interface{}
		fromAttr bool
	}{
		{"LabelStringOk", "/Physical/BASIC/A/R1/A01", "label", "string", false},
		{"LabelSingleAttrOk", "/Physical/BASIC/A/R1/A01", "label", "name", true},
		{"LabelStringWithOneAttrOk", "/Physical/BASIC/A/R1/A01", "label", "My name is #name", false},
		{"LabelStringWithMultipleAttrOk", "/Physical/BASIC/A/R1/A01", "label", "My name is #name and I am a #category", false},
		{"LabelSingleAttrAndStringOk", "/Physical/BASIC/A/R1/A01", "label", "name is my name", true},
		{"LabelSingleAttrAndStringWithAttrOk", "/Physical/BASIC/A/R1/A01", "label", "name\n#id", true},
		{"FontItalicOk", "/Physical/BASIC/A/R1/A01", "labelFont", "italic", false},
		{"LabelFontBoldOk", "/Physical/BASIC/A/R1/A01", "labelFont", "bold", false},
		{"LabelColorOk", "/Physical/BASIC/A/R1/A01", "labelFont", "color@C0FFEE", false},
		{"LabelBackgroundOk", "/Physical/BASIC/A/R1/A01", "labelBackground", "C0FFEE", false},
		{"ContentOk", "/Physical/BASIC/A/R1/A01", "displayContent", true, false},
		{"AlphaOk", "/Physical/BASIC/A/R1/A01", "alpha", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller, _, _ := interactLabelTestSetup(t)
			err := controller.InteractObject(tt.path, tt.keyword, tt.value, tt.fromAttr)
			assert.Nil(t, err)
		})
	}
}

func TestInteractObjectWithMock(t *testing.T) {
	tests := []struct {
		name       string
		mockObject map[string]any
		path       string
		keyword    string
		value      interface{}
		fromAttr   bool
	}{
		{"TilesNameOk", createRoom, "/Physical/BASIC/A/R1", "tilesName", true, false},
		{"TilesColorOk", createRoom, "/Physical/BASIC/A/R1", "tilesColor", true, false},
		{"UOk", createRoom, "/Physical/BASIC/A/R1", "U", true, false},
		{"SlotsOk", createRoom, "/Physical/BASIC/A/R1", "slots", true, false},
		{"LocalCSOk", createRoom, "/Physical/BASIC/A/R1", "localCS", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller, mockAPI, _ := interactTestSetup(t)

			test_utils.MockGetObject(mockAPI, tt.mockObject)
			err := controller.InteractObject(tt.path, tt.keyword, tt.value, tt.fromAttr)
			assert.Nil(t, err)
		})
	}
}

func TestSetLabel(t *testing.T) {
	controller, mockAPI, _, _ := test_utils.NewControllerWithMocks(t)

	room := test_utils.GetEntity("rack", "rack", "site.building.room", "domain")
	test_utils.MockGetObject(mockAPI, room)
	err := controller.SetLabel("/Physical/site/building/room/rack", []any{"myLabel"}, false)

	assert.Nil(t, err)
}
