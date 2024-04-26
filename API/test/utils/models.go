package utils

import (
	"p3/models"
	"p3/test/integration"
	"p3/utils"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func CreateTestUser(t *testing.T, role models.Role) (string, string) {
	// It creates a temporary user that will be deleted at the end of the test t
	email := "temporary_user@test.com"
	password := "fake_password"
	account := &models.Account{
		Name:     "Temporary User",
		Email:    email,
		Password: password,
		Roles: map[string]models.Role{
			"*": role,
		},
	}
	acc, err := account.Create(map[string]models.Role{"*": "manager"})
	assert.Nil(t, err)

	t.Cleanup(func() {
		// we get the user again as the user may have been deleted in a test
		user := models.GetUserByEmail(acc.Email)
		if user != nil {
			err := models.DeleteUser(user.ID)
			assert.Nil(t, err)
		}
	})
	return email, password
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
	entity, err := models.CreateEntity(utils.DOMAIN, domain, integration.ManagerUserRoles)
	assert.Nil(t, err)

	t.Cleanup(func() {
		// we get the domain again as it may have been deleted in a test
		filters := utils.RequestFilters{}
		domain, _ := models.GetObject(bson.M{"id": entity["id"]}, utils.EntityToString(utils.DOMAIN), filters, integration.ManagerUserRoles)
		if domain != nil {
			err := models.DeleteObject(utils.EntityToString(utils.DOMAIN), entity["id"].(string), integration.ManagerUserRoles)
			assert.Nil(t, err)
		}
	})
}

func CreateTestPhysicalEntity(t *testing.T, entityType int, name string, parentId string, recursive bool) {
	// It creates a temporary entity and its temporary parents if recursive is true
	if recursive && entityType != utils.SITE {
		ids := strings.Split(parentId, ".")
		CreateTestPhysicalEntity(t, entityType-1, ids[len(ids)-1], strings.Join(ids[:len(ids)-1], "."), recursive)
	}
	var entity map[string]any
	var err *utils.Error
	switch entityType {
	case utils.SITE:
		entity, err = integration.CreateSite(name)
	case utils.BLDG:
		entity, err = integration.CreateBuilding(parentId, name)
	case utils.ROOM:
		entity, err = integration.CreateRoom(parentId, name)
	case utils.RACK:
		entity, err = integration.CreateRack(parentId, name)
	default:
		t.Errorf("Invalid Entity type %d. Please verify", entityType)
	}
	assert.Nil(t, err)

	t.Cleanup(func() {
		// we get the room again as it may have been deleted in a test
		filters := utils.RequestFilters{}
		room, _ := models.GetObject(bson.M{"id": entity["id"]}, utils.EntityToString(entityType), filters, integration.ManagerUserRoles)
		if room != nil {
			err := models.DeleteObject(utils.EntityToString(entityType), entity["id"].(string), integration.ManagerUserRoles)
			assert.Nil(t, err)
		}
	})
}

func CreateTestProject(t *testing.T, name string) (models.Project, string) {
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
