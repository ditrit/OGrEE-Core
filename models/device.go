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

	if GetDB().Table("devices").
		Where("id = ?", device.Domain).First(&Device{}).Error != nil {

		return u.Message(false, "Domain should be correspond to Device ID"), false
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

//Get the device given the ID
func GetDevice(id uint) *Device {
	device := &Device{}
	err := GetDB().Table("devices").Where("id = ?", id).First(device).Error
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

func UpdateDevice(id uint, newDeviceInfo *Device) map[string]interface{} {
	device := &Device{}

	err := GetDB().Table("devices").Where("id = ?", id).First(device).Error
	if err != nil {
		return u.Message(false, "Device was not found")
	}

	if newDeviceInfo.Name != "" && newDeviceInfo.Name != device.Name {
		device.Name = newDeviceInfo.Name
	}

	if newDeviceInfo.Category != "" && newDeviceInfo.Category != device.Category {
		device.Category = newDeviceInfo.Category
	}

	if newDeviceInfo.Desc != "" && newDeviceInfo.Desc != device.Desc {
		device.Desc = newDeviceInfo.Desc
	}

	//Should it be possible to update domain
	// Will have to think about it more
	//if newDeviceInfo.Domain

	if newDeviceInfo.Color != "" && newDeviceInfo.Color != device.Color {
		device.Color = newDeviceInfo.Color
	}

	if newDeviceInfo.Orientation != "" {
		switch newDeviceInfo.Orientation {
		case "NE", "NW", "SE", "SW":
			device.Orientation = newDeviceInfo.Orientation

		default:
		}
	}

	//Successfully validated the new data
	GetDB().Table("devices").Save(device)
	return u.Message(true, "success")
}

func DeleteDevice(id uint) map[string]interface{} {

	//First check if the device exists
	err := GetDB().Table("devices").Where("id = ?", id).First(&Device{}).Error
	if err != nil {
		fmt.Println("Couldn't find the device to delete")
		return nil
	}

	//This is a hard delete!
	e := GetDB().Unscoped().Table("devices").Delete(&Device{}, id).Error

	//The command below is a soft delete
	//Meaning that the 'deleted_at' field will be set
	//the record will remain but unsearchable
	//e := GetDB().Table("tenants").Delete(Tenant{}, id).Error
	if e != nil {
		return u.Message(false, "There was an error in deleting the device")
	}

	return u.Message(true, "success")
}
