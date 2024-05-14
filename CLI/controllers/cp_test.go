package controllers_test

import (
	"cli/controllers"
	"cli/models"
	test_utils "cli/test"
	"cli/utils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCpObjectThatIsNotALayerCantBeCopied(t *testing.T) {
	controller, mockAPI, _, _ := lsSetup(t)

	test_utils.MockGetObjectByEntity(mockAPI, "tags", map[string]any{
		"slug": "asd",
	})

	err := controller.Cp("/Logical/Tags/asd", "asd2")
	assert.ErrorIs(t, err, controllers.ErrObjectCantBeCopied)
}

func TestCpLayerWithDestPathOrSlugCopiesSource(t *testing.T) {
	tests := []struct {
		name        string
		destination string
	}{
		{"WithPath", "/Logical/Layers/layer2"},
		{"WithSlug", "layer2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller, mockAPI, _, _ := lsSetup(t)

			layer1 := map[string]any{
				"slug":                    "layer1",
				models.LayerApplicability: "BASIC.A.R1",
				models.LayerFilters:       "category = device",
			}

			test_utils.MockGetObjectByEntity(mockAPI, "layers", layer1)
			test_utils.MockCreateObject(mockAPI, "layer", map[string]any{
				"slug":                    "layer2",
				models.LayerApplicability: "BASIC.A.R1",
				models.LayerFilters:       "category = device",
			})

			err := controller.Cp("/Logical/Layers/layer1", tt.destination)
			assert.Nil(t, err)
		})
	}
}

func TestCpLayerWhenSourceIsCachedCopiesSource(t *testing.T) {
	controller, mockAPI, _, mockClock := lsSetup(t)

	layer1 := map[string]any{
		"slug":                    "layer1",
		models.LayerApplicability: "BASIC.A.R1",
		models.LayerFilters:       "category = device",
	}

	now := time.Now()
	mockClock.On("Now").Return(now).Once()
	test_utils.MockGetObjectsByEntity(mockAPI, "layers", []any{layer1})

	objects, err := controller.Ls("/Logical/Layers", nil, nil)
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	utils.ContainsObjectNamed(t, objects, "layer1")

	test_utils.MockGetObjectByEntity(mockAPI, "layers", layer1)
	test_utils.MockCreateObject(mockAPI, "layer", map[string]any{
		"slug":                    "layer2",
		models.LayerApplicability: "BASIC.A.R1",
		models.LayerFilters:       "category = device",
	})

	err = controller.Cp("/Logical/Layers/layer1", "/Logical/Layers/layer2")
	assert.Nil(t, err)

	mockClock.On("Now").Return(now.Add(5 * time.Second)).Once()
	objects, err = controller.Ls("/Logical/Layers", nil, nil)
	assert.Nil(t, err)
	assert.Len(t, objects, 2)
	utils.ContainsObjectNamed(t, objects, "layer1")
	utils.ContainsObjectNamed(t, objects, "layer2")

	test_utils.MockGetObjectHierarchy(mockAPI, roomWithoutChildren)
	mockClock.On("Now").Return(now.Add(6 * time.Second)).Once()

	objects, err = controller.Ls("/Physical/BASIC/A/R1", nil, nil)
	assert.Nil(t, err)
	assert.Len(t, objects, 2)
	utils.ContainsObjectNamed(t, objects, "#layer1")
	utils.ContainsObjectNamed(t, objects, "#layer2")
}
