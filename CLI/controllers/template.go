package controllers

import (
	"cli/models"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
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
	case models.RACK, models.DEVICE:
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

// Used for importing data from templates
func attrSerialiser(someVal interface{}, idx string, ent int) string {
	if x, ok := someVal.(int); ok {
		if ent == models.DEVICE || ent == models.ROOM || ent == models.BLDG {
			return strconv.Itoa(x)
		}
		return strconv.Itoa(x / 10)
	} else if x, ok := someVal.(float64); ok {
		if ent == models.DEVICE || ent == models.ROOM || ent == models.BLDG {
			return strconv.FormatFloat(x, 'G', -1, 64)
		}
		return strconv.FormatFloat(x/10.0, 'G', -1, 64)
	} else {
		msg := "Warning: Invalid " + idx +
			" value detected in size." +
			" Resorting to default"
		println(msg)
		return "5"
	}
}

// If user provided templates, get the JSON
// and parse into templates
func (controller Controller) ApplyTemplate(attr, data map[string]interface{}, ent int) error {
	if templateName, hasTemplate := attr["template"].(string); hasTemplate {
		tmpl, err := controller.GetTemplate(templateName, ent)
		if err != nil {
			return err
		}

		//MergeMaps(attr, tmpl, true)
		key := determineStrKey(tmpl, []string{"sizeWDHmm", "sizeWDHm"})

		if sizeInf, ok := tmpl[key].([]interface{}); ok && len(sizeInf) == 3 {
			var xS, yS, zS string
			xS = attrSerialiser(sizeInf[0], "x", ent)
			yS = attrSerialiser(sizeInf[1], "y", ent)
			zS = attrSerialiser(sizeInf[2], "height", ent)

			attr["size"] = "{\"x\":" + xS + ", \"y\":" + yS + "}"
			attr["height"] = zS

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
										//Resort to default value
										msg := "Warning, invalid value provided for" +
											" sizeU. Defaulting to 5"
										println(msg)
										res = int((5 / 1000) / RACKUNIT)
									}
									attr["sizeU"] = strconv.Itoa(res)

								}
							}
						}
					}
				}

			} else if ent == models.ROOM {
				attr["sizeUnit"] = "m"
				attr["heightUnit"] = "m"

				//Copy additional Room specific attributes
				var tmp []byte
				CopyAttr(attr, tmpl, "technicalArea")
				if _, ok := attr["technicalArea"]; ok {
					//tmp, _ := json.Marshal(attr["technicalArea"])
					attr["technical"] = attr["technicalArea"]
					delete(attr, "technicalArea")
				}

				CopyAttr(attr, tmpl, "axisOrientation")

				CopyAttr(attr, tmpl, "reservedArea")
				if _, ok := attr["reservedArea"]; ok {
					//tmp, _ = json.Marshal(attr["reservedArea"])
					attr["reserved"] = attr["reservedArea"]
					delete(attr, "reservedArea")
				}

				parseReservedTech(attr)

				CopyAttr(attr, tmpl, "separators")
				if _, ok := attr["separators"]; ok {
					tmp, _ = json.Marshal(attr["separators"])
					attr["separators"] = string(tmp)
				}

				CopyAttr(attr, tmpl, "pillars")
				if _, ok := attr["pillars"]; ok {
					tmp, _ = json.Marshal(attr["pillars"])
					attr["pillars"] = string(tmp)
				}

				CopyAttr(attr, tmpl, "floorUnit")
				if _, ok := attr["floorUnit"]; ok {
					if floorUnit, ok := attr["floorUnit"].(string); ok {
						attr["floorUnit"] = floorUnit
					}
				}

				CopyAttr(attr, tmpl, "tiles")
				if _, ok := attr["tiles"]; ok {
					tmp, _ = json.Marshal(attr["tiles"])
					attr["tiles"] = string(tmp)
				}

				CopyAttr(attr, tmpl, "rows")
				if _, ok := attr["rows"]; ok {
					tmp, _ = json.Marshal(attr["rows"])
					attr["rows"] = string(tmp)
				}

				CopyAttr(attr, tmpl, "aisles")
				if _, ok := attr["aisles"]; ok {
					tmp, _ = json.Marshal(attr["aisles"])
					attr["aisles"] = string(tmp)
				}

				CopyAttr(attr, tmpl, "vertices")
				if _, ok := attr["vertices"]; ok {
					tmp, _ = json.Marshal(attr["vertices"])
					attr["vertices"] = string(tmp)
				}

				CopyAttr(attr, tmpl, "colors")
				if _, ok := attr["colors"]; ok {
					tmp, _ = json.Marshal(attr["colors"])
					attr["colors"] = string(tmp)
				}

				CopyAttr(attr, tmpl, "tileAngle")
				if _, ok := attr["tileAngle"]; ok {
					if tileAngle, ok := attr["tileAngle"].(int); ok {
						attr["tileAngle"] = strconv.Itoa(tileAngle)
					}

					if tileAngleF, ok := attr["tileAngle"].(float64); ok {
						tileAngleStr := strconv.FormatFloat(tileAngleF, 'f', -1, 64)
						attr["tileAngle"] = tileAngleStr
					}
				}

			} else if ent == models.BLDG {
				attr["sizeUnit"] = "m"
				attr["heightUnit"] = "m"

			} else {
				attr["sizeUnit"] = "cm"
				attr["heightUnit"] = "cm"
			}

			//Copy Description
			if _, ok := tmpl["description"]; ok {
				if descTable, ok := tmpl["description"].([]interface{}); ok {
					data["description"] = descTable
				} else {
					data["description"] = []interface{}{tmpl["description"]}
				}
			} else {
				data["description"] = []string{}
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
			if State.DebugLvl > 1 {
				println("Warning, invalid size value in template.",
					"Default values will be assigned")
			}
		}
	} else {
		if ent != models.CORRIDOR {
			attr["template"] = ""
		}
		//Serialise size and posXY if given
		serialiseVector(attr, "size")
	}

	return nil
}
