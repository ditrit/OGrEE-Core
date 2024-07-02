package controllers_test

import (
	"cli/controllers"
	"cli/models"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Tests ObjectUrlGeneric
func TestObjectUrlGenericInvalidPath(t *testing.T) {
	_, err := controllers.C.ObjectUrlGeneric("/invalid/path", 0, nil, nil)
	assert.NotNil(t, err)
	assert.Equal(t, "invalid object path", err.Error())
}

func TestObjectUrlGenericWithNoFilters(t *testing.T) {
	paths := []map[string]any{
		map[string]any{
			"basePath":  models.StrayPath,
			"objectId":  "stray-object",
			"endpoint":  "/api/objects",
			"idName":    "id",
			"namespace": "physical.stray",
		},
		map[string]any{
			"basePath":  models.PhysicalPath,
			"objectId":  "BASIC/A",
			"endpoint":  "/api/objects",
			"idName":    "id",
			"namespace": "physical.hierarchy",
		},
		map[string]any{
			"basePath":  models.ObjectTemplatesPath,
			"objectId":  "my-template",
			"endpoint":  "/api/objects",
			"idName":    "slug",
			"namespace": "logical.objtemplate",
		},
		map[string]any{
			"basePath":  models.RoomTemplatesPath,
			"objectId":  "my-room-template",
			"endpoint":  "/api/objects",
			"idName":    "slug",
			"namespace": "logical.roomtemplate",
		},
		map[string]any{
			"basePath":  models.BuildingTemplatesPath,
			"objectId":  "my-building-template",
			"endpoint":  "/api/objects",
			"idName":    "slug",
			"namespace": "logical.bldgtemplate",
		},
		map[string]any{
			"basePath":  models.GroupsPath,
			"objectId":  "group1",
			"endpoint":  "/api/objects",
			"idName":    "id",
			"namespace": "logical",
			"extraParams": map[string]any{
				"category": "group",
			},
		},
		map[string]any{
			"basePath":  models.TagsPath,
			"objectId":  "my-tag",
			"endpoint":  "/api/objects",
			"idName":    "slug",
			"namespace": "logical.tag",
		},
		map[string]any{
			"basePath":  models.LayersPath,
			"objectId":  "my-layer",
			"endpoint":  "/api/objects",
			"idName":    "slug",
			"namespace": "logical.layer",
		},
		map[string]any{
			"basePath":  models.DomainsPath,
			"objectId":  "domain1",
			"endpoint":  "/api/objects",
			"idName":    "id",
			"namespace": "organisational",
		},
		map[string]any{
			"basePath":  models.DomainsPath,
			"objectId":  "domain1/subdomain",
			"endpoint":  "/api/objects",
			"idName":    "id",
			"namespace": "organisational",
		},
	}
	for _, value := range paths {
		resultUrl, err := controllers.C.ObjectUrlGeneric(value["basePath"].(string)+value["objectId"].(string), 0, nil, nil)
		assert.Nil(t, err)
		assert.NotNil(t, resultUrl)

		parsedUrl, _ := url.Parse(resultUrl)
		assert.Equal(t, value["endpoint"], parsedUrl.Path)
		assert.Equal(t, strings.Replace(value["objectId"].(string), "/", ".", -1), parsedUrl.Query().Get(value["idName"].(string)))
		assert.Equal(t, value["namespace"], parsedUrl.Query().Get("namespace"))

		if extraParams, ok := value["extraParams"]; ok {
			for k, v := range extraParams.(map[string]any) {
				assert.Equal(t, v, parsedUrl.Query().Get(k))
			}
		}
	}
}

func TestObjectUrlGenericWithNormalFilters(t *testing.T) {
	filters := map[string]string{
		"color": "00ED00",
	}
	id := "BASIC/A"
	resultUrl, err := controllers.C.ObjectUrlGeneric(models.PhysicalPath+id, 0, filters, nil)
	assert.Nil(t, err)
	assert.NotNil(t, resultUrl)

	parsedUrl, _ := url.Parse(resultUrl)
	assert.Equal(t, "/api/objects", parsedUrl.Path)
	assert.Equal(t, strings.Replace(id, "/", ".", -1), parsedUrl.Query().Get("id"))
	assert.Equal(t, "physical.hierarchy", parsedUrl.Query().Get("namespace"))
	assert.Equal(t, "00ED00", parsedUrl.Query().Get("color"))
}

func TestObjectUrlGenericWithFilterField(t *testing.T) {
	filters := map[string]string{
		"filter": "color=00ED00",
	}
	id := "BASIC/A"
	resultUrl, err := controllers.C.ObjectUrlGeneric(models.PhysicalPath+id, 0, filters, nil)
	assert.Nil(t, err)
	assert.NotNil(t, resultUrl)

	parsedUrl, _ := url.Parse(resultUrl)
	assert.Equal(t, "/api/objects/search", parsedUrl.Path)
	assert.Equal(t, strings.Replace(id, "/", ".", -1), parsedUrl.Query().Get("id"))
	assert.Equal(t, "physical.hierarchy", parsedUrl.Query().Get("namespace"))
}
