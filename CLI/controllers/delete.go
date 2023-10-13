package controllers

import (
	"cli/models"
	"net/http"
)

func (controller Controller) DeleteObj(path string) error {
	objs, _, err := controller.GetObjectsWildcard(path)
	if err != nil {
		return err
	}
	url, err := ObjectUrlGeneric(path, 0, nil)
	if err != nil {
		return err
	}
	_, err = controller.API.Request(http.MethodDelete, url, nil, http.StatusOK)
	if err != nil {
		return err
	}
	for _, obj := range objs {
		if models.IsHierarchical(path) && IsInObjForUnity(obj["category"].(string)) {
			controller.Ogree3D.InformOptional("DeleteObj", -1, map[string]any{"type": "delete", "data": obj["id"].(string)})
		} else if models.IsTag(path) && IsEntityTypeForOGrEE3D(TAG) {
			controller.Ogree3D.InformOptional("DeleteObj", -1, map[string]any{"type": "delete-tag", "data": obj["slug"].(string)})
		}
	}
	if path == State.CurrPath {
		CD(TranslatePath(".."))
	}
	return nil
}
