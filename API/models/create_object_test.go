package models_test

import (
	"p3/models"
	"p3/test/integration"
	"p3/test/unit"
	u "p3/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

	_, err := integration.CreateGroup(room["id"].(string), "create-object-6", []string{"not-exists"})
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
		[]string{rack["name"].(string), corridor["name"].(string), generic["name"].(string)},
	)
	assert.Nil(t, err)
	unit.HasAttribute(t, group, "content", "create-object-7-rack,create-object-7-corridor,create-object-7-generic")
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
