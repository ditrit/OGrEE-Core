package controllers

import (
	"cli/models"
	"errors"
	"fmt"
	"net/http"
)

// GetTemplate gets a template for "entity" with "name".
//
// Returns error in case the template doesn't exist or the template
// category is not the correct for the entity
func (controller Controller) GetTemplate(name string, entity int) (map[string]any, error) {
	var location string

	switch entity {
	case models.BLDG:
		location = models.BuildingTemplatesPath + name
	case models.ROOM:
		location = models.RoomTemplatesPath + name
	case models.RACK, models.DEVICE, models.GENERIC:
		location = models.ObjectTemplatesPath + name
	default:
		return nil, fmt.Errorf("templates are not applicable to %s", models.EntityToString(entity))
	}

	node, err := controller.Tree(location, 0)
	if err != nil {
		if errors.Is(err, ErrObjectNotFound) {
			return nil, errors.New("template not found")
		}

		return nil, err
	}

	template := node.Obj.(map[string]any)

	if entity == models.BLDG || entity == models.ROOM {
		return template, nil
	}

	category := template["category"].(string)

	if category != models.EntityToString(entity) {
		return nil, fmt.Errorf("template of category %s is not applicable to %s", category, models.EntityToString(entity))
	}

	return template, nil
}

func (controller Controller) ApplyTemplateIfExists(attr, data map[string]any, ent int, isValidate bool) (bool, error) {
	if _, hasTemplate := attr["template"]; hasTemplate {
		if isValidate {
			return true, nil
		}
		// apply template
		return true, controller.ApplyTemplate(attr, data, ent)
	}
	return false, nil
}

func (controller Controller) ApplyTemplateOrSetSize(attr, data map[string]any, ent int, isValidate bool) (bool, error) {
	if hasTemplate, err := controller.ApplyTemplateIfExists(attr, data, ent,
		isValidate); !hasTemplate {
		// apply user input
		return hasTemplate, models.SetSize(attr)
	} else {
		return hasTemplate, err
	}
}

// If user provided templates, get the JSON
// and parse into templates
func (controller Controller) ApplyTemplate(attr, data map[string]interface{}, ent int) error {
	tmpl, err := controller.GetTemplate(attr["template"].(string), ent)
	if err != nil {
		return err
	}

	if err := models.ApplyTemplateToObj(attr, data, tmpl, ent); err != nil {
		return err
	}

	return nil
}

func (controller Controller) LoadTemplate(data map[string]interface{}) error {
	var URL string
	if cat := data["category"]; cat == "room" {
		//Room template
		URL = "/api/room_templates"
	} else if cat == "bldg" || cat == "building" {
		//Bldg template
		URL = "/api/bldg_templates"
	} else if cat == "rack" || cat == "device" || cat == "generic" {
		// Obj template
		URL = "/api/obj_templates"
	} else {
		return fmt.Errorf("this template does not have a valid category. Please add a category attribute with a value of building, room, rack, device or generic")
	}

	_, err := controller.API.Request(http.MethodPost, URL, data, http.StatusCreated)

	return err
}
