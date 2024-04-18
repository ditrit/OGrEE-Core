package controllers

import (
	l "cli/logger"
	"cli/models"
	"fmt"
	"net/http"
	pathutil "path"
	"strconv"
	"strings"
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

func (controller Controller) CreateObject(path string, ent int, data map[string]any) error {
	var parent map[string]any

	name := pathutil.Base(path)
	path = pathutil.Dir(path)
	if name == "." || name == "" {
		l.GetWarningLogger().Println("Invalid path name provided for OCLI object creation")
		return fmt.Errorf("Invalid path name provided for OCLI object creation")
	}

	data["name"] = name
	data["category"] = models.EntityToString(ent)
	data["description"] = ""

	//Retrieve Parent
	if ent != models.SITE && ent != models.STRAY_DEV {
		var err error
		parent, err = controller.PollObject(path)
		if err != nil {
			return err
		}
		if parent == nil && (ent != models.DOMAIN || path != "/Organisation/Domain") {
			return fmt.Errorf("parent not found")
		}
	}

	if ent != models.DOMAIN {
		if parent != nil {
			data["domain"] = parent["domain"]
		} else {
			data["domain"] = State.Customer
		}
	}

	attr, hasAttributes := data["attributes"].(map[string]any)
	if !hasAttributes {
		attr = map[string]any{}
	}

	var err error
	switch ent {
	case models.DOMAIN:
		if parent != nil {
			data["parentId"] = parent["id"]
		} else {
			data["parentId"] = ""
		}

	case models.SITE:
		break

	case models.BLDG:
		//Check for template
		if _, ok := attr["template"]; ok {
			err := controller.ApplyTemplate(attr, data, models.BLDG)
			if err != nil {
				return err
			}
		} else {
			//Serialise size and posXY manually instead
			attr["size"] = serialiseVector(attr, "size")
			fmt.Println(attr)
		}

		if _, ok := attr["size"].([]any); !ok {
			if _, ok = attr["size"].([]float64); !ok {
				if State.DebugLvl > 0 {
					l.GetErrorLogger().Println(
						"User gave invalid size value for creating building")
					return fmt.Errorf("Invalid size attribute provided." +
						" \nIt must be an array/list/vector with 3 elements." +
						" Please refer to the wiki or manual reference" +
						" for more details on how to create objects " +
						"using this syntax")
				}
				return nil
			}
		}

		attr["posXY"] = serialiseVector(attr, "posXY")

		if posXY, ok := attr["posXY"].([]float64); !ok || len(posXY) != 2 {
			if State.DebugLvl > 0 {
				l.GetErrorLogger().Println(
					"User gave invalid posXY value for creating building")
				return fmt.Errorf("Invalid posXY attribute provided." +
					" \nIt must be an array/list/vector with 2 elements." +
					" Please refer to the wiki or manual reference" +
					" for more details on how to create objects " +
					"using this syntax")
			}
			return nil
		}

		attr["posXYUnit"] = "m"
		attr["sizeUnit"] = "m"
		attr["heightUnit"] = "m"
		//attr["height"] = 0 //Should be set from parser by default
		data["parentId"] = parent["id"]

	case models.ROOM:
		baseAttrs := map[string]any{
			"floorUnit":  "t",
			"posXYUnit":  "m",
			"sizeUnit":   "m",
			"heightUnit": "m",
		}

		MergeMaps(attr, baseAttrs, false)

		//If user provided templates, get the JSON
		//and parse into templates
		//NOTE this function also assigns value for "size" attribute
		err := controller.ApplyTemplate(attr, data, ent)
		if err != nil {
			return err
		}

		attr["posXY"] = serialiseVector(attr, "posXY")

		if posXY, ok := attr["posXY"].([]float64); !ok || len(posXY) != 2 {
			if State.DebugLvl > 0 {
				l.GetErrorLogger().Println(
					"User gave invalid posXY value for creating room")
				return fmt.Errorf("Invalid posXY attribute provided." +
					" \nIt must be an array/list/vector with 2 elements." +
					" Please refer to the wiki or manual reference" +
					" for more details on how to create objects " +
					"using this syntax")
			}
			return nil
		}

		if _, ok := attr["size"].([]any); !ok {
			if _, ok = attr["size"].([]float64); !ok {
				if State.DebugLvl > 0 {
					l.GetErrorLogger().Println(
						"User gave invalid size value for creating room")
					return fmt.Errorf("Invalid size attribute provided." +
						" \nIt must be an array/list/vector with 3 elements." +
						" Please refer to the wiki or manual reference" +
						" for more details on how to create objects " +
						"using this syntax")
				}
				return nil
			}
		}

		data["parentId"] = parent["id"]
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

		MergeMaps(attr, baseAttrs, false)

		//If user provided templates, get the JSON
		//and parse into templates
		err := controller.ApplyTemplate(attr, data, ent)
		if err != nil {
			return err
		}

		if _, ok := attr["size"].([]any); !ok {
			if _, ok = attr["size"].([]float64); !ok {
				if State.DebugLvl > 0 {
					l.GetErrorLogger().Println(
						"User gave invalid size value for creating rack")
					return fmt.Errorf("Invalid size attribute/template provided." +
						" \nThe size must be an array/list/vector with " +
						"3 elements." + "\n\nIf you have provided a" +
						" template, please check that you are referring to " +
						"an existing template" +
						"\n\nFor more information " +
						"please refer to the wiki or manual reference" +
						" for more details on how to create objects " +
						"using this syntax")
				}
				return nil
			}
		}

		//Serialise posXY if given
		attr["posXYZ"] = serialiseVector(attr, "posXYZ")

		data["parentId"] = parent["id"]

	case models.DEVICE:
		//Special routine to perform on device
		//based on if the parent has a "slot" attribute

		//First check if attr has only posU & sizeU
		//reject if true while also converting sizeU to string if numeric
		//if len(attr) == 2 {
		if sizeU, ok := attr["sizeU"]; ok {
			sizeUValid := checkNumeric(attr["sizeU"])

			if _, ok := attr["template"]; !ok && !sizeUValid {
				l.GetWarningLogger().Println("Invalid template / sizeU parameter provided for device ")
				return fmt.Errorf("Please provide a valid device template or sizeU")
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
				return fmt.Errorf("Invalid device syntax: If you have provided a template, it was not found")
			}
		}
		//}

		//Process the posU/slot attribute
		if x, ok := attr["posU/slot"].([]string); ok && len(x) > 0 {
			delete(attr, "posU/slot")
			if posU, err := strconv.Atoi(x[0]); len(x) == 1 && err == nil {
				attr["posU"] = posU
			} else {
				if slots, err := ExpandSlotVector(x); err != nil {
					return err
				} else {
					attr["slot"] = slots
				}
			}
		}

		//If user provided templates, get the JSON
		//and parse into templates
		if _, ok := attr["template"]; ok {
			err := controller.ApplyTemplate(attr, data, models.DEVICE)
			if err != nil {
				return err
			}
		} else {
			var slot map[string]any
			if attr["slot"] != nil {
				slots := attr["slot"].([]string)
				if len(slots) != 1 {
					return fmt.Errorf("Invalid device syntax: only one slot can be provided if no template")
				}
				slot, err = GetSlot(parent, slots[0])
				if err != nil {
					return err
				}
			}
			if slot != nil {
				size := slot["elemSize"].([]any)
				attr["size"] = []float64{size[0].(float64) / 10., size[1].(float64) / 10.}
			} else {
				if parAttr, ok := parent["attributes"].(map[string]interface{}); ok {
					if rackSize, ok := parAttr["size"]; ok {
						attr["size"] = rackSize
					}
				}
			}
		}
		//End of device special routine

		baseAttrs := map[string]interface{}{
			"orientation": "front",
			"sizeUnit":    "mm",
			"heightUnit":  "mm",
		}

		MergeMaps(attr, baseAttrs, false)

		data["parentId"] = parent["id"]

	case models.GROUP:
		data["parentId"] = parent["id"]

	case models.STRAY_DEV:
		if _, ok := attr["template"]; ok {
			err := controller.ApplyTemplate(attr, data, models.DEVICE)
			if err != nil {
				return err
			}
		}

	default:
		//Execution should not reach here!
		return fmt.Errorf("Invalid Object Specified!")
	}

	data["attributes"] = attr

	//Because we already stored the string conversion in category
	//we can do the conversion for templates here
	data["category"] = strings.Replace(data["category"].(string), "_", "-", 1)

	return controller.PostObj(ent, data["category"].(string), data, path)
}

func CreateTag(slug, color string) error {
	return C.PostObj(models.TAG, models.EntityToString(models.TAG), map[string]any{
		"slug":        slug,
		"description": slug, // the description is initially set with the value of the slug
		"color":       color,
	}, models.TagsPath+slug)
}
