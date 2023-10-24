package controllers

import (
	"cli/utils"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

func Ls(path string, filters map[string]string, sortAttr string) ([]map[string]any, error) {
	var objects []map[string]any
	var err error

	if len(filters) == 0 {
		objects, err = lsObjectsWithoutFilters(path)
	} else {
		objects, err = lsObjectsWithFilters(path, filters)
	}

	if err != nil {
		return nil, err
	}

	if sortAttr != "" {
		objects = filterObjectsWithoutAttr(objects, sortAttr)
		if !objectsAreSortable(objects, sortAttr) {
			return nil, fmt.Errorf("objects cannot be sorted according to this attribute")
		}
	}

	less := func(i, j int) bool {
		if sortAttr != "" {
			vali, _ := utils.ObjectAttr(objects[i], sortAttr)
			valj, _ := utils.ObjectAttr(objects[j], sortAttr)
			res, _ := utils.CompareVals(vali, valj)
			return res
		}
		return utils.NameOrSlug(objects[i]) < utils.NameOrSlug(objects[j])
	}

	sort.Slice(objects, less)
	return objects, nil
}

func lsObjectsWithoutFilters(path string) ([]map[string]any, error) {
	n, err := Tree(path, 1)
	if err != nil {
		return nil, err
	}
	objects := []map[string]any{}
	for _, child := range n.Children {
		if child.Obj != nil {
			if strings.HasPrefix(path, "/Logical/Groups") {
				child.Obj["name"] = strings.ReplaceAll(child.Obj["id"].(string), ".", "/")
			}
			objects = append(objects, child.Obj)
		} else {
			objects = append(objects, map[string]any{"name": child.Name})
		}
	}
	return objects, nil
}

func lsObjectsWithFilters(path string, filters map[string]string) ([]map[string]any, error) {
	url, err := ObjectUrlGeneric(path+"/*", 0, filters)
	if err != nil {
		return nil, fmt.Errorf("cannot use filters at this location")
	}
	resp, err := API.Request("GET", url, nil, http.StatusOK)
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

func filterObjectsWithoutAttr(objects []map[string]any, attr string) []map[string]any {
	remainingObjects := []map[string]any{}
	for _, obj := range objects {
		_, hasAttr := utils.ObjectAttr(obj, attr)
		if hasAttr {
			remainingObjects = append(remainingObjects, obj)
		}
	}
	return remainingObjects
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
