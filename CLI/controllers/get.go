package controllers

import (
	"cli/models"
	"cli/utils"
	"fmt"
	"net/http"
	"strings"
)

func (controller Controller) GetObject(path string) (map[string]any, error) {
	return controller.GetObjectWithChildren(path, 0)
}

func (controller Controller) GetObjectsWildcard(path string) ([]map[string]any, []string, error) {
	url, err := ObjectUrlGeneric(path, 0, nil)
	if err != nil {
		return nil, nil, err
	}
	resp, err := controller.API.Request("GET", url, nil, http.StatusOK)
	if err != nil {
		return nil, nil, err
	}
	objsAny, ok := resp.Body["data"].([]any)
	if !ok {
		return nil, nil, fmt.Errorf("invalid response from API on GET %s", url)
	}
	prefix, _, _ := models.SplitPath(path)
	objs := infArrToMapStrinfArr(objsAny)
	paths := []string{}
	for _, obj := range objs {
		var suffix string
		objId, hasId := obj["id"].(string)
		if hasId {
			suffix = strings.Replace(objId, ".", "/", -1)
		} else {
			suffix = utils.NameOrSlug(obj)
		}
		objPath := prefix + suffix
		paths = append(paths, objPath)
	}
	return objs, paths, nil
}

func (controller Controller) GetObjectWithChildren(path string, depth int) (map[string]any, error) {
	obj, err := controller.PollObjectWithChildren(path, depth)
	if err != nil {
		return nil, err
	}

	if obj == nil {
		return nil, fmt.Errorf("object not found")
	}

	return obj, nil
}

func (controller Controller) PollObjectWithChildren(path string, depth int) (map[string]any, error) {
	url, err := ObjectUrl(path, depth)
	if err != nil {
		return nil, nil
	}
	resp, err := controller.API.Request(http.MethodGet, url, nil, http.StatusOK)
	if err != nil {
		if resp != nil && resp.status == http.StatusNotFound {
			return nil, nil
		}
		return nil, err
	}

	obj, ok := resp.Body["data"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid response from API on GET %s", url)
	}

	return obj, nil
}
