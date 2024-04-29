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
	case utils.DEVICE:
		entity, err = integration.CreateDevice(parentId, name)
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

func GetEntityMap(entity string, name string, parentId string, domain string) map[string]any {
	if domain == "" {
		domain = integration.TestDBName
	}
	// returns an entity to use in tests
	switch entity {
	case "site":
		return map[string]any{
			"attributes": map[string]any{
				"reservedColor":  "AAAAAA",
				"technicalColor": "D0FF78",
				"usableColor":    "5BDCFF",
			},
			"category":    "site",
			"description": "site",
			"domain":      domain,
			"name":        name,
			"tags":        []any{},
		}
	case "room":
		return map[string]any{
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
			"category":    "room",
			"description": "room",
			"domain":      domain,
			"name":        name,
			"parentId":    parentId,
		}
	case "rack":
		return map[string]any{
			"attributes": map[string]any{
				"height":     47,
				"heightUnit": "U",
				"rotation":   []any{45, 45, 45},
				"posXYZ":     []any{4.6666666666667, -2, 0},
				"posXYUnit":  "m",
				"size":       []any{80, 100.532442},
				"sizeUnit":   "cm",
				"template":   "",
			},
			"category":    "rack",
			"description": "rack",
			"domain":      domain,
			"name":        name,
			"parentId":    parentId,
		}
	case "device":
		return map[string]any{
			"parentId":    parentId,
			"name":        name,
			"category":    "device",
			"description": "device",
			"domain":      domain,
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
	default:
		return nil
	}
}
