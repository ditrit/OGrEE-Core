package utils

import (
	"cli/controllers"
	mocks "cli/mocks/controllers"
	"cli/readline"
	"encoding/json"
	"testing"
)

func CopyMap(toCopy map[string]any) map[string]any {
	jsonMap, _ := json.Marshal(toCopy)

	var newMap map[string]any

	json.Unmarshal(jsonMap, &newMap)

	return newMap
}

func EmptyChildren(object map[string]any) map[string]any {
	objectCopy := CopyMap(object)
	objectCopy["children"] = []any{}

	return objectCopy
}

func KeepOnlyDirectChildren(object map[string]any) map[string]any {
	objectCopy := CopyMap(object)

	for _, child := range objectCopy["children"].([]any) {
		delete(child.(map[string]any), "children")
	}

	return objectCopy
}

func RemoveChildren(object map[string]any) map[string]any {
	objectCopy := CopyMap(object)
	delete(objectCopy, "children")

	return objectCopy
}

func RemoveChildrenFromList(objects []any) []any {
	result := []any{}
	for _, object := range objects {
		result = append(result, RemoveChildren(object.(map[string]any)))
	}

	return result
}

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
	oldTerminal := controllers.State.Terminal
	controllers.State.DynamicSymbolTable = map[string]any{}
	controllers.State.FuncTable = map[string]any{}
	controllers.State.ClipBoard = []string{}
	controllers.State.DrawableObjs = []int{}
	rl, _ := readline.New("")
	controllers.State.Terminal = &rl

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
		controllers.State.Terminal = oldTerminal
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
	case "site":
		return map[string]any{
			"category":    "site",
			"children":    []any{},
			"id":          name,
			"name":        name,
			"description": "",
			"domain":      domain,
			"attributes":  map[string]any{},
		}
	case "building":
		return map[string]any{
			"category":    "building",
			"children":    []any{},
			"id":          id,
			"name":        name,
			"description": "",
			"parentId":    parentId,
			"domain":      domain,
			"attributes":  map[string]any{},
		}
	case "room":
		return map[string]any{
			"category":    "room",
			"children":    []any{},
			"id":          id,
			"name":        name,
			"description": "",
			"parentId":    parentId,
			"domain":      domain,
			"attributes":  map[string]any{},
		}
	case "rack":
		return map[string]any{
			"category":    "rack",
			"children":    []any{},
			"id":          id,
			"name":        name,
			"parentId":    parentId,
			"description": "",
			"domain":      domain,
			"attributes":  map[string]any{},
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
	case "genericTableTemplate":
		return map[string]any{
			"slug":        name,
			"description": "a table",
			"category":    "generic",
			"sizeWDHmm":   []any{447, 914.5, 263.3},
			"fbxModel":    "",
			"attributes": map[string]any{
				"type": "table",
			},
			"colors": []any{},
		}
	case "deviceChasisTemplate":
		return map[string]any{
			"slug":        name,
			"description": "",
			"category":    "device",
			"sizeWDHmm":   []any{216, 659, 100},
			"fbxModel":    "",
			"attributes": map[string]any{
				"type":   "chassis",
				"vendor": "IBM",
			},
			"colors":     []any{},
			"components": []any{},
		}
	default:
		return nil
	}
}
