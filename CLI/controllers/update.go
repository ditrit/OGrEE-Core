package controllers

import (
	"cli/models"
	"net/http"
)

func (controller Controller) UpdateObj(pathStr string, data map[string]any) (map[string]any, error) {
	attributes, hasAttributes := data["attributes"].(map[string]any)
	if hasAttributes {
		for key, val := range attributes {
			attributes[key] = Stringify(val)
		}
	}

	obj, err := controller.GetObject(pathStr)
	if err != nil {
		return nil, err
	}

	category := ""
	if obj["category"] != nil {
		category = obj["category"].(string)
	}

	url, err := controller.ObjectUrl(pathStr, 0)
	if err != nil {
		return nil, err
	}

	resp, err := controller.API.Request(http.MethodPatch, url, data, http.StatusOK)
	if err != nil {
		return nil, err
	}

	//Determine if Unity requires the message as
	//Interact or Modify
	entityType := models.EntityStrToInt(category)
	if models.IsTag(pathStr) {
		entityType = models.TAG
	} else if models.IsLayer(pathStr) {
		// For layers, update the object to the hierarchy in order to be cached
		data := resp.Body["data"].(map[string]any)
		_, err = State.Hierarchy.AddObjectInPath(data, pathStr)
		if err != nil {
			return nil, err
		}
		if len(data["filters"].(map[string]any)) == 0 {
			println("Attention: this layer is never shown as an option since it has no filters")
		}
	}

	message := map[string]any{}
	var key string

	if entityType == models.ROOM && (data["tilesName"] != nil || data["tilesColor"] != nil) {
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
	} else if entityType == models.RACK && data["U"] != nil {
		message["type"] = "interact"
		message["data"] = map[string]any{
			"id":    obj["id"],
			"param": "U",
			"value": data["U"],
		}
	} else if (entityType == models.DEVICE || entityType == models.RACK) &&
		(data["alpha"] != nil || data["slots"] != nil || data["localCS"] != nil) {

		//Get interactive key
		key = determineStrKey(data, []string{"alpha", "U", "slots", "localCS"})

		message["type"] = "interact"
		message["data"] = map[string]any{
			"id":    obj["id"],
			"param": key,
			"value": data[key],
		}
	} else if entityType == models.GROUP && data["content"] != nil {
		message["type"] = "interact"
		message["data"] = map[string]any{
			"id":    obj["id"],
			"param": "content",
			"value": data["content"],
		}
	} else if entityType == models.TAG {
		path, err := controller.SplitPath(pathStr)
		if err != nil {
			return nil, err
		}

		message["type"] = "modify-tag"
		message["data"] = map[string]any{
			"old-slug": path.ObjectID,
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
