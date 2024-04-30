package utils

import (
	"p3/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

var ManagerUserRoles = map[string]models.Role{
	models.ROOT_DOMAIN: models.Manager,
}

const TestDBName = "ogreeAutoTest"

func GetUserToken(email string, password string) string {
	// It executes the user login and returns tha auth token
	acc, e := models.Login(email, password)
	if e != nil {
		return ""
	}
	return acc.Token
}

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

func GetEntityMap(entity string, name string, parentId string, domain string) map[string]any {
	if domain == "" {
		domain = TestDBName
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
			"description": name,
			"domain":      domain,
			"name":        name,
			"tags":        []any{},
		}
	case "building":
		return map[string]any{
			"attributes": map[string]any{
				"height":     5,
				"heightUnit": "m",
				"posXY":      []any{50, 0},
				"posXYUnit":  "m",
				"size":       []any{49, 46.5},
				"sizeUnit":   "m",
				"rotation":   30.5,
			},
			"category":    "building",
			"description": name,
			"domain":      domain,
			"name":        name,
			"parentId":    parentId,
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
			"description": name,
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
			"description": name,
			"domain":      domain,
			"name":        name,
			"parentId":    parentId,
		}
	case "corridor":
		return map[string]any{
			"attributes": map[string]any{
				"color":       "000099",
				"content":     "B11,C19",
				"temperature": "cold",
				"height":      470,
				"heightUnit":  "cm",
				"rotation":    []any{45, 45, 45},
				"posXYUnit":   "m",
				"posXYZ":      []any{4.6666666666667, -2, 0},
				"size":        []any{80, 100.532442},
				"sizeUnit":    "cm",
			},
			"category":    "corridor",
			"description": "corridor",
			"domain":      domain,
			"name":        name,
			"parentId":    parentId,
		}
	case "generic":
		return map[string]any{
			"attributes": map[string]any{
				"height":     47,
				"heightUnit": "cm",
				"rotation":   []any{45, 45, 45},
				"posXYZ":     []any{4.6666666666667, -2, 0},
				"posXYUnit":  "m",
				"size":       []any{80, 100.532442},
				"shape":      "cube",
				"sizeUnit":   "cm",
				"template":   "",
				"type":       "box",
			},
			"category":    "generic",
			"description": name,
			"domain":      domain,
			"name":        name,
			"parentId":    parentId,
		}
	case "device":
		return map[string]any{
			"parentId":    parentId,
			"name":        name,
			"category":    "device",
			"description": name,
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
				"weightKg":    1.81,
			},
		}
	default:
		return nil
	}
}
