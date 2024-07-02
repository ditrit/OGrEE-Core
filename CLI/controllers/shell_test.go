package controllers_test

import (
	"cli/controllers"
	test_utils "cli/test"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test PWD
func TestPWD(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)
	controller.CD("/")
	location := controllers.PWD()
	assert.Equal(t, "/", location)

	test_utils.MockGetObject(mockAPI, rack1)
	path := "/Physical/" + strings.Replace(rack1["id"].(string), ".", "/", -1)
	err := controller.CD(path)
	assert.Nil(t, err)

	location = controllers.PWD()
	assert.Equal(t, path, location)
}
