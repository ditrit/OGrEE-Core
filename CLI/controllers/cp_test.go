package controllers_test

import (
	"cli/controllers"
	"cli/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCpObjectThatIsNotALayerCantBeCopied(t *testing.T) {
	controller, mockAPI, _, _ := lsSetup(t)

	mockGetObjectByEntity(mockAPI, "tags", map[string]any{
		"slug": "asd",
	})

	err := controller.Cp("/Logical/Tags/asd", "asd2")
	assert.ErrorIs(t, err, controllers.ErrObjectCantBeCopied)
}

func TestCpLayerWithDestPathCopiesSource(t *testing.T) {
	controller, mockAPI, _, _ := lsSetup(t)

	layer1 := map[string]any{
		"slug":                    "layer1",
		models.LayerApplicability: "BASIC.A.R1",
		models.LayerFilters:       map[string]any{"category": "device"},
	}

	mockGetObjectByEntity(mockAPI, "layers", layer1)
	mockCreateObject(mockAPI, "layer", map[string]any{
		"slug":                    "layer2",
		models.LayerApplicability: "BASIC.A.R1",
		models.LayerFilters:       map[string]any{"category": "device"},
	})

	err := controller.Cp("/Logical/Layers/layer1", "/Logical/Layers/layer2")
	assert.Nil(t, err)
}

func TestCpLayerWithDestSlugCopiesSource(t *testing.T) {
	controller, mockAPI, _, _ := lsSetup(t)

	layer1 := map[string]any{
		"slug":                    "layer1",
		models.LayerApplicability: "BASIC.A.R1",
		models.LayerFilters:       map[string]any{"category": "device"},
	}

	mockGetObjectByEntity(mockAPI, "layers", layer1)
	mockCreateObject(mockAPI, "layer", map[string]any{
		"slug":                    "layer2",
		models.LayerApplicability: "BASIC.A.R1",
		models.LayerFilters:       map[string]any{"category": "device"},
	})

	err := controller.Cp("/Logical/Layers/layer1", "layer2")
	assert.Nil(t, err)
}

func TestCpLayerWhenSourceIsCachedCopiesSource(t *testing.T) {
	controller, mockAPI, _, mockClock := lsSetup(t)

	layer1 := map[string]any{
		"slug":                    "layer1",
		models.LayerApplicability: "BASIC.A.R1",
		models.LayerFilters:       map[string]any{"category": "device"},
	}

	now := time.Now()
	mockClock.On("Now").Return(now).Once()
	mockGetObjectsByEntity(mockAPI, "layers", []any{layer1})

	objects, err := controller.Ls("/Logical/Layers", nil, "")
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	assert.Equal(t, "layer1", objects[0]["name"])

	mockGetObjectByEntity(mockAPI, "layers", layer1)
	mockCreateObject(mockAPI, "layer", map[string]any{
		"slug":                    "layer2",
		models.LayerApplicability: "BASIC.A.R1",
		models.LayerFilters:       map[string]any{"category": "device"},
	})

	err = controller.Cp("/Logical/Layers/layer1", "/Logical/Layers/layer2")
	assert.Nil(t, err)

	mockClock.On("Now").Return(now.Add(5 * time.Second)).Once()
	objects, err = controller.Ls("/Logical/Layers", nil, "")
	assert.Nil(t, err)
	assert.Len(t, objects, 2)
	assert.Equal(t, "layer1", objects[0]["name"])
	assert.Equal(t, "layer2", objects[1]["name"])

	mockGetObjectHierarchy(mockAPI, roomWithoutChildren)
	mockClock.On("Now").Return(now.Add(6 * time.Second)).Once()

	objects, err = controller.Ls("/Physical/BASIC/A/R1", nil, "")
	assert.Nil(t, err)
	assert.Len(t, objects, 2)
	assert.Equal(t, "#layer1", objects[0]["name"])
	assert.Equal(t, "#layer2", objects[1]["name"])
}
