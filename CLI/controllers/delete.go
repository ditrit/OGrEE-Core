package controllers

import (
	"cli/models"
	"fmt"
	"net/http"
	"strings"
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
