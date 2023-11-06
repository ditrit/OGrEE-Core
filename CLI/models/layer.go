package models

import (
	"strings"

	"github.com/elliotchance/pie/v2"
	"github.com/gertd/go-pluralize"
)

const (
	LayerApplicability = "applicability"
	LayerFilters       = "filters"
	LayerFiltersRemove = LayerFilters + "-"
)

var pluralizeClient = pluralize.NewClient()

type Layer struct {
	Name       string
	apiFilters map[string]string // filters that must be applied to the api request to get only the elements that are part of that layer
}

func (layer Layer) ApplyFilters(filters map[string]string) {
	for key, value := range layer.apiFilters {
		filters[key] = value
	}
}

type LayersFactory interface {
	// FromObjects returns the corresponding layers for the received objects
	FromObjects(objects []any) []Layer
}

type LayerByCategory struct {
	layer    Layer  // Layer to be returned in case of any object is of the category
	category string // category to which the objects must belong
}

// returns true if the object belongs to the category of the factory
func (factory LayerByCategory) objectIsPart(object any) bool {
	objectMap, isMap := object.(map[string]any)
	if !isMap {
		return false
	}

	objectCategory, isPresent := objectMap["category"].(string)

	return isPresent && objectCategory == factory.category
}

// FromObjects returns the layer of the factory is at least one object in the list is of the correct category
func (factory LayerByCategory) FromObjects(objects []any) []Layer {
	if pie.Any(objects, factory.objectIsPart) {
		return []Layer{factory.layer}
	}

	return []Layer{}
}

type LayerByAttribute struct {
	category  string // category to which the objects belong
	attribute string // attribute on which to create layers
}

// FromObjects returns one layer for each distinct layer.attribute value found in the list of objects.
func (factory LayerByAttribute) FromObjects(objects []any) []Layer {
	attributes := []string{}

	for _, object := range objects {
		objectMap, isMap := object.(map[string]any)
		if isMap {
			objectAttributes, hasAttributes := objectMap["attributes"].(map[string]any)
			if hasAttributes {
				objectAttribute, hasAttribute := objectAttributes[factory.attribute].(string)
				if hasAttribute {
					attributes = append(attributes, objectAttribute)
				}
			}
		}
	}

	attributes = pie.Unique(attributes)
	layers := []Layer{}

	for _, attribute := range attributes {
		layerName := "#" + pluralizeClient.Plural(attribute)
		layers = append(layers, Layer{
			Name: layerName,
			apiFilters: map[string]string{
				"category":        factory.category,
				factory.attribute: attribute,
			},
		})
	}

	return layers
}

var (
	RacksLayer = Layer{
		Name:       "#racks",
		apiFilters: map[string]string{"category": "rack"},
	}
	GroupsLayer = Layer{
		Name:       "#groups",
		apiFilters: map[string]string{"namespace": "logical", "category": "group"},
	}
	CorridorsLayer = Layer{
		Name:       "#corridors",
		apiFilters: map[string]string{"category": "corridor"},
	}
)

var (
	GroupsLayerFactory = LayerByCategory{
		layer:    GroupsLayer,
		category: "group",
	}
	DeviceTypeLayers = LayerByAttribute{
		category:  "device",
		attribute: "type",
	}
)

// LayerFactory to be executed for each entity
var LayersByEntity = map[int][]LayersFactory{
	ROOM: {
		LayerByCategory{
			layer:    CorridorsLayer,
			category: "corridor",
		},
		GroupsLayerFactory,
		LayerByCategory{
			layer:    RacksLayer,
			category: "rack",
		},
	},
	RACK: {
		GroupsLayerFactory,
		DeviceTypeLayers,
	},
	DEVICE: {
		DeviceTypeLayers,
	},
}

// Returns true if the received id element is a layer (starts with #)
func IsIDElementLayer(idElement string) bool {
	return strings.HasPrefix(idElement, "#")
}

// Returns true if the id is a layer folder (e.g. room1.#racks)
func IsObjectIDLayer(id string) bool {
	idSplit := strings.Split(id, ".")

	return IsIDElementLayer(idSplit[len(idSplit)-1])
}

// Returns true if the path is a layer (e.g. .../room1/#racks)
func PathIsLayer(path string) bool {
	pathSplit := strings.Split(path, "/")

	return IsIDElementLayer(pathSplit[len(pathSplit)-1])
}

// Returns true if the path has a layer (e.g. .../room1/#racks/rack1)
func PathHasLayer(path string) bool {
	return PathRemoveLayer(path) != path
}

// Removes a layer from the path.
// For example .../room1/#racks/rack1 is transformed into .../room1/rack1
func PathRemoveLayer(path string) string {
	pathSplit := strings.Split(path, "/")

	pathSplit = pie.Filter(pathSplit, func(pathElement string) bool {
		return !IsIDElementLayer(pathElement)
	})

	return strings.Join(pathSplit, "/")
}
