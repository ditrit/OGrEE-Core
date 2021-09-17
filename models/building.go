package models

import (
	"fmt"
	u "p3/utils"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

type Building_Attributes struct {
	ID      int    `json:"-" bson:"id"`
	PosXY   string `json:"posXY" bson:"bldg_pos_x_y"`
	PosXYU  string `json:"posXYUnit" bson:"bldg_pos_x_y_unit"`
	PosZ    string `json:"posZ" bson:"bldg_pos_z"`
	PosZU   string `json:"posZUnit" bson:"bldg_pos_z_unit"`
	Size    string `json:"size" bson:"bldg_size"`
	SizeU   string `json:"sizeUnit" bson:"bldg_size_unit"`
	Height  string `json:"height" bson:"bldg_height"`
	HeightU string `json:"heightUnit" bson:"bldg_height_unit"`
	Floors  string `json:"nbFloors" bson:"bldg_nb_floors"`
}

type Building struct {
	//gorm.Model
	ID              int                 `json:"-" bson:"id"`
	IDJSON          string              `json:"id" bson:"-"`
	Name            string              `json:"name" bson:"bldg_name"`
	ParentID        string              `json:"parentId" bson:"bldg_parent_id"`
	Category        string              `json:"category" bson:"-"`
	Domain          string              `json:"domain" bson:"bldg_domain"`
	DescriptionJSON []string            `json:"description" bson:"-"`
	DescriptionDB   string              `json:"-" bson:"bldg_description"`
	Attributes      Building_Attributes `json:"attributes" bson:"attributes"`

	Rooms []*Room `json:"rooms,omitempty" bson:"-"`
	//D is used to help the JSON marshalling
	//while Description will be used in
	//DB transactions
}

func (bldg *Building) Validate() (map[string]interface{}, bool) {
	if bldg.Name == "" {
		return u.Message(false, "Building Name should be on payload"), false
	}

	if bldg.Category == "" {
		return u.Message(false, "Category should be on the payload"), false
	}

	if bldg.Domain == "" {
		return u.Message(false, "Domain should should be on the payload"), false
	}

	if GetDB().Collection("sites").FindOne(GetCtx(), bson.M{"_id": bldg.ParentID}, nil).Decode(&Site{}) != nil {
		return u.Message(false, "ParentID should be correspond to site ID"), false
	}

	if bldg.Attributes.PosXY == "" {
		return u.Message(false, "XY coordinates should be on payload"), false
	}

	if bldg.Attributes.PosXYU == "" {
		return u.Message(false, "PositionXYU string should be on the payload"), false
	}

	if bldg.Attributes.PosZ == "" {
		return u.Message(false, "Z coordinates should be on payload"), false
	}

	if bldg.Attributes.PosZU == "" {
		return u.Message(false, "PositionZU string should be on the payload"), false
	}

	if bldg.Attributes.Size == "" {
		return u.Message(false, "Invalid building size on the payload"), false
	}

	if bldg.Attributes.SizeU == "" {
		return u.Message(false, "Building size string should be on the payload"), false
	}

	if bldg.Attributes.Height == "" {
		return u.Message(false, "Invalid Height on payload"), false
	}

	if bldg.Attributes.HeightU == "" {
		return u.Message(false, "Building Height string should be on the payload"), false
	}

	//Successfully validated bldg
	return u.Message(true, "success"), true
}

func (bldg *Building) Create() (map[string]interface{}, string) {
	if resp, ok := bldg.Validate(); !ok {
		return resp, "validate"
	}

	bldg.DescriptionDB = strings.Join(bldg.DescriptionJSON, "XYZ")
	if _, e := GetDB().Collection("buildings").InsertOne(GetCtx(), bldg); e != nil {
		return u.Message(false, "Internal Error while creating Bulding: "+e.Error()),
			e.Error()
	}
	bldg.IDJSON = strconv.Itoa(bldg.ID)

	resp := u.Message(true, "success")
	resp["data"] = bldg
	return resp, ""
}

func (b *Building) FormQuery() string {

	query := "SELECT * FROM building " + u.JoinQueryGen("building")
	if b.ParentID != "" {
		query += " AND bldg_parent_id = " + b.ParentID
	}
	if b.IDJSON != "" {
		query += " AND building.id = " + b.IDJSON
	}
	if b.Name != "" {
		query += " WHERE bldg_name = '" + b.Name + "'"
	}
	if b.Category != "" {
		query += " AND bldg_category = '" + b.Category + "'"
	}
	if b.Domain != "" {
		query += " AND bldg_domain = '" + b.Domain + "'"
	}
	if (Building_Attributes{}) != b.Attributes {
		if b.Attributes.PosXY != "" {
			query +=
				" AND building_attributes.bldg_bldg_pos_x_y = '" +
					b.Attributes.PosXY + "'"
		}
		if b.Attributes.PosXYU != "" {
			query +=
				" AND building_attributes.bldg_pos_x_y_unit = '" +
					b.Attributes.PosXYU + "'"
		}
		if b.Attributes.PosZ != "" {
			query +=
				" AND building_attributes.bldg_pos_z = '" +
					b.Attributes.PosZ + "'"
		}
		if b.Attributes.PosZU != "" {
			query +=
				" AND building_attributes.bldg_pos_z_unit = '" +
					b.Attributes.PosZU + "'"
		}
		if b.Attributes.Size != "" {
			query +=
				" AND building_attributes.bldg_size = '" +
					b.Attributes.Size + "'"
		}
		if b.Attributes.SizeU != "" {
			query +=
				" AND building_attributes.bldg_size_unit = '" +
					b.Attributes.SizeU + "'"
		}
		if b.Attributes.Height != "" {
			query +=
				" AND building_attributes.height = '" +
					b.Attributes.Height + "'"
		}
		if b.Attributes.HeightU != "" {
			query +=
				" AND building_attributes.bldg_height_unit = '" +
					b.Attributes.HeightU + "'"
		}
		if b.Attributes.Floors != "" {
			query +=
				" AND building_attributes.bldg_nb_floors = '" +
					b.Attributes.Floors + "'"
		}
	}
	return query
}

//Get Building by ID
func GetBuilding(id uint) (*Building, string) {
	bldg := &Building{}
	err := GetDB().Collection("buildings").FindOne(GetCtx(),
		bson.M{"_id": id}).Decode(bldg)
	if err != nil {
		fmt.Println(err)
		return nil, err.Error()
	}
	bldg.IDJSON = strconv.Itoa(bldg.ID)
	bldg.DescriptionJSON = strings.Split(bldg.DescriptionDB, "XYZ")
	bldg.Category = "building"
	return bldg, ""
}

//Get All Buildings
func GetAllBuildings() ([]*Building, string) {
	bldgs := make([]*Building, 0)
	//attrs := make([]*Building_Attributes, 0)
	c, err := GetDB().Collection("buildings").Find(GetCtx(), bson.M{})
	//err := GetDB().Find(&bldgs).Error
	if err != nil {
		fmt.Println(err)
		return nil, err.Error()
	}

	//err = GetDB().Find(&attrs).Error
	/*if err != nil {
		fmt.Println(err)
		return nil, err.Error()
	}*/
	for c.Next(GetCtx()) {
		b := &Building{}
		e := c.Decode(b)
		if e != nil {
			fmt.Println(err)
			return nil, err.Error()
		}
		bldgs = append(bldgs, b)
		//bldgs[i]
	}

	for i := range bldgs {
		bldgs[i].Category = "building"
		//bldgs[i].Attributes = *(attrs[i])
		bldgs[i].DescriptionJSON = strings.Split(bldgs[i].DescriptionDB, "XYZ")
		bldgs[i].IDJSON = strconv.Itoa(bldgs[i].ID)
	}

	return bldgs, ""
}

func GetBuildingsOfParent(id int) ([]*Building, string) {
	bldgs := make([]*Building, 0)
	//err := GetDB().Table("building").Where("bldg_parent_id = ?", id).Find(&bldgs).Error
	filter := bson.M{"bldg_parent_id": id}
	c, err := GetDB().Collection("buildings").Find(GetCtx(), filter)
	if err != nil {
		fmt.Println(err)
		return nil, err.Error()
	}

	for c.Next(GetCtx()) {
		b := &Building{}
		err := c.Decode(b)
		if err != nil {
			fmt.Println(err)
			return nil, err.Error()
		}
		b.Category = "building"
		bldgs = append(bldgs, b)
	}

	println("The length of bldg is: ", len(bldgs))
	/*for i := range bldgs {
		e := GetDB().Table("building_attributes").Where("id = ?", bldgs[i].ID).First(&(bldgs[i].Attributes)).Error

		if e != nil {
			fmt.Println(err)
			return nil, err.Error()
		}

		bldgs[i].Category = "building"
		bldgs[i].DescriptionJSON = strings.Split(bldgs[i].DescriptionDB, "XYZ")
		bldgs[i].IDJSON = strconv.Itoa(bldgs[i].ID)
	}*/

	return bldgs, ""
}

//Obtain all buildings of a site
func GetBuildings(site *Site) []*Building {
	bldgs := make([]*Building, 0)

	c, err := GetDB().Collection("buildings").Find(GetCtx(), bson.M{"bldg_parent_id": site.ID})
	//err := GetDB().Collection("buildings").Where("foreignkey = ?", site.ID).Find(&bldgs).Error
	if err != nil {
		fmt.Println(err)
		return nil
	}

	for c.Next(GetCtx()) {
		b := &Building{}
		e := c.Decode(b)
		if e != nil {
			fmt.Println(err)
			return nil
		}
		bldgs = append(bldgs, b)
	}

	return bldgs
}

func GetBuildingHierarchy(id uint) (*Building, string) {
	bldg, e := GetBuilding(id)
	if e != "" {
		return nil, e
	}

	bldg.Rooms, e = GetRoomsOfParent(id)
	if e != "" {
		return nil, e
	}

	for k, _ := range bldg.Rooms {
		bldg.Rooms[k], e = GetRoomHierarchy(uint(bldg.Rooms[k].ID))
		if e != "" {
			return nil, e
		}
	}
	return bldg, ""
}

func GetBuildingHierarchyNonStandard(id uint) (*Building,
	[]*Room, *[][]*Rack, *[][]*Device, string) {
	bldg, e := GetBuilding(id)
	if e != "" {
		return nil, nil, nil, nil, e
	}

	rooms, e1 := GetRoomsOfParent(id)
	if e1 != "" {
		return nil, nil, nil, nil, e1
	}
	devtree := make([][]*Device, 1)
	devices := make([][]*Device, 1)
	racktree := make([][]*Rack, 1)

	for k, _ := range rooms {
		_, racktree[k], devices, e = GetRoomHierarchyNonStandard(uint(rooms[k].ID))
		if e != "" {
			return nil, nil, nil, nil, e
		}
		devtree = append(devices, devtree...)
	}
	return bldg, rooms, &racktree, &devices, ""
}

func GetBuildingByQuery(b *Building) ([]*Building, string) {
	/*bldgs := make([]*Building, 0)
	attrs := make([]*Building_Attributes, 0)

	e := GetDB().Raw(b.FormQuery()).Find(&bldgs).
		Find(&attrs).Error

	if e != nil {
		return nil, e.Error()
	}

	for i := range bldgs {
		bldgs[i].Attributes = *(attrs[i])
		bldgs[i].IDJSON = strconv.Itoa(bldgs[i].ID)
		bldgs[i].DescriptionJSON =
			strings.Split(bldgs[i].DescriptionDB, "XYZ")
		bldgs[i].Category = "building"
	}*/

	return nil, ""
}

func UpdateBuilding(id uint, newBldgInfo *map[string]interface{}) (map[string]interface{}, string) {
	//bldg := &Building{}
	/*updateInf := map[string]interface{}

	json.Unmarshal()

	err := GetDB().Collection("building").FindOne(GetCtx(), bson.M{"_id":id}).Decode(bldg)
	if err != nil {
		return u.Message(false, "Building was not found"), err.Error()
	}

	if newBldgInfo.Name != "" && newBldgInfo.Name != bldg.Name {
		bldg.Name = newBldgInfo.Name
	}

	if newBldgInfo.Domain != "" && newBldgInfo.Domain != bldg.Domain {
		bldg.Domain = newBldgInfo.Domain
	}

	if dc := strings.Join(newBldgInfo.DescriptionJSON, "XYZ"); dc != "" && strings.Compare(dc, bldg.DescriptionDB) != 0 {
		bldg.DescriptionDB = dc
	}

	if newBldgInfo.Attributes.PosXY != "" && newBldgInfo.Attributes.PosXY != bldg.Attributes.PosXY {
		bldg.Attributes.PosXY = newBldgInfo.Attributes.PosXY
	}

	if newBldgInfo.Attributes.PosXYU != "" && newBldgInfo.Attributes.PosXYU != bldg.Attributes.PosXYU {
		bldg.Attributes.PosXYU = newBldgInfo.Attributes.PosXYU
	}

	if newBldgInfo.Attributes.PosZ != "" && newBldgInfo.Attributes.PosZ != bldg.Attributes.PosZ {
		bldg.Attributes.PosZ = newBldgInfo.Attributes.PosZ
	}

	if newBldgInfo.Attributes.PosZU != "" && newBldgInfo.Attributes.PosZU != bldg.Attributes.PosZU {
		bldg.Attributes.PosZU = newBldgInfo.Attributes.PosZU
	}

	if newBldgInfo.Attributes.Size != "" && newBldgInfo.Attributes.Size != bldg.Attributes.Size {
		bldg.Attributes.Size = newBldgInfo.Attributes.Size
	}

	if newBldgInfo.Attributes.SizeU != "" && newBldgInfo.Attributes.SizeU != bldg.Attributes.SizeU {
		bldg.Attributes.SizeU = newBldgInfo.Attributes.SizeU
	}

	if newBldgInfo.Attributes.Height != "" && newBldgInfo.Attributes.Height != bldg.Attributes.Height {
		bldg.Attributes.Height = newBldgInfo.Attributes.Height
	}

	if newBldgInfo.Attributes.HeightU != "" && newBldgInfo.Attributes.HeightU != bldg.Attributes.HeightU {
		bldg.Attributes.HeightU = newBldgInfo.Attributes.HeightU
	}

	if newBldgInfo.Attributes.Floors != "" && newBldgInfo.Attributes.Floors != bldg.Attributes.Floors {
		bldg.Attributes.Floors = newBldgInfo.Attributes.Floors
	}*/

	/*if e := GetDB().Table("building").Save(bldg).
		Table("building_attributes").Save(&(bldg.Attributes)).Error; e != nil {
		return u.Message(false, "Error while updating Building: "+e.Error()), e.Error()
	}*/
	c := GetDB().Collection("building").FindOneAndUpdate(GetCtx(), bson.M{"_id": id}, bson.M{"$set": *newBldgInfo}).Err()
	if c != nil {
		return u.Message(false, "failure: "+c.Error()), c.Error()
	}
	return u.Message(true, "success"), ""
}

func DeleteBuilding(id uint) map[string]interface{} {

	//This is a hard delete!
	c, _ := GetDB().Collection("building").DeleteOne(GetCtx(), bson.M{"_id": id})

	//The command below is a soft delete
	//Meaning that the 'deleted_at' field will be set
	//the record will remain but unsearchable
	//e := GetDB().Table("tenants").Delete(Tenant{}, id).Error
	if c.DeletedCount == 0 {
		return u.Message(false, "There was an error in deleting the building")
	}

	return u.Message(true, "success")
}

func GetBuildingByName(name string) (*Building, string) {
	bldg := &Building{}

	e := GetDB().Collection("building").
		FindOne(GetCtx(), bson.M{"bldg_name": name}).Decode(bldg)

	if e != nil {
		return nil, e.Error()
	}

	bldg.IDJSON = strconv.Itoa(bldg.ID)
	bldg.DescriptionJSON = strings.Split(bldg.DescriptionDB, "XYZ")
	bldg.Category = "building"
	return bldg, ""
}

func GetNamedRoomOfBuilding(id int, name string) (*Room, string) {
	if _, e := GetBuilding(uint(id)); e != "" {
		return nil, e
	}

	room, e := GetRoomByNameAndParentID(id, name)
	if e != "" {
		return nil, e
	}
	return room, ""

}

func GetRoomsOfBuilding(id int) ([]*Room, string) {
	if _, e := GetBuilding(uint(id)); e != "" {
		return nil, e
	}

	rooms, e := GetRoomsOfParent(uint(id))
	if e != "" {
		return nil, e
	}
	return rooms, ""
}

func GetRacksUsingNamedRoomOfBuilding(bldgid int, name string) ([]*Rack, string) {
	if _, e := GetBuilding(uint(bldgid)); e != "" {
		return nil, e
	}

	room, e := GetRoomByNameAndParentID(bldgid, name)
	if e != "" {
		return nil, e
	}

	racks, e1 := GetRacksOfParent(uint(room.ID))
	if e1 != "" {
		return nil, e1
	}

	return racks, ""
}

func GetNamedRackOfBuilding(id int, room_name, rack_name string) (*Rack, string) {
	if _, e := GetBuilding(uint(id)); e != "" {
		return nil, e
	}

	room, e := GetRoomByNameAndParentID(id, room_name)
	if e != "" {
		return nil, e
	}

	rack, e1 := GetRackByNameAndParentID(room.ID, rack_name)
	if e1 != "" {
		return nil, e1
	}
	return rack, ""
}

func GetDevicesUsingNamedRackOfBuilding(id int, room_name, rack_name string) ([]*Device, string) {
	if _, e := GetBuilding(uint(id)); e != "" {
		return nil, e
	}

	room, e := GetRoomByNameAndParentID(id, room_name)
	if e != "" {
		return nil, e
	}

	rack, e1 := GetRackByNameAndParentID(room.ID, rack_name)
	if e1 != "" {
		return nil, e
	}

	devices, e2 := GetDevicesOfParent(uint(rack.ID))
	if e2 != "" {
		return nil, e2
	}
	return devices, e2
}

func GetNamedDeviceOfBuilding(id int, room_name, rack_name, device_name string) (*Device, string) {
	if _, e := GetBuilding(uint(id)); e != "" {
		return nil, e
	}

	room, e := GetRoomByNameAndParentID(id, room_name)
	if e != "" {
		return nil, e
	}

	rack, e1 := GetRackByNameAndParentID(room.ID, rack_name)
	if e1 != "" {
		return nil, e1
	}

	device, e2 := GetDeviceByNameAndParentID(uint(rack.ID), device_name)
	if e2 != "" {
		return nil, e2
	}
	return device, ""
}

func GetBuildingByNameAndParentID(id int, name string) (*Building, string) {
	building := &Building{}
	/*err := GetDB().Raw(`SELECT * FROM building JOIN
	building_attributes ON building.id = building_attributes.id
	WHERE bldg_parent_id = ? AND bldg_name = ?`, id, name).
		Find(building).Find(&(building.Attributes)).Error*/

	err := GetDB().Collection("building").
		FindOne(GetCtx(),
			bson.M{"bldg_name": name, "bldg_parent_id": id}).Decode(building)

	if err != nil {
		fmt.Println(err)
		return nil, err.Error()
	}

	building.DescriptionJSON = strings.Split(building.DescriptionDB, "XYZ")
	building.Category = "building"
	building.IDJSON = strconv.Itoa(building.ID)
	return building, ""
}

func GetRoomsUsingNamedBldgOfSite(id int, name string) ([]*Room, string) {
	bldg, e := GetBuildingByNameAndParentID(id, name)
	if e != "" {
		return nil, e
	}

	rooms, e1 := GetRoomsOfBuilding(bldg.ID)
	if e1 != "" {
		return nil, e1
	}
	return rooms, ""
}

func GetBuildingHierarchyToRack(id int) (*Building, string) {
	bldg, e := GetBuilding(uint(id))
	if e != "" {
		return nil, e
	}

	bldg.Rooms, e = GetRoomsOfParent(uint(id))
	if e != "" {
		return nil, e
	}

	for idx, _ := range bldg.Rooms {
		bldg.Rooms[idx].Racks, e = GetRacksOfParent(uint(bldg.Rooms[idx].ID))
		if e != "" {
			return nil, e
		}
	}

	return bldg, ""
}

func GetBuildingHierarchyToDevice(id int) (*Building, string) {
	bldg, e := GetBuilding(uint(id))
	if e != "" {
		return nil, e
	}

	bldg.Rooms, e = GetRoomsOfParent(uint(id))
	if e != "" {
		return nil, e
	}

	for idx, _ := range bldg.Rooms {
		bldg.Rooms[idx].Racks, e = GetRacksOfParent(uint(bldg.Rooms[idx].ID))
		if e != "" {
			return nil, e
		}

		for k, _ := range bldg.Rooms[idx].Racks {
			bldg.Rooms[idx].Racks[k].Devices, e =
				GetDevicesOfParent(uint(bldg.Rooms[idx].Racks[k].ID))
			if e != "" {
				return nil, e
			}
		}
	}

	return bldg, ""
}

func GetBuildingHierarchyToSubdevice(id int) (*Building, string) {
	bldg, e := GetBuilding(uint(id))
	if e != "" {
		return nil, e
	}

	bldg.Rooms, e = GetRoomsOfParent(uint(id))
	if e != "" {
		return nil, e
	}

	for idx, _ := range bldg.Rooms {
		bldg.Rooms[idx].Racks, e = GetRacksOfParent(uint(bldg.Rooms[idx].ID))
		if e != "" {
			return nil, e
		}

		for k, _ := range bldg.Rooms[idx].Racks {
			bldg.Rooms[idx].Racks[k].Devices, e =
				GetDevicesOfParent(uint(bldg.Rooms[idx].Racks[k].ID))
			if e != "" {
				return nil, e
			}

			for h, _ := range bldg.Rooms[idx].Racks[k].Devices {
				bldg.Rooms[idx].Racks[k].Devices[h].Subdevices, e =
					GetSubdevicesOfParent(uint(bldg.Rooms[idx].
						Racks[k].Devices[h].ID))
				if e != "" {
					return nil, e
				}

			}
		}
	}

	return bldg, ""
}

func GetSubdevicesUsingNamedDeviceOfBuilding(id int, room_name,
	rack_name, device_name string) ([]*Subdevice, string) {

	room, err := GetRoomByNameAndParentID(id, room_name)
	if err != "" {
		return nil, err
	}
	rack, e := GetRackByNameAndParentID(room.ID, rack_name)
	if e != "" {
		return nil, e
	}

	device, e1 := GetDeviceByNameAndParentID(uint(rack.ID), device_name)
	if e1 != "" {
		return nil, e1
	}

	subdevices, e2 := GetSubdevicesOfParent(uint(device.ID))
	if e2 != "" {
		return nil, e2
	}

	return subdevices, ""
}

func GetNamedSubdeviceOfBuilding(id int, room_name,
	rack_name, device_name, subdevice_name string) (*Subdevice, string) {

	room, e := GetRoomByNameAndParentID(id, room_name)
	if e != "" {
		return nil, e
	}

	rack, e1 := GetRackByNameAndParentID(room.ID, rack_name)
	if e1 != "" {
		return nil, e1
	}

	device, e2 := GetDeviceByNameAndParentID(uint(rack.ID), device_name)
	if e2 != "" {
		return nil, e2
	}

	subdevice, e3 := GetSubdeviceByNameAndParentID(device.ID, subdevice_name)
	if e3 != "" {
		return nil, e3
	}
	return subdevice, ""
}

func GetSubdevice1sUsingNamedSubdeviceOfBuilding(id int, room_name,
	rack_name, device_name, subdevice_name string) ([]*Subdevice1, string) {

	room, err := GetRoomByNameAndParentID(id, room_name)
	if err != "" {
		return nil, err
	}
	rack, e := GetRackByNameAndParentID(room.ID, rack_name)
	if e != "" {
		return nil, e
	}

	device, e1 := GetDeviceByNameAndParentID(uint(rack.ID), device_name)
	if e1 != "" {
		return nil, e1
	}

	subdevice, e2 := GetSubdeviceByNameAndParentID((device.ID), subdevice_name)
	if e2 != "" {
		return nil, e2
	}

	return GetSubdevices1OfParent(subdevice.ID)
}

func GetNamedSubdevice1OfBuilding(id int, room_name, rack_name,
	device_name, subdevice_name, subdev1_name string) (*Subdevice1, string) {

	room, e := GetRoomByNameAndParentID(id, room_name)
	if e != "" {
		return nil, e
	}

	rack, e1 := GetRackByNameAndParentID(room.ID, rack_name)
	if e1 != "" {
		return nil, e1
	}

	device, e2 := GetDeviceByNameAndParentID(uint(rack.ID), device_name)
	if e2 != "" {
		return nil, e2
	}

	subdevice, e3 := GetSubdeviceByNameAndParentID(device.ID, subdevice_name)
	if e3 != "" {
		return nil, e3
	}
	return GetSubdevice1ByNameAndParentID(subdevice.ID, subdev1_name)
}
