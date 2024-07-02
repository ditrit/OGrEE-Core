package controllers_test

import (
	test_utils "cli/test"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test UI (UIDelay, UIToggle, UIHighlight)
func TestUIDelay(t *testing.T) {
	controller, _, ogree3D := layersSetup(t)
	// ogree3D.
	time := 15.0
	data := map[string]interface{}{
		"type": "ui",
		"data": map[string]interface{}{
			"command": "delay",
			"data":    time,
		},
	}
	ogree3D.On("Inform", "HandleUI", -1, data).Return(nil).Once() // The inform should be called once
	err := controller.UIDelay(time)
	assert.Nil(t, err)
}

func TestUIToggle(t *testing.T) {
	controller, _, ogree3D := layersSetup(t)
	// ogree3D.
	feature := "feature"
	enable := true
	data := map[string]interface{}{
		"type": "ui",
		"data": map[string]interface{}{
			"command": feature,
			"data":    enable,
		},
	}

	ogree3D.On("Inform", "HandleUI", -1, data).Return(nil).Once() // The inform should be called once
	err := controller.UIToggle(feature, enable)
	assert.Nil(t, err)
}

func TestUIHighlightObjectNotFound(t *testing.T) {
	controller, mockAPI, ogree3D := layersSetup(t)
	path := testRackObjPath

	test_utils.MockObjectNotFound(mockAPI, path)

	data := map[string]interface{}{
		"type": "ui",
		"data": map[string]interface{}{
			"command": "highlight",
			"data":    "BASIC.A.R1.A01",
		},
	}

	ogree3D.AssertNotCalled(t, "HandleUI", -1, data)
	err := controller.UIHighlight("/Physical/BASIC/A/R1/A01")
	assert.NotNil(t, err)
	assert.Equal(t, "object not found", err.Error())
}

func TestUIHighlightWorks(t *testing.T) {
	controller, mockAPI, ogree3D := layersSetup(t)
	data := map[string]interface{}{
		"type": "ui",
		"data": map[string]interface{}{
			"command": "highlight",
			"data":    rack1["id"],
		},
	}

	test_utils.MockGetObject(mockAPI, rack1)
	ogree3D.On("Inform", "HandleUI", -1, data).Return(nil).Once() // The inform should be called once
	err := controller.UIHighlight("/Physical/BASIC/A/R1/A01")
	assert.Nil(t, err)
}

func TestUIClearCache(t *testing.T) {
	controller, _, ogree3D := layersSetup(t)
	data := map[string]interface{}{
		"type": "ui",
		"data": map[string]interface{}{
			"command": "clearcache",
			"data":    "",
		},
	}

	ogree3D.On("Inform", "HandleUI", -1, data).Return(nil).Once() // The inform should be called once
	err := controller.UIClearCache()
	assert.Nil(t, err)
}

func TestCameraMove(t *testing.T) {
	controller, _, ogree3D := layersSetup(t)
	data := map[string]interface{}{
		"type": "camera",
		"data": map[string]interface{}{
			"command":  "move",
			"position": map[string]interface{}{"x": 0.0, "y": 1.0, "z": 2.0},
			"rotation": map[string]interface{}{"x": 0.0, "y": 0.0},
		},
	}

	ogree3D.On("Inform", "HandleUI", -1, data).Return(nil).Once() // The inform should be called once
	err := controller.CameraMove("move", []float64{0, 1, 2}, []float64{0, 0})
	assert.Nil(t, err)
}

func TestCameraWait(t *testing.T) {
	controller, _, ogree3D := layersSetup(t)
	time := 15.0
	data := map[string]interface{}{
		"type": "camera",
		"data": map[string]interface{}{
			"command":  "wait",
			"position": map[string]interface{}{"x": 0, "y": 0, "z": 0},
			"rotation": map[string]interface{}{"x": 999, "y": time},
		},
	}

	ogree3D.On("Inform", "HandleUI", -1, data).Return(nil).Once() // The inform should be called once
	err := controller.CameraWait(time)
	assert.Nil(t, err)
}

func TestFocusUIObjectNotFound(t *testing.T) {
	controller, mockAPI, ogree3D := layersSetup(t)

	test_utils.MockObjectNotFound(mockAPI, "/api/hierarchy_objects/"+rack1["id"].(string))
	err := controller.FocusUI("/Physical/" + strings.Replace(rack1["id"].(string), ".", "/", -1))
	ogree3D.AssertNotCalled(t, "Inform", "mock.Anything", "mock.Anything", "mock.Anything")
	assert.NotNil(t, err)
	assert.Equal(t, "object not found", err.Error())
}

func TestFocusUIEmptyPath(t *testing.T) {
	controller, mockAPI, ogree3D := layersSetup(t)
	data := map[string]interface{}{
		"type": "focus",
		"data": "",
	}

	ogree3D.On("Inform", "FocusUI", -1, data).Return(nil).Once() // The inform should be called once
	err := controller.FocusUI("")
	mockAPI.AssertNotCalled(t, "Request", "GET", "mock.Anything", "mock.Anything", "mock.Anything")
	assert.Nil(t, err)
}

func TestFocusUIErrorWithRoom(t *testing.T) {
	controller, mockAPI, ogree3D := layersSetup(t)
	errorMessage := "You cannot focus on this object. Note you cannot focus on Sites, Buildings and Rooms. "
	errorMessage += "For more information please refer to the help doc  (man >)"

	test_utils.MockGetObject(mockAPI, roomWithoutChildren)
	err := controller.FocusUI("/Physical/" + strings.Replace(roomWithoutChildren["id"].(string), ".", "/", -1))
	ogree3D.AssertNotCalled(t, "Inform", "mock.Anything", "mock.Anything", "mock.Anything")
	assert.NotNil(t, err)
	assert.Equal(t, errorMessage, err.Error())
}

func TestFocusUIWorks(t *testing.T) {
	controller, mockAPI, ogree3D := layersSetup(t)
	data := map[string]interface{}{
		"type": "focus",
		"data": rack1["id"],
	}

	ogree3D.On("Inform", "FocusUI", -1, data).Return(nil).Once() // The inform should be called once
	// Get Object will be called two times: Once in FocusUI and a second time in FocusUI->CD->Tree
	test_utils.MockGetObject(mockAPI, rack1)
	test_utils.MockGetObject(mockAPI, rack1)
	err := controller.FocusUI("/Physical/" + strings.Replace(rack1["id"].(string), ".", "/", -1))
	assert.Nil(t, err)
}
