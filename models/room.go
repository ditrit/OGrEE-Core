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
//Are in the same room
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

//Obtain all rooms of a bldg
func GetRooms(room *Building) []*Room {
	rooms := make([]*Room, 0)

	err := GetDB().Table("rooms").Where("foreignkey = ?", room.ID).Find(&rooms).Error
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return rooms
}

//Get all rooms
func GetAllRooms() []*Room {
	rooms := make([]*Room, 0)

	err := GetDB().Table("rooms").Find(&rooms).Error
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

func UpdateRoom(id uint, newRoomInfo *Room) map[string]interface{} {
	room := &Room{}

	err := GetDB().Table("rooms").Where("id = ?", id).First(room).Error
	if err != nil {
		return u.Message(false, "Room was not found")
	}

	if newRoomInfo.Name != "" && newRoomInfo.Name != room.Name {
		room.Name = newRoomInfo.Name
	}

	if newRoomInfo.Category != "" && newRoomInfo.Category != room.Category {
		room.Category = newRoomInfo.Category
	}

	if newRoomInfo.Desc != "" && newRoomInfo.Desc != room.Desc {
		room.Desc = newRoomInfo.Desc
	}

	if newRoomInfo.PosX > 0.0 && newRoomInfo.PosX != room.PosX {
		room.PosX = newRoomInfo.PosX
	}

	if newRoomInfo.PosY > 0.0 && newRoomInfo.PosY != room.PosX {
		room.PosY = newRoomInfo.PosY
	}

	if newRoomInfo.PosU != "" && newRoomInfo.PosU != room.PosU {
		room.PosU = newRoomInfo.PosU
	}

	if newRoomInfo.PosZ > 0.0 && newRoomInfo.PosZ != room.PosZ {
		room.PosZ = newRoomInfo.PosZ
	}

	if newRoomInfo.PosZU != "" && newRoomInfo.PosZU != room.PosZU {
		room.PosZU = newRoomInfo.PosZU
	}

	if newRoomInfo.Size > 0.0 && newRoomInfo.Size != room.Size {
		room.Size = newRoomInfo.Size
	}

	if newRoomInfo.SizeU != "" && newRoomInfo.SizeU != room.SizeU {
		room.SizeU = newRoomInfo.SizeU
	}

	if newRoomInfo.Height > 0.0 && newRoomInfo.Height != room.Height {
		room.Height = newRoomInfo.Height
	}

	if newRoomInfo.HeightU != "" && newRoomInfo.HeightU != room.HeightU {
		room.HeightU = newRoomInfo.HeightU
	}

	if newRoomInfo.Orientation != "" {
		switch newRoomInfo.Orientation {
		case "NE", "NW", "SE", "SW":
			room.Orientation = newRoomInfo.Orientation

		default:
		}
	}

	GetDB().Table("rooms").Save(room)
	return u.Message(true, "success")
}

func DeleteRoom(id uint) map[string]interface{} {

	//First check if the site exists
	err := GetDB().Table("rooms").Where("id = ?", id).First(&Room{}).Error
	if err != nil {
		fmt.Println("Couldn't find the room to delete")
		return nil
	}

	//This is a hard delete!
	e := GetDB().Unscoped().Table("rooms").Delete(&Room{}, id).Error

	//The command below is a soft delete
	//Meaning that the 'deleted_at' field will be set
	//the record will remain but unsearchable
	//e := GetDB().Table("tenants").Delete(Tenant{}, id).Error
	if e != nil {
		return u.Message(false, "There was an error in deleting the room")
	}

	return u.Message(true, "success")
}
