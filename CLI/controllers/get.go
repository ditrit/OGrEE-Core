package controllers

import (
	"cli/models"
	"cli/utils"
	"cli/views"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

var ErrObjectNotFound = errors.New("object not found")

func (controller Controller) GetObject(path string) (map[string]any, error) {
	return controller.GetObjectWithChildren(path, 0)
}

func (controller Controller) GetObjectsWildcard(pathStr string, filters map[string]string, recursive *RecursiveParams) ([]map[string]any, []string, error) {
	url, err := controller.ObjectUrlGeneric(pathStr, 0, filters, recursive)
	if err != nil {
		return nil, nil, err
	}

	var resp *Response
	var method string
	if models.PathHasLayer(pathStr) {
		if filters == nil {
			filters = map[string]string{}
		}
		path, err := controller.SplitPath(pathStr)
		if err != nil {
			return nil, nil, err
		}
		if path.Layer != nil {
			path.Layer.ApplyFilters(filters)
		}
	}

	if complexFilter, ok := filters["filter"]; ok {
		body := map[string]any{"filter": complexFilter}
		method = "POST "
		resp, err = controller.API.Request(http.MethodPost, url, body, http.StatusOK)
	} else {
		method = "GET "
		resp, err = controller.API.Request(http.MethodGet, url, nil, http.StatusOK)
	}

	if err != nil {
		return nil, nil, err
	}
	return controller.ParseWildcardResponse(resp, pathStr, method+url)
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

	objs := utils.AnyArrToMapArr(objsAny)
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
		return nil, ErrObjectNotFound
	}

	return obj, nil
}

func (controller Controller) PollObject(path string) (map[string]any, error) {
	return controller.PollObjectWithChildren(path, 0)
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
		if resp != nil && resp.Status == http.StatusNotFound {
			return nil, nil
		}
		return nil, err
	}

	obj, ok := resp.Body["data"].(map[string]any)
	if !ok {
		if listObj, ok := resp.Body["data"].([]any); !ok || len(listObj) < 1 {
			return nil, fmt.Errorf("invalid response from API on GET %s", url)
		} else {
			obj = listObj[0].(map[string]any)
		}
	}

	return obj, nil
}

func (controller Controller) GetByAttr(path string, u interface{}) error {
	obj, err := controller.GetObjectWithChildren(path, 1)
	if err != nil {
		return err
	}
	cat := obj["category"].(string)
	if cat != "rack" {
		return fmt.Errorf("command may only be performed on rack objects")
	}
	children := obj["children"].([]any)
	devices := utils.AnyArrToMapArr(children)
	switch u.(type) {
	case int:
		for i := range devices {
			if attr, ok := devices[i]["attributes"].(map[string]interface{}); ok {
				uStr := strconv.Itoa(u.(int))
				if attr["height"] == uStr {
					views.DisplayJson("", devices[i])
					return nil //What if the user placed multiple devices at same height?
				}
			}
		}
		if State.DebugLvl > NONE {
			println("The 'U' you provided does not correspond to any device in this rack")
		}
	default: //String
		for i := range devices {
			if attr, ok := devices[i]["attributes"].(map[string]interface{}); ok {
				if attr["slot"] == u.(string) {
					views.DisplayJson("", devices[i])
					return nil //What if the user placed multiple devices at same slot?
				}
			}
		}
		if State.DebugLvl > NONE {
			println("The slot you provided does not correspond to any device in this rack")
		}
	}
	return nil
}

func (controller Controller) GetSlot(rack map[string]any, location string) (map[string]any, error) {
	templateAny, ok := rack["attributes"].(map[string]any)["template"]
	if !ok {
		return nil, nil
	}
	template := templateAny.(string)
	if template == "" {
		return nil, nil
	}
	resp, err := controller.API.Request("GET", "/api/obj_templates/"+template, nil, http.StatusOK)
	if err != nil {
		return nil, err
	}
	slots, ok := resp.Body["data"].(map[string]any)["slots"]
	if !ok {
		return nil, nil
	}
	for _, slotAny := range slots.([]any) {
		slot := slotAny.(map[string]any)
		if slot["location"] == location {
			return slot, nil
		}
	}
	return nil, fmt.Errorf("the slot %s does not exist", location)
}
