package models_test

import (
	"log"
	"p3/models"
	"p3/test/integration"
	"p3/test/unit"
	test_utils "p3/test/utils"
	u "p3/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	// we create a rack with slots and all its parents
	integration.RequireCreateSite("siteA")
	integration.RequireCreateBuilding("siteA", "building-1")
	integration.RequireCreateRoom("siteA.building-1", "room-1")
	managerUserRoles := map[string]models.Role{
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
	rack := test_utils.GetEntityMap("rack", "rack-slots", "siteA.building-1.room-1", integration.TestDBName)
	rack["attributes"].(map[string]any)["template"] = "rack-with-slots"

	_, err := models.CreateEntity(u.OBJTMPL, rackTemplate, managerUserRoles)
	if err != nil {
		log.Fatalln("Error while creating template", err.Error())
	}
	_, err = models.CreateEntity(u.RACK, rack, managerUserRoles)
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
	integration.CreateTestPhysicalEntity(t, u.BLDG, "temporaryBuilding", "temporarySite", true)
	template := test_utils.GetEntityMap("room", "roomA", "temporarySite", integration.TestDBName)

	err := models.ValidateEntity(u.ROOM, template)
	assert.NotNil(t, err)
	assert.Equal(t, "ParentID should correspond to existing building ID", err.Message)

	template["parentId"] = "temporarySite.temporaryBuilding"
	err = models.ValidateEntity(u.ROOM, template)
	assert.Nil(t, err)
}

func TestValidateEntityDeviceParent(t *testing.T) {
	integration.CreateTestPhysicalEntity(t, u.DEVICE, "temporaryDevice", "temporarySite.temporaryBuilding.temporaryRoom.temporaryRack", true)
	template := test_utils.GetEntityMap("device", "deviceA", "temporarySite", integration.TestDBName)

	err := models.ValidateEntity(u.DEVICE, template)
	assert.NotNil(t, err)
	assert.Equal(t, "ParentID should correspond to existing rack or device ID", err.Message)

	template["parentId"] = "temporarySite.temporaryBuilding.temporaryRoom.temporaryRack"
	err = models.ValidateEntity(u.DEVICE, template)
	assert.Nil(t, err)

	template["parentId"] = "temporarySite.temporaryBuilding.temporaryRoom.temporaryRack.temporaryDevice"
	template["name"] = "deviceA"
	delete(template, "id")
	err = models.ValidateEntity(u.DEVICE, template)
	assert.Nil(t, err)
}

func TestValidateEntityDeviceSlot(t *testing.T) {
	template := test_utils.GetEntityMap("device", "deviceA", "siteA.building-1.room-1.rack-slots", integration.TestDBName)
	template["attributes"].(map[string]any)["slot"] = []any{"unknown"}

	err := models.ValidateEntity(u.DEVICE, template)
	assert.NotNil(t, err)
	assert.Equal(t, "Invalid slot: parent does not have all the requested slots", err.Message)

	// We add a valid slot
	template["attributes"].(map[string]any)["slot"] = []any{"u01"}
	err = models.ValidateEntity(u.DEVICE, template)
	assert.Nil(t, err)

	// We add a device to the slot
	delete(template, "id")
	_, err = models.CreateEntity(u.DEVICE, template, integration.ManagerUserRoles)
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
	integration.CreateTestPhysicalEntity(t, u.DEVICE, "device-1", "temporarySite.temporaryBuilding.temporaryRoom.temporaryRack", true)
	integration.CreateTestPhysicalEntity(t, u.DEVICE, "device-2", "temporarySite.temporaryBuilding.temporaryRoom.temporaryRack", false)
	integration.CreateTestPhysicalEntity(t, u.DEVICE, "device-2", "temporarySite.temporaryBuilding.temporaryRoom.temporaryRack.device-1", false)
	template := map[string]any{
		"parentId":    "temporarySite",
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

	template["parentId"] = "temporarySite.temporaryBuilding.temporaryRoom"
	template["name"] = "groupA"
	err = models.ValidateEntity(u.GROUP, template)
	assert.NotNil(t, err)
	assert.Equal(t, "All group objects must be directly under the parent (no . allowed)", err.Message)

	template["parentId"] = "temporarySite.temporaryBuilding.temporaryRoom"
	template["name"] = "groupA"
	template["attributes"].(map[string]any)["content"] = []any{"temporaryRack"}
	delete(template, "id")
	err = models.ValidateEntity(u.GROUP, template)
	assert.Nil(t, err)

	template["parentId"] = "temporarySite.temporaryBuilding.temporaryRoom.temporaryRack"
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
	siteName := "existingSite"
	integration.CreateTestPhysicalEntity(t, u.SITE, siteName, "", false)

	_, err := integration.CreateSite(siteName)
	assert.NotNil(t, err)
	assert.Equal(t, u.ErrDuplicate, err.Type)
	assert.Equal(t, "Error while creating site: Duplicates not allowed", err.Message)
}

func TestCreateCorridorOrGenericWithSameNameAsRackReturnsError(t *testing.T) {
	childName := "tempChild"
	roomId := "tempSite.tempBuilding.tempRoom"
	tests := []struct {
		name           string
		entityType     int
		createFunction func(string, string) (map[string]any, *u.Error)
	}{
		{"CorridorWithSameNameAsRack", u.RACK, integration.CreateCorridor},
		{"RackWithSameNameAsCorridor", u.CORRIDOR, integration.CreateRack},
		{"GenericWithSameNameAsRack", u.RACK, integration.CreateGeneric},
		{"RackWithSameNameAsGeneric", u.GENERIC, integration.CreateRack},
		{"GenericWithSameNameAsCorridor", u.CORRIDOR, integration.CreateGeneric},
		{"CorridorWithSameNameAsGeneric", u.GENERIC, integration.CreateCorridor},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			integration.CreateTestPhysicalEntity(t, tt.entityType, childName, roomId, true)
			_, err := tt.createFunction(roomId, childName)
			assert.NotNil(t, err)
			assert.Equal(t, u.ErrBadFormat, err.Type)
			assert.Equal(t, "Object name must be unique among corridors, racks and generic objects", err.Message)
		})
	}
}

func TestCreateGroupWithObjectThatNotExistsReturnsError(t *testing.T) {
	integration.CreateTestPhysicalEntity(t, u.ROOM, "tempRoom", "tempSite.tempBuilding", true)

	_, err := integration.CreateGroup("tempSite.tempBuilding.tempRoom", "create-object-6", []any{"not-exists"})
	assert.NotNil(t, err)
	assert.Equal(t, u.ErrBadFormat, err.Type)
	assert.Equal(t, "Some object(s) could not be found. Please check and try again", err.Message)
}

func TestCreateGroupWithCorridorsRacksAndGenericWorks(t *testing.T) {
	roomId := "tempSite.tempBuilding.tempRoom"
	rack := integration.CreateTestPhysicalEntity(t, u.RACK, "create-object-7-rack", roomId, true)
	corridor := integration.CreateTestPhysicalEntity(t, u.CORRIDOR, "create-object-7-corridor", roomId, false)
	generic := integration.CreateTestPhysicalEntity(t, u.GENERIC, "create-object-7-generic", roomId, false)

	group, err := integration.CreateGroup(
		roomId,
		"create-object-7",
		[]any{rack["name"].(string), corridor["name"].(string), generic["name"].(string)},
	)
	assert.Nil(t, err)
	unit.HasAttribute(t, group, "content", []any{"create-object-7-rack", "create-object-7-corridor", "create-object-7-generic"})
}

func TestCreateGenericWithParentNotRoomReturnsError(t *testing.T) {
	rack := integration.CreateTestPhysicalEntity(t, u.RACK, "create-object-8-rack", "tempSite.tempBuilding.tempRoom", true)

	_, err := integration.CreateGeneric(rack["id"].(string), "create-object-8-generic")
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "ParentID should correspond to existing room ID")
}

func TestCreateGenericWithParentRoomWorks(t *testing.T) {
	room := integration.CreateTestPhysicalEntity(t, u.ROOM, "create-object-9-room", "tempSite.tempBuilding", true)

	_, err := integration.CreateGeneric(room["id"].(string), "create-object-9-generic")
	assert.Nil(t, err)
}
