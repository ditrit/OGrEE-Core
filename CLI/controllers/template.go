package controllers

import (
	"cli/models"
	"cli/utils"
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

	key := determineStrKey(tmpl, []string{"sizeWDHmm", "sizeWDHm"})

	if sizeInf, hasSize := tmpl[key].([]any); hasSize && len(sizeInf) == 3 {
		attr["size"] = sizeInf[:2]
		attr["height"] = sizeInf[2]
		utils.CopyMapVal(attr, tmpl, "shape")

		if ent == models.DEVICE {
			if tmpx, ok := tmpl["attributes"]; ok {
				if x, ok := tmpx.(map[string]interface{}); ok {
					if tmp, ok := x["type"]; ok {
						if t, ok := tmp.(string); ok {
							if t == "chassis" || t == "server" {
								res := 0
								if val, ok := sizeInf[2].(float64); ok {
									res = int((val / 1000) / RACKUNIT)
								} else if val, ok := sizeInf[2].(int); ok {
									res = int((float64(val) / 1000) / RACKUNIT)
								} else {
									return errors.New("invalid size vector on given template")
								}
								attr["sizeU"] = res
							}
						}
					}
				}
			}

		} else if ent == models.ROOM {
			//Copy additional Room specific attributes
			utils.CopyMapVal(attr, tmpl, "technicalArea")
			if _, ok := attr["technicalArea"]; ok {
				attr["technical"] = attr["technicalArea"]
				delete(attr, "technicalArea")
			}

			utils.CopyMapVal(attr, tmpl, "reservedArea")
			if _, ok := attr["reservedArea"]; ok {
				attr["reserved"] = attr["reservedArea"]
				delete(attr, "reservedArea")
			}

			for _, attrName := range []string{"axisOrientation", "separators",
				"pillars", "floorUnit", "tiles", "rows", "aisles",
				"vertices", "colors", "tileAngle"} {
				utils.CopyMapVal(attr, tmpl, attrName)
			}

		} else {
			attr["sizeUnit"] = "mm"
			attr["heightUnit"] = "mm"
		}

		//Copy Description
		if _, ok := tmpl["description"]; ok {
			data["description"] = tmpl["description"]
		}

		//fbxModel section
		if check := utils.CopyMapVal(attr, tmpl, "fbxModel"); !check {
			if ent != models.BLDG {
				attr["fbxModel"] = ""
			}
		}

		//Copy orientation if available
		utils.CopyMapVal(attr, tmpl, "orientation")

		//Merge attributes if available
		if tmplAttrsInf, ok := tmpl["attributes"]; ok {
			if tmplAttrs, ok := tmplAttrsInf.(map[string]interface{}); ok {
				utils.MergeMaps(attr, tmplAttrs, false)
			}
		}
	} else {
		println("Warning, invalid size value in template.")
		return errors.New("invalid size vector on given template")
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
