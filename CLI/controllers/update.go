package controllers

import (
	"cli/models"
	"cli/utils"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

const invalidAttrNameMsg = "invalid attribute name"

// InnerAttrObjs
const (
	PillarAttr    = "pillar"
	SeparatorAttr = "separator"
	BreakerAttr   = "breaker"
)
const VirtualConfigAttr = "virtual_config"

func (controller Controller) UpdateObject(path, attr string, values []any) error {
	var err error
	switch attr {
	case "areas":
		_, err = controller.UpdateRoomAreas(path, values)
	case "separators+", "pillars+", "breakers+":
		_, err = controller.AddInnerAtrObj(strings.TrimSuffix(attr, "s+"), path, values)
	case "pillars-", "separators-", "breakers-":
		_, err = controller.DeleteInnerAttrObj(path, strings.TrimSuffix(attr, "-"), values[0].(string))
	case "vlinks+", "vlinks-":
		_, err = controller.UpdateVirtualLink(path, attr, values[0].(string))
	case "domain", "tags+", "tags-":
		isRecursive := len(values) > 1 && values[1] == "recursive"
		_, err = controller.PatchObj(path, map[string]any{attr: values[0]}, isRecursive)
	case "tags", "separators", "pillars", "vlinks", "breakers":
		err = fmt.Errorf(
			"object's %[1]s can not be updated directly, please use %[1]s+= and %[1]s-=",
			attr,
		)
	case "description":
		_, err = controller.UpdateDescription(path, attr, values)
	case VirtualConfigAttr:
		err = controller.AddVirtualConfig(path, values)
	default:
		if strings.Contains(attr, ".") {
			err = controller.UpdateInnerAtrObj(path, attr, values)
		} else {
			_, err = controller.UpdateAttributes(path, attr, values)
		}
	}
	return err
}

func (controller Controller) PatchObj(pathStr string, data map[string]any, withRecursive bool) (map[string]any, error) {
	obj, err := controller.GetObject(pathStr)
	if err != nil {
		return nil, err
	}

	category := ""
	if obj["category"] != nil {
		category = obj["category"].(string)
	}

	if category == models.EntityToString(models.DEVICE) {
		models.ComputeSizeUAndHeight(obj, data)
	}

	url, err := controller.ObjectUrl(pathStr, 0)
	if err != nil {
		return nil, err
	}
	if withRecursive {
		url = url + "?recursive=true"
	}

	resp, err := controller.API.Request(http.MethodPatch, url, data, http.StatusOK)
	if err != nil {
		return nil, err
	}

	if models.IsLayer(pathStr) {
		// For layers, update the object to the hierarchy in order to be cached
		data := resp.Body["data"].(map[string]any)
		_, err = State.Hierarchy.AddObjectInPath(data, pathStr)
		if err != nil {
			return nil, err
		}
	}

	return resp.Body, nil
}

// [obj]:[attributeName]=[values]
func (controller Controller) UpdateAttributes(path, attributeName string, values []any) (map[string]any, error) {
	var attributes map[string]any
	if attributeName == "slot" || attributeName == "content" {
		vecStr := []string{}
		for _, value := range values {
			vecStr = append(vecStr, value.(string))
		}
		var err error
		if vecStr, err = models.CheckExpandStrVector(vecStr); err != nil {
			return nil, err
		}
		attributes = map[string]any{attributeName: vecStr}
	} else {
		if len(values) > 1 {
			return nil, fmt.Errorf("attributes can only be assigned a single value")
		}
		attributes = map[string]any{attributeName: values[0]}
	}

	return controller.PatchObj(path, map[string]any{"attributes": attributes}, false)
}

// [obj]:description=values
func (controller Controller) UpdateDescription(path string, attr string, values []any) (map[string]any, error) {
	if len(values) != 1 {
		return nil, fmt.Errorf("a single value is expected to update a description")
	}
	newDesc, err := utils.ValToString(values[0], "description")
	if err != nil {
		return nil, err
	}
	data := map[string]any{"description": newDesc}
	return controller.PatchObj(path, data, false)
}

// vlinks+ and vlinks-
func (controller Controller) UpdateVirtualLink(path string, attr string, value string) (map[string]any, error) {
	if len(value) == 0 {
		return nil, fmt.Errorf("an empty string is not valid")
	}

	obj, err := controller.GetObject(path)
	if err != nil {
		return nil, err
	} else if obj["category"] != models.EntityToString(models.VIRTUALOBJ) {
		return nil, fmt.Errorf("only virtual objects can have vlinks")
	}

	vlinks, hasVlinks := obj["attributes"].(map[string]any)["vlinks"].([]any)
	if attr == "vlinks+" {
		if !hasVlinks {
			vlinks = []any{value}
		} else {
			vlinks = append(vlinks, value)
		}
	} else if attr == "vlinks-" {
		if !hasVlinks {
			return nil, fmt.Errorf("no vlinks defined for this object")
		}
		vlinks, err = removeVirtualLink(vlinks, value)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("invalid vlink update command")
	}

	data := map[string]any{"vlinks": vlinks}
	return controller.PatchObj(path, map[string]any{"attributes": data}, false)
}

func removeVirtualLink(vlinks []any, vlinkToRemove string) ([]any, error) {
	for i, vlink := range vlinks {
		if vlink == vlinkToRemove {
			vlinks = append(vlinks[:i], vlinks[i+1:]...)
			return vlinks, nil
		}
	}
	return nil, fmt.Errorf("vlink to remove not found")
}

// only to delete pillars, separators and breakers
func (controller Controller) DeleteInnerAttrObj(path, attribute, name string) (map[string]any, error) {
	obj, err := controller.GetObject(path)
	if err != nil {
		return nil, err
	}
	attributes := obj["attributes"].(map[string]any)
	attrMap, ok := attributes[attribute].(map[string]any)
	if !ok || attrMap[name] == nil {
		return nil, fmt.Errorf("%s %s does not exist", attribute, name)
	}
	delete(attrMap, name)
	attributes[attribute] = attrMap
	fmt.Println(attributes)
	return controller.PatchObj(path, map[string]any{"attributes": attributes}, false)
}

// only to add pillars, separators and breakers
func (controller Controller) AddInnerAtrObj(attrName, path string, values []any) (map[string]any, error) {
	// get object
	obj, err := controller.GetObject(path)
	if err != nil {
		return nil, err
	}
	attr := obj["attributes"].(map[string]any)
	if (attrName == PillarAttr || attrName == SeparatorAttr) && obj["category"] != models.EntityToString(models.ROOM) {
		return nil, fmt.Errorf("this attribute can only be added to rooms")
	} else if attrName == BreakerAttr && obj["category"] != models.EntityToString(models.RACK) {
		return nil, fmt.Errorf("this attribute can only be added to racks")
	}

	// check and create attr
	var name string
	var newAttrObject any
	if attrName == PillarAttr {
		name, newAttrObject, err = models.ValuesToPillar(values)
	} else if attrName == SeparatorAttr {
		name, newAttrObject, err = models.ValuesToSeparator(values)
	} else if attrName == BreakerAttr {
		name, newAttrObject, err = models.ValuesToBreaker(values)
	}
	if err != nil {
		return nil, err
	}

	// add attr to object
	var keyExist bool
	attr[attrName+"s"], keyExist = AddToMap(attr[attrName+"s"], name, newAttrObject)
	obj, err = controller.PatchObj(path, map[string]any{"attributes": attr}, false)
	if err != nil {
		return nil, err
	}
	if keyExist {
		fmt.Printf(attrName+" %s replaced\n", name)
	}
	return obj, nil
}

// [device]:virtual_config=type@clusterId@role
func (controller Controller) AddVirtualConfig(path string, values []any) error {
	if len(values) < 1 {
		return fmt.Errorf("invalid virtual_cofig values")
	}
	vconfig := map[string]any{"type": values[0]}
	if len(values) > 1 {
		vconfig["clusterId"] = values[1]
	}
	if len(values) > 2 {
		vconfig["role"] = values[2]
	}

	attributes := map[string]any{VirtualConfigAttr: vconfig}
	_, err := controller.PatchObj(path, map[string]any{"attributes": attributes}, false)
	return err
}

// [room]:areas=[r1,r2,r3,r4]@[t1,t2,t3,t4]
func (controller Controller) UpdateRoomAreas(path string, values []any) (map[string]any, error) {
	if attributes, e := models.SetRoomAreas(values); e != nil {
		return nil, e
	} else {
		return controller.PatchObj(path, map[string]any{"attributes": attributes}, false)
	}
}

func (controller Controller) UpdateInnerAtrObj(path, attr string, values []any) error {
	if regexp.MustCompile(`^breakers.([\w-]+).([\w-]+)$`).MatchString(attr) {
		return controller.UpdateRackBreakerData(path, attr, values)
	} else if regexp.MustCompile(`^virtual_config.([\w-]+)$`).MatchString(attr) {
		return controller.UpdateVirtualConfig(path, attr, values)
	} else {
		return fmt.Errorf(invalidAttrNameMsg)
	}
}

// [rack]:breakers.breakerName.attribute=value
func (controller Controller) UpdateRackBreakerData(path, attr string, values []any) error {
	// format attribute
	attrs := strings.Split(attr, ".") // breakers.name.attribute
	if len(attrs) != 3 {
		return fmt.Errorf(invalidAttrNameMsg)
	}
	// get rack and modify breakers
	obj, err := controller.GetObject(path)
	if err != nil {
		return err
	}
	attributes := obj["attributes"].(map[string]any)
	breakers, hasBreakers := attributes["breakers"].(map[string]any)
	notFoundErr := fmt.Errorf("rack does not have specified breaker")
	if !hasBreakers {
		return notFoundErr
	}
	breaker, hasBreaker := breakers[attrs[1]].(map[string]any)
	if !hasBreaker {
		return notFoundErr
	}
	breaker[attrs[2]] = values[0]
	_, err = controller.PatchObj(path, map[string]any{"attributes": attributes}, false)
	return err
}

// [obj]:virtual_config.attr=value
func (controller Controller) UpdateVirtualConfig(path, attr string, values []any) error {
	vconfigAttr, _ := strings.CutPrefix(attr, VirtualConfigAttr+".")
	if len(vconfigAttr) < 1 {
		return fmt.Errorf(invalidAttrNameMsg)
	}

	// get object and modify virtual config
	obj, err := controller.GetObject(path)
	if err != nil {
		return err
	}
	attributes := obj["attributes"].(map[string]any)
	vconfig, hasVconfig := attributes[VirtualConfigAttr].(map[string]any)
	if !hasVconfig {
		return fmt.Errorf("object does not have virtual config")
	}
	vconfig[vconfigAttr] = values[0]
	_, err = controller.PatchObj(path, map[string]any{"attributes": attributes}, false)
	return err
}

// Helpers
func AddToMap[T any](mapToAdd any, key string, val T) (map[string]any, bool) {
	attrMap, ok := mapToAdd.(map[string]any)
	if !ok {
		attrMap = map[string]any{}
	}
	_, keyExist := attrMap[key]
	attrMap[key] = val
	return attrMap, keyExist
}
