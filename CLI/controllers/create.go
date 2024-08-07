package controllers

import (
	"cli/models"
	"cli/utils"
	"fmt"
	"net/http"
	pathutil "path"
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
		// if true, dry run (no API requests)
		isValidate = validate[0]
	}

	// Object base data
	if err := models.SetObjectBaseData(ent, path, data); err != nil {
		return err
	}

	// Retrieve parent
	parentId, parent, err := controller.GetParentFromPath(pathutil.Dir(path), ent, isValidate)
	if err != nil {
		return err
	}
	if ent != models.SITE && ent != models.STRAY_DEV {
		data["parentId"] = parentId
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
	attr := data["attributes"].(map[string]any)
	switch ent {
	case models.BLDG, models.ROOM, models.RACK, models.CORRIDOR, models.GENERIC:
		utils.MergeMaps(attr, models.BaseAttrs[ent], false)
		if hasTemplate, err := controller.ApplyTemplateOrSetSize(attr, data, ent,
			isValidate); err != nil {
			return err
		} else if hasTemplate && isValidate {
			return nil
		}

		if err := models.SetPosAttr(ent, attr); err != nil {
			return err
		}
	case models.DEVICE:
		models.SetDeviceSizeUIfExists(attr)
		if err := models.SetDeviceSlotOrPosU(attr); err != nil {
			return err
		}

		if hasTemplate, err := controller.ApplyTemplateIfExists(attr, data, ent,
			isValidate); !hasTemplate {
			// apply user input
			setDevNoTemplateSize(attr, parent, isValidate)
		} else if err != nil {
			return err
		} else if isValidate {
			return nil
		}

		utils.MergeMaps(attr, models.DeviceBaseAttrs, false)
	case models.STRAY_DEV:
		if _, err := controller.ApplyTemplateIfExists(attr, data, ent,
			isValidate); err != nil {
			return err
		}
	default:
		break
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

func setDevNoTemplateSize(attr map[string]any, parent map[string]any, isValidate bool) error {
	var slot map[string]any
	var err error
	// get slot, if possible
	if slot, err = getSlotDevNoTemplate(attr, parent, isValidate); err != nil {
		return err
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

func getSlotDevNoTemplate(attr map[string]any, parent map[string]any, isValidate bool) (map[string]any, error) {
	// get slot (no template -> only one slot accepted)
	if attr["slot"] != nil {
		slots := attr["slot"].([]string)
		if len(slots) != 1 {
			return nil, fmt.Errorf("invalid device syntax: only one slot can be provided if no template")
		}
		if !isValidate {
			slot, err := C.GetSlot(parent, slots[0])
			if err != nil {
				return slot, err
			}
		}
	}
	return nil, nil
}
