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

func TestAddTagThatNotExistToEntityReturnsError(t *testing.T) {
	err := createSite("add-tag-1-site", nil)
	require.Nil(t, err)

	_, err = addTagToEntity(u.SITE, "add-tag-1-site", "not-exists")
	assert.NotNil(t, err)
	assert.Equal(t, "Tag to add not found", err.Message)
}

func TestAddTagToEntityAddItsToListAndEntityToEntities(t *testing.T) {
	err := createTag("add-tag-2")
	require.Nil(t, err)

	err = createSite("add-tag-2-site", nil)
	require.Nil(t, err)

	site, err := addTagToEntity(u.SITE, "add-tag-2-site", "add-tag-2")
	assert.Nil(t, err)
	assert.Len(t, site["tags"], 1)
	assert.Contains(t, site["tags"].(primitive.A), "add-tag-2")

	tag, err := getTag("add-tag-2")
	assert.Nil(t, err)
	assert.Len(t, tag["entities"], 1)
	assert.Contains(t, tag["entities"].(primitive.A), "add-tag-2-site")
}

func TestAddDuplicatedTagDoesNothing(t *testing.T) {
	err := createTag("add-tag-3")
	require.Nil(t, err)

	err = createSite("add-tag-3-site", []string{"add-tag-3"})
	require.Nil(t, err)

	site, err := addTagToEntity(u.SITE, "add-tag-3-site", "add-tag-3")
	assert.Nil(t, err)
	assert.Len(t, site["tags"], 1)
	assert.Contains(t, site["tags"].(primitive.A), "add-tag-3")
}

func TestAddTagToMultipleEntities(t *testing.T) {
	err := createTag("add-tag-4")
	require.Nil(t, err)

	err = createSite("add-tag-4-1-site", nil)
	require.Nil(t, err)

	err = createSite("add-tag-4-2-site", nil)
	require.Nil(t, err)

	_, err = addTagToEntity(u.SITE, "add-tag-4-1-site", "add-tag-4")
	assert.Nil(t, err)

	_, err = addTagToEntity(u.SITE, "add-tag-4-2-site", "add-tag-4")
	assert.Nil(t, err)

	tag, err := getTag("add-tag-4")
	assert.Nil(t, err)
	assert.Len(t, tag["entities"], 2)
	assert.Contains(t, tag["entities"].(primitive.A), "add-tag-4-1-site")
	assert.Contains(t, tag["entities"].(primitive.A), "add-tag-4-2-site")
}

func TestRemoveTagThatIsNotInListDoesNothing(t *testing.T) {
	err := createSite("remove-tag-1-site", nil)
	require.Nil(t, err)

	site, err := models.UpdateEntity(
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

func TestRemoveTagFromEntityThatHasOneTag(t *testing.T) {
	err := createTag("remove-tag-2")
	require.Nil(t, err)

	err = createSite("remove-tag-2-site", []string{"remove-tag-2"})
	require.Nil(t, err)

	site, err := models.UpdateEntity(
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

	tag, err := getTag("remove-tag-2")
	assert.Nil(t, err)
	assert.Len(t, tag["entities"], 0)
}

func TestRemoveTagFromEntityThatHasMultipleTags(t *testing.T) {
	err := createTag("remove-tag-3-1")
	require.Nil(t, err)

	err = createTag("remove-tag-3-2")
	require.Nil(t, err)

	err = createSite("remove-tag-3-site", []string{"remove-tag-3-1", "remove-tag-3-2"})
	require.Nil(t, err)

	site, err := models.UpdateEntity(
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

	tag, err := getTag("remove-tag-3-1")
	assert.Nil(t, err)
	assert.Len(t, tag["entities"], 0)
}

func TestRemoveThatIsInMultipleEntities(t *testing.T) {
	err := createTag("remove-tag-4")
	require.Nil(t, err)

	err = createSite("remove-tag-4-1-site", []string{"remove-tag-4"})
	require.Nil(t, err)

	err = createSite("remove-tag-4-2-site", []string{"remove-tag-4"})
	require.Nil(t, err)

	_, err = models.UpdateEntity(
		u.EntityToString(u.SITE),
		"remove-tag-4-1-site",
		map[string]any{
			"tags-": "remove-tag-4",
		},
		true,
		userRoles,
	)
	assert.Nil(t, err)

	tag, err := getTag("remove-tag-4")
	assert.Nil(t, err)
	assert.Len(t, tag["entities"], 1)
	assert.Contains(t, tag["entities"].(primitive.A), "remove-tag-4-2-site")
}

func TestDeleteTagNoExistentReturnsError(t *testing.T) {
	err := models.DeleteTag("delete-tag")
	assert.NotNil(t, err)
	assert.Equal(t, "Error deleting object: not found", err.Message)
}

func TestDeleteTagNotPresentInAnyEntityWorks(t *testing.T) {
	err := createTag("delete-tag-1")
	require.Nil(t, err)

	err = models.DeleteTag("delete-tag-1")
	assert.Nil(t, err)

	_, err = models.GetEntity(bson.M{"slug": "delete-tag-1"}, u.EntityToString(u.TAG), u.RequestFilters{}, nil)
	assert.NotNil(t, err)
	assert.Equal(t, "Nothing matches this request", err.Message)
}

func TestDeleteTagPresentInEntityRemovesItFromList(t *testing.T) {
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

	_, err = models.GetEntity(bson.M{"slug": "delete-tag-2"}, u.EntityToString(u.TAG), u.RequestFilters{}, nil)
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
			"slug":  slug,
			"name":  slug,
			"color": "aaaaaa",
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
		_, err = addTagToEntity(u.SITE, name, tag)
		if err != nil {
			return err
		}
	}

	return nil
}

func addTagToEntity(entityType int, entityID string, tagSlug string) (map[string]any, *u.Error) {
	return models.UpdateEntity(
		u.EntityToString(entityType),
		entityID,
		map[string]any{
			"tags+": tagSlug,
		},
		true,
		userRoles,
	)
}

func getSite(name string) (map[string]interface{}, *u.Error) {
	return models.GetEntity(
		bson.M{"name": name},
		u.EntityToString(u.SITE),
		u.RequestFilters{},
		userRoles,
	)
}

func getTag(slug string) (map[string]interface{}, *u.Error) {
	return models.GetEntity(
		bson.M{"slug": slug},
		u.EntityToString(u.TAG),
		u.RequestFilters{},
		userRoles,
	)
}
