package models

import (
	"p3/repository"
	u "p3/utils"
	"strings"

	"github.com/elliotchance/pie/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const RACKUNIT = 0.04445 //meter

func validateAttributes(entity int, data, parent map[string]any) *u.Error {
	attributes := data["attributes"].(map[string]any)
	switch entity {
	case u.CORRIDOR:
		setCorridorColor(attributes)
	case u.GROUP:
		if err := validateGroupContent(attributes["content"].([]any),
			data["parentId"].(string), parent["parent"].(string)); err != nil {
			return err
		}
	case u.DEVICE:
		var err *u.Error
		if attributes["sizeU"] != nil && attributes["height"] != nil {
			if err = checkSizeUAndHeight(attributes); err != nil {
				return err
			}
		}
		var deviceSlots []string
		if deviceSlots, err = slotToValidSlice(attributes); err != nil {
			return err
		}
		// check if all requested slots are free
		if err = validateDeviceSlots(deviceSlots,
			data["name"].(string), data["parentId"].(string)); err != nil {
			return err
		}
	case u.VIRTUALOBJ:
		if attributes["vlinks"] != nil {
			// check if all vlinks point to valid objects
			if err := validateVlinks(attributes["vlinks"].([]any)); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateDeviceSlots(deviceSlots []string, deviceName, deviceParentd string) *u.Error {
	// check if all requested slots are free
	var siblings []map[string]any
	var err *u.Error

	// find siblings
	idPattern := primitive.Regex{Pattern: "^" + deviceParentd +
		"(." + u.NAME_REGEX + "){1}$", Options: ""}
	if siblings, err = GetManyObjects(u.EntityToString(u.DEVICE), bson.M{"id": idPattern},
		u.RequestFilters{}, "", nil); err != nil {
		return err
	}

	for _, obj := range siblings {
		if obj["name"] == deviceName {
			// do not check itself
			continue
		}
		if siblingSlots, err := slotToValidSlice(obj["attributes"].(map[string]any)); err == nil {
			for _, requestedSlot := range deviceSlots {
				if pie.Contains(siblingSlots, requestedSlot) {
					return &u.Error{Type: u.ErrBadFormat,
						Message: "Invalid slot: one or more requested slots are already in use"}
				}
			}
		}
	}
	return nil
}

func validateVlinks(vlinks []any) *u.Error {
	for _, vlinkId := range vlinks {
		count, err := repository.CountObjectsManyEntities([]int{u.DEVICE, u.VIRTUALOBJ},
			bson.M{"id": strings.Split(vlinkId.(string), "#")[0]})
		if err != nil {
			return err
		}

		if count != 1 {
			return &u.Error{
				Type:    u.ErrBadFormat,
				Message: "One or more vlink objects could not be found. Note that it must be device or virtual obj",
			}
		}
	}
	return nil
}

func validateGroupContent(content []any, parentId, parentCategory string) *u.Error {
	if len(content) <= 1 && content[0] == "" {
		return &u.Error{
			Type:    u.ErrBadFormat,
			Message: "objects separated by a comma must be on the payload",
		}
	}

	// Ensure objects are all unique
	if !pie.AreUnique(content) {
		return &u.Error{
			Type:    u.ErrBadFormat,
			Message: "The group cannot have duplicate objects",
		}
	}

	// Ensure objects all exist
	if err := checkGroupContentExists(content, parentId, parentCategory); err != nil {
		return err
	}

	return nil
}

func checkGroupContentExists(content []any, parentId, parentCategory string) *u.Error {
	// Get filter
	filter := repository.GroupContentToOrFilter(content, parentId)

	// Get entities
	var siblingsEnts []int
	if parentCategory == "rack" {
		// If parent is rack, retrieve devices
		siblingsEnts = []int{u.DEVICE}
	} else {
		// If parent is room, retrieve room children
		siblingsEnts = u.RoomChildren
	}

	// Try to get the whole content
	count, err := repository.CountObjectsManyEntities(siblingsEnts, filter)
	if err != nil {
		return err
	}
	if count != len(content) {
		return &u.Error{
			Type:    u.ErrBadFormat,
			Message: "Some object(s) could not be found. Please check and try again",
		}
	}
	return nil
}

func setCorridorColor(attributes map[string]any) {
	// Set the color manually based on temp. as specified by client
	if attributes["temperature"] == "warm" {
		attributes["color"] = "990000"
	} else if attributes["temperature"] == "cold" {
		attributes["color"] = "000099"
	}
}

// Check if sizeU and height are coherents
func checkSizeUAndHeight(attributes map[string]any) *u.Error {
	sizeU, err := u.GetFloat(attributes["sizeU"])
	if err != nil {
		return &u.Error{
			Type:    u.ErrBadFormat,
			Message: err.Error(),
		}
	}
	height := attributes["height"]
	h := sizeU * RACKUNIT
	switch heightUnit := attributes["heightUnit"]; heightUnit {
	case "cm":
		h *= 100
	case "mm":
		h *= 1000
	}
	if height == h {
		return nil
	} else {
		return &u.Error{
			Type:    u.ErrBadFormat,
			Message: "sizeU and height are not consistent",
		}
	}
}
