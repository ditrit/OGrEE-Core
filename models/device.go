package models

import (
	u "p3/utils"

	"github.com/jinzhu/gorm"
)

type Device struct {
	gorm.Model
	Name        string          `json:"name"`
	Category    string          `json:"category"`
	Desc        string          `json:"description"`
	Domain      string          `json:"domain"`
	Color       string          `json:"color"`
	Orientation ECardinalOrient `json:"eorientation"`
}

func (device *Device) Validate() (map[string]interface{}, bool) {
	if device.Name == "" {
		return u.Message(false, "Device Name should be on payload"), false
	}

	if device.Category == "" {
		return u.Message(false, "Category should be on the payload"), false
	}

	if device.Desc == "" {
		return u.Message(false, "Description should be on the paylad"), false
	}

	if device.Domain == "" {
		return u.Message(false, "Domain should NULL!"), false
	}

	if device.Color == "" {
		return u.Message(false, "Color should be on the payload"), false
	}

	switch device.Orientation {
	case "NE", "NW", "SE", "SW":
	case "":
		return u.Message(false, "Orientation should be on the payload"), false

	default:
		return u.Message(false, "Orientation is invalid!"), false
	}

	//Successfully validated Device
	return u.Message(true, "success"), true
}
