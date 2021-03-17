package models

import (
	"fmt"
	u "p3/utils"
	"strings"
)

type Tenant_Attributes struct {
	ID          int    `json:"id" gorm:"column:id"`
	Color       string `json:"color" gorm:"column:tenant_color"`
	MainContact string `json:"mainContact" gorm:"column:main_contact"`
	MainPhone   string `json:"mainPhone" gorm:"column:main_phone"`
	MainEmail   string `json:"mainEmail" gorm:"column:main_email"`
}

type Tenant struct {
	//gorm.Model
	ID              int               `json:"id" gorm:"column:id"`
	Name            string            `json:"name" gorm:"column:tenant_name"`
	Category        string            `json:"category" gorm:"-"`
	Domain          string            `json:"domain" gorm:"column:tenant_domain"`
	ParentID        int               `json:"parentId" gorm:"column:tenant_parent_id"`
	DescriptionJSON []string          `json:"description" gorm:"-"`
	DescriptionDB   string            `json:"-" gorm:"column:tenant_description"`
	Attributes      Tenant_Attributes `json:"attributes"`
}

func (tenant *Tenant) Validate() (map[string]interface{}, bool) {

	if tenant.Name == "" {
		return u.Message(false, "Tenant Name should be on payload"), false
	}

	if tenant.Category == "" {
		return u.Message(false, "Category should be on the payload"), false
	}

	/*if tenant.Desc == "" {
		return u.Message(false, "Description should be on the paylad"), false
	}*/

	if tenant.Domain == "" {
		return u.Message(false, "Domain should be on the payload!"), false
	}

	if tenant.Attributes.Color == "" {
		return u.Message(false, "Color should be on the payload"), false
	}

	if tenant.Attributes.MainContact == "" {
		return u.Message(false, "MainContact should be on the payload"), false
	}

	if tenant.Attributes.MainPhone == "" {
		return u.Message(false, "MainPhone should be on the payload"), false
	}

	if tenant.Attributes.MainEmail == "" {
		return u.Message(false, "MainEmail should be on the payload"), false
	}

	//Successfully validated the Tenant
	return u.Message(true, "success"), true
}

func (tenant *Tenant) Create() (map[string]interface{}, string) {
	if resp, ok := tenant.Validate(); !ok {
		return resp, "validate"
	}
	//Strategy for inserting into both tables
	//Otherwise make 2 insert statements

	tenant.DescriptionDB = strings.Join(tenant.DescriptionJSON, "XYZ")
	if e := GetDB().Create(tenant).Error; e != nil {
		return u.Message(false, "Internal error while creating Tenant: "+e.Error()),
			"internal"
	}

	//This link explains JSON marshalling which will
	//be needed to merge the SQL Query below to the Query
	//above
	//Alot of code will be added to support the
	//custom JSON Marshalling
	//https://attilaolah.eu/2013/11/29/json-decoding-in-go/

	/*GetDB().Exec(`UPDATE tenant SET tenant_description = ?
	WHERE tenant.id = ?`, pq.Array(tenant.Description), tenant.ID)*/

	tenant.Attributes.ID = tenant.ID

	if e := GetDB().Table("tenant_attributes").Create(&tenant.Attributes).Error; e != nil {
		return u.Message(false, "Internal error while creating Tenant Attrs: "+e.Error()),
			"internal"
	}

	resp := u.Message(true, "success")
	resp["tenant"] = tenant
	return resp, ""
}

func GetTenant(id uint) (*Tenant, string) {
	tenant := &Tenant{}

	e := GetDB().Table("tenant").Where("id = ?", id).First(tenant).
		Table("tenant_attributes").Where("id = ?", id).First(&(tenant.Attributes)).
		Error

	if e != nil {
		//fmt.Println("BRUH")
		//fmt.Println(e)
		// e = record not found
		return nil, e.Error()
	}

	//r.Scan(tenant, &(tenant.Attributes))
	tenant.DescriptionJSON = strings.Split(tenant.DescriptionDB, "XYZ")
	return tenant, ""
}

func GetAllTenants() []*Tenant {
	tenants := make([]*Tenant, 0)
	attrs := make([]*Tenant_Attributes, 0)
	err := GetDB().Table("tenant").Find(&tenants).Error
	if err != nil {
		fmt.Println("There was an error in getting tenant by ID")
		return nil
	}

	err = GetDB().Table("tenant_attributes").Find(&attrs).Error
	if err != nil {
		fmt.Println("There was an error in getting tenant attrs by ID")
		return nil
	}

	for i := range tenants {
		tenants[i].Attributes = *(attrs[i])
		tenants[i].DescriptionJSON = strings.Split(tenants[i].DescriptionDB, "XYZ")
	}
	return tenants
}

//Only update valid fields
//If any fields are invalid
//Message will still be successful
func UpdateTenant(id uint, t *Tenant) (map[string]interface{}, string) {
	tenant := &Tenant{}
	err := GetDB().Table("tenant").Where("id = ?", id).First(tenant).
		Table("tenant_attributes").Where("id = ?", id).First(&(tenant.Attributes)).Error

	if err != nil {
		if err.Error() != "record not found" {
			return u.Message(false, "Internal Error"), "internal"
		}
		return u.Message(false, "Tenant was not found"), "record not found"
	}

	if t.Name != "" && t.Name != tenant.Name {
		tenant.Name = t.Name
	}

	if dc := strings.Join(t.DescriptionJSON, "XYZ"); dc != "" && strings.Compare(dc, tenant.DescriptionDB) != 0 {
		tenant.DescriptionDB = dc
	}

	/*if t.Category != "" && t.Category != tenant.Category {
		tenant.Category = t.Category
	}*/

	if t.Domain != "" && t.Domain != tenant.Domain {
		tenant.Domain = t.Domain
	}

	/*if t.Desc != "" && t.Desc != tenant.Desc {
		tenant.Desc = t.Desc
	}*/

	if t.Attributes.Color != "" && t.Attributes.Color != tenant.Attributes.Color {
		tenant.Attributes.Color = t.Attributes.Color
	}

	if t.Attributes.MainContact != "" && t.Attributes.MainContact != tenant.Attributes.MainContact {
		tenant.Attributes.Color = t.Attributes.Color
	}

	if t.Attributes.MainEmail != "" && t.Attributes.MainEmail != tenant.Attributes.MainEmail {
		tenant.Attributes.MainEmail = t.Attributes.MainEmail
	}

	if t.Attributes.MainPhone != "" && t.Attributes.MainPhone != tenant.Attributes.MainPhone {
		tenant.Attributes.MainPhone = t.Attributes.MainPhone
	}

	//fmt.Println(t.Description)
	if GetDB().Table("tenant").Save(tenant).Table("tenant_attributes").
		Save(&(tenant.Attributes)).Error != nil {
		return u.Message(false, "failure"), "internal"
	}

	return u.Message(true, "success"), ""
}

func DeleteTenant(id uint) map[string]interface{} {

	//This command is a hard delete!
	e := GetDB().Unscoped().Table("tenant").Delete(Tenant{}, id).RowsAffected

	//The command below is a soft delete
	//Meaning that the 'deleted_at' field will be set
	//the record will remain but unsearchable
	//e := GetDB().Table("tenants").Delete(Tenant{}, id).Error
	if e == 0 {
		return u.Message(false, "Tenant was not found")
	}

	return u.Message(true, "success")
}
