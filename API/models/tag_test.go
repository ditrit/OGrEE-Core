package models_test

import (
	"p3/models"
	u "p3/utils"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var userRoles = map[string]models.Role{
	models.ROOT_DOMAIN: models.Manager,
}

func TestAddTagThatNotExistReturnsError(t *testing.T) {
	err := createSite("add-tag-1-site", nil)
	require.Nil(t, err)

	_, err = addTagToObject("add-tag-1-site", "not-exists")
	assert.NotNil(t, err)
	assert.Equal(t, "Tag to add not found", err.Message)
}

func TestAddTagToObjectAddsItToList(t *testing.T) {
	err := createTag("add-tag-2")
	require.Nil(t, err)

	err = createSite("add-tag-2-site", nil)
	require.Nil(t, err)

	site, err := addTagToObject("add-tag-2-site", "add-tag-2")
	assert.Nil(t, err)
	assert.Len(t, site["tags"], 1)
	assert.Contains(t, site["tags"].(primitive.A), "add-tag-2")
}

func TestAddDuplicatedTagDoesNothing(t *testing.T) {
	err := createTag("add-tag-3")
	require.Nil(t, err)

	err = createSite("add-tag-3-site", []string{"add-tag-3"})
	require.Nil(t, err)

	site, err := addTagToObject("add-tag-3-site", "add-tag-3")
	assert.Nil(t, err)
	assert.Len(t, site["tags"], 1)
	assert.Contains(t, site["tags"].(primitive.A), "add-tag-3")
}

func TestRemoveTagThatIsNotInListDoesNothing(t *testing.T) {
	err := createSite("remove-tag-1-site", nil)
	require.Nil(t, err)

	site, err := models.UpdateObject(
		u.EntityToString(u.SITE),
		"remove-tag-1-site",
		map[string]any{
			"tags-": "not-present",
		},
		true,
		userRoles,
	)
	assert.Nil(t, err)
	assert.Len(t, site["tags"], 0)
}

func TestRemoveTagFromObjectThatHasOneTag(t *testing.T) {
	err := createTag("remove-tag-2")
	require.Nil(t, err)

	err = createSite("remove-tag-2-site", []string{"remove-tag-2"})
	require.Nil(t, err)

	site, err := models.UpdateObject(
		u.EntityToString(u.SITE),
		"remove-tag-2-site",
		map[string]any{
			"tags-": "remove-tag-2",
		},
		true,
		userRoles,
	)
	assert.Nil(t, err)
	assert.Len(t, site["tags"], 0)
}

func TestRemoveTagFromObjectThatHasMultipleTags(t *testing.T) {
	err := createTag("remove-tag-3-1")
	require.Nil(t, err)

	err = createTag("remove-tag-3-2")
	require.Nil(t, err)

	err = createSite("remove-tag-3-site", []string{"remove-tag-3-1", "remove-tag-3-2"})
	require.Nil(t, err)

	site, err := models.UpdateObject(
		u.EntityToString(u.SITE),
		"remove-tag-3-site",
		map[string]any{
			"tags-": "remove-tag-3-1",
		},
		true,
		userRoles,
	)
	assert.Nil(t, err)
	assert.Len(t, site["tags"], 1)
	assert.Contains(t, site["tags"].(primitive.A), "remove-tag-3-2")
}

func TestUpdateTagNoExistentReturnsError(t *testing.T) {
	_, err := models.UpdateObject(u.EntityToString(u.TAG), "update-tag", nil, false, nil)
	assert.NotNil(t, err)
	assert.Equal(t, "Nothing matches this request", err.Message)
}

func TestUpdateTagNotPresentInAnyObjectWorks(t *testing.T) {
	err := createTag("update-tag-1")
	require.Nil(t, err)

	updatedTag, err := models.UpdateObject(
		u.EntityToString(u.TAG),
		"update-tag-1",
		map[string]any{
			"slug": "update-tag-1-1",
		},
		true,
		nil,
	)
	assert.Nil(t, err)
	assert.Equal(t, "update-tag-1-1", updatedTag["slug"])
}

func TestUpdateTagPresentInOneObjectUpdatesItInList(t *testing.T) {
	err := createTag("update-tag-2")
	require.Nil(t, err)

	err = createTag("update-tag-3")
	require.Nil(t, err)

	err = createSite("update-tag-2-site", []string{"update-tag-2", "update-tag-3"})
	require.Nil(t, err)

	updatedTag, err := models.UpdateObject(
		u.EntityToString(u.TAG),
		"update-tag-2",
		map[string]any{
			"slug": "update-tag-2-2",
		},
		true,
		nil,
	)
	assert.Nil(t, err)
	assert.Equal(t, "update-tag-2-2", updatedTag["slug"])

	site, err := getSite("update-tag-2-site")
	assert.Nil(t, err)
	assert.Len(t, site["tags"], 2)
	assert.Contains(t, site["tags"].(primitive.A), "update-tag-3")
	assert.Contains(t, site["tags"].(primitive.A), "update-tag-2-2")
}

func TestDeleteTagNoExistentReturnsError(t *testing.T) {
	err := models.DeleteTag("delete-tag")
	assert.NotNil(t, err)
	assert.Equal(t, "Nothing matches this request", err.Message)
}

func TestDeleteTagNotPresentInAnyObjectWorks(t *testing.T) {
	err := createTag("delete-tag-1")
	require.Nil(t, err)

	err = models.DeleteTag("delete-tag-1")
	assert.Nil(t, err)

	_, err = models.GetObject(bson.M{"slug": "delete-tag-1"}, u.EntityToString(u.TAG), u.RequestFilters{}, nil)
	assert.NotNil(t, err)
	assert.Equal(t, "Nothing matches this request", err.Message)
}

func TestDeleteTagPresentInOneObjectRemovesItFromList(t *testing.T) {
	err := createTag("delete-tag-2")
	require.Nil(t, err)

	err = createTag("delete-tag-3")
	require.Nil(t, err)

	err = createSite("delete-tag-2-site", []string{"delete-tag-2", "delete-tag-3"})
	require.Nil(t, err)

	site, err := getSite("delete-tag-2-site")
	assert.Nil(t, err)
	assert.Len(t, site["tags"], 2)

	err = models.DeleteTag("delete-tag-2")
	assert.Nil(t, err)

	_, err = models.GetObject(bson.M{"slug": "delete-tag-2"}, u.EntityToString(u.TAG), u.RequestFilters{}, nil)
	assert.NotNil(t, err)
	assert.Equal(t, "Nothing matches this request", err.Message)

	site, err = getSite("delete-tag-2-site")
	assert.Nil(t, err)
	assert.Len(t, site["tags"], 1)
	assert.Equal(t, "delete-tag-3", site["tags"].(primitive.A)[0])
}

func createTag(slug string) *u.Error {
	_, err := models.CreateEntity(
		u.TAG,
		map[string]any{
			"slug":        slug,
			"description": slug,
			"color":       "aaaaaa",
		},
		nil,
	)

	return err
}

func createSite(name string, tags []string) *u.Error {
	_, err := models.CreateEntity(
		u.SITE,
		map[string]any{
			"attributes": map[string]any{
				"reservedColor":  "AAAAAA",
				"technicalColor": "D0FF78",
				"usableColor":    "5BDCFF",
			},
			"category":    "site",
			"description": []any{"site"},
			"domain":      "AutoTest",
			"name":        name,
		},
		userRoles,
	)
	if err != nil {
		return err
	}

	for _, tag := range tags {
		_, err = addTagToObject(name, tag)
		if err != nil {
			return err
		}
	}

	return nil
}

func addTagToObject(objectID string, tagSlug string) (map[string]any, *u.Error) {
	return models.UpdateObject(
		u.HIERARCHYOBJS_ENT,
		objectID,
		map[string]any{
			"tags+": tagSlug,
		},
		true,
		userRoles,
	)
}

func getSite(name string) (map[string]interface{}, *u.Error) {
	return models.GetObject(
		bson.M{"name": name},
		u.EntityToString(u.SITE),
		u.RequestFilters{},
		userRoles,
	)
}
