package models

import (
	"fmt"
	u "p3/utils"
)

type Rack_Attributes struct {
	ID          int    `json:"id" gorm:"column:id"`
	PosXY       string `json:"posXY" gorm:"column:rack_pos_x_y"`
	PosXYU      string `json:"posXYUnit" gorm:"column:rack_pos_x_y_unit"`
	PosZ        string `json:"posZ" gorm:"column:rack_pos_z"`
	PosZU       string `json:"posZUnit" gorm:"column:rack_pos_z_unit"`
	Template    string `json:"template" gorm:"column:rack_template"`
	Orientation string `json:"orientation" gorm:"column:rack_orientation"`
	Size        string `json:"size" gorm:"column:rack_size"`
	SizeU       string `json:"sizeUnit" gorm:"column:rack_size_unit"`
	Height      string `json:"height" gorm:"column:rack_height"`
	HeightU     string `json:"heightUnit" gorm:"column:rack_height_unit"`
	Vendor      string `json:"vendor" gorm:"column:rack_vendor"`
	Type        string `json:"type" gorm:"column:rack_type"`
	Model       string `json:"model" gorm:"column:rack_model"`
	Serial      string `json:"serial" gorm:"column:rack_serial"`
}

type Rack struct {
	//gorm.Model
	ID       int    `json:"id" gorm:"column:id"`
	Name     string `json:"name" gorm:"column:rack_name"`
	ParentID string `json:"parentId" gorm:"column:rack_parent_id"`
	Category string `json:"category" gorm:"-"`
	Domain   string `json:"domain" gorm:"column:rack_domain"`
	//D           []string        `json:"description" gorm:"-"`
	//Description string          `gorm:"-"`
	Attributes Rack_Attributes `json:"attributes"`

	//Site []Site
	//D is used to help the JSON marshalling
	//while Description will be used in
	//DB transactions
}

func (rack *Rack) Validate() (map[string]interface{}, bool) {
	if rack.Name == "" {
		return u.Message(false, "Rack Name should be on payload"), false
	}

	/*if rack.Category == "" {
		return u.Message(false, "Category should be on the payload"), false
	}

	if rack.Desc == "" {
		return u.Message(false, "Description should be on the payload"), false
	}*/

	if rack.Domain == "" {
		return u.Message(false, "Domain should should be on the payload"), false
	}

	if GetDB().Table("room").
		Where("id = ?", rack.ParentID).RecordNotFound() == true {

		return u.Message(false, "ParentID should be correspond to building ID"), false
	}

	if rack.Attributes.PosXY == "" {
		return u.Message(false, "XY coordinates should be on payload"), false
	}

	if rack.Attributes.PosXYU == "" {
		return u.Message(false, "PositionXYU string should be on the payload"), false
	}

	/*if rack.Attributes.PosZ == "" {
		return u.Message(false, "Z coordinates should be on payload"), false
	}

	if rack.Attributes.PosZU == "" {
		return u.Message(false, "PositionZU string should be on the payload"), false
	}*/

	/*if rack.Attributes.Template == "" {
		return u.Message(false, "Template should be on the payload"), false
	}*/

	switch rack.Attributes.Orientation {
	case "front", "rear", "left", "right":
	case "":
		return u.Message(false, "Orientation should be on the payload"), false

	default:
		return u.Message(false, "Orientation is invalid!"), false
	}

	if rack.Attributes.Size == "" {
		return u.Message(false, "Invalid size on the payload"), false
	}

	if rack.Attributes.SizeU == "" {
		return u.Message(false, "Rack size string should be on the payload"), false
	}

	if rack.Attributes.Height == "" {
		return u.Message(false, "Invalid Height on payload"), false
	}

	if rack.Attributes.HeightU == "" {
		return u.Message(false, "Rack Height string should be on the payload"), false
	}

	//Successfully validated Rack
	return u.Message(true, "success"), true
}

func (rack *Rack) Create() map[string]interface{} {
	if resp, ok := rack.Validate(); !ok {
		return resp
	}

	GetDB().Omit("rack_description").Create(rack)
	rack.Attributes.ID = rack.ID
	GetDB().Create(&(rack.Attributes))

	resp := u.Message(true, "success")
	resp["rack"] = rack
	return resp
}

//Get the rack using ID
func GetRack(id uint) *Rack {
	rack := &Rack{}
	err := GetDB().Table("rack").Where("id = ?", id).First(rack).
		Table("rack_attributes").Where("id = ?", id).First(&(rack.Attributes)).Error
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return rack
}

//Obtain all racks of a room
func GetRacks(room *Room) []*Rack {
	racks := make([]*Rack, 0)

	err := GetDB().Table("racks").Where("foreignkey = ?", room.ID).Find(&racks).Error
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return racks
}

//Obtain all racks
func GetAllRacks() []*Rack {
	racks := make([]*Rack, 0)

	err := GetDB().Table("racks").Find(&racks).Error
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return racks
}

//More methods should be made to
//Meet CRUD capabilities
//Need Update and Delete
//These would be a bit more complicated
//So leave them out for now

func UpdateRack(id uint, newRackInfo *Rack) map[string]interface{} {
	rack := &Rack{}

	err := GetDB().Table("racks").Where("id = ?", id).First(rack).Error
	if err != nil {
		return u.Message(false, "Rack was not found")
	}

	if newRackInfo.Name != "" && newRackInfo.Name != rack.Name {
		rack.Name = newRackInfo.Name
	}

	if newRackInfo.Category != "" && newRackInfo.Category != rack.Category {
		rack.Category = newRackInfo.Category
	}

	/*if newRackInfo.Desc != "" && newRackInfo.Desc != rack.Desc {
		rack.Desc = newRackInfo.Desc
	}*/

	//Should it be possible to update domain
	// Will have to think about it more
	//if newRackInfo.Domain

	/*if newRackInfo.Color != "" && newRackInfo.Color != rack.Color {
		rack.Color = newRackInfo.Color
	}

	if newRackInfo.Orientation != "" {
		switch newRackInfo.Orientation {
		case "NE", "NW", "SE", "SW":
			rack.Orientation = newRackInfo.Orientation

		default:
		}
	}*/

	//Successfully validated the new data
	GetDB().Table("racks").Save(rack)
	return u.Message(true, "success")
}

func DeleteRack(id uint) map[string]interface{} {

	//First check if the rack exists
	err := GetDB().Table("racks").Where("id = ?", id).First(&Rack{}).Error
	if err != nil {
		fmt.Println("Couldn't find the rack to delete")
		return nil
	}

	//This is a hard delete!
	e := GetDB().Unscoped().Table("racks").Delete(&Rack{}, id).Error

	//The command below is a soft delete
	//Meaning that the 'deleted_at' field will be set
	//the record will remain but unsearchable
	//e := GetDB().Table("tenants").Delete(Tenant{}, id).Error
	if e != nil {
		return u.Message(false, "There was an error in deleting the rack")
	}

	return u.Message(true, "success")
}
