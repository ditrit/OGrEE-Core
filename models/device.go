package models

import (
	"fmt"
	u "p3/utils"

	"github.com/jinzhu/gorm"
)

type Device struct {
	gorm.Model
	Name        string          `json:"name"`
	Category    string          `json:"category"`
	Desc        string          `json:"description"`
	Domain      int             `json:"domain"`
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

	if device.Domain == 0 {
		return u.Message(false, "Domain should be on the payload"), false
	}

	if GetDB().Table("racks").
		Where("id = ?", device.Domain).First(&Rack{}).Error != nil {

		return u.Message(false, "Domain should be correspond to Rack ID"), false
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

func (device *Device) Create() map[string]interface{} {
	if resp, ok := device.Validate(); !ok {
		return resp
	}

	GetDB().Create(device)

	resp := u.Message(true, "success")
	resp["device"] = device
	return resp
}

//Get the first device given the rack
func GetDevice(rack *Rack) *Device {
	device := &Device{}
	err := GetDB().Table("devices").Where("foreignkey = ?", rack.ID).First(device).Error
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return device
}

//Obtain all devices of a rack
func GetDevices(rack *Rack) []*Device {
	devices := make([]*Device, 0)

	err := GetDB().Table("devices").Where("foreignkey = ?", rack.ID).Find(&devices).Error
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return devices
}

//More methods should be made to
//Meet CRUD capabilities
//Need Update and Delete
//These would be a bit more complicated
//So leave them out for now
