package controllers

import (
	"cli/models"
	"net/http"
)

func (controller Controller) DeleteObj(path string) ([]string, error) {
	url, err := controller.ObjectUrlGeneric(path, 0, nil, nil)
	if err != nil {
		return nil, err
	}
	resp, err := controller.API.Request(http.MethodDelete, url, nil, http.StatusOK)
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
