package integration

import (
	"log"
	"p3/models"
	test_utils "p3/test/utils"
	"p3/utils"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

var ManagerUserRoles = map[string]models.Role{
	models.ROOT_DOMAIN: models.Manager,
}

func createObject(entity int, obj map[string]interface{}, require bool) (map[string]any, *utils.Error) {
	createdObj, err := models.CreateEntity(
		entity,
		obj,
		ManagerUserRoles,
	)

	if require && err != nil {
		log.Fatalln(err.Error())
	}

	return createdObj, err
}

func internalCreateSite(name string, require bool) (map[string]any, *utils.Error) {
	site := test_utils.GetEntityMap("site", name, "", TestDBName)
	return createObject(
		utils.SITE,
		site,
		require,
	)
}

func RequireCreateSite(name string) map[string]any {
	obj, _ := internalCreateSite(name, true)
	return obj
}

func CreateSite(name string) (map[string]any, *utils.Error) {
	return internalCreateSite(name, false)
}

func internalCreateBuilding(siteID, name string, require bool) (map[string]any, *utils.Error) {
	if siteID == "" {
		site := RequireCreateSite(name + "-site")
		siteID = site["id"].(string)
	}
	building := test_utils.GetEntityMap("building", name, siteID, TestDBName)

	return createObject(
		utils.BLDG,
		building,
		require,
	)
}

func RequireCreateBuilding(siteID, name string) map[string]any {
	obj, _ := internalCreateBuilding(siteID, name, true)
	return obj
}

func CreateBuilding(siteID, name string) (map[string]any, *utils.Error) {
	return internalCreateBuilding(siteID, name, false)
}

func internalCreateRoom(buildingID, name string, require bool) (map[string]any, *utils.Error) {
	if buildingID == "" {
		building := RequireCreateBuilding("", name+"-building")
		buildingID = building["id"].(string)
	}
	room := test_utils.GetEntityMap("room", name, buildingID, TestDBName)

	return createObject(
		utils.ROOM,
		room,
		require,
	)
}

func RequireCreateRoom(buildingID, name string) map[string]any {
	obj, _ := internalCreateRoom(buildingID, name, true)
	return obj
}

func CreateRoom(buildingID, name string) (map[string]any, *utils.Error) {
	return internalCreateRoom(buildingID, name, false)
}

func internalCreateRack(roomID, name string, require bool) (map[string]any, *utils.Error) {
	if roomID == "" {
		room := RequireCreateRoom("", name+"-room")
		roomID = room["id"].(string)
	}
	rack := test_utils.GetEntityMap("rack", name, roomID, TestDBName)

	return createObject(
		utils.RACK,
		rack,
		require,
	)
}

func RequireCreateRack(roomID, name string) map[string]any {
	obj, _ := internalCreateRack(roomID, name, true)
	return obj
}

func CreateRack(roomID, name string) (map[string]any, *utils.Error) {
	return internalCreateRack(roomID, name, false)
}

func internalCreateCorridor(roomID, name string, require bool) (map[string]any, *utils.Error) {
	if roomID == "" {
		room := RequireCreateRoom("", name+"-room")
		roomID = room["id"].(string)
	}
	corridor := test_utils.GetEntityMap("corridor", name, roomID, TestDBName)

	return createObject(
		utils.CORRIDOR,
		corridor,
		require,
	)
}

func RequireCreateCorridor(roomID, name string) map[string]any {
	obj, _ := internalCreateCorridor(roomID, name, true)
	return obj
}

func CreateCorridor(roomID, name string) (map[string]any, *utils.Error) {
	return internalCreateCorridor(roomID, name, false)
}

func internalCreateGeneric(roomID, name string, require bool) (map[string]any, *utils.Error) {
	if roomID == "" {
		room := RequireCreateRoom("", name+"-room")
		roomID = room["id"].(string)
	}
	generic := test_utils.GetEntityMap("generic", name, roomID, TestDBName)

	return createObject(
		utils.GENERIC,
		generic,
		require,
	)
}

func RequireCreateGeneric(roomID, name string) map[string]any {
	obj, _ := internalCreateGeneric(roomID, name, true)
	return obj
}

func CreateGeneric(roomID, name string) (map[string]any, *utils.Error) {
	return internalCreateGeneric(roomID, name, false)
}

func internalCreateDevice(parentID, name string, require bool) (map[string]any, *utils.Error) {
	device := test_utils.GetEntityMap("device", name, parentID, TestDBName)
	return createObject(
		utils.DEVICE,
		device,
		require,
	)
}

func RequireCreateDevice(parentID, name string) map[string]any {
	obj, _ := internalCreateDevice(parentID, name, true)
	return obj
}

func CreateDevice(parentID, name string) (map[string]any, *utils.Error) {
	return internalCreateDevice(parentID, name, false)
}

func internalCreateGroup(parentID, name string, content []any, require bool) (map[string]any, *utils.Error) {
	return createObject(
		utils.GROUP,
		map[string]any{
			"attributes": map[string]any{
				"content": content,
			},
			"category":    "group",
			"description": name,
			"domain":      TestDBName,
			"name":        name,
			"parentId":    parentID,
		},
		require,
	)
}

func RequireCreateGroup(parentID, name string, content []any) map[string]any {
	obj, _ := internalCreateGroup(parentID, name, content, true)
	return obj
}

func CreateGroup(parentID, name string, content []any) (map[string]any, *utils.Error) {
	return internalCreateGroup(parentID, name, content, false)
}

func CreateTestDomain(t *testing.T, name string, parentId string, color string) {
	// It creates a temporary domain that will be deleted at the end of the test t
	domainColor := "ffffff"
	if color != "" {
		domainColor = color
	}
	domain := map[string]any{
		"name":        name,
		"parentId":    parentId,
		"category":    "domain",
		"description": "temporary domain",
		"attributes": map[string]any{
			"color": domainColor,
		},
	}
	entity, err := models.CreateEntity(utils.DOMAIN, domain, ManagerUserRoles)
	assert.Nil(t, err)

	t.Cleanup(func() {
		// we get the domain again as it may have been deleted in a test
		filters := utils.RequestFilters{}
		domain, _ := models.GetObject(bson.M{"id": entity["id"]}, utils.EntityToString(utils.DOMAIN), filters, ManagerUserRoles)
		if domain != nil {
			err := models.DeleteObject(utils.EntityToString(utils.DOMAIN), entity["id"].(string), ManagerUserRoles)
			assert.Nil(t, err)
		}
	})
}

func CreateTestPhysicalEntity(t *testing.T, entityType int, name string, parentId string, recursive bool) map[string]any {
	// It creates a temporary entity and its temporary parents if recursive is true
	if recursive {
		ids := strings.Split(parentId, ".")
		if entityType == utils.CORRIDOR || entityType == utils.GENERIC {
			CreateTestPhysicalEntity(t, utils.ROOM, ids[len(ids)-1], strings.Join(ids[:len(ids)-1], "."), recursive)
		} else if entityType != utils.SITE {
			CreateTestPhysicalEntity(t, entityType-1, ids[len(ids)-1], strings.Join(ids[:len(ids)-1], "."), recursive)
		}
	}
	var entity map[string]any
	var err *utils.Error
	switch entityType {
	case utils.SITE:
		entity, err = CreateSite(name)
	case utils.BLDG:
		entity, err = CreateBuilding(parentId, name)
	case utils.ROOM:
		entity, err = CreateRoom(parentId, name)
	case utils.CORRIDOR:
		entity, err = CreateCorridor(parentId, name)
	case utils.GENERIC:
		entity, err = CreateGeneric(parentId, name)
	case utils.RACK:
		entity, err = CreateRack(parentId, name)
	case utils.DEVICE:
		entity, err = CreateDevice(parentId, name)
	default:
		t.Errorf("Invalid Entity type %d. Please verify", entityType)
	}
	assert.Nil(t, err)

	t.Cleanup(func() {
		// we get the room again as it may have been deleted in a test
		filters := utils.RequestFilters{}
		room, _ := models.GetObject(bson.M{"id": entity["id"]}, utils.EntityToString(entityType), filters, ManagerUserRoles)
		if room != nil {
			err := models.DeleteObject(utils.EntityToString(entityType), entity["id"].(string), ManagerUserRoles)
			assert.Nil(t, err)
		}
	})
	return entity
}

func CreateTestProject(t *testing.T, name string) (models.Project, string) {
	// Creates a temporary project that will be deleted at the end of the test
	adminUser := "admin@admin.com"
	project := models.Project{
		Name:        name,
		Attributes:  []string{"domain"},
		Namespace:   "physical",
		Permissions: []string{adminUser},
		ShowAvg:     false,
		ShowSum:     false,
	}
	err := models.AddProject(project)
	assert.Nil(t, err)

	// we get the project ID
	projects, _ := models.GetProjectsByUserEmail(adminUser)
	projectIndex := slices.IndexFunc(projects, func(p models.Project) bool {
		return p.Name == project.Name
	})
	projectId := projects[projectIndex].Id

	t.Cleanup(func() {
		// we get the room again as it may have been deleted in a test
		projects, _ := models.GetProjectsByUserEmail(adminUser)
		projectExists := slices.ContainsFunc(projects, func(p models.Project) bool {
			return p.Id == projectId
		})
		if projectExists {
			err := models.DeleteProject(projectId)
			assert.Nil(t, err)
		}
	})
	return project, projectId
}
