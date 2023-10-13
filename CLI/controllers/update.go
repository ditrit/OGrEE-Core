package controllers

import (
	"cli/models"
	"net/http"
)

var Update UpdateController = &UpdateControllerImpl{
	GetController: Get,
	API:           API,
	Ogree3D:       Ogree3D,
}

type UpdateController interface {
	UpdateObj(path string, data map[string]any) (map[string]any, error)
}

type UpdateControllerImpl struct {
	GetController GetController
	API           APIPort
	Ogree3D       Ogree3DPort
}

func (controller UpdateControllerImpl) UpdateObj(path string, data map[string]any) (map[string]any, error) {
	attributes, hasAttributes := data["attributes"].(map[string]any)
	if hasAttributes {
		for key, val := range attributes {
			attributes[key] = Stringify(val)
		}
	}

	obj, err := controller.GetController.GetObject(path)
	if err != nil {
		return nil, err
	}

	category := ""
	if obj["category"] != nil {
		category = obj["category"].(string)
	}

	url, err := ObjectUrlWithEntity(path, 0, category)
	if err != nil {
		return nil, err
	}

	resp, err := controller.API.Request(http.MethodPatch, url, data, http.StatusOK)
	if err != nil {
		return nil, err
	}

	//Determine if Unity requires the message as
	//Interact or Modify
	entityType := EntityStrToInt(category)
	if models.IsTag(path, nil) {
		entityType = TAG
	}

	message := map[string]any{}
	var key string

	if entityType == ROOM && (data["tilesName"] != nil || data["tilesColor"] != nil) {
		println("Room modifier detected")
		Disp(data)

		//Get the interactive key
		key = determineStrKey(data, []string{"tilesName", "tilesColor"})

		message["type"] = "interact"
		message["data"] = map[string]any{
			"id":    obj["id"],
			"param": key,
			"value": data[key],
		}
	} else if entityType == RACK && data["U"] != nil {
		message["type"] = "interact"
		message["data"] = map[string]any{
			"id":    obj["id"],
			"param": "U",
			"value": data["U"],
		}
	} else if (entityType == DEVICE || entityType == RACK) &&
		(data["alpha"] != nil || data["slots"] != nil || data["localCS"] != nil) {

		//Get interactive key
		key = determineStrKey(data, []string{"alpha", "U", "slots", "localCS"})

		message["type"] = "interact"
		message["data"] = map[string]any{
			"id":    obj["id"],
			"param": key,
			"value": data[key],
		}
	} else if entityType == GROUP && data["content"] != nil {
		message["type"] = "interact"
		message["data"] = map[string]any{
			"id":    obj["id"],
			"param": "content",
			"value": data["content"],
		}
	} else if entityType == TAG {
		oldSlug, err := models.ObjectSlug(path)
		if err != nil {
			return nil, err
		}

		message["type"] = "modify-tag"
		message["data"] = map[string]any{
			"old-slug": oldSlug,
			"tag":      resp.Body["data"],
		}
	} else {
		message["type"] = "modify"
		message["data"] = resp.Body["data"]
	}

	if IsEntityTypeForOGrEE3D(entityType) {
		err := controller.Ogree3D.InformOptional("UpdateObj", entityType, message)
		if err != nil {
			return nil, err
		}
	}

	return resp.Body, nil
}
