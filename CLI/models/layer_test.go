package models_test

import (
	"cli/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
)

func TestAutomaticLayerName(t *testing.T) {
	assert.Equal(t, "#racks", models.RacksLayer.Name())
}

func TestAutomaticLayerApplyFilters(t *testing.T) {
	filters := map[string]string{}
	models.RacksLayer.ApplyFilters(filters)
	assert.Contains(t, filters, "filter")
	assert.Equal(t, "category=rack", filters["filter"])

	filters["filter"] = "name=A01"
	models.RacksLayer.ApplyFilters(filters)
	assert.Contains(t, filters, "filter")
	assert.Equal(t, "(name=A01) & (category=rack)", filters["filter"])
}

func TestUserDefinedLayerName(t *testing.T) {
	layerName := "racks"
	layer := models.UserDefinedLayer{
		Slug:          layerName,
		Applicability: "BASIC.A.R1",
		Filter:        "category=rack",
	}
	assert.Equal(t, "#"+layerName, layer.Name())
}

func TestUserDefinedLayerMatches(t *testing.T) {
	layerName := "#racks"
	layer := models.UserDefinedLayer{
		Slug:          layerName,
		Applicability: "BASIC.A.R*.**",
		Filter:        "category=rack",
	}
	assert.True(t, layer.Matches(models.PhysicalPath+"BASIC/A/R1/A01"))
	assert.True(t, layer.Matches(models.PhysicalPath+"BASIC/A/R2/A01"))
	assert.True(t, layer.Matches(models.PhysicalPath+"BASIC/A/R1/A01/D01"))
	assert.False(t, layer.Matches(models.PhysicalPath+"BASIC/A/U1/A01"))
}

func TestUserDefinedLayerApplyFilters(t *testing.T) {
	layerName := "#racks"
	layer := models.UserDefinedLayer{
		Slug:          layerName,
		Applicability: "BASIC.A.R*.**",
		Filter:        "category=rack",
	}
	filters := map[string]string{}
	layer.ApplyFilters(filters)
	assert.Contains(t, filters, "filter")
	assert.Equal(t, layer.Filter, filters["filter"])

	filters["filter"] = "name=R01"
	layer.ApplyFilters(filters)
	assert.Contains(t, filters, "filter")
	assert.Equal(t, "(name=R01) & (category=rack)", filters["filter"])
}

func TestLayerByCategoryFromObjects(t *testing.T) {
	objects := []any{
		map[string]any{
			"category": "device",
		},
		map[string]any{
			"category": "group",
		},
	}
	automaticLayers := models.GroupsLayerFactory.FromObjects(objects)
	assert.Len(t, automaticLayers, 1)
	assert.Equal(t, "#groups", automaticLayers[0].Name())

	objects = slices.Delete(objects, 1, 2) // we delete the group
	automaticLayers = models.GroupsLayerFactory.FromObjects(objects)
	assert.Len(t, automaticLayers, 0)
}

func TestLayerByAttributeFromObjects(t *testing.T) {
	objects := []any{
		map[string]any{
			"category": "device",
			"attributes": map[string]any{
				"type": "blade",
			},
		},
		map[string]any{
			"category": "device",
			"attributes": map[string]any{
				"type": "blade",
			},
		},
		map[string]any{
			"category": "device",
			"attributes": map[string]any{
				"type": "table",
			},
		},
	}
	results := map[string]string{
		"#tables": "category=device&type=table",
		"#blades": "category=device&type=blade",
	}
	automaticLayers := models.DeviceTypeLayers.FromObjects(objects)
	assert.Len(t, automaticLayers, 2)
	for key, value := range results {
		filters := map[string]string{}
		index := slices.IndexFunc(automaticLayers, func(e models.AutomaticLayer) bool { return e.Name() == key })
		assert.True(t, index >= 0)
		automaticLayers[index].ApplyFilters(filters)
		assert.Equal(t, value, filters["filter"])
	}
}

func TestIs(t *testing.T) {
	tests := []struct {
		name          string
		isFunction    func(string) bool
		correctPath   string
		incorrectPath string
	}{
		{"IsIDElementLayer", models.IsIDElementLayer, "#layer-id", "id"},
		{"IsLayer", models.PathIsLayer, "/site/building/room1/room1/#racks", "/site/building/room1/room1/rack1"},
		{"HasLayer", models.PathHasLayer, "/site/building/room1/room1/#racks/rack1", "/site/building/room1/room1/rack1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.True(t, tt.isFunction(tt.correctPath))
			assert.False(t, tt.isFunction(tt.incorrectPath))
		})
	}
}

func TestPathRemoveLayer(t *testing.T) {
	expectedPath := "/site/building/room1/room1/rack1"
	assert.Equal(t, expectedPath, models.PathRemoveLayer("/site/building/room1/room1/#racks/rack1"))
	assert.Equal(t, expectedPath, models.PathRemoveLayer("/site/building/room1/room1/rack1"))
}
