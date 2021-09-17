package models

import (
	"fmt"
	u "p3/utils"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

type Subdevice_Attributes struct {
	ID          int    `json:"-" bson:"id"`
	PosXY       string `json:"posXY" bson:"subdevice_pos_x_y"`
	PosXYU      string `json:"posXYUnit" bson:"subdevice_pos_x_y_unit"`
	PosZ        string `json:"posZ" bson:"subdevice_pos_z"`
	PosZU       string `json:"posZUnit" bson:"subdevice_pos_z_unit"`
	Template    string `json:"template" bson:"subdevice_template"`
	Orientation string `json:"orientation" bson:"subdevice_orientation"`
	Size        string `json:"size" bson:"subdevice_size"`
	SizeUnit    string `json:"sizeUnit" bson:"subdevice_size_unit"`
	SizeU       string `json:"sizeU" bson:"subdevice_sizeu"`
	Slot        string `json:"slot" bson:"subdevice_slot"`
	PosU        string `json:"posU" bson:"subdevice_posu"`
	Height      string `json:"height" bson:"subdevice_height"`
	HeightU     string `json:"heightUnit" bson:"subdevice_height_unit"`
	Vendor      string `json:"vendor" bson:"subdevice_vendor"`
	Type        string `json:"type" bson:"subdevice_type"`
	Model       string `json:"model" bson:"subdevice_model"`
	Serial      string `json:"serial" bson:"subdevice_serial"`
}

type Subdevice struct {
	ID              int                  `json:"-" bson:"id"`
	IDJSON          string               `json:"id" bson:"-"`
	Name            string               `json:"name" bson:"subdevice_name"`
	ParentID        string               `json:"parentId" bson:"subdevice_parent_id"`
	Category        string               `json:"category" bson:"-"`
	Domain          string               `json:"domain" bson:"subdevice_domain"`
	DescriptionJSON []string             `json:"description" bson:"-"`
	DescriptionDB   string               `json:"-" bson:"subdevice_description"`
	Attributes      Subdevice_Attributes `json:"attributes"`

	Subdevs1 []*Subdevice1 `json:"subdevices1,omitempty", bson:"-"`
}

func (subdevice *Subdevice) Validate() (map[string]interface{}, bool) {
	if subdevice.Name == "" {
		return u.Message(false, "Subdevice Name should be on payload"), false
	}

	if subdevice.Category == "" {
		return u.Message(false, "Category should be on the payload"), false
	}

	if subdevice.Domain == "" {
		return u.Message(false, "Domain should be on the payload"), false
	}

	if GetDB().Collection("device").FindOne(GetCtx(), bson.M{"_id": subdevice.ParentID}).Err() != nil {
		return u.Message(false, "Domain should be correspond to Device ID"), false
	}

	switch subdevice.Attributes.Orientation {
	case "front", "rear", "frontflipped", "rearflipped":
	case "":
		return u.Message(false, "Orientation should be on the payload"), false

	default:
		return u.Message(false, "Orientation is invalid!"), false
	}

	if subdevice.Attributes.Size == "" {
		return u.Message(false, "Invalid size on the payload"), false
	}

	if subdevice.Attributes.SizeUnit == "" {
		return u.Message(false, "Subdevice size string should be on the payload"), false
	}

	if subdevice.Attributes.Height == "" {
		return u.Message(false, "Invalid Height on payload"), false
	}

	if subdevice.Attributes.HeightU == "" {
		return u.Message(false, "Subdevice Height string should be on the payload"), false
	}

	//Successfully validated SubDevice
	return u.Message(true, "success"), true
}

func (subdevice *Subdevice) Create() (map[string]interface{}, string) {
	if resp, ok := subdevice.Validate(); !ok {
		return resp, "validate"
	}

	subdevice.DescriptionDB = strings.Join(subdevice.DescriptionJSON, "XYZ")

	if _, e := GetDB().Collection("subdevice").InsertOne(GetCtx(), subdevice); e != nil {
		return u.Message(false, "Internal Error while creating Subdevice: "+e.Error()),
			"internal"
	}
	subdevice.IDJSON = strconv.Itoa(subdevice.ID)
	/*subdevice.Attributes.ID = subdevice.ID
	if e := GetDB().Create(&(subdevice.Attributes)).Error; e != nil {
		return u.Message(false, "Internal Error while creating Subdevice Attrs: "+e.Error()),
			"internal"
	}*/

	resp := u.Message(true, "success")
	resp["data"] = subdevice
	return resp, ""
}

func (d *Subdevice) FormQuery() string {

	query := "SELECT * FROM subdevice " + u.JoinQueryGen("subdevice")
	if d.ParentID != "" {
		//pid, _ := strconv.Atoi(d.ParentID)
		query += " AND subdevice_parent_id = " + d.ParentID
	}
	if d.IDJSON != "" {
		//id, _ := strconv.Atoi(d.IDJSON)
		query += " AND subdevice.id = " + d.IDJSON
	}
	if d.ParentID != "" {
		//pid, _ := strconv.Atoi(d.ParentID)
		query += " AND subdevice_parent_id = " + d.ParentID
	}
	if d.Name != "" {
		query += " WHERE subdevice_name = '" + d.Name + "'"
	}
	if d.Category != "" {
		query += " AND subdevice_category = '" + d.Category + "'"
	}
	if d.Domain != "" {
		query += " AND subdevice_domain = '" + d.Domain + "'"
	}
	if (Subdevice_Attributes{}) != d.Attributes {
		if d.Attributes.Template != "" {
			query +=
				" AND subdevice_attributes.subdevice_template = '" +
					d.Attributes.Template + "'"
		}
		if d.Attributes.Orientation != "" {
			query +=
				" AND subdevice_attributes.subdevice_orientation = '" +
					d.Attributes.Orientation + "'"
		}
		if d.Attributes.PosXY != "" {
			query +=
				" AND subdevice_attributes.subdevice_pos_x_y = '" +
					d.Attributes.PosXY + "'"
		}
		if d.Attributes.PosXYU != "" {
			query +=
				" AND subdevice_attributes.subdevice_pos_x_y_unit = '" +
					d.Attributes.PosXYU + "'"
		}
		if d.Attributes.PosZ != "" {
			query +=
				" AND subdevice_attributes.subdevice_pos_z = '" +
					d.Attributes.PosZ + "'"
		}
		if d.Attributes.PosZU != "" {
			query +=
				" AND subdevice_attributes.subdevice_pos_z_unit = '" +
					d.Attributes.PosZU + "'"
		}
		if d.Attributes.Size != "" {
			query +=
				" AND subdevice_attributes.subdevice_size = '" +
					d.Attributes.Size + "'"
		}
		if d.Attributes.SizeU != "" {
			query +=
				" AND subdevice_attributes.subdevice_sizeu = '" +
					d.Attributes.SizeU + "'"
		}
		if d.Attributes.SizeUnit != "" {
			query +=
				" AND subdevice_attributes.subdevice_size_unit = '" +
					d.Attributes.SizeUnit + "'"
		}
		if d.Attributes.Slot != "" {
			query +=
				" AND subdevice_attributes.subdevice_slot = '" +
					d.Attributes.Slot + "'"
		}
		if d.Attributes.PosU != "" {
			query +=
				" AND subdevice_attributes.subdevice_posu= '" +
					d.Attributes.PosU + "'"
		}
		if d.Attributes.Height != "" {
			query +=
				" AND subdevice_attributes.subdevice_height = '" +
					d.Attributes.Height + "'"
		}
		if d.Attributes.HeightU != "" {
			query +=
				" AND subdevice_attributes.subdevice_height_unit = '" +
					d.Attributes.HeightU + "'"
		}
		if d.Attributes.Vendor != "" {
			query +=
				" AND subdevice_attributes.subdevice_vendor = '" +
					d.Attributes.Vendor + "'"
		}
		if d.Attributes.Type != "" {
			query +=
				" AND subdevice_attributes.subdevice_type = '" +
					d.Attributes.Type + "'"
		}
		if d.Attributes.Model != "" {
			query +=
				" AND subdevice_attributes.subdevice_model = '" +
					d.Attributes.Model + "'"
		}
		if d.Attributes.Serial != "" {
			query +=
				" AND subdevice_attributes.subdevice_serial = '" +
					d.Attributes.Serial + "'"
		}
	}
	println(query)
	return query
}

//Get the subdevice given the ID
func GetSubdevice(id uint) (*Subdevice, string) {
	subdevice := &Subdevice{}
	/*err := GetDB().Collection("subdevice").Where("id = ?", id).First(subdevice).
	Table("subdevice_attributes").Where("id = ?", id).First(&(subdevice.Attributes)).Error
	*/
	err := GetDB().Collection("subdevice").FindOne(GetCtx(), bson.M{"_id": id}).Decode(subdevice)
	if err != nil {
		fmt.Println(err)
		return nil, err.Error()
	}
	subdevice.DescriptionJSON = strings.Split(subdevice.DescriptionDB, "XYZ")
	subdevice.Category = "device"
	subdevice.IDJSON = strconv.Itoa(subdevice.ID)
	return subdevice, ""
}

//Obtain all subdevices of a device
func GetSubdevicesOfParent(id uint) ([]*Subdevice, string) {
	subdevices := make([]*Subdevice, 0)
	c, err := GetDB().Collection("subdevice").Find(GetCtx(), bson.M{"subdevice_parent_id": id})
	if err != nil {
		fmt.Println(err)
		return nil, err.Error()
	}

	for c.Next(GetCtx()) {
		s := &Subdevice{}
		e := c.Decode(s)
		if e != nil {
			fmt.Println(err)
			return nil, err.Error()
		}
		subdevices = append(subdevices, s)
	}

	println("The length of subdevice is: ", len(subdevices))
	/*for i := range subdevices {
		e := GetDB().Collection("subdevice_attributes").Where("id = ?", subdevices[i].ID).First(&(subdevices[i].Attributes)).Error

		if e != nil {
			fmt.Println(err)
			return nil, err.Error()
		}

		subdevices[i].Category = "device"
		subdevices[i].DescriptionJSON = strings.Split(subdevices[i].DescriptionDB, "XYZ")
		subdevices[i].IDJSON = strconv.Itoa(subdevices[i].ID)
	}*/

	return subdevices, ""
}

func GetAllSubdevices() ([]*Subdevice, string) {
	subdevices := make([]*Subdevice, 0)
	//attrs := make([]*Subdevice_Attributes, 0)
	c, err := GetDB().Collection("subdevice").Find(GetCtx(), bson.D{{}})
	if err != nil {
		fmt.Println(err)
		return nil, err.Error()
	}

	for c.Next(GetCtx()) {
		s := &Subdevice{}
		if e := c.Decode(s); e != nil {
			fmt.Println(err)
			return nil, err.Error()
		}
		subdevices = append(subdevices, s)
	}

	/*for i := range subdevices {
		subdevices[i].Category = "device"
		subdevices[i].Attributes = *(attrs[i])
		subdevices[i].DescriptionJSON = strings.Split(subdevices[i].DescriptionDB, "XYZ")
		subdevices[i].IDJSON = strconv.Itoa(subdevices[i].ID)
	}*/

	return subdevices, ""
}

func GetSubdeviceByQuery(subdevice *Subdevice) ([]*Subdevice, string) {
	/*subdevices := make([]*Subdevice, 0)
	attrs := make([]*Subdevice_Attributes, 0)

	e := GetDB().Raw(subdevice.FormQuery()).Find(&subdevices).
		Find(&attrs).Error

	if e != nil {
		return nil, e.Error()
	}

	for i := range subdevices {
		subdevices[i].Attributes = *(attrs[i])
		subdevices[i].IDJSON = strconv.Itoa(subdevices[i].ID)
		subdevices[i].DescriptionJSON =
			strings.Split(subdevices[i].DescriptionDB, "XYZ")
		subdevices[i].Category = "device"
	}

	return subdevices, ""*/
	return nil, ""
}

func UpdateSubdevice(id uint, newSubdeviceInfo *map[string]interface{}) (map[string]interface{}, string) {
	/*subdevice := &Subdevice{}

	err := GetDB().Collection("subdevice").Where("id = ?", id).First(subdevice).
		Table("subdevice_attributes").Where("id = ?", id).First(&(subdevice.Attributes)).Error
	if err != nil {
		return u.Message(false, "Error while checking subdevice: "+err.Error()), err.Error()
	}

	if newSubdeviceInfo.Name != "" && newSubdeviceInfo.Name != subdevice.Name {
		subdevice.Name = newSubdeviceInfo.Name
	}

	if newSubdeviceInfo.Domain != "" && newSubdeviceInfo.Domain != subdevice.Domain {
		subdevice.Domain = newSubdeviceInfo.Domain
	}

	if dc := strings.Join(newSubdeviceInfo.DescriptionJSON, "XYZ"); dc != "" && strings.Compare(dc, subdevice.DescriptionDB) != 0 {
		subdevice.DescriptionDB = dc
	}

	if newSubdeviceInfo.Attributes.PosXY != "" && newSubdeviceInfo.Attributes.PosXY != subdevice.Attributes.PosXY {
		subdevice.Attributes.PosXY = newSubdeviceInfo.Attributes.PosXY
	}

	if newSubdeviceInfo.Attributes.PosXYU != "" && newSubdeviceInfo.Attributes.PosXYU != subdevice.Attributes.PosXYU {
		subdevice.Attributes.PosXYU = newSubdeviceInfo.Attributes.PosXYU
	}

	if newSubdeviceInfo.Attributes.PosZ != "" && newSubdeviceInfo.Attributes.PosZ != subdevice.Attributes.PosZ {
		subdevice.Attributes.PosZ = newSubdeviceInfo.Attributes.PosZ
	}

	if newSubdeviceInfo.Attributes.PosZU != "" && newSubdeviceInfo.Attributes.PosZU != subdevice.Attributes.PosZU {
		subdevice.Attributes.PosZU = newSubdeviceInfo.Attributes.PosZU
	}

	if newSubdeviceInfo.Attributes.Template != "" && newSubdeviceInfo.Attributes.Template != subdevice.Attributes.Template {
		subdevice.Attributes.Template = newSubdeviceInfo.Attributes.Template
	}

	if newSubdeviceInfo.Attributes.Orientation != "" {
		switch newSubdeviceInfo.Attributes.Orientation {
		case "front", "rear", "frontflipped", "rearflipped":
			subdevice.Attributes.Orientation = newSubdeviceInfo.Attributes.Orientation

		default:
		}
	}

	if newSubdeviceInfo.Attributes.Size != "" && newSubdeviceInfo.Attributes.Size != subdevice.Attributes.Size {
		subdevice.Attributes.Size = newSubdeviceInfo.Attributes.Size
	}

	if newSubdeviceInfo.Attributes.SizeUnit != "" && newSubdeviceInfo.Attributes.SizeUnit != subdevice.Attributes.SizeUnit {
		subdevice.Attributes.SizeUnit = newSubdeviceInfo.Attributes.SizeUnit
	}

	if newSubdeviceInfo.Attributes.Height != "" && newSubdeviceInfo.Attributes.Height != subdevice.Attributes.Height {
		subdevice.Attributes.Height = newSubdeviceInfo.Attributes.Height
	}

	if newSubdeviceInfo.Attributes.HeightU != "" && newSubdeviceInfo.Attributes.HeightU != subdevice.Attributes.HeightU {
		subdevice.Attributes.HeightU = newSubdeviceInfo.Attributes.HeightU
	}

	if newSubdeviceInfo.Attributes.SizeU != "" && newSubdeviceInfo.Attributes.SizeU != subdevice.Attributes.SizeU {
		subdevice.Attributes.SizeU = newSubdeviceInfo.Attributes.SizeU
	}

	if newSubdeviceInfo.Attributes.PosU != "" && newSubdeviceInfo.Attributes.PosU != subdevice.Attributes.PosU {
		subdevice.Attributes.PosU = newSubdeviceInfo.Attributes.PosU
	}

	if newSubdeviceInfo.Attributes.Slot != "" && newSubdeviceInfo.Attributes.Slot != subdevice.Attributes.Slot {
		subdevice.Attributes.Slot = newSubdeviceInfo.Attributes.Slot
	}

	if newSubdeviceInfo.Attributes.Vendor != "" && newSubdeviceInfo.Attributes.Vendor != subdevice.Attributes.Vendor {
		subdevice.Attributes.Vendor = newSubdeviceInfo.Attributes.Vendor
	}

	if newSubdeviceInfo.Attributes.Type != "" && newSubdeviceInfo.Attributes.Type != subdevice.Attributes.Type {
		subdevice.Attributes.Type = newSubdeviceInfo.Attributes.Type
	}

	if newSubdeviceInfo.Attributes.Model != "" && newSubdeviceInfo.Attributes.Model != subdevice.Attributes.Model {
		subdevice.Attributes.Model = newSubdeviceInfo.Attributes.Model
	}

	if newSubdeviceInfo.Attributes.Serial != "" && newSubdeviceInfo.Attributes.Serial != subdevice.Attributes.Serial {
		subdevice.Attributes.Serial = newSubdeviceInfo.Attributes.Serial
	}

	//Successfully validated the new data
	if e1 := GetDB().Collection("subdevice").Save(subdevice).
		Table("subdevice_attributes").Save(&(subdevice.Attributes)).Error; e1 != nil {
		return u.Message(false, "Error while updating subdevice: "+e1.Error()), e1.Error()
	}*/

	e := GetDB().Collection("subdevice").FindOneAndUpdate(GetCtx(), bson.M{"_id": id}, bson.M{"$set": *newSubdeviceInfo}).Err()
	if e != nil {
		return u.Message(false, "failure: "+e.Error()), e.Error()
	}
	return u.Message(true, "success"), ""
}

func DeleteSubdevice(id uint) map[string]interface{} {

	//This is a hard delete!
	c, _ := GetDB().Collection("subdevice").DeleteOne(GetCtx(), bson.M{"_id": id})

	if c.DeletedCount == 0 {
		return u.Message(false, "There was an error in deleting the subdevice")
	}

	return u.Message(true, "success")
}

func GetSubdeviceByName(name string) (*Subdevice, string) {
	subdevice := &Subdevice{}

	/*e := GetDB().Raw(`SELECT * FROM subdevice
	JOIN subdevice_attributes ON subdevice.id = subdevice_attributes.id
	WHERE subdevice_name = ?;`, name).Find(subdevice).Find(&subdevice.Attributes).Error
	*/
	e := GetDB().Collection("subdevice").FindOne(GetCtx(), bson.M{"subdevice_name": name}).Decode(subdevice)
	if e != nil {
		return nil, e.Error()
	}

	subdevice.IDJSON = strconv.Itoa(subdevice.ID)
	subdevice.DescriptionJSON = strings.Split(subdevice.DescriptionDB, "XYZ")
	subdevice.Category = "device"
	return subdevice, ""
}

func GetSubdeviceByNameAndParentID(id int, name string) (*Subdevice, string) {
	subdevice := &Subdevice{}
	/*err := GetDB().Raw(`SELECT * FROM subdevice JOIN
	subdevice_attributes ON subdevice.id = subdevice_attributes.id
	WHERE subdevice_parent_id = ? AND subdevice_name = ?`, id, name).
	Find(subdevice).Find(&(subdevice.Attributes)).Error*/
	err := GetDB().Collection("subdevice").FindOne(GetCtx(), bson.M{"subdevice_parent_id": id, "subdevice_name": name}).Decode(subdevice)

	if err != nil {
		return nil, "record not found"
	}

	subdevice.DescriptionJSON = strings.Split(subdevice.DescriptionDB, "XYZ")
	subdevice.Category = "device"
	subdevice.IDJSON = strconv.Itoa(subdevice.ID)
	return subdevice, ""
}

func GetSubdeviceHierarchy(id int) (*Subdevice, string) {
	subdev, e := GetSubdevice(uint(id))
	if e != "" {
		return nil, e
	}

	subdev.Subdevs1, e = GetSubdevices1OfParent(id)
	if e != "" {
		return nil, e
	}
	return subdev, ""
}
