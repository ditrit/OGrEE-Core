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
	Domain      string          `json:"domain"`
	Color       string          `json:"color"`
	Orientation ECardinalOrient `json:"eorientation"`
	Tenant      []Tenant        `gorm:"foreignKey:Tenant"`
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

	if site.Domain == "" {
		return u.Message(false, "Domain should be on the payload"), false
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
func GetSite(id uint) *Site {
	site := &Site{}

	err := GetDB().Table("sites").Where("id = ?", id).First(site).Error
	if err != nil {
		return nil
	}
	return site
}

//Getting the Sites related to tenant
//Would require Foreign Key (referring
// to tenant)
//I could design the controller
//to invoke get tenant then
//obtain sites
func GetSites(user uint) []*Site {
	sites := make([]*Site, 0)
	//err := GetDB().Table("contacts").Where("user_id = ?", user).Find(&).Error

	tenant := &Tenant{}
	err := GetDB().Table("tenants").Where("userid = ?", user).Find(tenant).Error
	if err != nil {
		fmt.Println(err)
		return nil
	}

	e := GetDB().Table("sites").Where("foriegnkey = ?", tenant.ID).Find(sites).Error
	if e != nil {
		fmt.Println(err)
		return nil
	}

	return sites
}

//More methods should be made to
//Meet CRUD capabilities
//Need Update and Delete
//These would be a bit more complicated
//So leave them out for now
