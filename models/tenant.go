package models

import (
	u "p3/utils"
)

type Tenant_Attributes struct {
	ID          int    `json:"-" gorm:"column:id"`
	Color       string `json:"color" gorm:"column:tenant_color"`
	MainContact string `json:"mainContact" gorm:"column:main_contact"`
	MainPhone   string `json:"mainPhone" gorm:"column:main_phone"`
	MainEmail   string `json:"mainEmail" gorm:"column:main_email"`
}

type Tenant struct {
	ID              int               `json:"-" gorm:"column:id"`
	IDJSON          string            `json:"id" gorm:"-"`
	Name            string            `json:"name" gorm:"column:tenant_name"`
	Category        string            `json:"category" gorm:"-"`
	Domain          string            `json:"domain" gorm:"column:tenant_domain"`
	ParentID        int               `json:"parentId" gorm:"column:tenant_parent_id"`
	DescriptionJSON []string          `json:"description" gorm:"-"`
	DescriptionDB   string            `json:"-" gorm:"column:tenant_description"`
	Attributes      Tenant_Attributes `json:"attributes" gorm:"-"`
	Sites           []*Site           `json:"sites,omitempty" gorm:"-"`
}

func (tenant *Tenant) Validate() (map[string]interface{}, bool) {

	if tenant.Name == "" {
		return u.Message(false, "Tenant Name should be on payload"), false
	}

	if tenant.Category == "" {
		return u.Message(false, "Category should be on the payload"), false
	}

	if tenant.Domain == "" {
		return u.Message(false, "Domain should be on the payload!"), false
	}

	if tenant.Attributes.Color == "" {
		return u.Message(false, "Color should be on the payload"), false
	}

	//Successfully validated the Tenant
	return u.Message(true, "success"), true
}
