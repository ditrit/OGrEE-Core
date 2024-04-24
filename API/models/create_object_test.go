package models_test

import (
	"log"
	"p3/models"
	"p3/test/integration"
	"p3/test/unit"
	u "p3/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	integration.RequireCreateSite("siteA")
	integration.RequireCreateBuilding("siteA", "building-1")
	integration.RequireCreateRoom("siteA.building-1", "room-1")
	integration.RequireCreateRack("siteA.building-1.room-1", "rack-1")
	integration.RequireCreateDevice("siteA.building-1.room-1.rack-1", "device-1")
	integration.RequireCreateDevice("siteA.building-1.room-1.rack-1", "device-2")
	integration.RequireCreateDevice("siteA.building-1.room-1.rack-1.device-1", "device-2")
	ManagerUserRoles := map[string]models.Role{
		models.ROOT_DOMAIN: models.Manager,
	}
	rackTemplate := map[string]any{
		"slug":        "rack-with-slots",
		"description": "rack with slots",
		"category":    "rack",
		"sizeWDHmm":   []any{605, 1200, 2003},
		"fbxModel":    "",
		"attributes": map[string]any{
			"vendor": "IBM",
			"model":  "9360-4PX",
		},
		"colors":     []any{},
		"components": []any{},
		"slots": []any{
			map[string]any{
				"location":   "u01",
				"type":       "u",
				"elemOrient": []any{33.3, -44.4, 107},
				"elemPos":    []any{58, 51, 44.45},
				"elemSize":   []any{482.6, 1138, 44.45},
				"mandatory":  "no",
				"labelPos":   "frontrear",
				"color":      "@color1",
			},
		},
		"sensors": []any{
			map[string]any{
				"location": "se1",
				"elemPos":  []any{"right", "rear", "upper"},
				"elemSize": []any{50, 20, 20},
			},
		},
	}
	rack := map[string]any{
		"attributes": map[string]any{
			"height":     47,
			"heightUnit": "U",
			"rotation":   []any{45, 45, 45},
			"posXYZ":     []any{4.6666666666667, -2, 0},
			"posXYUnit":  "m",
			"size":       []any{80, 100.532442},
			"sizeUnit":   "cm",
			"template":   "rack-with-slots",
		},
		"category":    "rack",
		"description": "rack with slots",
		"domain":      integration.TestDBName,
		"name":        "rack-slots",
		"parentId":    "siteA.building-1.room-1",
	}

	_, err := models.CreateEntity(u.OBJTMPL, rackTemplate, ManagerUserRoles)
	if err != nil {
		log.Fatalln("Error while creating template", err.Error())
	}
	_, err = models.CreateEntity(u.RACK, rack, ManagerUserRoles)
	if err != nil {
		log.Fatalln("Error while creating rack", err.Error())
	}
}

func TestValidateEntityDomainParent(t *testing.T) {
	template := map[string]any{
		"parentId":    "",
		"name":        "domainA",
		"category":    "domain",
		"description": "domainA",
		"attributes":  map[string]any{},
	}
	err := models.ValidateEntity(u.DOMAIN, template)
	assert.Nil(t, err)
}

func TestValidateEntityRoomParent(t *testing.T) {
	template := map[string]any{
		"parentId":    "siteA",
		"name":        "roomA",
		"category":    "room",
		"description": "roomA",
		"domain":      integration.TestDBName,
		"attributes": map[string]any{
			"floorUnit":       "t",
			"height":          2.8,
			"heightUnit":      "m",
			"axisOrientation": "+x+y",
			"rotation":        -90,
			"posXY":           []any{0, 0},
			"posXYUnit":       "m",
			"size":            []any{-13, -2.9},
			"sizeUnit":        "m",
			"template":        "",
		},
	}
	err := models.ValidateEntity(u.ROOM, template)
	assert.NotNil(t, err)
	assert.Equal(t, "ParentID should correspond to existing building ID", err.Message)

	template["parentId"] = "siteA.building-1"
	err = models.ValidateEntity(u.ROOM, template)
	assert.Nil(t, err)
}

func TestValidateEntityDeviceParent(t *testing.T) {
	template := map[string]any{
		"parentId":    "siteA",
		"name":        "deviceA",
		"category":    "device",
		"description": "deviceA",
		"domain":      integration.TestDBName,
		"attributes": map[string]any{
			"TDP":         "",
			"TDPmax":      "",
			"fbxModel":    "https://github.com/test.fbx",
			"height":      40.1,
			"heightUnit":  "mm",
			"model":       "TNF2LTX",
			"orientation": "front",
			"partNumber":  "0303XXXX",
			"size":        []any{388.4, 205.9},
			"sizeUnit":    "mm",
			"template":    "huawei-xxxxxx",
			"type":        "blade",
			"vendor":      "Huawei",
			"weightKg":    "1.81",
		},
	}
	err := models.ValidateEntity(u.DEVICE, template)
	assert.NotNil(t, err)
	assert.Equal(t, "ParentID should correspond to existing rack or device ID", err.Message)

	template["parentId"] = "siteA.building-1.room-1.rack-1"
	err = models.ValidateEntity(u.DEVICE, template)
	assert.Nil(t, err)

	template["parentId"] = "siteA.building-1.room-1.rack-1.device-1"
	template["name"] = "deviceA"
	delete(template, "id")
	err = models.ValidateEntity(u.DEVICE, template)
	assert.Nil(t, err)
}

func TestValidateEntityDeviceSlot(t *testing.T) {
	template := map[string]any{
		"parentId":    "siteA.building-1.room-1.rack-slots",
		"name":        "deviceA",
		"category":    "device",
		"description": "deviceA",
		"domain":      integration.TestDBName,
		"attributes": map[string]any{
			"slot":        []any{"unknown"},
			"TDP":         "",
			"TDPmax":      "",
			"fbxModel":    "https://github.com/test.fbx",
			"height":      40.1,
			"heightUnit":  "mm",
			"model":       "TNF2LTX",
			"orientation": "front",
			"partNumber":  "0303XXXX",
			"size":        []any{388.4, 205.9},
			"sizeUnit":    "mm",
			"template":    "huawei-xxxxxx",
			"type":        "blade",
			"vendor":      "Huawei",
			"weightKg":    "1.81",
		},
	}
	err := models.ValidateEntity(u.DEVICE, template)
	assert.NotNil(t, err)
	assert.Equal(t, "Invalid slot: parent does not have all the requested slots", err.Message)

	// We add a valid slot
	template["attributes"].(map[string]any)["slot"] = []any{"u01"}
	err = models.ValidateEntity(u.DEVICE, template)
	assert.Nil(t, err)

	// We add a device to the slot
	delete(template, "id")
	ManagerUserRoles := map[string]models.Role{
		models.ROOT_DOMAIN: models.Manager,
	}
	_, err = models.CreateEntity(u.DEVICE, template, ManagerUserRoles)
	assert.Nil(t, err, "The device")

	// we verify if we can add another device in the same slot
	template["attributes"].(map[string]any)["slot"] = []any{"u01"}
	template["name"] = "deviceB"
	delete(template, "id")
	delete(template, "createdDate")
	delete(template, "lastUpdated")
	err = models.ValidateEntity(u.DEVICE, template)
	assert.NotNil(t, err)
	assert.Equal(t, "Invalid slot: one or more requested slots are already in use", err.Message)
}

func TestValidateEntityGroupParent(t *testing.T) {
	template := map[string]any{
		"parentId":    "siteA",
		"name":        "groupA",
		"category":    "group",
		"description": "groupA",
		"domain":      integration.TestDBName,
		"attributes": map[string]any{
			"content": []any{"device-1", "device-1.device-2"},
		},
	}
	err := models.ValidateEntity(u.GROUP, template)
	assert.NotNil(t, err)
	assert.Equal(t, "Group parent should correspond to existing rack or room", err.Message)

	template["parentId"] = "siteA.building-1.room-1"
	template["name"] = "groupA"
	err = models.ValidateEntity(u.GROUP, template)
	assert.NotNil(t, err)
	assert.Equal(t, "All group objects must be directly under the parent (no . allowed)", err.Message)

	template["parentId"] = "siteA.building-1.room-1"
	template["name"] = "groupA"
	template["attributes"].(map[string]any)["content"] = []any{"rack-1"}
	delete(template, "id")
	err = models.ValidateEntity(u.GROUP, template)
	assert.Nil(t, err)

	template["parentId"] = "siteA.building-1.room-1.rack-1"
	template["name"] = "groupA"
	template["attributes"].(map[string]any)["content"] = []any{"device-1", "device-2"}
	delete(template, "id")
	err = models.ValidateEntity(u.GROUP, template)
	assert.Nil(t, err)
}

func TestCreateRackWithoutAttributesReturnsError(t *testing.T) {
	_, err := models.CreateEntity(
		u.RACK,
		map[string]any{
			"category":    "rack",
			"description": "rack",
			"domain":      integration.TestDBName,
			"name":        "create-object-1",
			"tags":        []any{},
		},
		integration.ManagerUserRoles,
	)
	assert.NotNil(t, err)
	assert.Equal(t, u.ErrBadFormat, err.Type)
	assert.Equal(t, "JSON body doesn't validate with the expected JSON schema", err.Message)
}

func TestCreateObjectWithDuplicatedNameReturnsError(t *testing.T) {
	site := integration.RequireCreateSite("create-object-1")

	_, err := integration.CreateSite(site["name"].(string))
	assert.NotNil(t, err)
	assert.Equal(t, u.ErrDuplicate, err.Type)
	assert.Equal(t, "Error while creating site: Duplicates not allowed", err.Message)
}

func TestCreateCorridorWithSameNameAsRackReturnsError(t *testing.T) {
	rack := integration.RequireCreateRack("", "create-object-2")

	_, err := integration.CreateCorridor(rack["parentId"].(string), "create-object-2")
	assert.NotNil(t, err)
	assert.Equal(t, u.ErrBadFormat, err.Type)
	assert.Equal(t, "Object name must be unique among corridors, racks and generic objects", err.Message)
}

func TestCreateRackWithSameNameAsCorridorReturnsError(t *testing.T) {
	corridor := integration.RequireCreateCorridor("", "create-object-3")

	_, err := integration.CreateRack(corridor["parentId"].(string), "create-object-3")
	assert.NotNil(t, err)
	assert.Equal(t, u.ErrBadFormat, err.Type)
	assert.Equal(t, "Object name must be unique among corridors, racks and generic objects", err.Message)
}

func TestCreateGenericWithSameNameAsRackReturnsError(t *testing.T) {
	rack := integration.RequireCreateRack("", "create-object-4")

	_, err := integration.CreateGeneric(rack["parentId"].(string), "create-object-4")
	assert.NotNil(t, err)
	assert.Equal(t, u.ErrBadFormat, err.Type)
	assert.Equal(t, "Object name must be unique among corridors, racks and generic objects", err.Message)
}

func TestCreateGenericWithSameNameAsCorridorReturnsError(t *testing.T) {
	corridor := integration.RequireCreateCorridor("", "create-object-5")

	_, err := integration.CreateGeneric(corridor["parentId"].(string), "create-object-5")
	assert.NotNil(t, err)
	assert.Equal(t, u.ErrBadFormat, err.Type)
	assert.Equal(t, "Object name must be unique among corridors, racks and generic objects", err.Message)
}

func TestCreateGroupWithObjectThatNotExistsReturnsError(t *testing.T) {
	room := integration.RequireCreateRoom("", "create-object-6-room")

	_, err := integration.CreateGroup(room["id"].(string), "create-object-6", []any{"not-exists"})
	assert.NotNil(t, err)
	assert.Equal(t, u.ErrBadFormat, err.Type)
	assert.Equal(t, "Some object(s) could not be found. Please check and try again", err.Message)
}

func TestCreateGroupWithCorridorsRacksAndGenericWorks(t *testing.T) {
	room := integration.RequireCreateRoom("", "create-object-7-room")
	rack := integration.RequireCreateRack(room["id"].(string), "create-object-7-rack")
	corridor := integration.RequireCreateCorridor(room["id"].(string), "create-object-7-corridor")
	generic := integration.RequireCreateGeneric(room["id"].(string), "create-object-7-generic")

	group, err := integration.CreateGroup(
		room["id"].(string),
		"create-object-7",
		[]any{rack["name"].(string), corridor["name"].(string), generic["name"].(string)},
	)
	assert.Nil(t, err)
	unit.HasAttribute(t, group, "content", []any{"create-object-7-rack", "create-object-7-corridor", "create-object-7-generic"})
}

func TestCreateGenericWithParentNotRoomReturnsError(t *testing.T) {
	rack := integration.RequireCreateRack("", "create-object-8-rack")

	_, err := integration.CreateGeneric(rack["id"].(string), "create-object-8-generic")
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "ParentID should correspond to existing room ID")
}

func TestCreateGenericWithParentRoomWorks(t *testing.T) {
	room := integration.RequireCreateRoom("", "create-object-9-room")

	_, err := integration.CreateGeneric(room["id"].(string), "create-object-9-generic")
	assert.Nil(t, err)
}
