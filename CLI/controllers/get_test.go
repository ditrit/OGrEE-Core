package controllers_test

import (
	"cli/controllers"
	"cli/models"
	test_utils "cli/test"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetWithSimpleFilters(t *testing.T) {
	tests := []struct {
		name            string
		mockQueryParams string
		mockResponse    []any
		path            string
		category        string
	}{
		{"WithoutStar", "category=room&id=BASIC.A.R1&namespace=physical.hierarchy", []any{roomWithChildren}, "/Physical/BASIC/A/R1", "room"},
		{"WithStar", "category=rack&id=BASIC.A.R1.*&namespace=physical.hierarchy", []any{rack1, rack2}, "/Physical/BASIC/A/R1/*", "rack"},
		{"SomethingStarWithFilters", "category=rack&id=BASIC.A.R1.A*&namespace=physical.hierarchy", []any{rack1}, "/Physical/BASIC/A/R1/A*", "rack"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller, mockAPI, _ := layersSetup(t)

			test_utils.MockGetObjects(mockAPI, tt.mockQueryParams, tt.mockResponse)

			objects, _, err := controller.GetObjectsWildcard(tt.path, map[string]string{
				"category": tt.category,
			}, nil)
			assert.Nil(t, err)
			assert.Len(t, objects, len(tt.mockResponse))
			for _, instance := range tt.mockResponse {
				assert.Contains(t, objects, test_utils.RemoveChildren(instance.(map[string]any)))
			}
		})
	}
}

func TestGetWithComplexFilters(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	test_utils.MockGetObjectsWithComplexFilters(
		mockAPI,
		"id=BASIC.A.R1&namespace=physical.hierarchy",
		map[string]any{
			"filter": "(category=room) & (name=R1 | height>3) ",
		},
		[]any{roomWithChildren},
	)

	objects, _, err := controller.GetObjectsWildcard("/Physical/BASIC/A/R1", map[string]string{
		"filter": "(category=room) & (name=R1 | height>3) ",
	}, nil)
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	assert.Contains(t, objects, test_utils.RemoveChildren(roomWithChildren))
}

func TestGetRecursiveSearchAllChildrenCalledInThatWay(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	test_utils.MockGetObjects(mockAPI, "id=BASIC.A.**.R1&namespace=physical.hierarchy", []any{roomWithChildren})

	objects, _, err := controller.GetObjectsWildcard(
		"/Physical/BASIC/A/R1",
		nil,
		&controllers.RecursiveParams{
			MaxDepth:    models.UnlimitedDepth,
			PathEntered: "R1",
		},
	)
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	assert.Contains(t, objects, test_utils.RemoveChildren(roomWithChildren))
}

func TestGetRecursiveWithFilters(t *testing.T) {
	tests := []struct {
		name            string
		mockQueryParams string
		mockResponse    []any
		path            string
		category        string
		recursiveParams controllers.RecursiveParams
	}{
		{"WithoutStar", "category=room&id=BASIC.A.**.R1&namespace=physical.hierarchy", []any{roomWithChildren}, "/Physical/BASIC/A/R1", "room", controllers.RecursiveParams{MaxDepth: models.UnlimitedDepth, PathEntered: "R1"}},
		{"WithStar", "category=device&id=BASIC.A.R1.**.*&namespace=physical.hierarchy", []any{chassis, pdu}, "/Physical/BASIC/A/R1/*", "device", controllers.RecursiveParams{MaxDepth: models.UnlimitedDepth}},
		{"SomethingStarRecursiveWithFilters", "category=device&id=BASIC.A.R1.**.ch*&namespace=physical.hierarchy", []any{chassis}, "/Physical/BASIC/A/R1/ch*", "device", controllers.RecursiveParams{MaxDepth: models.UnlimitedDepth, PathEntered: "ch*"}},
		{"FolderSomethingStarRecursiveWithFilters", "category=device&id=BASIC.A.**.R1.ch*&namespace=physical.hierarchy", []any{chassis}, "/Physical/BASIC/A/R1/ch*", "device", controllers.RecursiveParams{MaxDepth: models.UnlimitedDepth, PathEntered: "R1/ch*"}},
		{"PointRecursiveIsEqualToNotRecursive", "category=device&id=BASIC.A.R1&namespace=physical.hierarchy", []any{roomWithChildren}, "/Physical/BASIC/A/R1", "device", controllers.RecursiveParams{MaxDepth: models.UnlimitedDepth, PathEntered: "."}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller, mockAPI, _ := layersSetup(t)

			test_utils.MockGetObjects(mockAPI, tt.mockQueryParams, tt.mockResponse)

			objects, _, err := controller.GetObjectsWildcard(tt.path, map[string]string{
				"category": tt.category,
			}, &tt.recursiveParams)
			assert.Nil(t, err)
			assert.Len(t, objects, len(tt.mockResponse))
			for _, object := range tt.mockResponse {
				assert.Contains(t, objects, test_utils.RemoveChildren(object.(map[string]any)))
			}
		})
	}
}
