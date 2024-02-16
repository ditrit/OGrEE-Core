package controllers

import (
	"cli/models"
	"cli/utils"
	"net/http"
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
		body := utils.ComplexFilterToMap(filters["filter"])
		resp, err = controller.API.Request(http.MethodDelete, url, body, http.StatusOK)
	} else {
		resp, err = controller.API.Request(http.MethodDelete, url, nil, http.StatusOK)
	}
	if err != nil {
		return nil, err
	}
	objs, paths, err := controller.ParseWildcardResponse(resp, path, "DELETE "+url)
	if err != nil {
		return nil, err
	}
	for _, obj := range objs {
		if models.IsPhysical(path) && IsInObjForUnity(obj["category"].(string)) {
			controller.Ogree3D.InformOptional("DeleteObj", -1, map[string]any{"type": "delete", "data": obj["id"].(string)})
		} else if models.IsTag(path) && IsEntityTypeForOGrEE3D(models.TAG) {
			controller.Ogree3D.InformOptional("DeleteObj", -1, map[string]any{"type": "delete-tag", "data": obj["slug"].(string)})
		}
		if models.IsLayer(path) {
			State.Hierarchy.Children["Logical"].Children["Layers"].IsCached = false
			if IsEntityTypeForOGrEE3D(models.LAYER) {
				controller.Ogree3D.InformOptional("DeleteObj", -1, map[string]any{"type": "delete-layer", "data": obj["slug"].(string)})
			}
		}
	}
	if path == State.CurrPath {
		controller.CD(TranslatePath("..", false))
	}
	return paths, nil
}
