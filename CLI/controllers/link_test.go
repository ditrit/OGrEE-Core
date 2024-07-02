package controllers_test

import (
	"cli/models"
	test_utils "cli/test"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Tests LinkObject
func TestLinkObjectErrorNotStaryObject(t *testing.T) {
	controller, _, _ := layersSetup(t)

	err := controller.LinkObject(models.PhysicalPath+"BASIC/A/R1/A01", models.PhysicalPath+"BASIC/A/R1/A01", []string{}, []any{}, []string{})
	assert.NotNil(t, err)
	assert.Equal(t, "only stray objects can be linked", err.Error())
}

func TestLinkObjectWithoutSlots(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	strayDevice := test_utils.CopyMap(chassis)
	delete(strayDevice, "id")
	delete(strayDevice, "parentId")
	response := map[string]any{"message": "successfully linked"}
	body := map[string]any{"parentId": "BASIC.A.R1.A01", "slot": []string{}}

	test_utils.MockUpdateObject(mockAPI, body, response)

	slots := []string{}
	attributes := []string{}
	values := []any{}
	for key, value := range strayDevice["attributes"].(map[string]any) {
		attributes = append(attributes, key)
		values = append(values, value)
	}
	err := controller.LinkObject(models.StrayPath+"chT", models.PhysicalPath+"BASIC/A/R1/A01", attributes, values, slots)
	assert.Nil(t, err)
}

func TestLinkObjectWithInvalidSlots(t *testing.T) {
	controller, _, _ := layersSetup(t)

	strayDevice := test_utils.CopyMap(chassis)
	delete(strayDevice, "id")
	delete(strayDevice, "parentId")

	slots := []string{"slot01..slot03", "slot4"}
	attributes := []string{}
	values := []any{}
	for key, value := range strayDevice["attributes"].(map[string]any) {
		attributes = append(attributes, key)
		values = append(values, value)
	}
	err := controller.LinkObject(models.StrayPath+"chT", models.PhysicalPath+"BASIC/A/R1/A01", attributes, values, slots)
	assert.NotNil(t, err)
	assert.Equal(t, "Invalid device syntax: .. can only be used in a single element vector", err.Error())
}

func TestLinkObjectWithValidSlots(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	strayDevice := test_utils.CopyMap(chassis)
	delete(strayDevice, "id")
	delete(strayDevice, "parentId")
	response := map[string]any{"message": "successfully linked"}
	body := map[string]any{"parentId": "BASIC.A.R1.A01", "slot": []string{"slot01"}}

	test_utils.MockUpdateObject(mockAPI, body, response)

	slots := []string{"slot01"}
	attributes := []string{}
	values := []any{}
	for key, value := range strayDevice["attributes"].(map[string]any) {
		attributes = append(attributes, key)
		values = append(values, value)
	}
	err := controller.LinkObject(models.StrayPath+"chT", models.PhysicalPath+"BASIC/A/R1/A01", attributes, values, slots)
	assert.Nil(t, err)
}

// Tests UnlinkObject
func TestUnlinkObjectWithInvalidPath(t *testing.T) {
	controller, _, _ := layersSetup(t)

	err := controller.UnlinkObject("/invalid/path")
	assert.NotNil(t, err)
	assert.Equal(t, "invalid object path", err.Error())
}

func TestUnlinkObjectWithValidPath(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	test_utils.MockUpdateObject(mockAPI, nil, map[string]any{"message": "successfully unlinked"})

	err := controller.UnlinkObject(models.PhysicalPath + "BASIC/A/R1/A01")
	assert.Nil(t, err)
}
