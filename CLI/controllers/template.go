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

// If user provided templates, get the JSON
// and parse into templates
func (controller Controller) ApplyTemplate(attr, data map[string]interface{}, ent int) error {
	if templateName, hasTemplate := attr["template"].(string); hasTemplate {
		tmpl, err := controller.GetTemplate(templateName, ent)
		if err != nil {
			return err
		}

		key := determineStrKey(tmpl, []string{"sizeWDHmm", "sizeWDHm"})

		if sizeInf, hasSize := tmpl[key].([]any); hasSize && len(sizeInf) == 3 {
			attr["size"] = sizeInf[:2]
			attr["height"] = sizeInf[2]
			CopyAttr(attr, tmpl, "shape")

			if ent == models.DEVICE {
				attr["sizeUnit"] = "mm"
				attr["heightUnit"] = "mm"
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
				attr["sizeUnit"] = "m"
				attr["heightUnit"] = "m"

				//Copy additional Room specific attributes
				CopyAttr(attr, tmpl, "technicalArea")
				if _, ok := attr["technicalArea"]; ok {
					attr["technical"] = attr["technicalArea"]
					delete(attr, "technicalArea")
				}

				CopyAttr(attr, tmpl, "axisOrientation")

				CopyAttr(attr, tmpl, "reservedArea")
				if _, ok := attr["reservedArea"]; ok {
					attr["reserved"] = attr["reservedArea"]
					delete(attr, "reservedArea")
				}

				CopyAttr(attr, tmpl, "separators")
				CopyAttr(attr, tmpl, "pillars")
				CopyAttr(attr, tmpl, "floorUnit")
				CopyAttr(attr, tmpl, "tiles")
				CopyAttr(attr, tmpl, "rows")
				CopyAttr(attr, tmpl, "aisles")
				CopyAttr(attr, tmpl, "vertices")
				CopyAttr(attr, tmpl, "colors")
				CopyAttr(attr, tmpl, "tileAngle")

			} else if ent == models.BLDG {
				attr["sizeUnit"] = "m"
				attr["heightUnit"] = "m"

			} else {
				attr["sizeUnit"] = "mm"
				attr["heightUnit"] = "mm"
			}

			//Copy Description
			if _, ok := tmpl["description"]; ok {
				if descTable, ok := tmpl["description"].([]interface{}); ok {
					data["description"] = descTable[0]
					for _, desc := range descTable[1:] {
						data["description"] = data["description"].(string) + "\n" + desc.(string)
					}
				} else {
					data["description"] = tmpl["description"]
				}
			} else {
				data["description"] = ""
			}

			//fbxModel section
			if check := CopyAttr(attr, tmpl, "fbxModel"); !check {
				if ent != models.BLDG {
					attr["fbxModel"] = ""
				}
			}

			//Copy orientation if available
			CopyAttr(attr, tmpl, "orientation")

			//Merge attributes if available
			if tmplAttrsInf, ok := tmpl["attributes"]; ok {
				if tmplAttrs, ok := tmplAttrsInf.(map[string]interface{}); ok {
					MergeMaps(attr, tmplAttrs, false)
				}
			}
		} else {
			println("Warning, invalid size value in template.")
			return errors.New("invalid size vector on given template")
		}
	} else {
		//Serialise size and posXY if given
		attr["size"] = serialiseVector(attr, "size")
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
