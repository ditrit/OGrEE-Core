package models

import (
	"fmt"
	u "p3/utils"
	"strconv"
	"strings"
)

type Device_Attributes struct {
	ID          int    `json:"-" gorm:"column:id"`
	PosXY       string `json:"posXY" gorm:"column:device_pos_x_y"`
	PosXYU      string `json:"posXYUnit" gorm:"column:device_pos_x_y_unit"`
	PosZ        string `json:"posZ" gorm:"column:device_pos_z"`
	PosZU       string `json:"posZUnit" gorm:"column:device_pos_z_unit"`
	Template    string `json:"template" gorm:"column:device_template"`
	Orientation string `json:"orientation" gorm:"column:device_orientation"`
	Size        string `json:"size" gorm:"column:device_size"`
	SizeUnit    string `json:"sizeUnit" gorm:"column:device_size_unit"`
	SizeU       string `json:"sizeU" gorm:"column:device_sizeu"`
	Slot        string `json:"slot" gorm:"column:device_slot"`
	PosU        string `json:"posU" gorm:"column:device_posu"`
	Height      string `json:"height" gorm:"column:device_height"`
	HeightU     string `json:"heightUnit" gorm:"column:device_height_unit"`
	Vendor      string `json:"vendor" gorm:"column:device_vendor"`
	Type        string `json:"type" gorm:"column:device_type"`
	Model       string `json:"model" gorm:"column:device_model"`
	Serial      string `json:"serial" gorm:"column:device_serial"`
}

type Device struct {
	//gorm.Model
	ID              int               `json:"-" gorm:"column:id"`
	IDJSON          string            `json:"id" gorm:"-"`
	Name            string            `json:"name" gorm:"column:device_name"`
	ParentID        string            `json:"parentId" gorm:"column:device_parent_id"`
	Category        string            `json:"category" gorm:"-"`
	Domain          string            `json:"domain" gorm:"column:device_domain"`
	DescriptionJSON []string          `json:"description" gorm:"-"`
	DescriptionDB   string            `json:"-" gorm:"column:device_description"`
	Attributes      Device_Attributes `json:"attributes"`

	//Site []Site
	//DescriptionJSON is used to help the JSON marshalling
	//while DescriptionDB will be used in
	//DB transactions
}

func (device *Device) Validate() (map[string]interface{}, bool) {
	if device.Name == "" {
		return u.Message(false, "Device Name should be on payload"), false
	}

	if device.Category == "" {
		return u.Message(false, "Category should be on the payload"), false
	}

	if device.Domain == "" {
		return u.Message(false, "Domain should be on the payload"), false
	}

	if GetDB().Table("rack").
		Where("id = ?", device.ParentID).First(&Rack{}).Error != nil {

		return u.Message(false, "Domain should be correspond to Rack ID"), false
	}

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

	if device.Attributes.SizeUnit == "" {
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

func (device *Device) Create() (map[string]interface{}, string) {
	if resp, ok := device.Validate(); !ok {
		return resp, "validate"
	}

	device.DescriptionDB = strings.Join(device.DescriptionJSON, "XYZ")

	if e := GetDB().Create(device).Error; e != nil {
		return u.Message(false, "Internal Error while creating Device: "+e.Error()),
			"internal"
	}
	device.IDJSON = strconv.Itoa(device.ID)
	device.Attributes.ID = device.ID
	if e := GetDB().Create(&(device.Attributes)).Error; e != nil {
		return u.Message(false, "Internal Error while creating Device Attrs: "+e.Error()),
			"internal"
	}

	resp := u.Message(true, "success")
	resp["device"] = device
	return resp, ""
}

//Get the device given the ID
func GetDevice(id uint) (*Device, string) {
	device := &Device{}
	err := GetDB().Table("device").Where("id = ?", id).First(device).
		Table("device_attributes").Where("id = ?", id).First(&(device.Attributes)).Error
	if err != nil {
		fmt.Println(err)
		return nil, err.Error()
	}
	device.DescriptionJSON = strings.Split(device.DescriptionDB, "XYZ")
	device.Category = "device"
	device.IDJSON = strconv.Itoa(device.ID)
	return device, ""
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

func GetAllDevices() ([]*Device, string) {
	devices := make([]*Device, 0)
	attrs := make([]*Device_Attributes, 0)
	err := GetDB().Find(&devices).Find(&attrs).Error
	if err != nil {
		fmt.Println(err)
		return nil, err.Error()
	}

	for i := range devices {
		devices[i].Category = "device"
		devices[i].Attributes = *(attrs[i])
		devices[i].DescriptionJSON = strings.Split(devices[i].DescriptionDB, "XYZ")
		devices[i].IDJSON = strconv.Itoa(devices[i].ID)
	}

	return devices, ""
}

func UpdateDevice(id uint, newDeviceInfo *Device) (map[string]interface{}, string) {
	device := &Device{}

	err := GetDB().Table("device").Where("id = ?", id).First(device).
		Table("device_attributes").Where("id = ?", id).First(&(device.Attributes)).Error
	if err != nil {
		return u.Message(false, "Error while checking device: "+err.Error()), err.Error()
	}

	if newDeviceInfo.Name != "" && newDeviceInfo.Name != device.Name {
		device.Name = newDeviceInfo.Name
	}

	if newDeviceInfo.Domain != "" && newDeviceInfo.Domain != device.Domain {
		device.Domain = newDeviceInfo.Domain
	}

	if dc := strings.Join(newDeviceInfo.DescriptionJSON, "XYZ"); dc != "" && strings.Compare(dc, device.DescriptionDB) != 0 {
		device.DescriptionDB = dc
	}

	if newDeviceInfo.Attributes.PosXY != "" && newDeviceInfo.Attributes.PosXY != device.Attributes.PosXY {
		device.Attributes.PosXY = newDeviceInfo.Attributes.PosXY
	}

	if newDeviceInfo.Attributes.PosXYU != "" && newDeviceInfo.Attributes.PosXYU != device.Attributes.PosXYU {
		device.Attributes.PosXYU = newDeviceInfo.Attributes.PosXYU
	}

	if newDeviceInfo.Attributes.PosZ != "" && newDeviceInfo.Attributes.PosZ != device.Attributes.PosZ {
		device.Attributes.PosZ = newDeviceInfo.Attributes.PosZ
	}

	if newDeviceInfo.Attributes.PosZU != "" && newDeviceInfo.Attributes.PosZU != device.Attributes.PosZU {
		device.Attributes.PosZU = newDeviceInfo.Attributes.PosZU
	}

	if newDeviceInfo.Attributes.Template != "" && newDeviceInfo.Attributes.Template != device.Attributes.Template {
		device.Attributes.Template = newDeviceInfo.Attributes.Template
	}

	if newDeviceInfo.Attributes.Orientation != "" {
		switch newDeviceInfo.Attributes.Orientation {
		case "front", "rear", "frontflipped", "rearflipped":
			device.Attributes.Orientation = newDeviceInfo.Attributes.Orientation

		default:
		}
	}

	if newDeviceInfo.Attributes.Size != "" && newDeviceInfo.Attributes.Size != device.Attributes.Size {
		device.Attributes.Size = newDeviceInfo.Attributes.Size
	}

	if newDeviceInfo.Attributes.SizeUnit != "" && newDeviceInfo.Attributes.SizeUnit != device.Attributes.SizeUnit {
		device.Attributes.SizeUnit = newDeviceInfo.Attributes.SizeUnit
	}

	if newDeviceInfo.Attributes.Height != "" && newDeviceInfo.Attributes.Height != device.Attributes.Height {
		device.Attributes.Height = newDeviceInfo.Attributes.Height
	}

	if newDeviceInfo.Attributes.HeightU != "" && newDeviceInfo.Attributes.HeightU != device.Attributes.HeightU {
		device.Attributes.HeightU = newDeviceInfo.Attributes.HeightU
	}

	if newDeviceInfo.Attributes.SizeU != "" && newDeviceInfo.Attributes.SizeU != device.Attributes.SizeU {
		device.Attributes.SizeU = newDeviceInfo.Attributes.SizeU
	}

	if newDeviceInfo.Attributes.PosU != "" && newDeviceInfo.Attributes.PosU != device.Attributes.PosU {
		device.Attributes.PosU = newDeviceInfo.Attributes.PosU
	}

	if newDeviceInfo.Attributes.Slot != "" && newDeviceInfo.Attributes.Slot != device.Attributes.Slot {
		device.Attributes.Slot = newDeviceInfo.Attributes.Slot
	}

	if newDeviceInfo.Attributes.Vendor != "" && newDeviceInfo.Attributes.Vendor != device.Attributes.Vendor {
		device.Attributes.Vendor = newDeviceInfo.Attributes.Vendor
	}

	if newDeviceInfo.Attributes.Type != "" && newDeviceInfo.Attributes.Type != device.Attributes.Type {
		device.Attributes.Type = newDeviceInfo.Attributes.Type
	}

	if newDeviceInfo.Attributes.Model != "" && newDeviceInfo.Attributes.Model != device.Attributes.Model {
		device.Attributes.Model = newDeviceInfo.Attributes.Model
	}

	if newDeviceInfo.Attributes.Serial != "" && newDeviceInfo.Attributes.Serial != device.Attributes.Serial {
		device.Attributes.Serial = newDeviceInfo.Attributes.Serial
	}

	//Successfully validated the new data
	if e1 := GetDB().Table("device").Save(device).
		Table("device_attributes").Save(&(device.Attributes)).Error; e1 != nil {
		return u.Message(false, "Error while updating device: "+e1.Error()), e1.Error()
	}
	return u.Message(true, "success"), ""
}

func DeleteDevice(id uint) map[string]interface{} {

	//This is a hard delete!
	e := GetDB().Unscoped().Table("device").Delete(&Device{}, id).RowsAffected

	//The command below is a soft delete
	//Meaning that the 'deleted_at' field will be set
	//the record will remain but unsearchable
	//e := GetDB().Table("tenants").Delete(Tenant{}, id).Error
	if e == 0 {
		return u.Message(false, "There was an error in deleting the device")
	}

	return u.Message(true, "success")
}

func GetDeviceByName(name string) (*Device, string) {
	device := &Device{}

	e := GetDB().Raw(`SELECT * FROM device 
	JOIN device_attributes ON device.id = device_attributes.id 
	WHERE device_name = ?;`, name).Find(device).Find(&device.Attributes).Error

	if e != nil {
		//fmt.Println(e)
		return nil, e.Error()
	}

	device.IDJSON = strconv.Itoa(device.ID)
	device.DescriptionJSON = strings.Split(device.DescriptionDB, "XYZ")
	device.Category = "device"
	return device, ""
}
