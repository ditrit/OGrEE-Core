package models

import (
	u "p3/utils"
	"time"
)

type Tenant_Attributes struct {
	ID uint `gorm:\"primary_key\" gorm: "id"`
	/*CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *(time.Time) "sql:\"index\""*/
	Color       string `gorm:"column:tenant_color" json:"color"`
	MainContact string `json:"mainContact"`
	MainPhone   string `json:"mainPhone"`
	MainEmail   string `json:"mainEmail"`
}

type Tenant struct {
	//gorm.Model
	ID        uint `gorm:\"primary_key\" gorm: "id"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *(time.Time) "sql:\"index\""
	//gorm:"column:beast_id"
	Name       string            `gorm:"column:tenant_name" json:"name"`
	ParentID   string            `gorm:"column:tenant_parent_id" json:"parentId"`
	Category   string            `gorm:"column:tenant_category" json:"category" gorm:"-"`
	Domain     string            `gorm:"column:tenant_domain" json:"domain"`
	Attributes Tenant_Attributes `json:"attributes"`
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

func (tenant *Tenant) Create() map[string]interface{} {
	if resp, ok := tenant.Validate(); !ok {
		return resp
	}
	//Strategy for inserting into both tables
	//Otherwise make 2 insert statements
	GetDB().Table("tenant").Select("tenant_name",
		"tenant_domain", "tenant_description").Create(&tenant)

	tenant.Attributes.ID = tenant.ID

	/*GetDB().Table("tenant_attributes").Select("id", "tenant_color", "main_contact",
	"main_phone", "main_email").Create(&tenant.Tenant_Attributes)*/

	GetDB().Table("tenant_attributes").Create(&tenant.Attributes)

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

	/*if t.Name != "" && t.Name != tenant.Name {
		tenant.Name = t.Name
	}

	if t.Category != "" && t.Category != tenant.Category {
		tenant.Category = t.Category
	}*/

	/*if t.Desc != "" && t.Desc != tenant.Desc {
		tenant.Desc = t.Desc
	}

	if t.Color != "" && t.Color != tenant.Color {
		tenant.Color = t.Color
	}*/

	GetDB().Table("tenants").Save(tenant)
	//.Update(tenant)
	return u.Message(true, "success")
}

func DeleteTenant(id uint) map[string]interface{} {

	//This command is a hard delete!
	e := GetDB().Unscoped().Table("tenants").Delete(Tenant{}, id).RowsAffected

	//The command below is a soft delete
	//Meaning that the 'deleted_at' field will be set
	//the record will remain but unsearchable
	//e := GetDB().Table("tenants").Delete(Tenant{}, id).Error
	if e == 0 {
		return u.Message(false, "Tenant was not found")
	}

	return u.Message(true, "success")
}
