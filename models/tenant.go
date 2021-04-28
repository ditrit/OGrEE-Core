package models

import (
	"fmt"
	u "p3/utils"
	"strconv"
	"strings"
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

func (tenant *Tenant) Create() (map[string]interface{}, string) {
	if resp, ok := tenant.Validate(); !ok {
		return resp, "validate"
	}
	//Strategy for inserting into both tables
	//Otherwise make 2 insert statements

	tenant.DescriptionDB = strings.Join(tenant.DescriptionJSON, "XYZ")
	if e := GetDB().Create(tenant).Error; e != nil {
		return u.Message(false, "Internal error while creating Tenant: "+e.Error()),
			e.Error()
	}

	/*GetDB().Exec(`UPDATE tenant SET tenant_description = ?
	WHERE tenant.id = ?`, pq.Array(tenant.Description), tenant.ID)*/
	//tenant.ID, _ = strconv.Atoi(tenant.IDJSON)
	tenant.IDJSON = strconv.Itoa(tenant.ID)
	println("Tenant id is: ", tenant.ID)
	tenant.Attributes.ID = tenant.ID

	if e := GetDB().Table("tenant_attributes").Create(&tenant.Attributes).Error; e != nil {
		return u.Message(false, "Internal error while creating Tenant Attrs: "+e.Error()),
			e.Error()
	}

	resp := u.Message(true, "success")
	resp["tenant"] = tenant
	return resp, ""
}

func (t *Tenant) FormQuery() string {

	query := "SELECT * FROM tenant " + u.JoinQueryGen("tenant")
	if t.Name != "" {
		query += " WHERE tenant_name = '" + t.Name + "'"
	}
	if t.Category != "" {
		query += " AND tenant_category = '" + t.Category + "'"
	}
	if t.Domain != "" {
		query += " AND tenant_domain = '" + t.Domain + "'"
	}
	if (Tenant_Attributes{}) != t.Attributes {
		if t.Attributes.Color != "" {
			query +=
				" AND tenant_attributes.tenant_color = '" +
					t.Attributes.Color + "'"
		}
		if t.Attributes.MainContact != "" {
			query +=
				" AND tenant_attributes.main_contact = '" +
					t.Attributes.MainContact + "'"
		}
		if t.Attributes.MainEmail != "" {
			query +=
				" AND tenant_attributes.main_email = '" +
					t.Attributes.MainEmail + "'"
		}
		if t.Attributes.MainPhone != "" {
			query +=
				" AND tenant_attributes.main_phone = '" +
					t.Attributes.MainPhone + "'"
		}
	}
	return query
}

func GetTenant(id uint) (*Tenant, string) {
	tenant := &Tenant{}

	e := GetDB().Table("tenant").Where("id = ?", id).First(tenant).
		Table("tenant_attributes").Where("id = ?", id).First(&(tenant.Attributes)).
		Error

	if e != nil {
		//fmt.Println(e)
		return nil, e.Error()
	}

	tenant.DescriptionJSON = strings.Split(tenant.DescriptionDB, "XYZ")
	tenant.IDJSON = strconv.Itoa(tenant.ID)
	tenant.Category = "tenant"
	return tenant, ""
}

func GetAllTenants() ([]*Tenant, string) {
	tenants := make([]*Tenant, 0)
	attrs := make([]*Tenant_Attributes, 0)
	err := GetDB().Table("tenant").Find(&tenants).Error
	if err != nil {
		fmt.Println("There was an error in getting tenants")
		return nil, err.Error()
	}

	err = GetDB().Table("tenant_attributes").Find(&attrs).Error
	if err != nil {
		fmt.Println("There was an error in getting tenant attrs")
		return nil, err.Error()
	}

	for i := range tenants {
		tenants[i].Category = "tenant"
		tenants[i].Attributes = *(attrs[i])
		tenants[i].DescriptionJSON = strings.Split(tenants[i].DescriptionDB, "XYZ")
		tenants[i].IDJSON = strconv.Itoa(tenants[i].ID)
	}
	return tenants, ""
}

func GetTenantHierarchy(id int) (*Tenant, string) {
	tn, e := GetTenant(uint(id))
	if e != "" {
		return nil, e
	}

	tn.Sites, e = GetSitesOfParent(id)
	if e != "" {
		return nil, e
	}

	for k, _ := range tn.Sites {
		tn.Sites[k], e = GetSiteHierarchy(tn.Sites[k].ID)
		if e != "" {
			return nil, e
		}
	}
	return tn, ""
}

func GetTenantHierarchyNonStandard(id int) (*Tenant, []*Site,
	*[][]*Building, *[][]*Room, *[][]*Rack, *[][]*Device, string) {
	tn, e := GetTenant(uint(id))
	if e != "" {
		return nil, nil, nil, nil, nil, nil, e
	}

	sites, e := GetSitesOfParent(id)
	if e != "" {
		return nil, nil, nil, nil, nil, nil, e
	}
	buildings := make([][]*Building, len(sites))
	rooms := make([][]*Room, 1)
	racks := make([][]*Rack, 1)
	devices := make([][]*Device, 1)
	tmpbuildings := new([][]*Building)
	tmprooms := new([][]*Room)
	tmpracks := new([][]*Rack)
	tmpdevices := new([][]*Device)

	for k, _ := range sites {
		_, buildings[k], tmprooms, tmpracks,
			tmpdevices, e = GetSiteHierarchyNonStandard(sites[k].ID)
		if e != "" {
			return nil, nil, nil, nil, nil, nil, e
		}
		*tmpbuildings = append(*tmpbuildings, buildings...)
		rooms = append(rooms, *tmprooms...)
		racks = append(racks, *tmpracks...)
		devices = append(devices, *tmpdevices...)
	}
	return tn, sites, tmpbuildings, &rooms, &racks, &devices, ""
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
			return u.Message(false, "Internal Error: "+err.Error()), "internal"
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
		tenant.Attributes.MainContact = t.Attributes.MainContact
	}

	if t.Attributes.MainEmail != "" && t.Attributes.MainEmail != tenant.Attributes.MainEmail {
		tenant.Attributes.MainEmail = t.Attributes.MainEmail
	}

	if t.Attributes.MainPhone != "" && t.Attributes.MainPhone != tenant.Attributes.MainPhone {
		tenant.Attributes.MainPhone = t.Attributes.MainPhone
	}

	//fmt.Println(t.Description)
	if e := GetDB().Table("tenant").Save(tenant).Table("tenant_attributes").
		Save(&(tenant.Attributes)).Error; e != nil {
		return u.Message(false, "Failed to update Tenant: "+e.Error()), "internal"
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

func GetTenantByName(name string) (*Tenant, string) {
	tenant := &Tenant{}

	/*e := GetDB().Raw(`SELECT * FROM tenant
	JOIN tenant_attributes ON tenant.id = tenant_attributes.id
	WHERE tenant_name = ?;`, name).Find(tenant).Find(&tenant.Attributes).Error*/

	e := GetDB().Find(tenant).Find(&tenant.Attributes).Error

	if e != nil {
		//fmt.Println(e)
		return nil, e.Error()
	}

	tenant.IDJSON = strconv.Itoa(tenant.ID)
	tenant.DescriptionJSON = strings.Split(tenant.DescriptionDB, "XYZ")
	tenant.Category = "tenant"
	return tenant, ""
}

func GetTenantByQuery(q *Tenant) ([]*Tenant, string) {
	tenants := make([]*Tenant, 0)
	attrs := make([]*Tenant_Attributes, 0)

	e := GetDB().Raw(q.FormQuery()).Find(&tenants).
		Find(&attrs).Error

	if e != nil {
		return nil, e.Error()
	}

	for i := range tenants {
		tenants[i].Attributes = *(attrs[i])
		tenants[i].IDJSON = strconv.Itoa(tenants[i].ID)
		tenants[i].DescriptionJSON =
			strings.Split(tenants[i].DescriptionDB, "XYZ")
		tenants[i].Category = "tenant"
	}

	return tenants, ""
}

func GetSitesOfTenant(name string) ([]*Site, string) {
	tenant, e := GetTenantByName(name)
	if e != "" {
		return nil, e
	}
	sites, e1 := GetSitesOfParent(tenant.ID)
	if e1 != "" {
		return nil, e1
	}
	return sites, ""
}

func GetNamedSiteOfTenant(tenant_name, site_name string) (*Site, string) {
	tenant, e := GetTenantByName(tenant_name)
	if e != "" {
		return nil, e
	}

	site, e1 := GetSiteByNameAndParentID(tenant.ID, site_name)
	if e1 != "" {
		return nil, e1
	}
	return site, ""
}

func GetBuildingsUsingNamedSiteOfTenant(tenant_name, site_name string) ([]*Building, string) {
	site, e := GetNamedSiteOfTenant(tenant_name, site_name)
	if e != "" {
		return nil, e
	}
	bldgs, e2 := GetBuildingsOfParent(site.ID)
	if e2 != "" {
		return nil, e2
	}
	return bldgs, ""
}

func GetNamedBuildingOfTenant(tenant_name, site_name, bldg_name string) (*Building, string) {
	site, e := GetNamedSiteOfTenant(tenant_name, site_name)
	if e != "" {
		return nil, e
	}

	bldg, e2 := GetBuildingByNameAndParentID(site.ID, bldg_name)
	if e2 != "" {
		return nil, e2
	}
	return bldg, ""
}

func GetRoomsUsingNamedBuildingOfTenant(tenant_name, site_name, bldg_name string) ([]*Room, string) {
	bldg, e := GetNamedBuildingOfTenant(tenant_name, site_name, bldg_name)
	if e != "" {
		return nil, e
	}

	rooms, e2 := GetRoomsOfParent(uint(bldg.ID))
	if e2 != "" {
		return nil, e2
	}
	return rooms, ""
}

func GetNamedRoomOfTenant(tenant_name, site_name, bldg_name, room_name string) (*Room, string) {
	bldg, e := GetNamedBuildingOfTenant(tenant_name, site_name, bldg_name)
	if e != "" {
		return nil, e
	}

	room, e2 := GetRoomByNameAndParentID(bldg.ID, room_name)
	if e2 != "" {
		return nil, e2
	}
	return room, ""
}

func GetRacksUsingNamedRoomOfTenant(tenant_name, site_name, bldg_name, room_name string) ([]*Rack, string) {
	room, e := GetNamedRoomOfTenant(tenant_name, site_name, bldg_name, room_name)
	if e != "" {
		return nil, e
	}

	racks, e2 := GetRacksOfParent(uint(room.ID))
	if e2 != "" {
		return nil, e2
	}
	return racks, ""
}

func GetNamedRackOfTenant(tenant_name, site_name, bldg_name, room_name, rack_name string) (*Rack, string) {
	room, e := GetNamedRoomOfTenant(tenant_name, site_name, bldg_name, room_name)
	if e != "" {
		return nil, e
	}

	rack, e2 := GetRackByNameAndParentID(room.ID, rack_name)
	if e2 != "" {
		return nil, e2
	}
	return rack, ""
}

func GetDevicesUsingNamedRackOfTenant(tenant_name, site_name, bldg_name, room_name, rack_name string) ([]*Device, string) {
	rack, e := GetNamedRackOfTenant(tenant_name, site_name, bldg_name, room_name, rack_name)
	if e != "" {
		return nil, e
	}

	devices, e2 := GetDevicesOfParent(uint(rack.ID))
	if e2 != "" {
		return nil, e2
	}
	return devices, ""
}

func GetNamedDeviceOfTenant(tenant_name, site_name, bldg_name, room_name, rack_name, dev_name string) (*Device, string) {
	rack, e := GetNamedRackOfTenant(tenant_name, site_name, bldg_name, room_name, rack_name)
	if e != "" {
		return nil, e
	}

	device, e2 := GetDeviceByNameAndParentID(uint(rack.ID), dev_name)
	if e2 != "" {
		return nil, e2
	}
	return device, ""
}
