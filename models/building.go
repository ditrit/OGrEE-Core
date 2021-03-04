package models

import (
	"fmt"
	u "p3/utils"
)

type Building_Attributes struct {
	ID      int    `json:"id" gorm:"column:id"`
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
	ID          int                 `json:"id" gorm:"column:id"`
	Name        string              `json:"name" gorm:"column:bldg_name"`
	ParentID    string              `json:"parentId" gorm:"column:bldg_parent_id"`
	Category    string              `json:"category" gorm:"-"`
	Domain      string              `json:"domain" gorm:"column:bldg_domain"`
	D           []string            `json:"description" gorm:"-"`
	Description string              `gorm:"column:bldg_description"`
	Attributes  Building_Attributes `json:"attributes"`

	//Site []Site
	//D is used to help the JSON marshalling
	//while Description will be used in
	//DB transactions
}

func (bldg *Building) Validate() (map[string]interface{}, bool) {
	if bldg.Name == "" {
		return u.Message(false, "Building Name should be on payload"), false
	}

	/*if bldg.Category == "" {
		return u.Message(false, "Category should be on the payload"), false
	}*/

	/*if bldg.Desc == "" {
		return u.Message(false, "Description should be on the payload"), false
	}*/

	if bldg.Domain == "" {
		return u.Message(false, "Domain should should be on the payload"), false
	}

	if GetDB().Table("site").
		Where("id = ?", bldg.ParentID).First(&Site{}).Error != nil {

		return u.Message(false, "Domain should be correspond to site ID"), false
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

	if bldg.Attributes.Floors == "" {
		return u.Message(false, "Floors string should be on the payload"), false
	}

	//Successfully validated bldg
	return u.Message(true, "success"), true
}

func (bldg *Building) Create() map[string]interface{} {
	if resp, ok := bldg.Validate(); !ok {
		return resp
	}

	GetDB().Omit("bldg_description").Create(bldg)
	bldg.Attributes.ID = bldg.ID
	GetDB().Create(&(bldg.Attributes))

	resp := u.Message(true, "success")
	resp["building"] = bldg
	return resp
}

//Get Building by ID
func GetBuilding(id uint) *Building {
	bldg := &Building{}
	err := GetDB().Table("building").Where("id = ?", id).First(bldg).
		Table("building_attributes").Where("id = ?", id).First(&(bldg.Attributes)).Error
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return bldg
}

//Get All Buildings
func GetAllBuildings() []*Building {
	bldgs := make([]*Building, 0)
	attrs := make([]*Building_Attributes, 0)
	err := GetDB().Find(&bldgs).Error
	if err != nil {
		fmt.Println(err)
		return nil
	}

	err = GetDB().Find(&attrs).Error
	if err != nil {
		fmt.Println(err)
		return nil
	}

	for i := range bldgs {
		bldgs[i].Attributes = *(attrs[i])
	}

	return bldgs
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

//More methods should be made to
//Meet CRUD capabilities
//Need Update and Delete
//These would be a bit more complicated
//So leave them out for now

func UpdateBuilding(id uint, newBldgInfo *Building) map[string]interface{} {
	bldg := &Building{}

	err := GetDB().Table("buildings").Where("id = ?", id).First(bldg).Error
	if err != nil {
		return u.Message(false, "Building was not found")
	}

	if newBldgInfo.Name != "" && newBldgInfo.Name != bldg.Name {
		bldg.Name = newBldgInfo.Name
	}

	if newBldgInfo.Category != "" && newBldgInfo.Category != bldg.Category {
		bldg.Category = newBldgInfo.Category
	}

	/*if newBldgInfo.Desc != "" && newBldgInfo.Desc != bldg.Desc {
		bldg.Desc = newBldgInfo.Desc
	}*/

	/*if newBldgInfo.PosX > 0.0 && newBldgInfo.PosX != bldg.PosX {
		bldg.PosX = newBldgInfo.PosX
	}

	if newBldgInfo.PosY > 0.0 && newBldgInfo.PosY != bldg.PosX {
		bldg.PosY = newBldgInfo.PosY
	}

	if newBldgInfo.PosU != "" && newBldgInfo.PosU != bldg.PosU {
		bldg.PosU = newBldgInfo.PosU
	}

	if newBldgInfo.PosZ > 0.0 && newBldgInfo.PosZ != bldg.PosZ {
		bldg.PosZ = newBldgInfo.PosZ
	}

	if newBldgInfo.PosZU != "" && newBldgInfo.PosZU != bldg.PosZU {
		bldg.PosZU = newBldgInfo.PosZU
	}

	if newBldgInfo.Size > 0.0 && newBldgInfo.Size != bldg.Size {
		bldg.Site = newBldgInfo.Site
	}

	if newBldgInfo.SizeU != "" && newBldgInfo.SizeU != bldg.SizeU {
		bldg.SizeU = newBldgInfo.SizeU
	}

	if newBldgInfo.Height > 0.0 && newBldgInfo.Height != bldg.Height {
		bldg.Height = newBldgInfo.Height
	}

	if newBldgInfo.HeightU != "" && newBldgInfo.HeightU != bldg.HeightU {
		bldg.HeightU = newBldgInfo.HeightU
	}*/

	GetDB().Table("buildings").Save(bldg)
	return u.Message(true, "success")
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
