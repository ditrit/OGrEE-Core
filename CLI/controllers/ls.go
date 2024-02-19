package controllers

import (
	"cli/models"
	"cli/utils"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/elliotchance/pie/v2"
	"golang.org/x/exp/maps"
)

var errLayerNotFound = errors.New("the layer used does not exist")
var ErrRecursiveOnlyFiltersLayers = errors.New("-r can only be applied to ls with filters or layers")

type RecursiveParams struct {
	PathEntered string
	MinDepth    int
	MaxDepth    int
}

func (controller Controller) Ls(path string, filters map[string]string, recursive *RecursiveParams) ([]map[string]any, error) {
	if len(filters) == 0 && !models.PathIsLayer(path) {
		if recursive != nil {
			return nil, ErrRecursiveOnlyFiltersLayers
		}
		return controller.lsObjectsWithoutFilters(path)
	}

	return controller.lsObjectsWithFilters(path, filters, recursive)
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

func (controller Controller) lsObjectsWithFilters(path string, filters map[string]string, recursive *RecursiveParams) ([]map[string]any, error) {
	url, err := controller.ObjectUrlGeneric(path+"/*", 0, filters, recursive)
	if err != nil {
		if errors.Is(err, errLayerNotFound) || errors.Is(err, models.ErrMaxLessMin) {
			return nil, err
		}

		return nil, fmt.Errorf("cannot use filters at this location")
	}

	var resp *Response
	if complexFilter, ok := filters["filter"]; ok {
		body := utils.ComplexFilterToMap(complexFilter)
		resp, err = controller.API.Request(http.MethodPost, url, body, http.StatusOK)
	} else {
		resp, err = controller.API.Request(http.MethodGet, url, map[string]any{}, http.StatusOK)
	}

	if err != nil {
		return nil, err
	}

	objectsAny := resp.Body["data"].([]any)
	objects := []map[string]any{}

	for _, objAny := range objectsAny {
		obj, ok := objAny.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid response from API on POST %s", url)
		}

		objects = append(objects, obj)
	}

	return objects, nil
}

// Obtains a HierarchyNode using Tree and adds the layers
func (controller Controller) lsNode(path string) (*HierarchyNode, error) {
	n, err := controller.Tree(path, 1)
	if err != nil {
		return nil, err
	}

	err = controller.addLayers(path, n)
	if err != nil {
		return nil, err
	}

	return n, nil
}

func (controller Controller) addLayers(path string, rootNode *HierarchyNode) error {
	addAutomaticLayers(rootNode)
	return controller.addUserDefinedLayers(path, rootNode)
}

func (controller Controller) addUserDefinedLayers(path string, rootNode *HierarchyNode) error {
	if models.IsPhysical(path) && !models.PathIsLayer(path) {
		layersNode, err := controller.Tree(models.LayersPath, 1)
		if err != nil {
			return err
		}

		for _, layerNode := range layersNode.Children {
			layer := layerNode.Obj.(models.UserDefinedLayer)
			if layer.Matches(path) {
				// layer in hierarchy has a pointer to the layer stored in /Logical/Layers
				layerInHierarchyNode := NewLayerNode(layer.Name(), &layerNode.Obj)
				rootNode.AddChild(layerInHierarchyNode)
			}
		}
	}

	return nil
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
				rootNode.AddChild(NewLayerNode(layer.Name(), layer))
			}
		}
	}
}
