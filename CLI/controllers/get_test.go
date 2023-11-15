package controllers_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetWithFilters(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjects(mockAPI, "category=room&id=BASIC.A.R1&namespace=physical.hierarchy", []any{roomWithChildren})

	objects, _, err := controller.GetObjectsWildcard("/Physical/BASIC/A/R1", map[string]string{
		"category": "room",
	}, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	assert.Contains(t, objects, removeChildren(roomWithChildren))
}

func TestGetStarWithFilters(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjects(mockAPI, "category=rack&id=BASIC.A.R1.*&namespace=physical.hierarchy", []any{rack1, rack2})

	objects, _, err := controller.GetObjectsWildcard("/Physical/BASIC/A/R1/*", map[string]string{
		"category": "rack",
	}, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 2)
	assert.Contains(t, objects, removeChildren(rack1))
	assert.Contains(t, objects, removeChildren(rack2))
}

func TestGetSomethingStarWithFilters(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjects(mockAPI, "category=rack&id=BASIC.A.R1.A*&namespace=physical.hierarchy", []any{rack1})

	objects, _, err := controller.GetObjectsWildcard("/Physical/BASIC/A/R1/A*", map[string]string{
		"category": "rack",
	}, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	assert.Contains(t, objects, removeChildren(rack1))
}

func TestGetRecursiveSearchAllChildrenCalledInThatWay(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjects(mockAPI, "id=BASIC.A.**R1&namespace=physical.hierarchy", []any{roomWithChildren})

	objects, _, err := controller.GetObjectsWildcard("/Physical/BASIC/A/R1", nil, true)
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	assert.Contains(t, objects, removeChildren(roomWithChildren))
}

func TestGetRecursiveWithFilters(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjects(mockAPI, "category=room&id=BASIC.A.**R1&namespace=physical.hierarchy", []any{roomWithChildren})

	objects, _, err := controller.GetObjectsWildcard("/Physical/BASIC/A/R1", map[string]string{
		"category": "room",
	}, true)
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	assert.Contains(t, objects, removeChildren(roomWithChildren))
}

func TestGetStarRecursiveWithFilters(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjects(mockAPI, "category=device&id=BASIC.A.R1.**&namespace=physical.hierarchy", []any{chassis, pdu})

	objects, _, err := controller.GetObjectsWildcard("/Physical/BASIC/A/R1/*", map[string]string{
		"category": "device",
	}, true)
	assert.Nil(t, err)
	assert.Len(t, objects, 2)
	assert.Contains(t, objects, removeChildren(chassis))
	assert.Contains(t, objects, removeChildren(pdu))
}

func TestGetSomethingStarRecursiveWithFilters(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjects(mockAPI, "category=device&id=BASIC.A.R1.**ch*&namespace=physical.hierarchy", []any{chassis})

	objects, _, err := controller.GetObjectsWildcard("/Physical/BASIC/A/R1/ch*", map[string]string{
		"category": "device",
	}, true)
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	assert.Contains(t, objects, removeChildren(chassis))
}
