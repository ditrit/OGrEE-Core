package models_test

import (
	"cli/models"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeRecursive(t *testing.T) {
	id := "site.building.room.rack"
	path := models.Path{
		Prefix:   models.PhysicalPath,
		ObjectID: id,
	}
	err := path.MakeRecursive(2, 1, models.PhysicalPath+"site")
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "max depth cannot be less than the min depth")

	err = path.MakeRecursive(1, 2, models.PhysicalPath+"site")
	assert.Nil(t, err)
	assert.Equal(t, "**{1,2}."+id, path.ObjectID)

	path.ObjectID = id
	err = path.MakeRecursive(1, -1, models.PhysicalPath+"site")
	assert.Nil(t, err)
	assert.Equal(t, "**{1,}."+id, path.ObjectID)

	id += ".*"
	path.ObjectID = id
	err = path.MakeRecursive(1, 2, models.PhysicalPath+"site")
	assert.Nil(t, err)
	assert.Equal(t, strings.Replace(id, ".*", ".**{1,2}.*", -1), path.ObjectID)
}

func TestSss(t *testing.T) {
	domainPath := models.OrganisationPath + "domain1"
	tests := []struct {
		name          string
		isFunction    func(string) bool
		correctPath   string
		incorrectPath string
	}{
		{"IsPhysical", models.IsPhysical, models.PhysicalPath + "site1", domainPath},
		{"IsStray", models.IsStray, models.StrayPath + "stray-device1", domainPath},
		{"IsObjectTemplate", models.IsObjectTemplate, models.ObjectTemplatesPath + "template1", domainPath},
		{"IsRoomTemplate", models.IsRoomTemplate, models.RoomTemplatesPath + "template1", domainPath},
		{"IsBuildingTemplate", models.IsBuildingTemplate, models.BuildingTemplatesPath + "template1", domainPath},
		{"IsTag", models.IsTag, models.TagsPath + "tag1", domainPath},
		{"IsLayer", models.IsLayer, models.LayersPath + "layer1", domainPath},
		{"IsGroup", models.IsGroup, models.GroupsPath + "group1", domainPath},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.True(t, tt.isFunction(tt.correctPath))
			assert.False(t, tt.isFunction(tt.incorrectPath))
		})
	}
}

func TestSplitPath(t *testing.T) {
	parts := models.SplitPath(models.PhysicalPath + "site1/building")
	assert.Len(t, parts, 4) // first element is nil -> [,Physical,site1,room]
	assert.Equal(t, "site1", parts[2])
	assert.Equal(t, "building", parts[3])
}

func TestJoinPath(t *testing.T) {
	path := models.JoinPath([]string{models.PhysicalPath, "site1", "building"})
	assert.Equal(t, models.PhysicalPath+"/site1/building", path)
}

func TestPhysicalPathToObjectID(t *testing.T) {
	path := models.PhysicalPathToObjectID(models.PhysicalPath + "site1/building")
	assert.Equal(t, "site1.building", path)
}

func TestPhysicalIDToPath(t *testing.T) {
	path := models.PhysicalIDToPath("site1.building")
	assert.Equal(t, models.PhysicalPath+"site1/building", path)
}

func TestPathRemoveLast(t *testing.T) {
	path := models.PhysicalPath + "site1/building/room"
	assert.Equal(t, models.PhysicalPath+"site1/building", models.PathRemoveLast(path, 1))
	assert.Equal(t, path, models.PathRemoveLast(path, 0))
	assert.Equal(t, models.PhysicalPath+"site1", models.PathRemoveLast(path, 2))
}

func TestObjectIDToRelativePath(t *testing.T) {
	id := "site1.building.room"
	assert.Equal(t, "building/room", models.ObjectIDToRelativePath(id, models.PhysicalPath+"site1"))
	assert.Equal(t, "room", models.ObjectIDToRelativePath(id, models.PhysicalPath+"site1/building"))
	assert.Equal(t, "site1/building/room", models.ObjectIDToRelativePath(id, models.PhysicalPath))
}
