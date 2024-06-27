package controllers

import (
	l "cli/logger"
	"cli/models"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/exp/slices"
)

func (controller Controller) DeleteObj(path string) ([]string, error) {
	url, err := controller.ObjectUrlGeneric(path, 0, nil, nil)
	if err != nil {
		return nil, err
	}

	var resp *Response
	if models.PathHasLayer(path) {
		filters := map[string]string{}
		pathSplit, err := controller.SplitPath(path)
		if err != nil {
			return nil, err
		}
		if pathSplit.Layer != nil {
			pathSplit.Layer.ApplyFilters(filters)
		}
		body := map[string]any{"filter": filters["filter"]}
		resp, err = controller.API.Request(http.MethodDelete, url, body, http.StatusOK)
	} else {
		resp, err = controller.API.Request(http.MethodDelete, url, nil, http.StatusOK)
	}
	if err != nil {
		return nil, err
	}
	_, paths, err := controller.ParseWildcardResponse(resp, path, "DELETE "+url)
	if err != nil {
		return nil, err
	}
	if models.IsLayer(path) {
		State.Hierarchy.Children["Logical"].Children["Layers"].IsCached = false
	}
	if path == State.CurrPath {
		controller.CD(TranslatePath("..", false))
	}
	return paths, nil
}

func (controller Controller) UnsetAttribute(path string, attr string) error {
	obj, err := controller.GetObject(path)
	if err != nil {
		return err
	}
	delete(obj, "id")
	delete(obj, "lastUpdated")
	delete(obj, "createdDate")
	attributes, hasAttributes := obj["attributes"].(map[string]any)
	if !hasAttributes {
		return fmt.Errorf("object has no attributes")
	}
	if vconfigAttr, found := strings.CutPrefix(attr, VIRTUALCONFIG+"."); found {
		if len(vconfigAttr) < 1 {
			return fmt.Errorf("invalid attribute name")
		} else if vAttrs, ok := attributes[VIRTUALCONFIG].(map[string]any); !ok {
			return fmt.Errorf("object has no " + VIRTUALCONFIG)
		} else {
			delete(vAttrs, vconfigAttr)
		}
	} else {
		delete(attributes, attr)
	}
	url, err := controller.ObjectUrl(path, 0)
	if err != nil {
		return err
	}
	_, err = controller.API.Request("PUT", url, obj, http.StatusOK)
	return err
}

// Specific update for deleting elements in an array of an obj
func (controller Controller) UnsetInObj(Path, attr string, idx int) (map[string]interface{}, error) {
	var arr []interface{}

	//Check for valid idx
	if idx < 0 {
		return nil,
			fmt.Errorf("Index out of bounds. Please provide an index greater than 0")
	}

	//Get the object
	obj, err := controller.GetObject(Path)
	if err != nil {
		return nil, err
	}

	//Check if attribute exists in object
	existing, nested := AttrIsInObj(obj, attr)
	if !existing {
		if State.DebugLvl > ERROR {
			l.GetErrorLogger().Println("Attribute :" + attr + " was not found")
		}
		return nil, fmt.Errorf("Attribute :" + attr + " was not found")
	}

	//Check if attribute is an array
	if nested {
		objAttributes := obj["attributes"].(map[string]interface{})
		if _, ok := objAttributes[attr].([]interface{}); !ok {
			if State.DebugLvl > ERROR {
				println("Attribute is not an array")
			}
			return nil, fmt.Errorf("Attribute is not an array")

		}
		arr = objAttributes[attr].([]interface{})

	} else {
		if _, ok := obj[attr].([]interface{}); !ok {
			if State.DebugLvl > ERROR {
				l.GetErrorLogger().Println("Attribute :" + attr + " was not found")
			}
			return nil, fmt.Errorf("Attribute :" + attr + " was not found")
		}
		arr = obj[attr].([]interface{})
	}

	//Ensure that we can delete elt in array
	if len(arr) == 0 {
		if State.DebugLvl > ERROR {
			println("Cannot delete anymore elements")
		}
		return nil, fmt.Errorf("Cannot delete anymore elements")
	}

	//Perform delete
	if idx >= len(arr) {
		idx = len(arr) - 1
	}
	arr = slices.Delete(arr, idx, idx+1)

	//Save back into obj
	if nested {
		obj["attributes"].(map[string]interface{})[attr] = arr
	} else {
		obj[attr] = arr
	}

	URL, err := controller.ObjectUrl(Path, 0)
	if err != nil {
		return nil, err
	}

	_, err = controller.API.Request("PUT", URL, obj, http.StatusOK)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
