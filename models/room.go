package models

import (
	"fmt"
	u "p3/utils"
	"strings"
)

type Room_Attributes struct {
	ID          int    `json:"-" gorm:"column:id"`
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
	ID              int             `json:"id" gorm:"column:id"`
	Name            string          `json:"name" gorm:"column:room_name"`
	ParentID        string          `json:"parentId" gorm:"column:room_parent_id"`
	Category        string          `json:"category" gorm:"-"`
	Domain          string          `json:"domain" gorm:"column:room_domain"`
	DescriptionJSON []string        `json:"description" gorm:"-"`
	DescriptionDB   string          `json:"-" gorm:"column:room_description"`
	Attributes      Room_Attributes `json:"attributes"`

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

	if room.Category == "" {
		return u.Message(false, "Category should be on the payload"), false
	}

	if room.Domain == "" {
		return u.Message(false, "Domain should should be on the payload"), false
	}

	if GetDB().Table("building").
		Where("id = ?", room.ParentID).First(&Building{}).Error != nil {

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

	switch room.Attributes.Orientation {
	case "-E-N", "-E+N", "+E-N", "+E+N":
	case "-N-W", "-N+W", "+N-W", "+N+W":
	case "-W-S", "-W+S", "+W-S", "+W+S":
	case "-S-E", "-S+E", "+S-E", "+S+E":
	case "":
		return u.Message(false, "Orientation should be on the payload"), false

	default:
		return u.Message(false, "Orientation is invalid!"), false
	}

	if room.Attributes.Size == "" {
		return u.Message(false, "Invalid size on the payload"), false
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

func (room *Room) Create() (map[string]interface{}, string) {
	if resp, ok := room.Validate(); !ok {
		return resp, "validate"
	}

	room.DescriptionDB = strings.Join(room.DescriptionJSON, "XYZ")

	if e := GetDB().Create(room).Error; e != nil {
		return u.Message(false, "Internal Error while creating Room: "+e.Error()),
			e.Error()
	}
	room.Attributes.ID = room.ID
	if e := GetDB().Create(&(room.Attributes)).Error; e != nil {
		return u.Message(false, "Internal Error while creating Room Attrs: "+e.Error()),
			e.Error()
	}

	resp := u.Message(true, "success")
	resp["room"] = room
	return resp, ""
}

//Get the room by ID
func GetRoom(id uint) (*Room, string) {
	room := &Room{}
	err := GetDB().Table("room").Where("id = ?", id).First(room).
		Table("room_attributes").Where("id = ?", id).First(&(room.Attributes)).Error
	if err != nil {
		fmt.Println(err)
		return nil, err.Error()
	}

	room.DescriptionJSON = strings.Split(room.DescriptionDB, "XYZ")
	room.Category = "room"
	return room, ""
}

//Obtain all rooms of a bldg
func GetRooms(room *Building) ([]*Room, string) {
	rooms := make([]*Room, 0)

	err := GetDB().Table("rooms").Where("foreignkey = ?", room.ID).Find(&rooms).Error
	if err != nil {
		fmt.Println(err)
		return nil, err.Error()
	}

	return rooms, ""
}

//Get all rooms
func GetAllRooms() ([]*Room, string) {
	rooms := make([]*Room, 0)
	attrs := make([]*Room_Attributes, 0)
	err := GetDB().Find(&rooms).Find(&attrs).Error
	if err != nil {
		fmt.Println(err)
		return nil, err.Error()
	}

	for i := range rooms {
		rooms[i].Category = "room"
		rooms[i].Attributes = *(attrs[i])
		rooms[i].DescriptionJSON = strings.Split(rooms[i].DescriptionDB, "XYZ")
	}

	return rooms, ""
}

func UpdateRoom(id uint, newRoomInfo *Room) (map[string]interface{}, string) {
	room := &Room{}

	err := GetDB().Table("room").Where("id = ?", id).First(room).
		Table("room_attributes").Where("id = ?", id).First(&(room.Attributes)).Error
	if err != nil {
		return u.Message(false, "Error while checking Room: "+err.Error()), err.Error()
	}

	if newRoomInfo.Name != "" && newRoomInfo.Name != room.Name {
		room.Name = newRoomInfo.Name
	}

	if newRoomInfo.Domain != "" && newRoomInfo.Domain != room.Domain {
		room.Domain = newRoomInfo.Domain
	}

	if dc := strings.Join(newRoomInfo.DescriptionJSON, "XYZ"); dc != "" && strings.Compare(dc, room.DescriptionDB) != 0 {
		room.DescriptionDB = dc
	}

	/*if newRoomInfo.Category != "" && newRoomInfo.Category != room.Category {
		room.Category = newRoomInfo.Category
	}*/

	if newRoomInfo.Attributes.PosXY != "" && newRoomInfo.Attributes.PosXY != room.Attributes.PosXY {
		room.Attributes.PosXY = newRoomInfo.Attributes.PosXY
	}

	if newRoomInfo.Attributes.PosXYU != "" && newRoomInfo.Attributes.PosXYU != room.Attributes.PosXYU {
		room.Attributes.PosXYU = newRoomInfo.Attributes.PosXYU
	}

	if newRoomInfo.Attributes.PosZ != "" && newRoomInfo.Attributes.PosZ != room.Attributes.PosZ {
		room.Attributes.PosZ = newRoomInfo.Attributes.PosZ
	}

	if newRoomInfo.Attributes.PosZU != "" && newRoomInfo.Attributes.PosZU != room.Attributes.PosZU {
		room.Attributes.PosZU = newRoomInfo.Attributes.PosZU
	}

	if newRoomInfo.Attributes.Template != "" && newRoomInfo.Attributes.Template != room.Attributes.Template {
		room.Attributes.Template = newRoomInfo.Attributes.Template
	}

	if newRoomInfo.Attributes.Orientation != "" {
		switch newRoomInfo.Attributes.Orientation {
		case "-E-N", "-E+N", "+E-N", "+E+N":
		case "-N-W", "-N+W", "+N-W", "+N+W":
		case "-W-S", "-W+S", "+W-S", "+W+S":
		case "-S-E", "-S+E", "+S-E", "+S+E":
			room.Attributes.Orientation = newRoomInfo.Attributes.Orientation

		default:
		}
	}

	if newRoomInfo.Attributes.Size != "" && newRoomInfo.Attributes.Size != room.Attributes.Size {
		room.Attributes.Size = newRoomInfo.Attributes.Size
	}

	if newRoomInfo.Attributes.SizeU != "" && newRoomInfo.Attributes.SizeU != room.Attributes.SizeU {
		room.Attributes.SizeU = newRoomInfo.Attributes.SizeU
	}

	if newRoomInfo.Attributes.Height != "" && newRoomInfo.Attributes.Height != room.Attributes.Height {
		room.Attributes.Height = newRoomInfo.Attributes.Height
	}

	if newRoomInfo.Attributes.HeightU != "" && newRoomInfo.Attributes.HeightU != room.Attributes.HeightU {
		room.Attributes.HeightU = newRoomInfo.Attributes.HeightU
	}

	if e1 := GetDB().Table("room").Save(room).
		Table("room_attributes").Save(&(room.Attributes)).Error; e1 != nil {
		return u.Message(false, "Error while updating Room: "+e1.Error()), e1.Error()
	}

	return u.Message(true, "success"), ""
}

func DeleteRoom(id uint) map[string]interface{} {

	//This is a hard delete!
	e := GetDB().Unscoped().Table("room").Delete(&Room{}, id).RowsAffected

	//The command below is a soft delete
	//Meaning that the 'deleted_at' field will be set
	//the record will remain but unsearchable
	//e := GetDB().Table("tenants").Delete(Tenant{}, id).Error
	if e == 0 {
		return u.Message(false, "There was an error in deleting the room")
	}

	return u.Message(true, "success")
}

func GetRoomByName(name string) (*Room, string) {
	room := &Room{}

	e := GetDB().Raw(`SELECT * FROM room 
	JOIN room_attributes ON room.id = room_attributes.id 
	WHERE room_name = ?;`, name).Find(room).Find(&room.Attributes).Error

	if e != nil {
		//fmt.Println(e)
		return nil, e.Error()
	}

	room.DescriptionJSON = strings.Split(room.DescriptionDB, "XYZ")
	room.Category = "room"
	return room, ""
}
