package controllers

import (
	"cli/models"
	"cli/utils"
	"fmt"
	"net/http"
	"strings"
)

var Get GetController = &GetControllerImpl{}

type GetController interface {
	GetObject(path string) (map[string]any, error)
	GetObjectsWildcard(path string) ([]map[string]any, []string, error)
}

type GetControllerImpl struct{}

func (controller GetControllerImpl) GetObject(path string) (map[string]any, error) {
	return GetObjectWithChildren(path, 0)
}

func (controller GetControllerImpl) GetObjectsWildcard(path string) ([]map[string]any, []string, error) {
	url, err := ObjectUrlGeneric(path, 0, nil)
	if err != nil {
		return nil, nil, err
	}
	resp, err := API.Request("GET", url, nil, http.StatusOK)
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
