package utils

import (
	"cli/controllers"
	mocks "cli/mocks/controllers"
	"testing"
)

func NewControllerWithMocks(t *testing.T) (controllers.Controller, *mocks.APIPort, *mocks.Ogree3DPort, *mocks.ClockPort) {
	// Returns a Mock controller
	mockAPI := mocks.NewAPIPort(t)
	mockOgree3D := mocks.NewOgree3DPort(t)
	mockClock := mocks.NewClockPort(t)
	return controllers.Controller{
		API:     mockAPI,
		Ogree3D: mockOgree3D,
		Clock:   mockClock,
	}, mockAPI, mockOgree3D, mockClock
}

func SetMainEnvironmentMock(t *testing.T) (controllers.Controller, *mocks.APIPort, *mocks.Ogree3DPort, *mocks.ClockPort) {
	// Sets the CLI environment with the mocks. At the end of the test, it will reset the environment
	oldDynamicSymbolTable := controllers.State.DynamicSymbolTable
	oldFuncTable := controllers.State.FuncTable
	oldClipboard := controllers.State.ClipBoard
	oldPrevPath := controllers.State.PrevPath
	oldCurrPath := controllers.State.CurrPath
	oldDrawableObjs := controllers.State.DrawableObjs
	controllers.State.DynamicSymbolTable = map[string]any{}
	controllers.State.FuncTable = map[string]any{}
	controllers.State.ClipBoard = []string{}
	controllers.State.DrawableObjs = []int{}

	controller, mockAPI, mockOgree3D, mockClock := NewControllerWithMocks(t)

	oldControllerValue := controllers.C
	controllers.C = controller
	oldHierarchy := controllers.State.Hierarchy
	controllers.State.Hierarchy = controllers.BuildBaseTree(controller)

	t.Cleanup(func() {
		controllers.State.DynamicSymbolTable = oldDynamicSymbolTable
		controllers.State.FuncTable = oldFuncTable
		controllers.C = oldControllerValue
		controllers.State.Hierarchy = oldHierarchy
		controllers.State.ClipBoard = oldClipboard
		controllers.State.DrawableObjs = oldDrawableObjs
		controllers.State.PrevPath = oldPrevPath
		controllers.State.CurrPath = oldCurrPath
	})

	return controller, mockAPI, mockOgree3D, mockClock
}

func GetTestDrawableJson() map[string]map[string]any {
	return map[string]map[string]any{
		"rack": map[string]any{
			"name":        true,
			"parentId":    true,
			"category":    true,
			"description": false,
			"domain":      true,
			"attributes": map[string]any{
				"color": true,
			},
		},
	}
}

func GetEntity(entityName string, name string, parentId string, domain string) map[string]any {
	id := name
	if parentId != "" {
		id = parentId + "." + id
	}
	switch entityName {
	case "domain":
		return map[string]any{
			"category":    "domain",
			"id":          id,
			"name":        name,
			"parentId":    parentId,
			"description": "",
			"attributes":  map[string]any{},
		}
	case "rack":
		return map[string]any{
			"category": "rack",
			"children": []any{},
			"id":       id,
			"name":     name,
			"parentId": parentId,
			"domain":   domain,
		}
	case "device":
		return map[string]any{
			"category":    "device",
			"id":          id,
			"name":        name,
			"parentId":    parentId,
			"domain":      domain,
			"description": "",
			"attributes": map[string]any{
				"height":      47,
				"heightUnit":  "U",
				"orientation": "front",
				"size":        []float64{1, 1},
				"sizeUnit":    "cm",
			},
		}
	case "generic":
		return map[string]any{
			"attributes": map[string]any{
				"height":     1.0,
				"heightUnit": "cm",
				"rotation":   []float64{0, 0, 0},
				"posXYZ":     []float64{1, 1, 1},
				"posXYUnit":  "m",
				"size":       []float64{1, 1},
				"shape":      "cube",
				"sizeUnit":   "cm",
				"type":       "box",
			},
			"category":    "generic",
			"description": "",
			"domain":      domain,
			"name":        name,
			"parentId":    parentId,
		}
	default:
		return nil
	}
}
