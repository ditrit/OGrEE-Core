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

	if models.EntityCreationMustBeInformed(ent) && IsInObjForUnity(entity) {
		entInt := models.EntityStrToInt(entity)
		Ogree3D.InformOptional("PostObj", entInt, map[string]any{"type": "create", "data": resp.Body["data"]})
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

func CreateObject(path string, ent int, data map[string]interface{}) error {
	var attr map[string]interface{}
	var parent map[string]interface{}

	name := pathutil.Base(path)
	path = pathutil.Dir(path)
	if name == "." || name == "" {
		l.GetWarningLogger().Println("Invalid path name provided for OCLI object creation")
		return fmt.Errorf("Invalid path name provided for OCLI object creation")
	}

	data["name"] = name
	data["category"] = models.EntityToString(ent)
	data["description"] = []interface{}{}

	//Retrieve Parent
	if ent != models.SITE && ent != models.STRAY_DEV {
		var err error
		parent, err = PollObject(path)
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

	var err error
	switch ent {
	case models.DOMAIN:
		if parent != nil {
			data["parentId"] = parent["id"]
		} else {
			data["parentId"] = ""
		}

	case models.SITE:
		//Default values
		//data["parentId"] = parent["id"]
		data["attributes"] = map[string]interface{}{}

	case models.BLDG:
		attr = data["attributes"].(map[string]interface{})

		//Check for template
		if _, ok := attr["template"]; ok {
			err := GetOCLIAtributesTemplateHelper(attr, data, models.BLDG)
			if err != nil {
				return err
			}
		} else {
			//Serialise size and posXY manually instead
			serialiseVector(attr, "size")

			//Since template was not provided, set it empty
			attr["template"] = ""
		}

		if attr["size"] == "" {
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

		serialiseVector(attr, "posXY")

		if attr["posXY"] == "" {
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

		//Check rotation
		if _, ok := attr["rotation"].(float64); ok {
			attr["rotation"] =
				strconv.FormatFloat(attr["rotation"].(float64), 'f', -1, 64)
		}

		attr["posXYUnit"] = "m"
		attr["sizeUnit"] = "m"
		attr["heightUnit"] = "m"
		//attr["height"] = 0 //Should be set from parser by default
		data["parentId"] = parent["id"]

	case models.ROOM:
		attr = data["attributes"].(map[string]interface{})

		baseAttrs := map[string]interface{}{
			"floorUnit": "t",
			"posXYUnit": "m", "sizeUnit": "m",
			"heightUnit": "m"}

		MergeMaps(attr, baseAttrs, false)

		//If user provided templates, get the JSON
		//and parse into templates
		//NOTE this function also assigns value for "size" attribute
		err := GetOCLIAtributesTemplateHelper(attr, data, ent)
		if err != nil {
			return err
		}

		serialiseVector(attr, "posXY")

		if attr["posXY"] == "" {
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

		//Check rotation
		if _, ok := attr["rotation"].(float64); ok {
			attr["rotation"] =
				strconv.FormatFloat(attr["rotation"].(float64), 'f', -1, 64)
		}

		if attr["size"] == "" {
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

		data["parentId"] = parent["id"]
		data["attributes"] = attr
		if State.DebugLvl >= 3 {
			println("DEBUG VIEW THE JSON")
			Disp(data)
		}

	case models.RACK, models.CORRIDOR:
		attr = data["attributes"].(map[string]interface{})
		//Save rotation because it gets overwritten by
		//GetOCLIAtrributesTemplateHelper()
		rotation := attr["rotation"].([]float64)

		baseAttrs := map[string]interface{}{
			"sizeUnit":   "cm",
			"heightUnit": "U",
		}
		if ent == models.CORRIDOR {
			baseAttrs["heightUnit"] = "cm"
		}

		MergeMaps(attr, baseAttrs, false)

		//If user provided templates, get the JSON
		//and parse into templates
		err := GetOCLIAtributesTemplateHelper(attr, data, ent)
		if err != nil {
			return err
		}

		if attr["size"] == "" {
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

		//Serialise posXY if given
		serialiseVector(attr, "posXYZ")

		//Restore the rotation overwritten
		//by the helper func
		attr["rotation"] = fmt.Sprintf("{\"x\":%v, \"y\":%v, \"z\":%v}", rotation[0], rotation[1], rotation[2])

		if attr["posXYZ"] == "" {
			if State.DebugLvl > 0 {
				l.GetErrorLogger().Println(
					"User gave invalid posXYZ value for creating rack")
				return fmt.Errorf("Invalid posXYZ attribute provided." +
					" \nIt must be an array/list/vector with 2 or 3 elements." +
					" Please refer to the wiki or manual reference" +
					" for more details on how to create objects " +
					"using this syntax")
			}
			return nil
		}

		data["parentId"] = parent["id"]
		data["attributes"] = attr

	case models.DEVICE:
		attr = data["attributes"].(map[string]interface{})

		//Special routine to perform on device
		//based on if the parent has a "slot" attribute

		//First check if attr has only posU & sizeU
		//reject if true while also converting sizeU to string if numeric
		//if len(attr) == 2 {
		if sizeU, ok := attr["sizeU"]; ok {
			sizeUValid := checkNumeric(attr["sizeU"])

			if _, ok := attr["template"]; !ok && sizeUValid == false {
				l.GetWarningLogger().Println("Invalid template / sizeU parameter provided for device ")
				return fmt.Errorf("Please provide a valid device template or sizeU")
			}

			//Convert block
			//And Set height
			if _, ok := sizeU.(int); ok {
				attr["sizeU"] = strconv.Itoa(sizeU.(int))
				attr["height"] = strconv.FormatFloat(
					(float64(sizeU.(int)) * 44.5), 'G', -1, 64)
			} else if _, ok := sizeU.(float64); ok {
				attr["sizeU"] = strconv.FormatFloat(sizeU.(float64), 'G', -1, 64)
				attr["height"] = strconv.FormatFloat(sizeU.(float64)*44.5, 'G', -1, 64)
			}
			//End of convert block
			if _, ok := attr["slot"]; ok {
				l.GetWarningLogger().Println("Invalid device syntax encountered")
				return fmt.Errorf("Invalid device syntax: If you have provided a template, it was not found")
			}
		}
		//}

		var slot map[string]any
		//Process the posU/slot attribute
		if x, ok := attr["posU/slot"]; ok {
			delete(attr, "posU/slot")
			if _, err := strconv.Atoi(x.(string)); err == nil {
				attr["posU"] = x
				attr["slot"] = ""
			} else {
				attr["slot"] = x
			}
			slot, err = GetSlot(parent, x.(string))
			if err != nil {
				return err
			}
		}

		//If user provided templates, get the JSON
		//and parse into templates
		if _, ok := attr["template"]; ok {
			err := GetOCLIAtributesTemplateHelper(attr, data, models.DEVICE)
			if err != nil {
				return err
			}
		} else {
			attr["template"] = ""
			if slot != nil {
				size := slot["elemSize"].([]any)
				attr["size"] = fmt.Sprintf(
					"{\"x\":%f, \"y\":%f}", size[0].(float64)/10., size[1].(float64)/10.)
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
		data["attributes"] = attr

	case models.GROUP:
		//name, category, domain, pid
		data["parentId"] = parent["id"]
		attr := data["attributes"].(map[string]interface{})

		groups := strings.Join(attr["content"].([]string), ",")
		attr["content"] = groups

	case models.STRAY_DEV:
		attr = data["attributes"].(map[string]interface{})
		if _, ok := attr["template"]; ok {
			err := GetOCLIAtributesTemplateHelper(attr, data, models.DEVICE)
			if err != nil {
				return err
			}
		} else {
			attr["template"] = ""
		}

	default:
		//Execution should not reach here!
		return fmt.Errorf("Invalid Object Specified!")
	}

	//Stringify the attributes if not already
	if _, ok := data["attributes"]; ok {
		if attributes, ok := data["attributes"].(map[string]interface{}); ok {
			for i := range attributes {
				attributes[i] = Stringify(attributes[i])
			}
		}
	}

	//Because we already stored the string conversion in category
	//we can do the conversion for templates here
	data["category"] = strings.Replace(data["category"].(string), "_", "-", 1)

	return C.PostObj(ent, data["category"].(string), data, path)
}

func CreateTag(slug, color string) error {
	return C.PostObj(models.TAG, models.EntityToString(models.TAG), map[string]any{
		"slug":        slug,
		"description": slug, // the description is initially set with the value of the slug
		"color":       color,
	}, models.TagsPath+slug)
}
