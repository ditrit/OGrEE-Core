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
	Domain      string          `json:"domain"`
	Color       string          `json:"color"`
	Orientation ECardinalOrient `json:"eorientation"`
}

//Validate needs to ensure that the rack coords
//Are in the same bldg
//This is not yet implemented
func (rack *Rack) Validate() (map[string]interface{}, bool) {
	if rack.Name == "" {
		return u.Message(false, "Rack Name should be on payload"), false
	}

	if rack.Category == "" {
		return u.Message(false, "Category should be on the payload"), false
	}

	if rack.Desc == "" {
		return u.Message(false, "Description should be on the paylad"), false
	}

	if rack.Domain != "" {
		return u.Message(false, "Domain should NULL!"), false
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

//Get the first rack given the room
func GetRack(room *Room) *Rack {
	rack := &Rack{}
	err := GetDB().Table("racks").Where("foreignkey = ?", room.ID).First(rack).Error
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

//More methods should be made to
//Meet CRUD capabilities
//Need Update and Delete
//These would be a bit more complicated
//So leave them out for now
