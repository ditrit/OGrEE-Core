package models

import (
	"fmt"
	u "p3/utils"

	"github.com/jinzhu/gorm"
)

type ECardinalOrient string

type Site struct {
	gorm.Model
	Name        string          `json:"name"`
	Category    string          `json:"category"`
	Desc        string          `json:"description"`
	Domain      int             `json:"domain"`
	Color       string          `json:"color"`
	Orientation ECardinalOrient `json:"eorientation"`
	Building    []Building
}

func (site *Site) Validate() (map[string]interface{}, bool) {
	if site.Name == "" {
		return u.Message(false, "site Name should be on payload"), false
	}

	if site.Category == "" {
		return u.Message(false, "Category should be on the payload"), false
	}

	if site.Desc == "" {
		return u.Message(false, "Description should be on the payload"), false
	}

	if site.Domain == 0 {
		return u.Message(false, "Domain should be on the payload"), false
	}

	if GetDB().Table("tenants").
		Where("id = ?", site.Domain).First(&Tenant{}).Error != nil {

		return u.Message(false, "Domain should be correspond to tenant ID"), false
	}

	if site.Color == "" {
		return u.Message(false, "Color should be on the payload"), false
	}

	switch site.Orientation {
	case "NE", "NW", "SE", "SW":
	case "":
		return u.Message(false, "Orientation should be on the payload"), false

	default:
		return u.Message(false, "Orientation is invalid!"), false
	}

	//Successfully validated Site
	return u.Message(true, "success"), true
}

func (site *Site) Create() map[string]interface{} {
	if resp, ok := site.Validate(); !ok {
		return resp
	}

	GetDB().Create(site)
	resp := u.Message(true, "success")
	resp["site"] = site
	return resp
}

//Would have to think about
//these functions more
//since I set it up
//to just obtain the first site
//The GORM command might be
//wrong too
func GetSites(id uint) []*Site {
	tenant := &Tenant{}
	site := make([]*Site, 0)

	err := GetDB().Table("tenants").Where("id = ?", id).First(tenant).Error
	if err != nil {
		fmt.Println("yo the tenant wasnt found here")
		return nil
	}

	e := GetDB().Table("sites").Where("domain = ?", id).Find(&site).Error
	if e != nil {
		fmt.Println("yo the there isnt any site matching the foreign key")
		return nil
	}

	return site
}

func GetSite(id uint) *Site {
	site := &Site{}

	err := GetDB().Table("sites").Where("id = ?", id).First(site).Error
	if err != nil {
		fmt.Println("There was an error in getting site by ID")
		return nil
	}
	return site
}

//More methods should be made to
//Meet CRUD capabilities
//Need Update and Delete
//These would be a bit more complicated
//So leave them out for now

func DeleteSite(id uint) map[string]interface{} {
	//This command is a hard delete!
	e := GetDB().Unscoped().Table("sites").Delete(Tenant{}, id).Error

	//The command below is a soft delete
	//Meaning that the 'deleted_at' field will be set
	//the record will remain but unsearchable
	//e := GetDB().Table("tenants").Delete(Tenant{}, id).Error
	if e != nil {
		return u.Message(false, "Tenant was not found")
	}

	return u.Message(true, "success")
}
