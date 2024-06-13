package models

import (
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/elliotchance/pie/v2"
	"github.com/gertd/go-pluralize"
)

const (
	LayerApplicability = "applicability"
	LayerFilters       = "filter"
	LayerFiltersAdd    = LayerFilters + "+"
)

var pluralizeClient = pluralize.NewClient()

type Layer interface {
	Name() string
	ApplyFilters(filters map[string]string)
}

type AutomaticLayer struct {
	name       string
	apiFilters string // filters that must be applied to the api request to get only the elements that are part of that layer
}

func (layer AutomaticLayer) Name() string {
	return layer.name
}

func (layer AutomaticLayer) ApplyFilters(filters map[string]string) {
	if complexFilter, ok := filters["filter"]; ok {
		filters["filter"] = "(" + complexFilter + ") & (" + layer.apiFilters + ")"
	} else {
		filters["filter"] = layer.apiFilters
	}
}

type UserDefinedLayer struct {
	Slug          string
	Applicability string
	Filter        string
}

func (layer UserDefinedLayer) Name() string {
	return toLayerName(layer.Slug)
}

func (layer UserDefinedLayer) Matches(path string) bool {
	applicability := strings.TrimSuffix(
		PhysicalIDToPath(layer.Applicability),
		"/",
	)

	match, err := doublestar.Match(applicability, path)

	return err == nil && match
}

func (layer UserDefinedLayer) ApplyFilters(filters map[string]string) {
	if complexFilter, ok := filters["filter"]; ok {
		filters["filter"] = "(" + complexFilter + ") & (" + layer.Filter + ")"
	} else {
		filters["filter"] = layer.Filter
	}
}

type LayersFactory interface {
	// FromObjects returns the corresponding layers for the received objects
	FromObjects(objects []any) []AutomaticLayer
}

type LayerByCategory struct {
	layer    AutomaticLayer // Layer to be returned in case of any object is of the category
	category string         // category to which the objects must belong
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
func (factory LayerByCategory) FromObjects(objects []any) []AutomaticLayer {
	if pie.Any(objects, factory.objectIsPart) {
		return []AutomaticLayer{factory.layer}
	}

	return []AutomaticLayer{}
}

type LayerByAttribute struct {
	category  string // category to which the objects belong
	attribute string // attribute on which to create layers
}

// FromObjects returns one layer for each distinct layer.attribute value found in the list of objects.
func (factory LayerByAttribute) FromObjects(objects []any) []AutomaticLayer {
	attributes := []string{}

	for _, object := range objects {
		objectMap, isMap := object.(map[string]any)
		if isMap {
			objectAttributes, hasAttributes := objectMap["attributes"].(map[string]any)
			if hasAttributes {
				if prefix, suffix, found := strings.Cut(factory.attribute, "."); found {
					// the attribute is an object (prefix), get value from key (suffix), e.g. virtual_config.type
					if objAttrPrefix, hasAttribute := objectAttributes[prefix].(map[string]any); hasAttribute {
						if objectAttribute, hasAttribute := objAttrPrefix[suffix].(string); hasAttribute {
							attributes = append(attributes, objectAttribute)
						}
					}
				} else {
					objectAttribute, hasAttribute := objectAttributes[factory.attribute].(string)
					if hasAttribute {
						attributes = append(attributes, objectAttribute)
					}
				}
			}
		}
	}

	attributes = pie.Unique(attributes)
	layers := []AutomaticLayer{}

	for _, attribute := range attributes {
		layerName := toLayerName(pluralizeClient.Plural(attribute))
		layers = append(layers, AutomaticLayer{
			name:       layerName,
			apiFilters: "category=" + factory.category + "&" + factory.attribute + "=" + attribute,
		})
	}

	return layers
}

func toLayerName(name string) string {
	return "#" + name
}

var (
	RacksLayer = AutomaticLayer{
		name:       "#racks",
		apiFilters: "category=rack",
	}
	GroupsLayer = AutomaticLayer{
		name:       "#groups",
		apiFilters: "category=group",
	}
	CorridorsLayer = AutomaticLayer{
		name:       "#corridors",
		apiFilters: "category=corridor",
	}
	GenericsLayer = AutomaticLayer{
		name:       "#generics",
		apiFilters: "category=generic",
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
	VirtualTypeLayers = LayerByAttribute{
		category:  "virtual_obj",
		attribute: "virtual_config.type",
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
		LayerByCategory{
			layer:    GenericsLayer,
			category: "generic",
		},
		LayerByAttribute{
			category:  "generic",
			attribute: "type",
		},
	},
	RACK: {
		GroupsLayerFactory,
		DeviceTypeLayers,
	},
	DEVICE: {
		DeviceTypeLayers,
	},
	VIRTUALOBJ: {
		VirtualTypeLayers,
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
