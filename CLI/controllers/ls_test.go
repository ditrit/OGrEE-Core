package controllers_test

import (
	"cli/controllers"
	mocks "cli/mocks/controllers"
	"cli/models"
	test_utils "cli/test"
	"cli/utils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func lsSetup(t *testing.T) (controllers.Controller, *mocks.APIPort, *mocks.Ogree3DPort, *mocks.ClockPort) {
	controller, mockAPI, mockOgree3d, clockMock := test_utils.NewControllerWithMocks(t)
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

	objects, err := controller.Ls("/Physical/BASIC/A/R1", nil, nil)
	assert.Nil(t, err)
	assert.Len(t, objects, 0)
}

func TestLsOnElementNotAsksForLayersIfTheyAreUpdated(t *testing.T) {
	controller, mockAPI, _, mockClock := lsSetup(t)

	now := time.Now()
	mockClock.On("Now").Return(now).Once()
	mockGetObjectsByEntity(mockAPI, "layers", []any{})

	objects, err := controller.Ls("/Logical/Layers", nil, nil)
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

	objects, err = controller.Ls("/Physical/BASIC/A/R1", nil, nil)
	assert.Nil(t, err)
	assert.Len(t, objects, 0)
}

func TestLsOnElementAsksForLayersIfTheyAreNotUpdated(t *testing.T) {
	controller, mockAPI, _, mockClock := lsSetup(t)

	now := time.Now()
	mockClock.On("Now").Return(now).Once()
	mockGetObjectsByEntity(mockAPI, "layers", []any{})

	objects, err := controller.Ls("/Logical/Layers", nil, nil)
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

	objects, err = controller.Ls("/Physical/BASIC/A/R1", nil, nil)
	assert.Nil(t, err)
	assert.Len(t, objects, 0)
}

func TestLsWithFilters(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjects(mockAPI, "category=rack&id=BASIC.A.R1.*&namespace=physical.hierarchy", []any{rack1, rack2})

	objects, err := controller.Ls("/Physical/BASIC/A/R1", map[string]string{
		"category": "rack",
	}, nil)
	assert.Nil(t, err)
	assert.Len(t, objects, 2)
	utils.ContainsObjectNamed(t, objects, "A01")
	utils.ContainsObjectNamed(t, objects, "B01")
}

func TestLsWithComplexFilters(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectsWithComplexFilters(
		mockAPI,
		"id=BASIC.A.R1.%2A&namespace=physical.hierarchy",
		map[string]any{
			"filter": "(category=corridor) | (category=rack & name!=A01) ",
		},
		[]any{corridor, rack2},
	)

	objects, err := controller.Ls("/Physical/BASIC/A/R1", map[string]string{
		"filter": "(category=corridor) | (category=rack & name!=A01) ",
	}, nil)
	assert.Nil(t, err)
	assert.Len(t, objects, 2)
	utils.ContainsObjectNamed(t, objects, "CO1")
	utils.ContainsObjectNamed(t, objects, "B01")
}

func TestLsRecursiveReturnsError(t *testing.T) {
	controller, _, _ := layersSetup(t)

	_, err := controller.Ls("/Physical/BASIC/A/R1", nil, &controllers.RecursiveParams{MaxDepth: models.UnlimitedDepth})
	assert.ErrorIs(t, err, controllers.ErrRecursiveOnlyFiltersLayers)
}

func TestLsRecursiveWithFilters(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjects(mockAPI, "category=rack&id=BASIC.A.R1.**.*&namespace=physical.hierarchy", []any{rack1, rack2})

	objects, err := controller.Ls("/Physical/BASIC/A/R1", map[string]string{
		"category": "rack",
	}, &controllers.RecursiveParams{MaxDepth: models.UnlimitedDepth})
	assert.Nil(t, err)
	assert.Len(t, objects, 2)
	utils.ContainsObjectNamed(t, objects, "A01")
	utils.ContainsObjectNamed(t, objects, "B01")
}

func TestLsPointRecursiveMaxLessThatMinReturnsError(t *testing.T) {
	controller, _, _ := layersSetup(t)

	_, err := controller.Ls("/Physical/BASIC/A/R1", map[string]string{
		"category": "rack",
	}, &controllers.RecursiveParams{MinDepth: 3, MaxDepth: 2})
	assert.ErrorIs(t, err, models.ErrMaxLessMin)
}

func TestLsRecursiveWithMinButNotMax(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjects(mockAPI, "category=device&id=BASIC.A.R1.**{1,}.*&namespace=physical.hierarchy", []any{chassis, pdu})

	objects, err := controller.Ls("/Physical/BASIC/A/R1", map[string]string{
		"category": "device",
	}, &controllers.RecursiveParams{MinDepth: 1, MaxDepth: models.UnlimitedDepth})
	assert.Nil(t, err)
	assert.Len(t, objects, 2)
	utils.ContainsObjectNamed(t, objects, "chT")
	utils.ContainsObjectNamed(t, objects, "pdu")
}

func TestLsRecursiveWithMinAndMax(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjects(mockAPI, "category=device&id=BASIC.A.R1.**{1,2}.*&namespace=physical.hierarchy", []any{chassis, pdu})

	objects, err := controller.Ls("/Physical/BASIC/A/R1", map[string]string{
		"category": "device",
	}, &controllers.RecursiveParams{MinDepth: 1, MaxDepth: 2})
	assert.Nil(t, err)
	assert.Len(t, objects, 2)
	utils.ContainsObjectNamed(t, objects, "chT")
	utils.ContainsObjectNamed(t, objects, "pdu")
}
