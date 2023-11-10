package controllers_test

import (
	"cli/controllers"
	mocks "cli/mocks/controllers"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func lsSetup(t *testing.T) (controllers.Controller, *mocks.APIPort, *mocks.Ogree3DPort, *mocks.ClockPort) {
	controller, mockAPI, mockOgree3d, clockMock := newControllerWithMocks(t)
	controllers.State.Hierarchy = controllers.BuildBaseTree(controller)

	return controller, mockAPI, mockOgree3d, clockMock
}

func TestLsOnElementAsksForLayersIfTheyHaveNeverBeenLoaded(t *testing.T) {
	controller, mockAPI, _, mockClock := lsSetup(t)

	mockClock.On("Now").Return(time.Now()).Once()
	mockGetObjectsByEntity(mockAPI, "layers", []any{})
	mockGetObjectHierarchy(mockAPI, map[string]any{
		"category": "room",
		"children": []any{},
		"id":       "BASIC.A.R1",
		"name":     "R1",
		"parentId": "BASIC.A",
	})

	objects, err := controller.Ls("/Physical/BASIC/A/R1", nil, "", false)
	assert.Nil(t, err)
	assert.Len(t, objects, 0)
}

func TestLsOnElementNotAsksForLayersIfTheyAreUpdated(t *testing.T) {
	controller, mockAPI, _, mockClock := lsSetup(t)

	now := time.Now()
	mockClock.On("Now").Return(now).Once()
	mockGetObjectsByEntity(mockAPI, "layers", []any{})

	objects, err := controller.Ls("/Logical/Layers", nil, "", false)
	assert.Nil(t, err)
	assert.Len(t, objects, 0)

	mockClock.On("Now").Return(now.Add(5 * time.Second)).Once()
	mockGetObjectHierarchy(mockAPI, map[string]any{
		"category": "room",
		"children": []any{},
		"id":       "BASIC.A.R1",
		"name":     "R1",
		"parentId": "BASIC.A",
	})

	objects, err = controller.Ls("/Physical/BASIC/A/R1", nil, "", false)
	assert.Nil(t, err)
	assert.Len(t, objects, 0)
}

func TestLsOnElementAsksForLayersIfTheyAreNotUpdated(t *testing.T) {
	controller, mockAPI, _, mockClock := lsSetup(t)

	now := time.Now()
	mockClock.On("Now").Return(now).Once()
	mockGetObjectsByEntity(mockAPI, "layers", []any{})

	objects, err := controller.Ls("/Logical/Layers", nil, "", false)
	assert.Nil(t, err)
	assert.Len(t, objects, 0)

	mockClock.On("Now").Return(now.Add(50 * time.Minute)).Once()
	mockGetObjectsByEntity(mockAPI, "layers", []any{})
	mockGetObjectHierarchy(mockAPI, map[string]any{
		"category": "room",
		"children": []any{},
		"id":       "BASIC.A.R1",
		"name":     "R1",
		"parentId": "BASIC.A",
	})

	objects, err = controller.Ls("/Physical/BASIC/A/R1", nil, "", false)
	assert.Nil(t, err)
	assert.Len(t, objects, 0)
}
