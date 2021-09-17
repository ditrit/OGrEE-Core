package models

import (
	"fmt"
	u "p3/utils"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

type Device_Attributes struct {
	ID          int    `json:"-" bson:"id"`
	PosXY       string `json:"posXY" bson:"device_pos_x_y"`
	PosXYU      string `json:"posXYUnit" bson:"device_pos_x_y_unit"`
	PosZ        string `json:"posZ" bson:"device_pos_z"`
	PosZU       string `json:"posZUnit" bson:"device_pos_z_unit"`
	Template    string `json:"template" bson:"device_template"`
	Orientation string `json:"orientation" bson:"device_orientation"`
	Size        string `json:"size" bson:"device_size"`
	SizeUnit    string `json:"sizeUnit" bson:"device_size_unit"`
	SizeU       string `json:"sizeU" bson:"device_sizeu"`
	Slot        string `json:"slot" bson:"device_slot"`
	PosU        string `json:"posU" bson:"device_posu"`
	Height      string `json:"height" bson:"device_height"`
	HeightU     string `json:"heightUnit" bson:"device_height_unit"`
	Vendor      string `json:"vendor" bson:"device_vendor"`
	Type        string `json:"type" bson:"device_type"`
	Model       string `json:"model" bson:"device_model"`
	Serial      string `json:"serial" bson:"device_serial"`
}

type Device struct {
	ID              int               `json:"-" bson:"id"`
	IDJSON          string            `json:"id" bson:"-"`
	Name            string            `json:"name" bson:"device_name"`
	ParentID        string            `json:"parentId" bson:"device_parent_id"`
	Category        string            `json:"category" bson:"-"`
	Domain          string            `json:"domain" bson:"device_domain"`
	DescriptionJSON []string          `json:"description" bson:"-"`
	DescriptionDB   string            `json:"-" bson:"device_description"`
	Attributes      Device_Attributes `json:"attributes"`
	Subdevices      []*Subdevice      `json:"subdevices,omitempty" bson:"-"`
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

	if GetDB().Collection("rack").FindOne(GetCtx(), bson.M{"_id": device.ParentID}).Err() != nil {
		return u.Message(false, "ParentID should be correspond to Rack ID"), false
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

	if _, e := GetDB().Collection("device").InsertOne(GetCtx(), device); e != nil {
		return u.Message(false, "Internal Error while creating Device: "+e.Error()),
			"internal"
	}
	device.IDJSON = strconv.Itoa(device.ID)
	/*device.Attributes.ID = device.ID
	if e := GetDB().Create(&(device.Attributes)).Error; e != nil {
		return u.Message(false, "Internal Error while creating Device Attrs: "+e.Error()),
			"internal"
	}*/

	resp := u.Message(true, "success")
	resp["data"] = device
	return resp, ""
}

func (d *Device) FormQuery() string {

	query := "SELECT * FROM device " + u.JoinQueryGen("device")
	if d.ParentID != "" {
		//pid, _ := strconv.Atoi(d.ParentID)
		query += " AND device_parent_id = " + d.ParentID
	}
	if d.IDJSON != "" {
		//id, _ := strconv.Atoi(d.IDJSON)
		query += " AND device.id = " + d.IDJSON
	}
	if d.ParentID != "" {
		//pid, _ := strconv.Atoi(d.ParentID)
		query += " AND device_parent_id = " + d.ParentID
	}
	if d.Name != "" {
		query += " WHERE device_name = '" + d.Name + "'"
	}
	if d.Category != "" {
		query += " AND device_category = '" + d.Category + "'"
	}
	if d.Domain != "" {
		query += " AND device_domain = '" + d.Domain + "'"
	}
	if (Device_Attributes{}) != d.Attributes {
		if d.Attributes.Template != "" {
			query +=
				" AND device_attributes.device_template = '" +
					d.Attributes.Template + "'"
		}
		if d.Attributes.Orientation != "" {
			query +=
				" AND device_attributes.device_orientation = '" +
					d.Attributes.Orientation + "'"
		}
		if d.Attributes.PosXY != "" {
			query +=
				" AND device_attributes.device_pos_x_y = '" +
					d.Attributes.PosXY + "'"
		}
		if d.Attributes.PosXYU != "" {
			query +=
				" AND device_attributes.device_pos_x_y_unit = '" +
					d.Attributes.PosXYU + "'"
		}
		if d.Attributes.PosZ != "" {
			query +=
				" AND device_attributes.device_pos_z = '" +
					d.Attributes.PosZ + "'"
		}
		if d.Attributes.PosZU != "" {
			query +=
				" AND device_attributes.device_pos_z_unit = '" +
					d.Attributes.PosZU + "'"
		}
		if d.Attributes.Size != "" {
			query +=
				" AND device_attributes.device_size = '" +
					d.Attributes.Size + "'"
		}
		if d.Attributes.SizeU != "" {
			query +=
				" AND device_attributes.device_sizeu = '" +
					d.Attributes.SizeU + "'"
		}
		if d.Attributes.SizeUnit != "" {
			query +=
				" AND device_attributes.device_size_unit = '" +
					d.Attributes.SizeUnit + "'"
		}
		if d.Attributes.Slot != "" {
			query +=
				" AND device_attributes.device_slot = '" +
					d.Attributes.Slot + "'"
		}
		if d.Attributes.PosU != "" {
			query +=
				" AND device_attributes.device_posu= '" +
					d.Attributes.PosU + "'"
		}
		if d.Attributes.Height != "" {
			query +=
				" AND device_attributes.device_height = '" +
					d.Attributes.Height + "'"
		}
		if d.Attributes.HeightU != "" {
			query +=
				" AND device_attributes.device_height_unit = '" +
					d.Attributes.HeightU + "'"
		}
		if d.Attributes.Vendor != "" {
			query +=
				" AND device_attributes.device_vendor = '" +
					d.Attributes.Vendor + "'"
		}
		if d.Attributes.Type != "" {
			query +=
				" AND device_attributes.device_type = '" +
					d.Attributes.Type + "'"
		}
		if d.Attributes.Model != "" {
			query +=
				" AND device_attributes.device_model = '" +
					d.Attributes.Model + "'"
		}
		if d.Attributes.Serial != "" {
			query +=
				" AND device_attributes.device_serial = '" +
					d.Attributes.Serial + "'"
		}
	}
	println(query)
	return query
}

//Get the device given the ID
func GetDevice(id uint) (*Device, string) {
	device := &Device{}
	err := GetDB().Collection("device").FindOne(GetCtx(), bson.M{"_id": id}).Decode(device).Error()
	if err != "" {
		fmt.Println(err)
		return nil, err
	}
	device.DescriptionJSON = strings.Split(device.DescriptionDB, "XYZ")
	device.Category = "device"
	device.IDJSON = strconv.Itoa(device.ID)
	return device, ""
}

//Obtain all devices of a rack
func GetDevicesOfParent(id uint) ([]*Device, string) {
	devices := make([]*Device, 0)
	c, err := GetDB().Collection("device").Find(GetCtx(), bson.M{"device_parent_id": id})
	if err != nil {
		fmt.Println(err)
		return nil, err.Error()
	}

	for c.Next(GetCtx()) {
		d := &Device{}
		e := c.Decode(d)
		if e != nil {
			fmt.Println(e)
			return nil, e.Error()
		}
		devices = append(devices, d)
	}

	println("The length of device is: ", len(devices))
	/*for i := range devices {
		e := GetDB().Collection("device_attributes").Where("id = ?", devices[i].ID).First(&(devices[i].Attributes)).Error

		if e != nil {
			fmt.Println(err)
			return nil, err.Error()
		}

		devices[i].Category = "device"
		devices[i].DescriptionJSON = strings.Split(devices[i].DescriptionDB, "XYZ")
		devices[i].IDJSON = strconv.Itoa(devices[i].ID)
	}*/

	return devices, ""
}

func GetAllDevices() ([]*Device, string) {
	devices := make([]*Device, 0)
	//attrs := make([]*Device_Attributes, 0)
	c, err := GetDB().Collection("device").Find(GetCtx(), bson.D{{}})
	if err != nil {
		fmt.Println(err)
		return nil, err.Error()
	}

	for c.Next(GetCtx()) {
		d := &Device{}
		e := c.Decode(d)
		if e != nil {
			fmt.Println(err)
			return nil, err.Error()
		}
		d.Category = "device"
		devices = append(devices, d)
	}

	/*for i := range devices {
		devices[i].Category = "device"
		devices[i].Attributes = *(attrs[i])
		devices[i].DescriptionJSON = strings.Split(devices[i].DescriptionDB, "XYZ")
		devices[i].IDJSON = strconv.Itoa(devices[i].ID)
	}*/

	return devices, ""
}

func GetDeviceByQuery(device *Device) ([]*Device, string) {
	/*devices := make([]*Device, 0)
	attrs := make([]*Device_Attributes, 0)

	e := GetDB().Raw(device.FormQuery()).Find(&devices).
		Find(&attrs).Error

	if e != nil {
		return nil, e.Error()
	}

	for i := range devices {
		devices[i].Attributes = *(attrs[i])
		devices[i].IDJSON = strconv.Itoa(devices[i].ID)
		devices[i].DescriptionJSON =
			strings.Split(devices[i].DescriptionDB, "XYZ")
		devices[i].Category = "device"
	}

	return devices, ""*/
	return nil, ""
}

func UpdateDevice(id uint, newDeviceInfo *map[string]interface{}) (map[string]interface{}, string) {
	/*device := &Device{}

	err := GetDB().Collection("device").Where("id = ?", id).First(device).
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
	}*/

	//Successfully validated the new data
	/*if e1 := GetDB().Collection("device").Save(device).
		Table("device_attributes").Save(&(device.Attributes)).Error; e1 != nil {
		return u.Message(false, "Error while updating device: "+e1.Error()), e1.Error()
	}*/
	e := GetDB().Collection("device").FindOneAndUpdate(GetCtx(), bson.M{"_id": id}, bson.M{"$set": *newDeviceInfo}).Err()
	if e != nil {
		return u.Message(false, "failure: "+e.Error()), e.Error()
	}
	return u.Message(true, "success"), ""
}

func DeleteDevice(id uint) map[string]interface{} {

	//This is a hard delete!
	c, _ := GetDB().Collection("device").DeleteOne(GetCtx(), bson.M{"_id": id})
	//The command below is a soft delete
	//Meaning that the 'deleted_at' field will be set
	//the record will remain but unsearchable
	//e := GetDB().Table("tenants").Delete(Tenant{}, id).Error
	if c.DeletedCount == 0 {
		return u.Message(false, "There was an error in deleting the device")
	}

	return u.Message(true, "success")
}

func GetDeviceByName(name string) (*Device, string) {
	device := &Device{}

	/*e := GetDB().Raw(`SELECT * FROM device
	JOIN device_attributes ON device.id = device_attributes.id
	WHERE device_name = ?;`, name).Find(device).Find(&device.Attributes).Error*/
	e := GetDB().Collection("device").FindOne(GetCtx(), bson.M{"name": name}).Decode(device)

	if e != nil {
		return nil, e.Error()
	}

	device.IDJSON = strconv.Itoa(device.ID)
	device.DescriptionJSON = strings.Split(device.DescriptionDB, "XYZ")
	device.Category = "device"
	return device, ""
}

func GetDeviceByNameAndParentID(id uint, name string) (*Device, string) {
	device := &Device{}
	/*err := GetDB().Raw(`SELECT * FROM device JOIN
	device_attributes ON device.id = device_attributes.id
	WHERE device_parent_id = ? AND device_name = ?`, id, name).
		Find(device).Find(&(device.Attributes)).Error*/
	err := GetDB().Collection("device").FindOne(GetCtx(), bson.M{"device_parent_id": id, "name": name}).Decode(device)
	if err != nil {
		fmt.Println(err)
		return nil, err.Error()
	}

	device.DescriptionJSON = strings.Split(device.DescriptionDB, "XYZ")
	device.Category = "device"
	device.IDJSON = strconv.Itoa(device.ID)
	return device, ""
}

//Obtain device and all subdevices
func GetDeviceHierarchy(id int) (*Device, string) {
	dev, e := GetDevice(uint(id))
	if e != "" {
		return nil, e
	}

	dev.Subdevices, e = GetSubdevicesOfParent(uint(id))
	if e != "" {
		return nil, e
	}

	for i, _ := range dev.Subdevices {
		dev.Subdevices[i], e = GetSubdeviceHierarchy(dev.Subdevices[i].ID)
	}

	return dev, ""
}

func GetNamedSubdevice1OfDevice(id int, sd, sd1 string) (*Subdevice1, string) {
	subdev, e := GetSubdeviceByNameAndParentID(id, sd)
	if e != "" {
		return nil, ""
	}

	subdev1, e := GetSubdevice1ByNameAndParentID(subdev.ID, sd1)
	if e != "" {
		return nil, ""
	}

	return subdev1, ""
}

func GetSubdevicesOfDevice(id int) ([]*Subdevice, string) {
	subdevs, e := GetSubdevicesOfParent(uint(id))
	if e != "" {
		return nil, ""
	}

	return subdevs, ""
}

func GetSubdevice1sUsingNamedSubdeviceOfDevice(id int, name string) ([]*Subdevice1, string) {
	subdev, e := GetSubdeviceByNameAndParentID(id, name)
	if e != "" {
		return nil, e
	}

	sd1s, e1 := GetSubdevices1OfParent(subdev.ID)
	if e1 != "" {
		return nil, e1
	}
	return sd1s, ""
}
