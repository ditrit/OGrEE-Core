package models

import (
	"fmt"
	u "p3/utils"
	"strconv"
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
	Technical   string `json:"technical" gorm:"column:room_technical"`
	Reserved    string `json:"reserved" gorm:"column:room_reserved"`
}

type Room struct {
	ID              int             `json:"-" gorm:"column:id"`
	IDJSON          string          `json:"id" gorm:"-"`
	Name            string          `json:"name" gorm:"column:room_name"`
	ParentID        string          `json:"parentId" gorm:"column:room_parent_id"`
	Category        string          `json:"category" gorm:"-"`
	Domain          string          `json:"domain" gorm:"column:room_domain"`
	DescriptionJSON []string        `json:"description" gorm:"-"`
	DescriptionDB   string          `json:"-" gorm:"column:room_description"`
	Attributes      Room_Attributes `json:"attributes"`

	Racks []*Rack `json:"racks,omitempty", gorm:"-"`
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
	room.IDJSON = strconv.Itoa(room.ID)
	room.Attributes.ID = room.ID
	if e := GetDB().Create(&(room.Attributes)).Error; e != nil {
		return u.Message(false, "Internal Error while creating Room Attrs: "+e.Error()),
			e.Error()
	}

	resp := u.Message(true, "success")
	resp["room"] = room
	return resp, ""
}

func (r *Room) FormQuery() string {

	query := "SELECT * FROM room " + u.JoinQueryGen("room")
	if r.ParentID != "" {
		query += " AND room_parent_id = " + r.ParentID
	}
	if r.IDJSON != "" {
		query += " AND room.id = " + r.IDJSON
	}
	if r.Name != "" {
		query += " WHERE room_name = '" + r.Name + "'"
	}
	if r.Category != "" {
		query += " AND room_category = '" + r.Category + "'"
	}
	if r.Domain != "" {
		query += " AND room_domain = '" + r.Domain + "'"
	}
	if (Room_Attributes{}) != r.Attributes {
		if r.Attributes.Template != "" {
			query +=
				" AND room_attributes.room_template = '" +
					r.Attributes.Template + "'"
		}
		if r.Attributes.PosXY != "" {
			query +=
				" AND room_attributes.room_pos_x_y = '" +
					r.Attributes.PosXY + "'"
		}
		if r.Attributes.PosXYU != "" {
			query +=
				" AND room_attributes.room_pos_x_y_unit = '" +
					r.Attributes.PosXYU + "'"
		}
		if r.Attributes.PosZ != "" {
			query +=
				" AND room_attributes.room_pos_z = '" +
					r.Attributes.PosZ + "'"
		}
		if r.Attributes.PosZU != "" {
			query +=
				" AND room_attributes.room_pos_z_unit = '" +
					r.Attributes.PosZU + "'"
		}
		if r.Attributes.Size != "" {
			query +=
				" AND room_attributes.room_size = '" +
					r.Attributes.Size + "'"
		}
		if r.Attributes.SizeU != "" {
			query +=
				" AND room_attributes.room_size_unit = '" +
					r.Attributes.SizeU + "'"
		}
		if r.Attributes.Height != "" {
			query +=
				" AND room_attributes.room_height = '" +
					r.Attributes.Height + "'"
		}
		if r.Attributes.HeightU != "" {
			query +=
				" AND room_attributes.room_height_unit = '" +
					r.Attributes.HeightU + "'"
		}
		if r.Attributes.Template != "" {
			query +=
				" AND room_attributes.room_template = '" +
					r.Attributes.Template + "'"
		}
		if r.Attributes.Orientation != "" {
			query +=
				" AND room_attributes.room_orientation = '" +
					r.Attributes.Orientation + "'"
		}
		if r.Attributes.Technical != "" {
			query +=
				" AND room_attributes.room_orientation = '" +
					r.Attributes.Technical + "'"
		}
		if r.Attributes.Reserved != "" {
			query +=
				" AND room_attributes.room_reserved = '" +
					r.Attributes.Reserved + "'"
		}
	}
	return query
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
	room.IDJSON = strconv.Itoa(room.ID)
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

func GetRoomHierarchy(id uint) (*Room, string) {

	room, e := GetRoom(id)
	if e != "" {
		return nil, e
	}

	room.Racks, e = GetRacksOfParent(id)
	if e != "" {
		return nil, e
	}

	for i, _ := range room.Racks {
		room.Racks[i], e = GetRackHierarchy(uint(room.Racks[i].ID))
		if e != "" {
			return nil, e
		}

	}

	return room, ""
}

func GetRoomHierarchyNonStandard(id uint) (*Room, []*Rack, [][]*Device, string) {

	room, e := GetRoom(id)
	if e != "" {
		return nil, nil, nil, e
	}

	racks, e1 := GetRacksOfParent(id)
	if e1 != "" {
		return nil, nil, nil, e1
	}

	devtree := make([][]*Device, len(racks))

	for i, _ := range racks {
		devtree[i], e = GetDevicesOfParent(uint(racks[i].ID))
		if e != "" {
			return nil, nil, nil, e
		}

	}

	return room, racks, devtree, ""
}

func GetRoomsOfParent(id uint) ([]*Room, string) {
	rooms := make([]*Room, 0)
	err := GetDB().Table("room").Where("room_parent_id = ?", id).Find(&rooms).Error
	if err != nil {
		fmt.Println(err)
		return nil, err.Error()
	}

	println("The length of room is: ", len(rooms))
	for i := range rooms {
		e := GetDB().Table("room_attributes").Where("id = ?", rooms[i].ID).First(&(rooms[i].Attributes)).Error

		if e != nil {
			fmt.Println(err)
			return nil, err.Error()
		}

		rooms[i].Category = "room"
		rooms[i].DescriptionJSON = strings.Split(rooms[i].DescriptionDB, "XYZ")
		rooms[i].IDJSON = strconv.Itoa(rooms[i].ID)
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
		rooms[i].IDJSON = strconv.Itoa(rooms[i].ID)
	}

	return rooms, ""
}

func GetRoomByQuery(r *Room) ([]*Room, string) {
	rooms := make([]*Room, 0)
	attrs := make([]*Room_Attributes, 0)

	e := GetDB().Raw(r.FormQuery()).Find(&rooms).
		Find(&attrs).Error

	if e != nil {
		return nil, e.Error()
	}

	for i := range rooms {
		rooms[i].Attributes = *(attrs[i])
		rooms[i].IDJSON = strconv.Itoa(rooms[i].ID)
		rooms[i].DescriptionJSON =
			strings.Split(rooms[i].DescriptionDB, "XYZ")
		rooms[i].Category = "room"
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
		case "-E-N", "-E+N", "+E-N", "+E+N",
			"-N-W", "-N+W", "+N-W", "+N+W",
			"-W-S", "-W+S", "+W-S", "+W+S",
			"-S-E", "-S+E", "+S-E", "+S+E":
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

	if newRoomInfo.Attributes.Technical != "" && newRoomInfo.Attributes.Technical != room.Attributes.Technical {
		room.Attributes.Technical = newRoomInfo.Attributes.Technical
	}

	if newRoomInfo.Attributes.Reserved != "" && newRoomInfo.Attributes.Reserved != room.Attributes.Reserved {
		room.Attributes.Reserved = newRoomInfo.Attributes.Reserved
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
		return nil, e.Error()
	}

	room.IDJSON = strconv.Itoa(room.ID)
	room.DescriptionJSON = strings.Split(room.DescriptionDB, "XYZ")
	room.Category = "room"
	return room, ""
}

func GetDevicesUsingNamedRackOfRoom(roomID int, rack_name string) ([]*Device, string) {
	if _, e := GetRoom(uint(roomID)); e != "" {
		return nil, e
	}

	rack, e := GetRackByNameAndParentID(roomID, rack_name)
	if e != "" {
		return nil, e
	}

	rack.Devices, e = GetDevicesOfParent(uint(rack.ID))
	if e != "" {
		return nil, e
	}

	return rack.Devices, ""
}

func GetNamedDeviceOfRoom(roomID int, rack_name, device_name string) (*Device, string) {
	if _, e := GetRoom(uint(roomID)); e != "" {
		return nil, e
	}

	rack, e1 := GetRackByNameAndParentID(roomID, rack_name)
	if e1 != "" {
		return nil, e1
	}

	device, e2 := GetDeviceByNameAndParentID(uint(rack.ID), device_name)
	if e2 != "" {
		return nil, e2
	}

	return device, ""
}

func GetRoomByNameAndParentID(id int, name string) (*Room, string) {
	room := &Room{}
	err := GetDB().Raw(`SELECT * FROM room JOIN 
	room_attributes ON room.id = room_attributes.id
	WHERE room_parent_id = ? AND room_name = ?`, id, name).
		Find(room).Find(&(room.Attributes)).Error
	if err != nil {
		fmt.Println(err)
		return nil, err.Error()
	}

	room.DescriptionJSON = strings.Split(room.DescriptionDB, "XYZ")
	room.Category = "room"
	room.IDJSON = strconv.Itoa(room.ID)
	return room, ""
}

func GetNamedSubdeviceOfRoom(id int, rack_name, device_name, subdev_name string) (*Subdevice, string) {
	rack, e := GetRackByNameAndParentID(id, rack_name)
	if e != "" {
		return nil, e
	}

	dev, e1 := GetDeviceByNameAndParentID(uint(rack.ID), device_name)
	if e1 != "" {
		return nil, e1
	}

	subdev, e2 := GetSubdeviceByNameAndParentID(dev.ID, subdev_name)
	if e2 != "" {
		return nil, e2
	}

	return subdev, ""
}

func GetSubdevicesUsingNamedDeviceOfRoom(id int, rack_name, device_name string) ([]*Subdevice, string) {
	rack, e := GetRackByNameAndParentID(id, rack_name)
	if e != "" {
		return nil, e
	}

	dev, e1 := GetDeviceByNameAndParentID(uint(rack.ID), device_name)
	if e1 != "" {
		return nil, e1
	}

	subdevices, e2 := GetSubdevicesOfParent(uint(dev.ID))
	if e2 != "" {
		return nil, e2
	}

	return subdevices, ""
}
