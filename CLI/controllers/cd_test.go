package controllers_test

import (
	"cli/controllers"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCdToALayer(t *testing.T) {
	controller, _, _ := layersSetup(t)

	path := "/Physical/" + strings.Replace(roomWithoutChildren["id"].(string), ".", "/", -1) + "/#my-layer"
	oldCurrentPath := controllers.State.CurrPath

	err := controller.CD(path)
	assert.NotNil(t, err)
	assert.Equal(t, "it is not possible to cd into a layer", err.Error())
	assert.Equal(t, oldCurrentPath, controllers.State.PrevPath)
	assert.Equal(t, oldCurrentPath, controllers.State.CurrPath)
}

func TestCdObjectNotFound(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	path := "/Physical/" + strings.Replace(rack1["id"].(string), ".", "/", -1)
	oldCurrentPath := controllers.State.CurrPath
	mockObjectNotFound(mockAPI, "/api/hierarchy-objects/"+rack1["id"].(string))

	err := controller.CD(path)
	assert.NotNil(t, err)
	assert.Equal(t, "object not found", err.Error())
	assert.Equal(t, oldCurrentPath, controllers.State.PrevPath)
	assert.Equal(t, oldCurrentPath, controllers.State.CurrPath)
}

func TestCdWorks(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObject(mockAPI, rack1)
	path := "/Physical/" + strings.Replace(rack1["id"].(string), ".", "/", -1)
	oldCurrentPath := controllers.State.CurrPath

	err := controller.CD(path)
	assert.Nil(t, err)
	assert.Equal(t, oldCurrentPath, controllers.State.PrevPath)
	assert.Equal(t, path, controllers.State.CurrPath)
}
