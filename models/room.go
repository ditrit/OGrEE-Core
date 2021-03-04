package models

import (
	"fmt"
	u "p3/utils"
)

type Room_Attributes struct {
	ID          int    `json:"id" gorm:"column:id"`
	PosXY       string `json:"posXY" gorm:"column:room_pos_x_y"`
	PosXYU      string `json:"posXYUnit" gorm:"column:room_pos_x_y_unit"`
	PosZ        string `json:"posZ" gorm:"column:room_pos_z"`
	PosZU       string `json:"posZUnit" gorm:"column:room_pos_z_unit"`
	Template    string `json:"template" gorm:"column:room_template"`
	Orientation string `json:"orientation" gorm:"column:room_orientation"`
	Size        string `json:"size" gorm:"column:room_size"`
	SizeU       string `json:"sizeUnit" gorm:"column:room_size_unit"`
	Height      string `json:"height" gorm:"column:room_height"`
	HeightU     string `json:"heightUnit" gorm:"column:room_height_unit"`
}

type Room struct {
	//gorm.Model
	ID          int             `json:"id" gorm:"column:id"`
	Name        string          `json:"name" gorm:"column:room_name"`
	ParentID    string          `json:"parentId" gorm:"column:room_parent_id"`
	Category    string          `json:"category" gorm:"-"`
	Domain      string          `json:"domain" gorm:"column:room_domain"`
	D           []string        `json:"description" gorm:"-"`
	Description string          `gorm:"column:room_description"`
	Attributes  Room_Attributes `json:"attributes"`

	//Site []Site
	//D is used to help the JSON marshalling
	//while Description will be used in
	//DB transactions
}

//Validate needs to ensure that the room coords
//Are in the same room
//This is not yet implemented
func (room *Room) Validate() (map[string]interface{}, bool) {
	if room.Name == "" {
		return u.Message(false, "Room Name should be on payload"), false
	}

	/*if room.Category == "" {
		return u.Message(false, "Category should be on the payload"), false
	}

	if room.Desc == "" {
		return u.Message(false, "Description should be on the payload"), false
	}*/

	if room.Domain == "" {
		return u.Message(false, "Domain should should be on the payload"), false
	}

	if GetDB().Table("building").
		Where("id = ?", room.ParentID).RecordNotFound() == true {

		return u.Message(false, "ParentID should be correspond to building ID"), false
	}

	if room.Attributes.PosXY == "" {
		return u.Message(false, "XY coordinates should be on payload"), false
	}

	if room.Attributes.PosXYU == "" {
		return u.Message(false, "PositionXYU string should be on the payload"), false
	}

	if room.Attributes.PosZ == "" {
		return u.Message(false, "Z coordinates should be on payload"), false
	}

	if room.Attributes.PosZU == "" {
		return u.Message(false, "PositionZU string should be on the payload"), false
	}

	if room.Attributes.Template == "" {
		return u.Message(false, "Invalid building size on the payload"), false
	}

	switch room.Attributes.Orientation {
	case "NE", "NW", "SE", "SW":
	case "":
		return u.Message(false, "Orientation should be on the payload"), false

	default:
		return u.Message(false, "Orientation is invalid!"), false
	}

	if room.Attributes.Size == "" {
		return u.Message(false, "Invalid building size on the payload"), false
	}

	if room.Attributes.SizeU == "" {
		return u.Message(false, "Room size string should be on the payload"), false
	}

	if room.Attributes.Height == "" {
		return u.Message(false, "Invalid Height on payload"), false
	}

	if room.Attributes.HeightU == "" {
		return u.Message(false, "Room Height string should be on the payload"), false
	}

	//Successfully validated Room
	return u.Message(true, "success"), true
}

func (room *Room) Create() map[string]interface{} {
	if resp, ok := room.Validate(); !ok {
		return resp
	}

	//GetDB().Create(room)
	GetDB().Omit("room_description").Create(room)
	room.Attributes.ID = room.ID
	GetDB().Create(&(room.Attributes))

	resp := u.Message(true, "success")
	resp["room"] = room
	return resp
}

//Get the room by ID
func GetRoom(id uint) *Room {
	room := &Room{}
	err := GetDB().Table("room").Where("id = ?", id).First(room).
		Table("room_attributes").Where("id = ?", id).First(&(room.Attributes)).Error
	if err != nil {
		fmt.Println(err)
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
	attrs := make([]*Room_Attributes, 0)
	err := GetDB().Find(&rooms).Find(&attrs).Error
	if err != nil {
		fmt.Println(err)
		return nil
	}

	for i := range rooms {
		rooms[i].Attributes = *(attrs[i])
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

	/*if newRoomInfo.Desc != "" && newRoomInfo.Desc != room.Desc {
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
	}*/

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
