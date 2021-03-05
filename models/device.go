package models

import (
	"fmt"
	u "p3/utils"
)

type Device_Attributes struct {
	ID          int    `json:"id" gorm:"column:id"`
	PosXY       string `json:"posXY" gorm:"column:device_pos_x_y"`
	PosXYU      string `json:"posXYUnit" gorm:"column:device_pos_x_y_unit"`
	PosZ        string `json:"posZ" gorm:"column:device_pos_z"`
	PosZU       string `json:"posZUnit" gorm:"column:device_pos_z_unit"`
	Template    string `json:"template" gorm:"column:device_template"`
	Orientation string `json:"orientation" gorm:"column:device_orientation"`
	Size        string `json:"size" gorm:"column:device_size"`
	SizeU       string `json:"sizeUnit" gorm:"column:device_size_unit"`
	Height      string `json:"height" gorm:"column:device_height"`
	HeightU     string `json:"heightUnit" gorm:"column:device_height_unit"`
	Vendor      string `json:"vendor" gorm:"column:device_vendor"`
	Type        string `json:"type" gorm:"column:device_type"`
	Model       string `json:"model" gorm:"column:device_model"`
	Serial      string `json:"serial" gorm:"column:device_serial"`
}

type Device struct {
	//gorm.Model
	ID       int    `json:"id" gorm:"column:id"`
	Name     string `json:"name" gorm:"column:device_name"`
	ParentID string `json:"parentId" gorm:"column:device_parent_id"`
	Category string `json:"category" gorm:"-"`
	Domain   string `json:"domain" gorm:"column:device_domain"`
	//D           []string        `json:"description" gorm:"-"`
	//Description string          `gorm:"-"`
	Attributes Device_Attributes `json:"attributes"`

	//Site []Site
	//D is used to help the JSON marshalling
	//while Description will be used in
	//DB transactions
}

func (device *Device) Validate() (map[string]interface{}, bool) {
	if device.Name == "" {
		return u.Message(false, "Device Name should be on payload"), false
	}

	/*if device.Category == "" {
		return u.Message(false, "Category should be on the payload"), false
	}*/

	if device.Domain == "" {
		return u.Message(false, "Domain should be on the payload"), false
	}

	if GetDB().Table("rack").
		Where("id = ?", device.ParentID).RecordNotFound() == true {

		return u.Message(false, "Domain should be correspond to Rack ID"), false
	}

	if device.Attributes.PosXY == "" {
		return u.Message(false, "XY coordinates should be on payload"), false
	}

	if device.Attributes.PosXYU == "" {
		return u.Message(false, "PositionXYU string should be on the payload"), false
	}

	/*if device.Attributes.PosZ == "" {
		return u.Message(false, "Z coordinates should be on payload"), false
	}

	if device.Attributes.PosZU == "" {
		return u.Message(false, "PositionZU string should be on the payload"), false
	}*/

	/*if device.Attributes.Template == "" {
		return u.Message(false, "Template should be on the payload"), false
	}*/

	switch device.Attributes.Orientation {
	case "front", "rear", "frontflipped", "rearflipped":
	case "":
		return u.Message(false, "Orientation should be on the payload"), false

	default:
		return u.Message(false, "Orientation is invalid!"), false
	}

	if device.Attributes.Size == "" {
		return u.Message(false, "Invalid size on the payload"), false
	}

	if device.Attributes.SizeU == "" {
		return u.Message(false, "Rack size string should be on the payload"), false
	}

	if device.Attributes.Height == "" {
		return u.Message(false, "Invalid Height on payload"), false
	}

	if device.Attributes.HeightU == "" {
		return u.Message(false, "Rack Height string should be on the payload"), false
	}

	//Successfully validated Device
	return u.Message(true, "success"), true
}

func (device *Device) Create() map[string]interface{} {
	if resp, ok := device.Validate(); !ok {
		return resp
	}

	GetDB().Create(device)
	device.Attributes.ID = device.ID
	GetDB().Create(&(device.Attributes))

	resp := u.Message(true, "success")
	resp["device"] = device
	return resp
}

//Get the device given the ID
func GetDevice(id uint) *Device {
	device := &Device{}
	err := GetDB().Table("device").Where("id = ?", id).First(device).
		Table("device_attributes").Where("id = ?", id).First(&(device.Attributes)).Error
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

func GetAllDevices() []*Device {
	devices := make([]*Device, 0)
	attrs := make([]*Device_Attributes, 0)
	err := GetDB().Find(&devices).Find(&attrs).Error
	if err != nil {
		fmt.Println(err)
		return nil
	}

	for i := range devices {
		devices[i].Attributes = *(attrs[i])
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

	/*if newDeviceInfo.Desc != "" && newDeviceInfo.Desc != device.Desc {
		device.Desc = newDeviceInfo.Desc
	}*/

	//Should it be possible to update domain
	// Will have to think about it more
	//if newDeviceInfo.Domain

	/*if newDeviceInfo.Color != "" && newDeviceInfo.Color != device.Color {
		device.Color = newDeviceInfo.Color
	}

	if newDeviceInfo.Orientation != "" {
		switch newDeviceInfo.Orientation {
		case "NE", "NW", "SE", "SW":
			device.Orientation = newDeviceInfo.Orientation

		default:
		}
	}*/

	//Successfully validated the new data
	GetDB().Table("devices").Save(device)
	return u.Message(true, "success")
}

func DeleteDevice(id uint) map[string]interface{} {

	//This is a hard delete!
	e := GetDB().Unscoped().Table("device").Delete(&Device{}, id).Error

	//The command below is a soft delete
	//Meaning that the 'deleted_at' field will be set
	//the record will remain but unsearchable
	//e := GetDB().Table("tenants").Delete(Tenant{}, id).Error
	if e != nil {
		return u.Message(false, "There was an error in deleting the device")
	}

	return u.Message(true, "success")
}
