package models

import (
	u "p3/utils"

	"github.com/jinzhu/gorm"
)

type Tenant struct {
	gorm.Model
	Name     string `json:"name"`
	Category string `json:"category"`
	Desc     string `json:"description"`
	Domain   string `json:"domain"`
	Color    string `json:"color"`
	Site     []Site
}

func (tenant *Tenant) Validate() (map[string]interface{}, bool) {

	if tenant.Name == "" {
		return u.Message(false, "Tenant Name should be on payload"), false
	}

	if tenant.Category == "" {
		return u.Message(false, "Category should be on the payload"), false
	}

	if tenant.Desc == "" {
		return u.Message(false, "Description should be on the paylad"), false
	}

	if tenant.Domain != "" {
		return u.Message(false, "Domain should be NULL!"), false
	}

	if tenant.Color == "" {
		return u.Message(false, "Color should be on the payload"), false
	}

	//Successfully validated the Tenant
	return u.Message(true, "success"), true
}

func (tenant *Tenant) Create() map[string]interface{} {
	if resp, ok := tenant.Validate(); !ok {
		return resp
	}

	GetDB().Create(tenant)

	resp := u.Message(true, "success")
	resp["tenant"] = tenant
	return resp
}

func GetTenant(id uint) *Tenant {
	tenant := &Tenant{}

	err := GetDB().Table("tenants").Where("id = ?", id).First(tenant).Error
	if err != nil {
		return nil
	}
	return tenant
}

//More methods should be made to
//Meet CRUD capabilities
//Need Update and Delete
//These would be a bit more complicated
//So leave them out for now

//It may be possible for an API
//user to have more than 1 tenant
//This isn't realistic so this
//will not be implemented
func GetAllTenants() []*Tenant {
	tenants := make([]*Tenant, 0)

	err := GetDB().Table("tenants").Find(&tenants).Error
	if err != nil {
		return nil
	}
	return tenants
}

//Only update valid fields
//If any fields are invalid
//Message will still be successful
func UpdateTenant(id uint, t *Tenant) map[string]interface{} {
	tenant := &Tenant{}

	err := GetDB().Table("tenants").Where("id = ?", id).First(tenant).Error
	if err != nil {
		return u.Message(false, "Tenant was not found")
	}

	if t.Name != "" && t.Name != tenant.Name {
		tenant.Name = t.Name
	}

	if t.Category != "" && t.Category != tenant.Category {
		tenant.Category = t.Category
	}

	if t.Desc != "" && t.Desc != tenant.Desc {
		tenant.Desc = t.Desc
	}

	if t.Domain != "" && t.Domain != tenant.Domain {
		tenant.Domain = t.Domain
	}

	if t.Color != "" && t.Color != tenant.Color {
		tenant.Color = t.Color
	}

	GetDB().Table("tenants").Save(tenant)
	//.Update(tenant)
	return u.Message(true, "success")
}

func DeleteTenant(id uint) map[string]interface{} {

	//This command is a hard delete!
	e := GetDB().Unscoped().Table("tenants").Delete(Tenant{}, id).Error

	//The command below is a soft delete
	//Meaning that the 'deleted_at' field will be set
	//the record will remain but unsearchable
	//e := GetDB().Table("tenants").Delete(Tenant{}, id).Error
	if e != nil {
		return u.Message(false, "Tenant was not found")
	}

	return u.Message(true, "success")
}
