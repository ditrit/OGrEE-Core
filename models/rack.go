package models

import (
	"fmt"
	u "p3/utils"

	"github.com/jinzhu/gorm"
)

type Rack struct {
	gorm.Model
	Name        string          `json:"name"`
	Category    string          `json:"category"`
	Desc        string          `json:"description"`
	Domain      int             `json:"domain"`
	Color       string          `json:"color"`
	Orientation ECardinalOrient `json:"eorientation"`
	Room        []Room          `gorm:"foreignKey:Room"`
}

func (rack *Rack) Validate() (map[string]interface{}, bool) {
	if rack.Name == "" {
		return u.Message(false, "Rack Name should be on payload"), false
	}

	if rack.Category == "" {
		return u.Message(false, "Category should be on the payload"), false
	}

	if rack.Desc == "" {
		return u.Message(false, "Description should be on the payload"), false
	}

	if rack.Domain == 0 {
		return u.Message(false, "Domain should should be on the payload"), false
	}

	if GetDB().Table("rooms").
		Where("id = ?", rack.Domain).First(&Room{}).Error != nil {

		return u.Message(false, "Domain should be correspond to Room ID"), false
	}

	if rack.Color == "" {
		return u.Message(false, "Color should be on the payload"), false
	}

	switch rack.Orientation {
	case "NE", "NW", "SE", "SW":
	case "":
		return u.Message(false, "Orientation should be on the payload"), false

	default:
		return u.Message(false, "Orientation is invalid!"), false
	}

	//Successfully validated Rack
	return u.Message(true, "success"), true
}

func (rack *Rack) Create() map[string]interface{} {
	if resp, ok := rack.Validate(); !ok {
		return resp
	}

	GetDB().Create(rack)

	resp := u.Message(true, "success")
	resp["rack"] = rack
	return resp
}

//Get the rack using ID
func GetRack(id uint) *Rack {
	rack := &Rack{}
	err := GetDB().Table("racks").Where("id = ?", id).First(rack).Error
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

	if newRackInfo.Desc != "" && newRackInfo.Desc != rack.Desc {
		rack.Desc = newRackInfo.Desc
	}

	//Should it be possible to update domain
	// Will have to think about it more
	//if newRackInfo.Domain

	if newRackInfo.Color != "" && newRackInfo.Color != rack.Color {
		rack.Color = newRackInfo.Color
	}

	if newRackInfo.Orientation != "" {
		switch newRackInfo.Orientation {
		case "NE", "NW", "SE", "SW":
			rack.Orientation = newRackInfo.Orientation

		default:
		}
	}

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
