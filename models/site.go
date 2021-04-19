package models

import (
	"fmt"
	u "p3/utils"
	"strconv"
	"strings"
)

type ECardinalOrient string

//Desc        string          `json:"description"`

type Site_Attributes struct {
	ID             int    `json:"-" gorm:"column:id"`
	Orientation    string `json:"orientation" gorm:"column:site_orientation"`
	UsableColor    string `json:"usableColor" gorm:"column:usable_color"`
	ReservedColor  string `json:"reservedColor" gorm:"column:reserved_color"`
	TechnicalColor string `json:"technicalColor" gorm:"column:technical_color"`
	Address        string `json:"address" gorm:"column:address"`
	Zipcode        string `json:"zipcode" gorm:"column:zipcode"`
	City           string `json:"city" gorm:"column:city"`
	Country        string `json:"country" gorm:"column:country"`
	Gps            string `json:"gps" gorm:"column:gps"`
}

type Site struct {
	//gorm.Model
	ID              int             `json:"-" gorm:"column:id"`
	IDJSON          string          `json:"id" gorm:"-"`
	Name            string          `json:"name" gorm:"column:site_name"`
	Category        string          `json:"category" gorm:"-"`
	Domain          string          `json:"domain" gorm:"column:site_domain"`
	ParentID        string          `json:"parentId" gorm:"column:site_parent_id"`
	DescriptionJSON []string        `json:"description" gorm:"-"`
	DescriptionDB   string          `json:"-" gorm:"column:site_description"`
	Attributes      Site_Attributes `json:"attributes"`

	Buildings []*Building `json:"buildings,omitempty"`
}

func (site *Site) Validate() (map[string]interface{}, bool) {
	if site.Name == "" {
		return u.Message(false, "site Name should be on payload"), false

	}

	if site.Category == "" {
		return u.Message(false, "Category should be on the payload"), false

	}

	if site.Domain == "" {
		return u.Message(false, "Domain should be on the payload"), false

	}

	if GetDB().Table("tenant").
		Where("id = ?", site.ParentID).First(&Tenant{}).Error != nil {

		return u.Message(false, "SiteParentID should be correspond to tenant ID"), false

	}

	switch site.Attributes.Orientation {
	case "EN", "NW", "WS", "SE":
	case "":
		return u.Message(false, "Orientation should be on the payload"), false

	default:
		return u.Message(false, "Orientation is invalid!"), false
	}

	if site.Attributes.UsableColor == "" {
		return u.Message(false, "Usable Color should be on the payload"), false
	}

	if site.Attributes.ReservedColor == "" {
		return u.Message(false, "Reserved Color should be on the payload"), false
	}

	if site.Attributes.TechnicalColor == "" {
		return u.Message(false, "Technical Color should be on the payload"), false
	}

	//Successfully validated Site
	return u.Message(true, "success"), true
}

func (site *Site) Create() (map[string]interface{}, string) {
	if resp, ok := site.Validate(); !ok {
		return resp, "validate"
	}

	//GetDB().Create(site)
	site.DescriptionDB = strings.Join(site.DescriptionJSON, "XYZ")

	e := GetDB().Create(site).Error
	if e != nil {
		return u.Message(false, "Internal Error while creating Site: "+e.Error()),
			e.Error()
	}

	site.IDJSON = strconv.Itoa(site.ID)
	site.Attributes.ID = site.ID

	e = GetDB().Table("site_attributes").Create(&(site.Attributes)).Error
	if e != nil {
		return u.Message(false, "Internal Error while creating Site Attrs: "+e.Error()),
			e.Error()
	}
	resp := u.Message(true, "success")
	resp["site"] = site
	return resp, ""
}

//Get sites of a Tenant
func GetSites(id uint) []*Site {
	site := make([]*Site, 0)

	err := GetDB().Table("tenant").Where("id = ?", id).First(&Tenant{}).Error
	if err != nil {
		fmt.Println("yo the tenant wasnt found here")
		return nil
	}

	e := GetDB().Table("site").Where("site_parent_id = ?", id).Find(&site).Error
	if e != nil {
		fmt.Println("yo the there isnt any site matching the foreign key")
		return nil
	}

	//This can be an efficiency issue which
	//can be compared to making a Attribute
	//struct slice then make the same query as above
	//then iterate and assign attributes from the
	//attribute slice
	for i := range site {
		GetDB().Raw("SELECT * FROM site_attributes WHERE id = ?",
			site[i].ID).Scan(&(site[i].Attributes))

		fmt.Println("ITER ID: ", site[i].ID)
		if err != nil {
			return nil
		}
	}

	return site
}

func GetSite(id uint) (*Site, string) {
	site := &Site{}

	err := GetDB().Table("site").Where("id = ?", id).First(site).
		Table("site_attributes").Where("id = ?", id).First(&(site.Attributes)).Error
	if err != nil {
		fmt.Println("There was an error in getting site by ID: " + err.Error())
		return nil, err.Error()
	}
	site.DescriptionJSON = strings.Split(site.DescriptionDB, "XYZ")
	site.Category = "site"
	site.IDJSON = strconv.Itoa(site.ID)
	return site, ""
}

func GetSiteHierarchy(id int) (*Site, string) {
	site, e := GetSite(uint(id))
	if e != "" {
		return nil, e
	}

	site.Buildings, e = GetBuildingsOfParent(id)
	if e != "" {
		return nil, e
	}

	for k, _ := range site.Buildings {
		site.Buildings[k], e = GetBuildingHierarchy(uint(site.Buildings[k].ID))
		if e != "" {
			return nil, e
		}
	}
	return site, ""
}

func GetAllSites() ([]*Site, string) {
	sites := make([]*Site, 0)
	attrs := make([]*Site_Attributes, 0)
	err := GetDB().Table("site").Find(&sites).Error
	if err != nil {
		fmt.Println("There was an error in getting sites by ID: " + err.Error())
		return nil, err.Error()
	}

	err = GetDB().Table("site_attributes").Find(&attrs).Error
	if err != nil {
		fmt.Println("There was an error in getting site attrs by ID: " + err.Error())
		return nil, err.Error()
	}

	for i := range sites {
		sites[i].Category = "site"
		sites[i].Attributes = *(attrs[i])
		sites[i].DescriptionJSON = strings.Split(sites[i].DescriptionDB, "XYZ")
		sites[i].IDJSON = strconv.Itoa(sites[i].ID)
	}
	return sites, ""
}

func DeleteSite(id uint) map[string]interface{} {
	//This is a hard delete!
	e := GetDB().Unscoped().Table("site").Delete(&Site{}, id).RowsAffected

	//The command below is a soft delete
	//Meaning that the 'deleted_at' field will be set
	//the record will remain but unsearchable
	//e := GetDB().Table("tenants").Delete(Tenant{}, id).Error
	if e == 0 {
		return u.Message(false, "There was an error in deleting the site")
	}

	return u.Message(true, "success")
}

func DeleteSitesOfTenant(id uint) map[string]interface{} {

	//First check if the domain is valid
	if GetDB().Table("site").Where("site_parent_id = ?", id).First(&Site{}).Error != nil {
		return u.Message(false, "The parent, tenant, was not found")
	}

	//This is a hard delete!
	e := GetDB().Unscoped().Table("site").
		Where("site_parent_id = ?", id).Delete(&Site{}).Error

	//The command below is a soft delete
	//Meaning that the 'deleted_at' field will be set
	//the record will remain but unsearchable
	//e := GetDB().Table("tenants").Delete(Tenant{}, id).Error
	if e != nil {
		return u.Message(false, "There was an error in deleting the site")
	}

	return u.Message(true, "success")
}

func UpdateSite(id uint, newSiteInfo *Site) (map[string]interface{}, string) {
	site := &Site{}

	err := GetDB().Table("site").Where("id = ?", id).First(site).
		Table("site_attributes").Where("id = ?", id).First(&(site.Attributes)).Error
	if err != nil {
		return u.Message(false, "Site was not found: "+err.Error()), err.Error()
	}

	if newSiteInfo.Name != "" && newSiteInfo.Name != site.Name {
		site.Name = newSiteInfo.Name
	}

	/*if newSiteInfo.Category != "" && newSiteInfo.Category != site.Category {
		site.Category = newSiteInfo.Category
	}*/

	/*if newSiteInfo.Desc != "" && newSiteInfo.Desc != site.Desc {
		site.Desc = newSiteInfo.Desc
	}*/

	if newSiteInfo.Domain != "" && newSiteInfo.Domain != site.Domain {
		site.Domain = newSiteInfo.Domain
	}

	if dc := strings.Join(newSiteInfo.DescriptionJSON, "XYZ"); dc != "" && strings.Compare(dc, site.DescriptionDB) != 0 {
		site.DescriptionDB = dc
	}

	if newSiteInfo.Attributes.Orientation != "" {
		switch newSiteInfo.Attributes.Orientation {
		case "NE", "NW", "SE", "SW":
			site.Attributes.Orientation = newSiteInfo.Attributes.Orientation

		default:
		}
	}

	if newSiteInfo.Attributes.UsableColor != "" && newSiteInfo.Attributes.UsableColor != site.Attributes.UsableColor {
		site.Attributes.UsableColor = newSiteInfo.Attributes.UsableColor
	}

	if newSiteInfo.Attributes.ReservedColor != "" && newSiteInfo.Attributes.ReservedColor != site.Attributes.ReservedColor {
		site.Attributes.ReservedColor = newSiteInfo.Attributes.ReservedColor
	}

	if newSiteInfo.Attributes.TechnicalColor != "" && newSiteInfo.Attributes.TechnicalColor != site.Attributes.TechnicalColor {
		site.Attributes.TechnicalColor = newSiteInfo.Attributes.TechnicalColor
	}

	if newSiteInfo.Attributes.Address != "" && newSiteInfo.Attributes.Address != site.Attributes.Address {
		site.Attributes.Address = newSiteInfo.Attributes.Address
	}

	if newSiteInfo.Attributes.Zipcode != "" && newSiteInfo.Attributes.Zipcode != site.Attributes.Zipcode {
		site.Attributes.Zipcode = newSiteInfo.Attributes.Zipcode
	}

	if newSiteInfo.Attributes.City != "" && newSiteInfo.Attributes.City != site.Attributes.City {
		site.Attributes.City = newSiteInfo.Attributes.City
	}

	if newSiteInfo.Attributes.Country != "" && newSiteInfo.Attributes.Country != site.Attributes.Country {
		site.Attributes.Country = newSiteInfo.Attributes.Country
	}

	if newSiteInfo.Attributes.Gps != "" && newSiteInfo.Attributes.Gps != site.Attributes.Gps {
		site.Attributes.Gps = newSiteInfo.Attributes.Gps
	}

	//Successfully validated the new data
	if e := GetDB().Table("site").Save(site).Table("site_attributes").
		Save(&(site.Attributes)).Error; e != nil {
		return u.Message(false, "Error while updating Site: "+e.Error()), e.Error()
	}
	return u.Message(true, "success"), ""
}

func GetSiteByName(name string) (*Site, string) {
	site := &Site{}

	e := GetDB().Raw(`SELECT * FROM site 
	JOIN site_attributes ON site.id = site_attributes.id 
	WHERE site_name = ?;`, name).Find(site).Find(&site.Attributes).Error

	if e != nil {
		//fmt.Println(e)
		return nil, e.Error()
	}

	site.IDJSON = strconv.Itoa(site.ID)
	site.DescriptionJSON = strings.Split(site.DescriptionDB, "XYZ")
	site.Category = "site"
	return site, ""
}
