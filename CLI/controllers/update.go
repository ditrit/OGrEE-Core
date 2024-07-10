package controllers

import (
	"cli/models"
	"net/http"
)

func (controller Controller) UpdateObj(pathStr string, data map[string]any, withRecursive bool) (map[string]any, error) {
	obj, err := controller.GetObject(pathStr)
	if err != nil {
		return nil, err
	}

	category := ""
	if obj["category"] != nil {
		category = obj["category"].(string)
	}

	if category == models.EntityToString(models.DEVICE) {
		models.ComputeSizeUAndHeight(obj, data)
	}

	url, err := controller.ObjectUrl(pathStr, 0)
	if err != nil {
		return nil, err
	}
	if withRecursive {
		url = url + "?recursive=true"
	}

	resp, err := controller.API.Request(http.MethodPatch, url, data, http.StatusOK)
	if err != nil {
		return nil, err
	}

	if models.IsLayer(pathStr) {
		// For layers, update the object to the hierarchy in order to be cached
		data := resp.Body["data"].(map[string]any)
		_, err = State.Hierarchy.AddObjectInPath(data, pathStr)
		if err != nil {
			return nil, err
		}
	}

	return resp.Body, nil
}
