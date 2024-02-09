package controllers

import (
	"cli/models"
	"fmt"
	"regexp"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/elliotchance/pie/v2"
)

// Splits a path or an id using the separator, to obtain:
//  1. The path/id without the layer
//  2. The layer
//  3. errLayerNotFound, in case a found layer does not exist
//
// If the path or id has more than one layer, all are remove but only the name of the last one is returned
func (controller Controller) splitLayer(pathOrID, separator string) (string, models.Layer, error) {
	idSplit := strings.Split(pathOrID, separator)

	return controller.splitLayerRecursive("", idSplit, separator)
}

// Recursive function to make splitLayer able to support nested layers.
// previousRealID is the real id obtained until this function is called.
// idSplit is the list of id elements that remain to be determined if they are a layer.
func (controller Controller) splitLayerRecursive(previousRealID string, idSplit []string, separator string) (string, models.Layer, error) {
	layerIndex := pie.FindFirstUsing(idSplit, func(idElement string) bool {
		return models.IsIDElementLayer(idElement)
	})

	var realID string
	var layer models.Layer

	if layerIndex == -1 {
		realID = strings.Join(idSplit, separator)
	} else {
		realID = strings.Join(idSplit[:layerIndex], separator)
	}

	if previousRealID != "" {
		realID = previousRealID + separator + realID
	}

	if layerIndex != -1 {
		layerName := idSplit[layerIndex]

		var err error

		layer, err = controller.getLayerFromHierarchy(realID, layerName, separator)
		if err != nil {
			return "", nil, err
		}

		if layerIndex != len(idSplit)-1 {
			var newLayer models.Layer

			realID, newLayer, err = controller.splitLayerRecursive(realID, idSplit[layerIndex+1:], separator)
			if err != nil {
				return "", nil, err
			}

			if newLayer != nil {
				layer = newLayer
			}
		}
	}

	return realID, layer, nil
}

// Given the parent id, the layer name and the separator used in the parent id,
// finds the layer in the hierarchy.
//
// If the layer is not present, Tree is executed on the parent to try to find the layer in its children.
//
// If the layer is not present, errLayerNotFound is returned.
func (controller Controller) getLayerFromHierarchy(parent, layerName, separator string) (models.Layer, error) {
	if separator != "/" {
		parent = strings.ReplaceAll(models.PhysicalIDToPath(parent), "/*", "")
	}

	layerNode := State.Hierarchy.FindNode(parent + "/" + layerName)
	if layerNode == nil {
		parentNode, err := controller.lsNode(parent)
		if err != nil {
			return nil, err
		}

		var isPresent bool

		layerNode, isPresent = parentNode.Children[layerName]
		if !isPresent {
			return nil, errLayerNotFound
		}
	}

	layerPointer, isPointer := layerNode.Obj.(*any)
	if isPointer {
		return (*layerPointer).(models.Layer), nil
	}

	return layerNode.Obj.(models.Layer), nil
}

// Obtains the layer from an object ID to obtain:
//
// If the object is not inside a layer (e.g. room1.rack1), the same object ID, nil and nil.
//
// If the object is inside a layer (e.g. room1.#racks.rack1):
//  1. The real object ID, without the layer (room1.rack1 for room1.#racks.rack1).
//  2. If the object is a layer (e.g. room1.#racks, room1.#racks.*), the layer object.
//  3. errLayerNotFound, in case a found layer does not exist
func (controller Controller) GetLayer(id string) (string, models.Layer, error) {
	realID, layer, err := controller.splitLayer(id, ".")
	if err != nil {
		return "", nil, err
	}

	if layer == nil {
		return realID, nil, nil
	}

	idSplit := strings.Split(id, ".")
	layerIndex := pie.FindFirstUsing(idSplit, func(idElement string) bool {
		return idElement == layer.Name()
	})

	// only in case the layer is the second to last or laster position of the id the filter is applied (e.g. get room1/#racks, get room1/#racks/*)
	// to avoid applying filters when we are inside a layer element (e.g. get room1.#racks.rack1.*)
	if layerIndex < len(idSplit)-2 {
		return realID, nil, nil
	}

	if layerIndex == len(idSplit)-1 {
		// a layer is not an object, it is a reference to all objects that meet a condition
		realID = realID + ".*"
	}

	return realID, layer, nil
}

func (controller Controller) CreateLayer(slug, applicability, filterName, filterValue string) error {
	applicability, err := TranslateApplicability(applicability)
	if err != nil {
		return err
	}

	return controller.PostObj(models.LAYER, models.EntityToString(models.LAYER), map[string]any{
		"slug":                    slug,
		models.LayerFilters:       map[string]any{filterName: filterValue},
		models.LayerApplicability: applicability,
	}, models.LayersPath+slug)
}

func (controller Controller) UpdateLayer(path string, attributeName string, value any) error {
	var err error
	switch attributeName {
	case models.LayerApplicability:
		var applicability string
		applicability, err = TranslateApplicability(value.(string))
		if err != nil {
			return err
		}

		_, err = controller.UpdateObj(path, map[string]any{attributeName: applicability})
	case models.LayerFiltersRemove:
		_, err = controller.UpdateObj(path, map[string]any{attributeName: value})
	case models.LayerFiltersAdd:
		var filters map[string]any
		reg := regexp.MustCompile(`^(\w(\w|\-)*)=[^!<>=&|]+$`)
		if strings.HasPrefix(value.(string), "filter=") {
			filters = map[string]any{"filter": value.(string)[7:]}
		} else if reg.MatchString(value.(string)) {
			values := strings.Split(value.(string), "=")
			filters = map[string]any{values[0]: values[1]}
		} else {
			filters = map[string]any{"filter": value.(string)}
		}
		_, err = controller.UpdateObj(path, map[string]any{models.LayerFilters: filters})
	default:
		_, err = controller.UpdateObj(path, map[string]any{attributeName: value})
		if attributeName == "slug" {
			State.Hierarchy.Children["Logical"].Children["Layers"].IsCached = false
		}
	}

	return err
}

func TranslateApplicability(applicability string) (string, error) {
	applicability = TranslatePath(applicability, false)

	if !models.IsPhysical(applicability) {
		return "", fmt.Errorf("applicability must be an hierarchical path, found: %s", applicability)
	}

	if !doublestar.ValidatePattern(applicability) {
		return "", fmt.Errorf("applicability pattern is not valid")
	}

	return models.PhysicalPathToObjectID(applicability), nil
}
