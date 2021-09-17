package models

import (
	"fmt"
	u "p3/utils"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

type Subdevice1_Attributes struct {
	ID          int    `json:"-" bson:"id"`
	PosXY       string `json:"posXY" bson:"subdevice1_pos_x_y"`
	PosXYU      string `json:"posXYUnit" bson:"subdevice1_pos_x_y_unit"`
	PosZ        string `json:"posZ" bson:"subdevice1_pos_z"`
	PosZU       string `json:"posZUnit" bson:"subdevice1_pos_z_unit"`
	Template    string `json:"template" bson:"subdevice1_template"`
	Orientation string `json:"orientation" bson:"subdevice1_orientation"`
	Size        string `json:"size" bson:"subdevice1_size"`
	SizeUnit    string `json:"sizeUnit" bson:"subdevice1_size_unit"`
	SizeU       string `json:"sizeU" bson:"subdevice1_sizeu"`
	Slot        string `json:"slot" bson:"subdevice1_slot"`
	PosU        string `json:"posU" bson:"subdevice1_posu"`
	Height      string `json:"height" bson:"subdevice1_height"`
	HeightU     string `json:"heightUnit" bson:"subdevice1_height_unit"`
	Vendor      string `json:"vendor" bson:"subdevice1_vendor"`
	Type        string `json:"type" bson:"subdevice1_type"`
	Model       string `json:"model" bson:"subdevice1_model"`
	Serial      string `json:"serial" bson:"subdevice1_serial"`
}

type Subdevice1 struct {
	ID              int                   `json:"-" bson:"id"`
	IDJSON          string                `json:"id" bson:"-"`
	Name            string                `json:"name" bson:"subdevice1_name"`
	ParentID        string                `json:"parentId" bson:"subdevice1_parent_id"`
	Category        string                `json:"category" bson:"-"`
	Domain          string                `json:"domain" bson:"subdevice1_domain"`
	DescriptionJSON []string              `json:"description" bson:"-"`
	DescriptionDB   string                `json:"-" bson:"subdevice1_description"`
	Attributes      Subdevice1_Attributes `json:"attributes"`
}

func (subdevice1 *Subdevice1) Validate() (map[string]interface{}, bool) {
	if subdevice1.Name == "" {
		return u.Message(false, "Subdevice1 Name should be on payload"), false
	}

	if subdevice1.Category == "" {
		return u.Message(false, "Category should be on the payload"), false
	}

	if subdevice1.Domain == "" {
		return u.Message(false, "Domain should be on the payload"), false
	}

	if GetDB().Collection("subdevice").FindOne(GetCtx(), bson.M{"_id": subdevice1.ParentID}).Err() != nil {
		return u.Message(false, "ParentID should be correspond to Subdevice ID"), false
	}

	switch subdevice1.Attributes.Orientation {
	case "front", "rear", "frontflipped", "rearflipped":
	case "":
		return u.Message(false, "Orientation should be on the payload"), false

	default:
		return u.Message(false, "Orientation is invalid!"), false
	}

	if subdevice1.Attributes.Size == "" {
		return u.Message(false, "Invalid size on the payload"), false
	}

	if subdevice1.Attributes.SizeUnit == "" {
		return u.Message(false, "Subdevice size string should be on the payload"), false
	}

	if subdevice1.Attributes.Height == "" {
		return u.Message(false, "Invalid Height on payload"), false
	}

	if subdevice1.Attributes.HeightU == "" {
		return u.Message(false, "Subdevice Height string should be on the payload"), false
	}

	//Successfully validated SubDevice
	return u.Message(true, "success"), true
}

func (subdevice1 *Subdevice1) Create() (map[string]interface{}, string) {
	if resp, ok := subdevice1.Validate(); !ok {
		return resp, "validate"
	}

	subdevice1.DescriptionDB = strings.Join(subdevice1.DescriptionJSON, "XYZ")

	if _, e := GetDB().Collection("subdevice1").InsertOne(GetCtx(), subdevice1); e != nil {
		return u.Message(false, "Internal Error while creating Subdevice1: "+e.Error()),
			"internal"
	}
	subdevice1.IDJSON = strconv.Itoa(subdevice1.ID)
	/*subdevice1.Attributes.ID = subdevice1.ID
	if e := GetDB().Create(&(subdevice1.Attributes)).Error; e != nil {
		return u.Message(false, "Internal Error while creating Subdevice1 Attrs: "+e.Error()),
			"internal"
	}*/

	resp := u.Message(true, "success")
	resp["data"] = subdevice1
	return resp, ""
}

func (d *Subdevice1) FormQuery() string {

	query := "SELECT * FROM subdevice1 " + u.JoinQueryGen("subdevice1")
	if d.ParentID != "" {
		//pid, _ := strconv.Atoi(d.ParentID)
		query += " AND subdevice1_parent_id = " + d.ParentID
	}
	if d.IDJSON != "" {
		//id, _ := strconv.Atoi(d.IDJSON)
		query += " AND subdevice1.id = " + d.IDJSON
	}
	if d.ParentID != "" {
		//pid, _ := strconv.Atoi(d.ParentID)
		query += " AND subdevice1_parent_id = " + d.ParentID
	}
	if d.Name != "" {
		query += " WHERE subdevice1_name = '" + d.Name + "'"
	}
	if d.Category != "" {
		query += " AND subdevice1_category = '" + d.Category + "'"
	}
	if d.Domain != "" {
		query += " AND subdevice1_domain = '" + d.Domain + "'"
	}
	if (Subdevice1_Attributes{}) != d.Attributes {
		if d.Attributes.Template != "" {
			query +=
				" AND subdevice1_attributes.subdevice1_template = '" +
					d.Attributes.Template + "'"
		}
		if d.Attributes.Orientation != "" {
			query +=
				" AND subdevice1_attributes.subdevice1_orientation = '" +
					d.Attributes.Orientation + "'"
		}
		if d.Attributes.PosXY != "" {
			query +=
				" AND subdevice1_attributes.subdevice1_pos_x_y = '" +
					d.Attributes.PosXY + "'"
		}
		if d.Attributes.PosXYU != "" {
			query +=
				" AND subdevice1_attributes.subdevice1_pos_x_y_unit = '" +
					d.Attributes.PosXYU + "'"
		}
		if d.Attributes.PosZ != "" {
			query +=
				" AND subdevice1_attributes.subdevice1_pos_z = '" +
					d.Attributes.PosZ + "'"
		}
		if d.Attributes.PosZU != "" {
			query +=
				" AND subdevice1_attributes.subdevice1_pos_z_unit = '" +
					d.Attributes.PosZU + "'"
		}
		if d.Attributes.Size != "" {
			query +=
				" AND subdevice1_attributes.subdevice1_size = '" +
					d.Attributes.Size + "'"
		}
		if d.Attributes.SizeU != "" {
			query +=
				" AND subdevice1_attributes.subdevice1_sizeu = '" +
					d.Attributes.SizeU + "'"
		}
		if d.Attributes.SizeUnit != "" {
			query +=
				" AND subdevice1_attributes.subdevice1_size_unit = '" +
					d.Attributes.SizeUnit + "'"
		}
		if d.Attributes.Slot != "" {
			query +=
				" AND subdevice1_attributes.subdevice1_slot = '" +
					d.Attributes.Slot + "'"
		}
		if d.Attributes.PosU != "" {
			query +=
				" AND subdevice1_attributes.subdevice1_posu= '" +
					d.Attributes.PosU + "'"
		}
		if d.Attributes.Height != "" {
			query +=
				" AND subdevice1_attributes.subdevice1_height = '" +
					d.Attributes.Height + "'"
		}
		if d.Attributes.HeightU != "" {
			query +=
				" AND subdevice1_attributes.subdevice1_height_unit = '" +
					d.Attributes.HeightU + "'"
		}
		if d.Attributes.Vendor != "" {
			query +=
				" AND subdevice1_attributes.subdevice1_vendor = '" +
					d.Attributes.Vendor + "'"
		}
		if d.Attributes.Type != "" {
			query +=
				" AND subdevice1_attributes.subdevice1_type = '" +
					d.Attributes.Type + "'"
		}
		if d.Attributes.Model != "" {
			query +=
				" AND subdevice1_attributes.subdevice1_model = '" +
					d.Attributes.Model + "'"
		}
		if d.Attributes.Serial != "" {
			query +=
				" AND subdevice1_attributes.subdevice1_serial = '" +
					d.Attributes.Serial + "'"
		}
	}
	println(query)
	return query
}

func GetSubdevice1(id int) (*Subdevice1, string) {
	subdevice1 := &Subdevice1{}
	err := GetDB().Collection("subdevice1").FindOne(GetCtx(), bson.M{"id": id}).Decode(subdevice1)
	if err != nil {
		fmt.Println(err)
		return nil, err.Error()
	}
	subdevice1.DescriptionJSON = strings.Split(subdevice1.DescriptionDB, "XYZ")
	subdevice1.Category = "device"
	subdevice1.IDJSON = strconv.Itoa(subdevice1.ID)
	return subdevice1, ""
}

//Obtain all subdevices1 of a subdevice
func GetSubdevices1OfParent(id int) ([]*Subdevice1, string) {
	subdevices1 := make([]*Subdevice1, 0)
	c, err := GetDB().Collection("subdevice1").Find(GetCtx(), bson.M{"subdevice1_parent_id": id})
	if err != nil {
		fmt.Println(err)
		return nil, err.Error()
	}

	for c.Next(GetCtx()) {
		s := &Subdevice1{}
		if e := c.Decode(s); e != nil {
			fmt.Println(err)
			return nil, err.Error()
		}
		subdevices1 = append(subdevices1, s)
	}

	println("The length of subdevice1 is: ", len(subdevices1))
	/*for i := range subdevices1 {
		e := GetDB().Collection("subdevice1_attributes").Where("id = ?", subdevices1[i].ID).First(&(subdevices1[i].Attributes)).Error

		if e != nil {
			fmt.Println(err)
			return nil, err.Error()
		}

		subdevices1[i].Category = "device"
		subdevices1[i].DescriptionJSON = strings.Split(subdevices1[i].DescriptionDB, "XYZ")
		subdevices1[i].IDJSON = strconv.Itoa(subdevices1[i].ID)
	}*/

	return subdevices1, ""
}

func GetAllSubdevices1() ([]*Subdevice1, string) {
	subdevices1 := make([]*Subdevice1, 0)
	//attrs := make([]*Subdevice1_Attributes, 0)
	c, err := GetDB().Collection("subdevice1").Find(GetCtx(), bson.D{{}})
	if err != nil {
		fmt.Println(err)
		return nil, err.Error()
	}

	for c.Next(GetCtx()) {
		d := &Subdevice1{}
		e := c.Decode(d)
		if e != nil {
			fmt.Println(err)
			return nil, err.Error()
		}
		d.Category = "subdevice1"
		subdevices1 = append(subdevices1, d)
	}

	/*for i := range subdevices1 {
		subdevices1[i].Category = "device"
		subdevices1[i].Attributes = *(attrs[i])
		subdevices1[i].DescriptionJSON = strings.Split(subdevices1[i].DescriptionDB, "XYZ")
		subdevices1[i].IDJSON = strconv.Itoa(subdevices1[i].ID)
	}*/

	return subdevices1, ""
}

func GetSubdevice1ByQuery(subdevice1 *Subdevice1) ([]*Subdevice1, string) {
	/*subdevices1 := make([]*Subdevice1, 0)
	attrs := make([]*Subdevice1_Attributes, 0)

	e := GetDB().Raw(subdevice1.FormQuery()).Find(&subdevices1).
		Find(&attrs).Error

	if e != nil {
		return nil, e.Error()
	}

	for i := range subdevices1 {
		subdevices1[i].Attributes = *(attrs[i])
		subdevices1[i].IDJSON = strconv.Itoa(subdevices1[i].ID)
		subdevices1[i].DescriptionJSON =
			strings.Split(subdevices1[i].DescriptionDB, "XYZ")
		subdevices1[i].Category = "device"
	}

	return subdevices1, ""*/
	return nil, ""
}

func UpdateSubdevice1(id int, newSubdeviceInfo *map[string]interface{}) (map[string]interface{}, string) {
	/*subdevice1 := &Subdevice1{}

	err := GetDB().Collection("subdevice1").Where("id = ?", id).First(subdevice1).
		Table("subdevice1_attributes").Where("id = ?", id).First(&(subdevice1.Attributes)).Error
	if err != nil {
		return u.Message(false, "Error while checking subdevice1: "+err.Error()), err.Error()
	}

	if newSubdeviceInfo.Name != "" && newSubdeviceInfo.Name != subdevice1.Name {
		subdevice1.Name = newSubdeviceInfo.Name
	}

	if newSubdeviceInfo.Domain != "" && newSubdeviceInfo.Domain != subdevice1.Domain {
		subdevice1.Domain = newSubdeviceInfo.Domain
	}

	if dc := strings.Join(newSubdeviceInfo.DescriptionJSON, "XYZ"); dc != "" && strings.Compare(dc, subdevice1.DescriptionDB) != 0 {
		subdevice1.DescriptionDB = dc
	}

	if newSubdeviceInfo.Attributes.PosXY != "" && newSubdeviceInfo.Attributes.PosXY != subdevice1.Attributes.PosXY {
		subdevice1.Attributes.PosXY = newSubdeviceInfo.Attributes.PosXY
	}

	if newSubdeviceInfo.Attributes.PosXYU != "" && newSubdeviceInfo.Attributes.PosXYU != subdevice1.Attributes.PosXYU {
		subdevice1.Attributes.PosXYU = newSubdeviceInfo.Attributes.PosXYU
	}

	if newSubdeviceInfo.Attributes.PosZ != "" && newSubdeviceInfo.Attributes.PosZ != subdevice1.Attributes.PosZ {
		subdevice1.Attributes.PosZ = newSubdeviceInfo.Attributes.PosZ
	}

	if newSubdeviceInfo.Attributes.PosZU != "" && newSubdeviceInfo.Attributes.PosZU != subdevice1.Attributes.PosZU {
		subdevice1.Attributes.PosZU = newSubdeviceInfo.Attributes.PosZU
	}

	if newSubdeviceInfo.Attributes.Template != "" && newSubdeviceInfo.Attributes.Template != subdevice1.Attributes.Template {
		subdevice1.Attributes.Template = newSubdeviceInfo.Attributes.Template
	}

	if newSubdeviceInfo.Attributes.Orientation != "" {
		switch newSubdeviceInfo.Attributes.Orientation {
		case "front", "rear", "frontflipped", "rearflipped":
			subdevice1.Attributes.Orientation = newSubdeviceInfo.Attributes.Orientation

		default:
		}
	}

	if newSubdeviceInfo.Attributes.Size != "" && newSubdeviceInfo.Attributes.Size != subdevice1.Attributes.Size {
		subdevice1.Attributes.Size = newSubdeviceInfo.Attributes.Size
	}

	if newSubdeviceInfo.Attributes.SizeUnit != "" && newSubdeviceInfo.Attributes.SizeUnit != subdevice1.Attributes.SizeUnit {
		subdevice1.Attributes.SizeUnit = newSubdeviceInfo.Attributes.SizeUnit
	}

	if newSubdeviceInfo.Attributes.Height != "" && newSubdeviceInfo.Attributes.Height != subdevice1.Attributes.Height {
		subdevice1.Attributes.Height = newSubdeviceInfo.Attributes.Height
	}

	if newSubdeviceInfo.Attributes.HeightU != "" && newSubdeviceInfo.Attributes.HeightU != subdevice1.Attributes.HeightU {
		subdevice1.Attributes.HeightU = newSubdeviceInfo.Attributes.HeightU
	}

	if newSubdeviceInfo.Attributes.SizeU != "" && newSubdeviceInfo.Attributes.SizeU != subdevice1.Attributes.SizeU {
		subdevice1.Attributes.SizeU = newSubdeviceInfo.Attributes.SizeU
	}

	if newSubdeviceInfo.Attributes.PosU != "" && newSubdeviceInfo.Attributes.PosU != subdevice1.Attributes.PosU {
		subdevice1.Attributes.PosU = newSubdeviceInfo.Attributes.PosU
	}

	if newSubdeviceInfo.Attributes.Slot != "" && newSubdeviceInfo.Attributes.Slot != subdevice1.Attributes.Slot {
		subdevice1.Attributes.Slot = newSubdeviceInfo.Attributes.Slot
	}

	if newSubdeviceInfo.Attributes.Vendor != "" && newSubdeviceInfo.Attributes.Vendor != subdevice1.Attributes.Vendor {
		subdevice1.Attributes.Vendor = newSubdeviceInfo.Attributes.Vendor
	}

	if newSubdeviceInfo.Attributes.Type != "" && newSubdeviceInfo.Attributes.Type != subdevice1.Attributes.Type {
		subdevice1.Attributes.Type = newSubdeviceInfo.Attributes.Type
	}

	if newSubdeviceInfo.Attributes.Model != "" && newSubdeviceInfo.Attributes.Model != subdevice1.Attributes.Model {
		subdevice1.Attributes.Model = newSubdeviceInfo.Attributes.Model
	}

	if newSubdeviceInfo.Attributes.Serial != "" && newSubdeviceInfo.Attributes.Serial != subdevice1.Attributes.Serial {
		subdevice1.Attributes.Serial = newSubdeviceInfo.Attributes.Serial
	}

	//Successfully validated the new data
	if e1 := GetDB().Collection("subdevice1").Save(subdevice1).
		Table("subdevice1_attributes").Save(&(subdevice1.Attributes)).Error; e1 != nil {
		return u.Message(false, "Error while updating subdevice1: "+e1.Error()), e1.Error()
	}*/
	e := GetDB().Collection("subdevice1").FindOneAndUpdate(GetCtx(), bson.M{"_id": id}, bson.M{"$set": *newSubdeviceInfo}).Err()
	if e != nil {
		return u.Message(false, "failure: "+e.Error()), e.Error()
	}
	return u.Message(true, "success"), ""
}

func DeleteSubdevice1(id int) map[string]interface{} {

	//This is a hard delete!
	c, _ := GetDB().Collection("subdevice1").DeleteOne(GetCtx(), bson.M{"_id": id})

	if c.DeletedCount == 0 {
		return u.Message(false, "There was an error in deleting the subdevice1")
	}

	return u.Message(true, "success")
}

func GetSubdevice1ByName(name string) (*Subdevice1, string) {
	subdevice1 := &Subdevice1{}

	/*e := GetDB().Raw(`SELECT * FROM subdevice1
	JOIN subdevice1_attributes ON subdevice1.id = subdevice1_attributes.id
	WHERE subdevice1_name = ?;`, name).Find(subdevice1).Find(&subdevice1.Attributes).Error
	*/
	e := GetDB().Collection("subdevice1").FindOne(GetCtx(), bson.M{"subdevice1_name": name}).Decode(subdevice1)

	if e != nil {
		return nil, e.Error()
	}

	subdevice1.IDJSON = strconv.Itoa(subdevice1.ID)
	subdevice1.DescriptionJSON = strings.Split(subdevice1.DescriptionDB, "XYZ")
	subdevice1.Category = "device"
	return subdevice1, ""
}

func GetSubdevice1ByNameAndParentID(id int, name string) (*Subdevice1, string) {
	subdevice1 := &Subdevice1{}
	/*err := GetDB().Raw(`SELECT * FROM subdevice1 JOIN
	subdevice1_attributes ON subdevice1.id = subdevice1_attributes.id
	WHERE subdevice1_parent_id = ? AND subdevice1_name = ?`, id, name).
	Find(subdevice1).Find(&(subdevice1.Attributes)).Error*/

	err := GetDB().Collection("subdevice1").FindOne(GetCtx(), bson.M{"subdevice1_parent_id": id, "subdevice1_name": name}).Decode(subdevice1)
	if err != nil {
		return nil, "record not found"
	}

	subdevice1.DescriptionJSON = strings.Split(subdevice1.DescriptionDB, "XYZ")
	subdevice1.Category = "device"
	subdevice1.IDJSON = strconv.Itoa(subdevice1.ID)
	return subdevice1, ""
}
