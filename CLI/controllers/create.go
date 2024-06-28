package controllers

import (
	l "cli/logger"
	"cli/models"
	"cli/utils"
	"fmt"
	"net/http"
	pathutil "path"
	"strconv"
)

func (controller Controller) PostObj(ent int, entity string, data map[string]any, path string) error {
	resp, err := controller.API.Request(http.MethodPost, "/api/"+entity+"s", data, http.StatusCreated)
	if err != nil {
		return err
	}

	if entInt := models.EntityStrToInt(entity); models.EntityCreationMustBeInformed(ent) &&
		IsInObjForUnity(entity) && entInt != models.LAYER {
		createType := "create"
		Ogree3D.InformOptional("PostObj", entInt, map[string]any{"type": createType, "data": resp.Body["data"]})
	}

	if models.IsLayer(path) {
		// For layers, add the object to the hierarchy in order to be cached
		_, err = State.Hierarchy.AddObjectInPath(data, path)
		if err != nil {
			return err
		}
	}

	return nil
}

func (controller Controller) ValidateObj(ent int, entity string, data map[string]any, path string) error {
	fmt.Println(data)
	resp, err := controller.API.Request(http.MethodPost, "/api/validate/"+entity+"s", data, http.StatusOK)
	if err != nil {
		fmt.Println(err)
		fmt.Println("RESP:")
		fmt.Println(resp)
		return err
	}

	return nil
}

func (controller Controller) CreateObject(path string, ent int, data map[string]any, validate ...bool) error {
	isValidate := false
	if len(validate) > 0 {
		isValidate = validate[0]
	}
	name := pathutil.Base(path)
	path = pathutil.Dir(path)
	if name == "." || name == "" {
		l.GetWarningLogger().Println("Invalid path name provided for OCLI object creation")
		return fmt.Errorf("invalid path name provided for OCLI object creation")
	}
	data["name"] = name
	data["category"] = models.EntityToString(ent)
	data["description"] = ""

	//Retrieve Parent
	parentId, parent, err := controller.GetParentFromPath(path, ent, isValidate)
	if err != nil {
		return err
	}

	// Set domain
	if ent != models.DOMAIN {
		if parent == nil || isValidate {
			data["domain"] = State.Customer
		} else {
			data["domain"] = parent["domain"]
		}
	}

	// Attributes
	attr, hasAttributes := data["attributes"].(map[string]any)
	if !hasAttributes {
		attr = map[string]any{}
	}

	switch ent {
	case models.DOMAIN:
		data["parentId"] = parentId

	case models.SITE:
		break

	case models.BLDG:
		//Check for template
		if _, ok := attr["template"]; ok {
			if isValidate {
				return nil
			}
			err := controller.ApplyTemplate(attr, data, models.BLDG)
			if err != nil {
				return err
			}
		} else {
			//Serialise size and posXY manually instead
			attr["size"] = models.SerialiseVector(attr, "size")
		}

		if err := models.CheckSize(attr); err != nil {
			return err
		}

		if err := models.SetPosXY(attr); err != nil {
			return err
		}

		attr["posXYUnit"] = "m"
		attr["sizeUnit"] = "m"
		attr["heightUnit"] = "m"
		//attr["height"] = 0 //Should be set from parser by default
		data["parentId"] = parentId

	case models.ROOM:
		baseAttrs := map[string]any{
			"floorUnit":  "t",
			"posXYUnit":  "m",
			"sizeUnit":   "m",
			"heightUnit": "m",
		}

		utils.MergeMaps(attr, baseAttrs, false)

		//If user provided templates, get the JSON
		//and parse into templates
		//NOTE this function also assigns value for "size" attribute
		if _, ok := attr["template"]; ok && isValidate {
			return nil
		}
		err := controller.ApplyTemplate(attr, data, ent)
		if err != nil {
			return err
		}

		models.SetPosXY(attr)

		if err := models.CheckSize(attr); err != nil {
			return err
		}

		data["parentId"] = parentId
		if State.DebugLvl >= 3 {
			println("DEBUG VIEW THE JSON")
			Disp(data)
		}

	case models.RACK, models.CORRIDOR, models.GENERIC:
		baseAttrs := map[string]any{
			"sizeUnit":   "cm",
			"heightUnit": "U",
		}
		if ent == models.CORRIDOR || ent == models.GENERIC {
			baseAttrs["heightUnit"] = "cm"
		}

		utils.MergeMaps(attr, baseAttrs, false)

		//If user provided templates, get the JSON
		//and parse into templates
		if _, ok := attr["template"]; ok && isValidate {
			return nil
		}
		err := controller.ApplyTemplate(attr, data, ent)
		if err != nil {
			return err
		}

		if err := models.CheckSize(attr); err != nil {
			return err
		}

		//Serialise posXY if given
		attr["posXYZ"] = models.SerialiseVector(attr, "posXYZ")

		data["parentId"] = parentId

	case models.DEVICE:
		//Special routine to perform on device
		//based on if the parent has a "slot" attribute

		//First check if attr has only posU & sizeU
		//reject if true while also converting sizeU to string if numeric
		//if len(attr) == 2 {
		_, hasTemplate := attr["template"]
		if sizeU, ok := attr["sizeU"]; ok {
			sizeUValid := utils.IsNumeric(attr["sizeU"])

			if hasTemplate && isValidate {
				return nil
			}
			if !hasTemplate && !sizeUValid {
				l.GetWarningLogger().Println("Invalid template / sizeU parameter provided for device ")
				return fmt.Errorf("please provide a valid device template or sizeU")
			}

			//Convert block
			//And Set height
			if sizeUInt, ok := sizeU.(int); ok {
				attr["sizeU"] = sizeUInt
				attr["height"] = float64(sizeUInt) * 44.5
			} else if sizeUFloat, ok := sizeU.(float64); ok {
				attr["sizeU"] = sizeUFloat
				attr["height"] = sizeUFloat * 44.5
			}
			//End of convert block
			if _, ok := attr["slot"]; ok {
				l.GetWarningLogger().Println("Invalid device syntax encountered")
				return fmt.Errorf("invalid device syntax: If you have provided a template, it was not found")
			}
		}
		//}

		//Process the posU/slot attribute
		if x, ok := attr["posU/slot"].([]string); ok && len(x) > 0 {
			delete(attr, "posU/slot")
			if posU, err := strconv.Atoi(x[0]); len(x) == 1 && err == nil {
				attr["posU"] = posU
			} else {
				if slots, err := models.ExpandStrVector(x); err != nil {
					return err
				} else {
					attr["slot"] = slots
				}
			}
		}

		//If user provided templates, get the JSON
		//and parse into templates
		if hasTemplate {
			if isValidate {
				return nil
			}
			err := controller.ApplyTemplate(attr, data, models.DEVICE)
			if err != nil {
				return err
			}
		} else {
			setDeviceNoTemplateSlotSize(attr, parent, isValidate)
		}
		//End of device special routine

		baseAttrs := map[string]interface{}{
			"orientation": "front",
			"sizeUnit":    "mm",
			"heightUnit":  "mm",
		}

		utils.MergeMaps(attr, baseAttrs, false)

		data["parentId"] = parentId

	case models.GROUP:
		data["parentId"] = parentId

	case models.STRAY_DEV:
		if _, ok := attr["template"]; ok {
			if isValidate {
				return nil
			}
			err := controller.ApplyTemplate(attr, data, models.DEVICE)
			if err != nil {
				return err
			}
		}

	case models.VIRTUALOBJ:
		if parent != nil {
			data["parentId"] = parentId
		}
	default:
		//Execution should not reach here!
		return fmt.Errorf("invalid Object Specified!")
	}

	data["attributes"] = attr

	if isValidate {
		return controller.ValidateObj(ent, data["category"].(string), data, path)
	}
	return controller.PostObj(ent, data["category"].(string), data, path)
}

func (controller Controller) CreateTag(slug, color string) error {
	return controller.PostObj(models.TAG, models.EntityToString(models.TAG), map[string]any{
		"slug":        slug,
		"description": slug, // the description is initially set with the value of the slug
		"color":       color,
	}, models.TagsPath+slug)
}

func setDeviceNoTemplateSlotSize(attr map[string]any, parent map[string]any, isValidate bool) error {
	var slot map[string]any
	var err error
	// get slot (no template -> only one slot accepted)
	if attr["slot"] != nil {
		slots := attr["slot"].([]string)
		if len(slots) != 1 {
			return fmt.Errorf("Invalid device syntax: only one slot can be provided if no template")
		}
		if !isValidate {
			slot, err = C.GetSlot(parent, slots[0])
			if err != nil {
				return err
			}
		}
	}
	if slot != nil {
		// apply size from slot
		size := slot["elemSize"].([]any)
		attr["size"] = []float64{size[0].(float64) / 10., size[1].(float64) / 10.}
	} else {
		if isValidate {
			// apply random size to validate
			attr["size"] = []float64{10, 10, 10}
		} else if parAttr, ok := parent["attributes"].(map[string]interface{}); ok {
			if rackSize, ok := parAttr["size"]; ok {
				// apply size from rack
				attr["size"] = rackSize
			}
		}
	}
	return nil
}
