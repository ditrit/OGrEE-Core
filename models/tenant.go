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
func GetTenants(user uint) []*Tenant {
	//tenants := make([]*Tenant, 0)
	return nil
}
