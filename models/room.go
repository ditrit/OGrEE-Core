package models

import (
	"fmt"
	u "p3/utils"

	"github.com/jinzhu/gorm"
)

type Room struct {
	gorm.Model
	Name        string          `json:"name"`
	Category    string          `json:"category"`
	Desc        string          `json:"description"`
	Domain      int             `json:"domain"`
	Orientation ECardinalOrient `json:"eorientation"`

	PosX     float64    `json:"posx"`
	PosY     float64    `json:"posy"`
	PosU     string     `json:"posxyu"`
	PosZ     float64    `json:"posz"`
	PosZU    string     `json:"poszu"`
	Size     float64    `json:"size"`
	SizeU    string     `json:"sizeu"`
	Height   float64    `json:"height"`
	HeightU  string     `json:"heightu"`
	Building []Building `gorm:"foreignKey:Building"`
}

//Validate needs to ensure that the room coords
//Are in the same bldg
//This is not yet implemented
func (room *Room) Validate() (map[string]interface{}, bool) {
	if room.Name == "" {
		return u.Message(false, "Room Name should be on payload"), false
	}

	if room.Category == "" {
		return u.Message(false, "Category should be on the payload"), false
	}

	if room.Desc == "" {
		return u.Message(false, "Description should be on the payload"), false
	}

	if room.Domain == 0 {
		return u.Message(false, "Domain should should be on the payload"), false
	}

	if GetDB().Table("buildings").
		Where("id = ?", room.Domain).First(&Building{}).Error != nil {

		return u.Message(false, "Domain should be correspond to building ID"), false
	}

	if room.PosX < 0.0 || room.PosY < 0.0 {
		return u.Message(false, "Invalid XYcoordinates on payload"), false
	}

	if room.PosU == "" {
		return u.Message(false, "PositionXY string should be on the payload"), false
	}

	if room.PosZ < 0.0 {
		return u.Message(false, "Invalid Z coordinates on payload"), false
	}

	if room.PosZU == "" {
		return u.Message(false, "PositionZ string should be on the payload"), false
	}

	if room.Size <= 0.0 {
		return u.Message(false, "Invalid room size on the payload"), false
	}

	if room.SizeU == "" {
		return u.Message(false, "Room size string should be on the payload"), false
	}

	if room.Height <= 0.0 {
		return u.Message(false, "Invalid Height on payload"), false
	}

	if room.HeightU == "" {
		return u.Message(false, "Room Height string should be on the payload"), false
	}

	switch room.Orientation {
	case "NE", "NW", "SE", "SW":
	case "":
		return u.Message(false, "Orientation should be on the payload"), false

	default:
		return u.Message(false, "Orientation is invalid!"), false
	}

	//Successfully validated Room
	return u.Message(true, "success"), true
}

func (room *Room) Create() map[string]interface{} {
	if resp, ok := room.Validate(); !ok {
		return resp
	}

	GetDB().Create(room)

	resp := u.Message(true, "success")
	resp["room"] = room
	return resp
}

//Get the room by ID
func GetRoom(id uint) *Room {
	room := &Room{}
	err := GetDB().Table("rooms").Where("id = ?", id).First(room).Error
	if err != nil {
		fmt.Println("There was an error in getting room by ID")
		return nil
	}
	return room
}

//Obtain all rooms of a site
func GetRooms(bldg *Building) []*Room {
	rooms := make([]*Room, 0)

	err := GetDB().Table("rooms").Where("foreignkey = ?", bldg.ID).Find(&rooms).Error
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return rooms
}

//More methods should be made to
//Meet CRUD capabilities
//Need Update and Delete
//These would be a bit more complicated
//So leave them out for now
