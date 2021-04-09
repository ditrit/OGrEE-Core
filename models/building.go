package models

import (
	"fmt"
	u "p3/utils"
	"strconv"
	"strings"
)

type Building_Attributes struct {
	ID      int    `json:"-" gorm:"column:id"`
	PosXY   string `json:"posXY" gorm:"column:bldg_pos_x_y"`
	PosXYU  string `json:"posXYUnit" gorm:"column:bldg_pos_x_y_unit"`
	PosZ    string `json:"posZ" gorm:"column:bldg_pos_z"`
	PosZU   string `json:"posZUnit" gorm:"column:bldg_pos_z_unit"`
	Size    string `json:"size" gorm:"column:bldg_size"`
	SizeU   string `json:"sizeUnit" gorm:"column:bldg_size_unit"`
	Height  string `json:"height" gorm:"column:bldg_height"`
	HeightU string `json:"heightUnit" gorm:"column:bldg_height_unit"`
	Floors  string `json:"nbFloors" gorm:"column:bldg_nb_floors"`
}

type Building struct {
	//gorm.Model
	ID              int                 `json:"-" gorm:"column:id"`
	IDJSON          string              `json:"id" gorm:"-"`
	Name            string              `json:"name" gorm:"column:bldg_name"`
	ParentID        string              `json:"parentId" gorm:"column:bldg_parent_id"`
	Category        string              `json:"category" gorm:"-"`
	Domain          string              `json:"domain" gorm:"column:bldg_domain"`
	DescriptionJSON []string            `json:"description" gorm:"-"`
	DescriptionDB   string              `json:"-" gorm:"column:bldg_description"`
	Attributes      Building_Attributes `json:"attributes"`

	//Site []Site
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

	if GetDB().Table("site").
		Where("id = ?", bldg.ParentID).First(&Site{}).Error != nil {

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
	if e := GetDB().Create(bldg).Error; e != nil {
		return u.Message(false, "Internal Error while creating Bulding: "+e.Error()),
			e.Error()
	}
	bldg.IDJSON = strconv.Itoa(bldg.ID)
	bldg.Attributes.ID = bldg.ID
	if e := GetDB().Create(&(bldg.Attributes)).Error; e != nil {
		return u.Message(false, "Internal Error while creating Bulding Attrs: "+e.Error()),
			e.Error()
	}

	resp := u.Message(true, "success")
	resp["building"] = bldg
	return resp, ""
}

//Get Building by ID
func GetBuilding(id uint) (*Building, string) {
	bldg := &Building{}
	err := GetDB().Table("building").Where("id = ?", id).First(bldg).
		Table("building_attributes").Where("id = ?", id).First(&(bldg.Attributes)).Error
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
	attrs := make([]*Building_Attributes, 0)
	err := GetDB().Find(&bldgs).Error
	if err != nil {
		fmt.Println(err)
		return nil, err.Error()
	}

	err = GetDB().Find(&attrs).Error
	if err != nil {
		fmt.Println(err)
		return nil, err.Error()
	}

	for i := range bldgs {
		bldgs[i].Category = "building"
		bldgs[i].Attributes = *(attrs[i])
		bldgs[i].DescriptionJSON = strings.Split(bldgs[i].DescriptionDB, "XYZ")
		bldgs[i].IDJSON = strconv.Itoa(bldgs[i].ID)
	}

	return bldgs, ""
}

//Obtain all buildings of a site
func GetBuildings(site *Site) []*Building {
	bldgs := make([]*Building, 0)

	err := GetDB().Table("buildings").Where("foreignkey = ?", site.ID).Find(&bldgs).Error
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return bldgs
}

func UpdateBuilding(id uint, newBldgInfo *Building) (map[string]interface{}, string) {
	bldg := &Building{}

	err := GetDB().Table("building").Where("id = ?", id).First(bldg).
		Table("building_attributes").Where("id = ?", id).First(&(bldg.Attributes)).Error
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

	/*if newBldgInfo.Category != "" && newBldgInfo.Category != bldg.Category {
		bldg.Category = newBldgInfo.Category
	}*/

	/*if newBldgInfo.Desc != "" && newBldgInfo.Desc != bldg.Desc {
		bldg.Desc = newBldgInfo.Desc
	}*/

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

	if e := GetDB().Table("building").Save(bldg).
		Table("building_attributes").Save(&(bldg.Attributes)).Error; e != nil {
		return u.Message(false, "Error while updating Building: "+e.Error()), e.Error()
	}
	return u.Message(true, "success"), ""
}

func DeleteBuilding(id uint) map[string]interface{} {

	//This is a hard delete!
	e := GetDB().Unscoped().Table("building").
		Where("id = ?", id).Delete(&Building{}).RowsAffected

	//The command below is a soft delete
	//Meaning that the 'deleted_at' field will be set
	//the record will remain but unsearchable
	//e := GetDB().Table("tenants").Delete(Tenant{}, id).Error
	if e == 0 {
		return u.Message(false, "There was an error in deleting the building")
	}

	return u.Message(true, "success")
}

func GetBuildingByName(name string) (*Building, string) {
	bldg := &Building{}

	e := GetDB().Raw(`SELECT * FROM building 
	JOIN building_attributes ON building.id = building_attributes.id 
	WHERE bldg_name = ?;`, name).Find(bldg).Find(&bldg.Attributes).Error

	if e != nil {
		//fmt.Println(e)
		return nil, e.Error()
	}

	bldg.IDJSON = strconv.Itoa(bldg.ID)
	bldg.DescriptionJSON = strings.Split(bldg.DescriptionDB, "XYZ")
	bldg.Category = "bldg"
	return bldg, ""
}
