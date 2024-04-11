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

func TestIsPhysical(t *testing.T) {
	assert.True(t, models.IsPhysical(models.PhysicalPath+"site1"))
	assert.False(t, models.IsPhysical(models.OrganisationPath+"domain1/"))
}

func TestIsStray(t *testing.T) {
	assert.True(t, models.IsStray(models.StrayPath+"stray-device1"))
	assert.False(t, models.IsStray(models.OrganisationPath+"domain1"))
}

func TestIsObjectTemplate(t *testing.T) {
	assert.True(t, models.IsObjectTemplate(models.ObjectTemplatesPath+"template1"))
	assert.False(t, models.IsObjectTemplate(models.OrganisationPath+"domain1"))
}

func TestIsRoomTemplate(t *testing.T) {
	assert.True(t, models.IsRoomTemplate(models.RoomTemplatesPath+"template1"))
	assert.False(t, models.IsRoomTemplate(models.OrganisationPath+"domain1"))
}

func TestIsBuildingTemplate(t *testing.T) {
	assert.True(t, models.IsBuildingTemplate(models.BuildingTemplatesPath+"template1"))
	assert.False(t, models.IsBuildingTemplate(models.OrganisationPath+"domain1"))
}

func TestIsTag(t *testing.T) {
	assert.True(t, models.IsTag(models.TagsPath+"tag1"))
	assert.False(t, models.IsTag(models.OrganisationPath+"domain1"))
}

func TestIsLayer(t *testing.T) {
	assert.True(t, models.IsLayer(models.LayersPath+"layer1"))
	assert.False(t, models.IsLayer(models.OrganisationPath+"domain1"))
}

func TestIsGroup(t *testing.T) {
	assert.True(t, models.IsGroup(models.GroupsPath+"group1"))
	assert.False(t, models.IsGroup(models.OrganisationPath+"domain1"))
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
