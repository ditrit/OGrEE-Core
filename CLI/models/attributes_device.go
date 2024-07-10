package models

import (
	"cli/utils"
	"fmt"
)

// Compute coherent sizeU or height according to given data
func ComputeSizeUAndHeight(obj, data map[string]any) error {
	errMsg := "unknown heightUnit value"
	if data["attributes"] == nil {
		return nil
	}
	newAttrs, err := MapStringAny(data["attributes"])
	if err != nil {
		return err
	}
	currentAttrs, err := MapStringAny(obj["attributes"])
	if err != nil {
		return err
	}
	if newAttrs["sizeU"] != nil {
		sizeU, err := utils.GetFloat(newAttrs["sizeU"])
		if err != nil {
			return err
		}
		var height = sizeU * RACKUNIT
		switch heightUnit := currentAttrs["heightUnit"]; heightUnit {
		case "cm":
			height *= 100
		case "mm":
			height *= 1000
		default:
			return fmt.Errorf(errMsg)
		}
		newAttrs["height"] = height
	}
	if newAttrs["height"] != nil {
		height, err := utils.GetFloat(newAttrs["height"])
		if err != nil {
			return err
		}
		var sizeU = height / RACKUNIT
		switch heightUnit := currentAttrs["heightUnit"]; heightUnit {
		case "cm":
			sizeU /= 100
		case "mm":
			sizeU /= 1000
		default:
			return fmt.Errorf(errMsg)
		}
		newAttrs["sizeU"] = sizeU
	}
	return nil
}
