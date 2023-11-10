package controllers

import (
	"cli/models"
	"cli/utils"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var errRecursiveOnlyToLayers = errors.New("recursive (-r) is only applicable to layers")

func (controller Controller) GetObject(path string) (map[string]any, error) {
	return controller.GetObjectWithChildren(path, 0)
}

func (controller Controller) GetObjectsWildcard(pathStr string, recursive bool) ([]map[string]any, []string, error) {
	if recursive && !models.PathIsLayer(pathStr) {
		return nil, nil, errRecursiveOnlyToLayers
	}

	url, err := controller.ObjectUrlGeneric(pathStr, 0, nil, recursive)
	if err != nil {
		return nil, nil, err
	}
	resp, err := controller.API.Request(http.MethodGet, url, nil, http.StatusOK)
	if err != nil {
		return nil, nil, err
	}
	return controller.ParseWildcardResponse(resp, pathStr, "GET "+url)
}

func (controller Controller) ParseWildcardResponse(resp *Response, pathStr string, route string) ([]map[string]any, []string, error) {
	objsAny, ok := resp.Body["data"].([]any)
	if !ok {
		return nil, nil, fmt.Errorf("invalid response from API on %s", route)
	}

	path, err := controller.SplitPath(pathStr)
	if err != nil {
		return nil, nil, err
	}

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
		objPath := path.Prefix + suffix
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
	url, err := controller.ObjectUrl(path, depth)
	if err != nil {
		if errors.Is(err, errLayerNotFound) {
			return nil, err
		}

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
