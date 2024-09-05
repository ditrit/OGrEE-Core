package controllers_test

import (
	"cli/controllers"
	test_utils "cli/test"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test UnfoldPath
func TestUnfoldPath(t *testing.T) {
	controller, mockAPI, _, _ := test_utils.SetMainEnvironmentMock(t)
	wildcardPath := "/Physical/site/building/room/rack*"
	firstRackPath := "/Physical/site/building/room/rack1"
	secondRackPath := "/Physical/site/building/room/rack2"
	rack1 := test_utils.GetEntity("rack", "rack1", "site.building.room", "")
	rack2 := test_utils.GetEntity("rack", "rack2", "site.building.room", "")
	test_utils.MockGetObjects(mockAPI, "id=site.building.room.rack*&namespace=physical.hierarchy", []any{rack1, rack2})
	controllers.State.ClipBoard = []string{firstRackPath}
	tests := []struct {
		name          string
		path          string
		expectedValue []string
	}{
		{"StringWithStar", wildcardPath, []string{firstRackPath, secondRackPath}},
		{"Clipboard", "_", controllers.State.ClipBoard},
		{"SimplePath", secondRackPath, []string{secondRackPath}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := controller.UnfoldPath(tt.path)
			assert.Nil(t, err)
			assert.Equal(t, tt.expectedValue, results)
		})
	}
}
