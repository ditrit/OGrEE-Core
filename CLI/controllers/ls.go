package controllers

import (
	"cli/models"
	"cli/utils"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/elliotchance/pie/v2"
	"golang.org/x/exp/maps"
)

var errLayerNotFound = errors.New("the layer used does not exist")

func (controller Controller) Ls(path string, filters map[string]string, sortAttr string) ([]map[string]any, error) {
	var objects []map[string]any
	var err error

	if len(filters) == 0 && !models.PathIsLayer(path) {
		objects, err = controller.lsObjectsWithoutFilters(path)
	} else {
		objects, err = controller.lsObjectsWithFilters(path, filters)
	}

	if err != nil {
		return nil, err
	}

	if sortAttr != "" {
		objects = pie.Filter(objects, func(object map[string]any) bool {
			_, hasAttr := utils.ObjectAttr(object, sortAttr)
			return hasAttr
		})

		if !objectsAreSortable(objects, sortAttr) {
			return nil, fmt.Errorf("objects cannot be sorted according to this attribute")
		}

		sort.Slice(objects, func(i, j int) bool {
			vali, _ := utils.ObjectAttr(objects[i], sortAttr)
			valj, _ := utils.ObjectAttr(objects[j], sortAttr)
			res, _ := utils.CompareVals(vali, valj)
			return res
		})
	} else {
		sort.Slice(objects, func(i, j int) bool {
			if isObjectLayer(objects[i]) {
				if !isObjectLayer(objects[j]) {
					return false
				}
			} else if isObjectLayer(objects[j]) {
				return true
			}

			return utils.NameOrSlug(objects[i]) < utils.NameOrSlug(objects[j])
		})
	}

	return objects, nil
}

func isObjectLayer(object map[string]any) bool {
	name, hasName := object["name"].(string)
	if !hasName {
		return false
	}

	return models.IsObjectIDLayer(name)
}

func (controller Controller) lsObjectsWithoutFilters(path string) ([]map[string]any, error) {
	n, err := controller.lsNode(path)
	if err != nil {
		return nil, err
	}

	objects := []map[string]any{}
	for _, child := range n.Children {
		if child.Obj != nil {
			childObj, isMap := child.Obj.(map[string]any)
			if isMap {
				if models.IsGroup(path) {
					childObj["name"] = strings.ReplaceAll(childObj["id"].(string), ".", "/")
				}
				objects = append(objects, childObj)
				continue
			}
		}

		objects = append(objects, map[string]any{"name": child.Name})
	}

	return objects, nil
}

func (controller Controller) lsObjectsWithFilters(path string, filters map[string]string) ([]map[string]any, error) {
	url, err := controller.ObjectUrlGeneric(path+"/*", 0, filters)
	if err != nil {
		if errors.Is(err, errLayerNotFound) {
			return nil, err
		}

		return nil, fmt.Errorf("cannot use filters at this location")
	}

	resp, err := controller.API.Request("GET", url, nil, http.StatusOK)
	if err != nil {
		return nil, err
	}
	objectsAny := resp.Body["data"].([]any)
	objects := []map[string]any{}
	for _, objAny := range objectsAny {
		obj, ok := objAny.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid response from API on GET %s", url)
		}
		objects = append(objects, obj)
	}
	return objects, nil
}

func objectsAreSortable(objects []map[string]any, attr string) bool {
	for i := 1; i < len(objects); i++ {
		val0, _ := utils.ObjectAttr(objects[0], attr)
		vali, _ := utils.ObjectAttr(objects[i], attr)
		_, comparable := utils.CompareVals(val0, vali)
		if !comparable {
			return false
		}
	}
	return true
}

// Obtains a HierarchyNode using Tree and adds the layers
func (controller Controller) lsNode(path string) (*HierarchyNode, error) {
	n, err := controller.Tree(path, 1)
	if err != nil {
		return nil, err
	}

	addAutomaticLayers(n)

	return n, nil
}

// Adds to the children the automatic layers, depending of the category of the rootObject
// and if any of the children is part of that layer (to avoid displaying empty layers)
func addAutomaticLayers(rootNode *HierarchyNode) {
	rootObject, objIsMap := rootNode.Obj.(map[string]any)
	if !objIsMap {
		return
	}

	children := pie.Map(maps.Values(rootNode.Children), func(node *HierarchyNode) any {
		return node.Obj
	})

	category, categoryPresent := rootObject["category"].(string)
	if categoryPresent {
		entity := models.EntityStrToInt(category)
		layerFactories := models.LayersByEntity[entity]

		for _, factory := range layerFactories {
			for _, layer := range factory.FromObjects(children) {
				rootNode.AddChild(NewLayerNode(layer.Name, layer))
			}
		}
	}
}
