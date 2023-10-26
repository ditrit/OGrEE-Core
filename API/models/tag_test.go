package models_test

import (
	"context"
	"encoding/json"
	"net/http"
	"p3/models"
	"p3/repository"
	"p3/test/e2e"
	"p3/test/integration"
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
	assert.Equal(t, "Error deleting object: not found", err.Message)
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

func TestTagWithImageReturnsImagePathOnGet(t *testing.T) {
	err := createTagWithImage("create-tag-1", "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABgAAAAYCAYAAADgdz34AAAABHNCSVQICAgIfAhkiAAAAAlwSFlzAAAApgAAAKYB3X3/OAAAABl0RVh0U29mdHdhcmUAd3d3Lmlua3NjYXBlLm9yZ5vuPBoAAANCSURBVEiJtZZPbBtFFMZ/M7ubXdtdb1xSFyeilBapySVU8h8OoFaooFSqiihIVIpQBKci6KEg9Q6H9kovIHoCIVQJJCKE1ENFjnAgcaSGC6rEnxBwA04Tx43t2FnvDAfjkNibxgHxnWb2e/u992bee7tCa00YFsffekFY+nUzFtjW0LrvjRXrCDIAaPLlW0nHL0SsZtVoaF98mLrx3pdhOqLtYPHChahZcYYO7KvPFxvRl5XPp1sN3adWiD1ZAqD6XYK1b/dvE5IWryTt2udLFedwc1+9kLp+vbbpoDh+6TklxBeAi9TL0taeWpdmZzQDry0AcO+jQ12RyohqqoYoo8RDwJrU+qXkjWtfi8Xxt58BdQuwQs9qC/afLwCw8tnQbqYAPsgxE1S6F3EAIXux2oQFKm0ihMsOF71dHYx+f3NND68ghCu1YIoePPQN1pGRABkJ6Bus96CutRZMydTl+TvuiRW1m3n0eDl0vRPcEysqdXn+jsQPsrHMquGeXEaY4Yk4wxWcY5V/9scqOMOVUFthatyTy8QyqwZ+kDURKoMWxNKr2EeqVKcTNOajqKoBgOE28U4tdQl5p5bwCw7BWquaZSzAPlwjlithJtp3pTImSqQRrb2Z8PHGigD4RZuNX6JYj6wj7O4TFLbCO/Mn/m8R+h6rYSUb3ekokRY6f/YukArN979jcW+V/S8g0eT/N3VN3kTqWbQ428m9/8k0P/1aIhF36PccEl6EhOcAUCrXKZXXWS3XKd2vc/TRBG9O5ELC17MmWubD2nKhUKZa26Ba2+D3P+4/MNCFwg59oWVeYhkzgN/JDR8deKBoD7Y+ljEjGZ0sosXVTvbc6RHirr2reNy1OXd6pJsQ+gqjk8VWFYmHrwBzW/n+uMPFiRwHB2I7ih8ciHFxIkd/3Omk5tCDV1t+2nNu5sxxpDFNx+huNhVT3/zMDz8usXC3ddaHBj1GHj/As08fwTS7Kt1HBTmyN29vdwAw+/wbwLVOJ3uAD1wi/dUH7Qei66PfyuRj4Ik9is+hglfbkbfR3cnZm7chlUWLdwmprtCohX4HUtlOcQjLYCu+fzGJH2QRKvP3UNz8bWk1qMxjGTOMThZ3kvgLI5AzFfo379UAAAAASUVORK5CYII=")
	assert.Nil(t, err)

	tag, err := getTag("create-tag-1")
	assert.Nil(t, err)

	tagImage, imagePresent := tag["image"]
	assert.True(t, imagePresent)
	assert.NotEmpty(t, tagImage)

	response := e2e.MakeRequest(http.MethodGet, "/api/objects?namespace=logical.tag&slug=create-tag-1", nil)
	assert.Equal(t, http.StatusOK, response.Code)

	var responseBody map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &responseBody)
	objects := responseBody["data"].([]any)
	assert.Len(t, objects, 1)
	imagePath, imagePresent := objects[0].(map[string]any)["image"].(string)
	assert.True(t, imagePresent)
	assert.Equal(t, "/api/images/"+tagImage.(primitive.ObjectID).Hex(), imagePath)

	response = e2e.MakeRequest(http.MethodOptions, imagePath, nil)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "image/png", response.Header().Get("Content-Type"))
}

func TestUpdateTagWithImageWorks(t *testing.T) {
	err := createTagWithImage("update-tag-4", "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABgAAAAYCAYAAADgdz34AAAABHNCSVQICAgIfAhkiAAAAAlwSFlzAAAApgAAAKYB3X3/OAAAABl0RVh0U29mdHdhcmUAd3d3Lmlua3NjYXBlLm9yZ5vuPBoAAANCSURBVEiJtZZPbBtFFMZ/M7ubXdtdb1xSFyeilBapySVU8h8OoFaooFSqiihIVIpQBKci6KEg9Q6H9kovIHoCIVQJJCKE1ENFjnAgcaSGC6rEnxBwA04Tx43t2FnvDAfjkNibxgHxnWb2e/u992bee7tCa00YFsffekFY+nUzFtjW0LrvjRXrCDIAaPLlW0nHL0SsZtVoaF98mLrx3pdhOqLtYPHChahZcYYO7KvPFxvRl5XPp1sN3adWiD1ZAqD6XYK1b/dvE5IWryTt2udLFedwc1+9kLp+vbbpoDh+6TklxBeAi9TL0taeWpdmZzQDry0AcO+jQ12RyohqqoYoo8RDwJrU+qXkjWtfi8Xxt58BdQuwQs9qC/afLwCw8tnQbqYAPsgxE1S6F3EAIXux2oQFKm0ihMsOF71dHYx+f3NND68ghCu1YIoePPQN1pGRABkJ6Bus96CutRZMydTl+TvuiRW1m3n0eDl0vRPcEysqdXn+jsQPsrHMquGeXEaY4Yk4wxWcY5V/9scqOMOVUFthatyTy8QyqwZ+kDURKoMWxNKr2EeqVKcTNOajqKoBgOE28U4tdQl5p5bwCw7BWquaZSzAPlwjlithJtp3pTImSqQRrb2Z8PHGigD4RZuNX6JYj6wj7O4TFLbCO/Mn/m8R+h6rYSUb3ekokRY6f/YukArN979jcW+V/S8g0eT/N3VN3kTqWbQ428m9/8k0P/1aIhF36PccEl6EhOcAUCrXKZXXWS3XKd2vc/TRBG9O5ELC17MmWubD2nKhUKZa26Ba2+D3P+4/MNCFwg59oWVeYhkzgN/JDR8deKBoD7Y+ljEjGZ0sosXVTvbc6RHirr2reNy1OXd6pJsQ+gqjk8VWFYmHrwBzW/n+uMPFiRwHB2I7ih8ciHFxIkd/3Omk5tCDV1t+2nNu5sxxpDFNx+huNhVT3/zMDz8usXC3ddaHBj1GHj/As08fwTS7Kt1HBTmyN29vdwAw+/wbwLVOJ3uAD1wi/dUH7Qei66PfyuRj4Ik9is+hglfbkbfR3cnZm7chlUWLdwmprtCohX4HUtlOcQjLYCu+fzGJH2QRKvP3UNz8bWk1qMxjGTOMThZ3kvgLI5AzFfo379UAAAAASUVORK5CYII=")
	assert.Nil(t, err)

	updatedTag, err := models.UpdateObject(
		u.EntityToString(u.TAG),
		"update-tag-4",
		map[string]any{
			"name": "update tag 4",
		},
		true,
		nil,
	)
	assert.Nil(t, err)
	assert.NotNil(t, updatedTag["image"])
}

func TestUpdateSetEmptyImageRemovesOldImage(t *testing.T) {
	err := createTagWithImage("update-tag-5", "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABgAAAAYCAYAAADgdz34AAAABHNCSVQICAgIfAhkiAAAAAlwSFlzAAAApgAAAKYB3X3/OAAAABl0RVh0U29mdHdhcmUAd3d3Lmlua3NjYXBlLm9yZ5vuPBoAAANCSURBVEiJtZZPbBtFFMZ/M7ubXdtdb1xSFyeilBapySVU8h8OoFaooFSqiihIVIpQBKci6KEg9Q6H9kovIHoCIVQJJCKE1ENFjnAgcaSGC6rEnxBwA04Tx43t2FnvDAfjkNibxgHxnWb2e/u992bee7tCa00YFsffekFY+nUzFtjW0LrvjRXrCDIAaPLlW0nHL0SsZtVoaF98mLrx3pdhOqLtYPHChahZcYYO7KvPFxvRl5XPp1sN3adWiD1ZAqD6XYK1b/dvE5IWryTt2udLFedwc1+9kLp+vbbpoDh+6TklxBeAi9TL0taeWpdmZzQDry0AcO+jQ12RyohqqoYoo8RDwJrU+qXkjWtfi8Xxt58BdQuwQs9qC/afLwCw8tnQbqYAPsgxE1S6F3EAIXux2oQFKm0ihMsOF71dHYx+f3NND68ghCu1YIoePPQN1pGRABkJ6Bus96CutRZMydTl+TvuiRW1m3n0eDl0vRPcEysqdXn+jsQPsrHMquGeXEaY4Yk4wxWcY5V/9scqOMOVUFthatyTy8QyqwZ+kDURKoMWxNKr2EeqVKcTNOajqKoBgOE28U4tdQl5p5bwCw7BWquaZSzAPlwjlithJtp3pTImSqQRrb2Z8PHGigD4RZuNX6JYj6wj7O4TFLbCO/Mn/m8R+h6rYSUb3ekokRY6f/YukArN979jcW+V/S8g0eT/N3VN3kTqWbQ428m9/8k0P/1aIhF36PccEl6EhOcAUCrXKZXXWS3XKd2vc/TRBG9O5ELC17MmWubD2nKhUKZa26Ba2+D3P+4/MNCFwg59oWVeYhkzgN/JDR8deKBoD7Y+ljEjGZ0sosXVTvbc6RHirr2reNy1OXd6pJsQ+gqjk8VWFYmHrwBzW/n+uMPFiRwHB2I7ih8ciHFxIkd/3Omk5tCDV1t+2nNu5sxxpDFNx+huNhVT3/zMDz8usXC3ddaHBj1GHj/As08fwTS7Kt1HBTmyN29vdwAw+/wbwLVOJ3uAD1wi/dUH7Qei66PfyuRj4Ik9is+hglfbkbfR3cnZm7chlUWLdwmprtCohX4HUtlOcQjLYCu+fzGJH2QRKvP3UNz8bWk1qMxjGTOMThZ3kvgLI5AzFfo379UAAAAASUVORK5CYII=")
	assert.Nil(t, err)

	tag, err := getTag("update-tag-5")
	assert.Nil(t, err)

	tagImage, imagePresent := tag["image"]
	assert.True(t, imagePresent)
	assert.NotEmpty(t, tagImage)

	updatedTag, err := models.UpdateObject(
		u.EntityToString(u.TAG),
		"update-tag-5",
		map[string]any{
			"image": "",
		},
		true,
		nil,
	)
	assert.Nil(t, err)
	assert.Empty(t, updatedTag["image"])

	_, err = repository.GetImage(context.Background(), tagImage.(primitive.ObjectID).Hex())
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "Nothing matches this request")
}

func TestUpdateWithNewImageRemovesOldImage(t *testing.T) {
	err := createTagWithImage("update-tag-6", "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABgAAAAYCAYAAADgdz34AAAABHNCSVQICAgIfAhkiAAAAAlwSFlzAAAApgAAAKYB3X3/OAAAABl0RVh0U29mdHdhcmUAd3d3Lmlua3NjYXBlLm9yZ5vuPBoAAANCSURBVEiJtZZPbBtFFMZ/M7ubXdtdb1xSFyeilBapySVU8h8OoFaooFSqiihIVIpQBKci6KEg9Q6H9kovIHoCIVQJJCKE1ENFjnAgcaSGC6rEnxBwA04Tx43t2FnvDAfjkNibxgHxnWb2e/u992bee7tCa00YFsffekFY+nUzFtjW0LrvjRXrCDIAaPLlW0nHL0SsZtVoaF98mLrx3pdhOqLtYPHChahZcYYO7KvPFxvRl5XPp1sN3adWiD1ZAqD6XYK1b/dvE5IWryTt2udLFedwc1+9kLp+vbbpoDh+6TklxBeAi9TL0taeWpdmZzQDry0AcO+jQ12RyohqqoYoo8RDwJrU+qXkjWtfi8Xxt58BdQuwQs9qC/afLwCw8tnQbqYAPsgxE1S6F3EAIXux2oQFKm0ihMsOF71dHYx+f3NND68ghCu1YIoePPQN1pGRABkJ6Bus96CutRZMydTl+TvuiRW1m3n0eDl0vRPcEysqdXn+jsQPsrHMquGeXEaY4Yk4wxWcY5V/9scqOMOVUFthatyTy8QyqwZ+kDURKoMWxNKr2EeqVKcTNOajqKoBgOE28U4tdQl5p5bwCw7BWquaZSzAPlwjlithJtp3pTImSqQRrb2Z8PHGigD4RZuNX6JYj6wj7O4TFLbCO/Mn/m8R+h6rYSUb3ekokRY6f/YukArN979jcW+V/S8g0eT/N3VN3kTqWbQ428m9/8k0P/1aIhF36PccEl6EhOcAUCrXKZXXWS3XKd2vc/TRBG9O5ELC17MmWubD2nKhUKZa26Ba2+D3P+4/MNCFwg59oWVeYhkzgN/JDR8deKBoD7Y+ljEjGZ0sosXVTvbc6RHirr2reNy1OXd6pJsQ+gqjk8VWFYmHrwBzW/n+uMPFiRwHB2I7ih8ciHFxIkd/3Omk5tCDV1t+2nNu5sxxpDFNx+huNhVT3/zMDz8usXC3ddaHBj1GHj/As08fwTS7Kt1HBTmyN29vdwAw+/wbwLVOJ3uAD1wi/dUH7Qei66PfyuRj4Ik9is+hglfbkbfR3cnZm7chlUWLdwmprtCohX4HUtlOcQjLYCu+fzGJH2QRKvP3UNz8bWk1qMxjGTOMThZ3kvgLI5AzFfo379UAAAAASUVORK5CYII=")
	assert.Nil(t, err)

	tag, err := getTag("update-tag-6")
	assert.Nil(t, err)

	tagOldImage, imagePresent := tag["image"]
	assert.True(t, imagePresent)
	assert.NotEmpty(t, tagOldImage)

	updatedTag, err := models.UpdateObject(
		u.EntityToString(u.TAG),
		"update-tag-6",
		map[string]any{
			// the base64 is the same as the previous one, but we can't detect it
			"image": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABgAAAAYCAYAAADgdz34AAAABHNCSVQICAgIfAhkiAAAAAlwSFlzAAAApgAAAKYB3X3/OAAAABl0RVh0U29mdHdhcmUAd3d3Lmlua3NjYXBlLm9yZ5vuPBoAAANCSURBVEiJtZZPbBtFFMZ/M7ubXdtdb1xSFyeilBapySVU8h8OoFaooFSqiihIVIpQBKci6KEg9Q6H9kovIHoCIVQJJCKE1ENFjnAgcaSGC6rEnxBwA04Tx43t2FnvDAfjkNibxgHxnWb2e/u992bee7tCa00YFsffekFY+nUzFtjW0LrvjRXrCDIAaPLlW0nHL0SsZtVoaF98mLrx3pdhOqLtYPHChahZcYYO7KvPFxvRl5XPp1sN3adWiD1ZAqD6XYK1b/dvE5IWryTt2udLFedwc1+9kLp+vbbpoDh+6TklxBeAi9TL0taeWpdmZzQDry0AcO+jQ12RyohqqoYoo8RDwJrU+qXkjWtfi8Xxt58BdQuwQs9qC/afLwCw8tnQbqYAPsgxE1S6F3EAIXux2oQFKm0ihMsOF71dHYx+f3NND68ghCu1YIoePPQN1pGRABkJ6Bus96CutRZMydTl+TvuiRW1m3n0eDl0vRPcEysqdXn+jsQPsrHMquGeXEaY4Yk4wxWcY5V/9scqOMOVUFthatyTy8QyqwZ+kDURKoMWxNKr2EeqVKcTNOajqKoBgOE28U4tdQl5p5bwCw7BWquaZSzAPlwjlithJtp3pTImSqQRrb2Z8PHGigD4RZuNX6JYj6wj7O4TFLbCO/Mn/m8R+h6rYSUb3ekokRY6f/YukArN979jcW+V/S8g0eT/N3VN3kTqWbQ428m9/8k0P/1aIhF36PccEl6EhOcAUCrXKZXXWS3XKd2vc/TRBG9O5ELC17MmWubD2nKhUKZa26Ba2+D3P+4/MNCFwg59oWVeYhkzgN/JDR8deKBoD7Y+ljEjGZ0sosXVTvbc6RHirr2reNy1OXd6pJsQ+gqjk8VWFYmHrwBzW/n+uMPFiRwHB2I7ih8ciHFxIkd/3Omk5tCDV1t+2nNu5sxxpDFNx+huNhVT3/zMDz8usXC3ddaHBj1GHj/As08fwTS7Kt1HBTmyN29vdwAw+/wbwLVOJ3uAD1wi/dUH7Qei66PfyuRj4Ik9is+hglfbkbfR3cnZm7chlUWLdwmprtCohX4HUtlOcQjLYCu+fzGJH2QRKvP3UNz8bWk1qMxjGTOMThZ3kvgLI5AzFfo379UAAAAASUVORK5CYII=",
		},
		true,
		nil,
	)
	assert.Nil(t, err)
	assert.NotEmpty(t, updatedTag["image"])

	_, err = repository.GetImage(context.Background(), tagOldImage.(primitive.ObjectID).Hex())
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "Nothing matches this request")
}

func TestDeleteTagAlsoDeletesTagImage(t *testing.T) {
	err := createTagWithImage("delete-tag-4", "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABgAAAAYCAYAAADgdz34AAAABHNCSVQICAgIfAhkiAAAAAlwSFlzAAAApgAAAKYB3X3/OAAAABl0RVh0U29mdHdhcmUAd3d3Lmlua3NjYXBlLm9yZ5vuPBoAAANCSURBVEiJtZZPbBtFFMZ/M7ubXdtdb1xSFyeilBapySVU8h8OoFaooFSqiihIVIpQBKci6KEg9Q6H9kovIHoCIVQJJCKE1ENFjnAgcaSGC6rEnxBwA04Tx43t2FnvDAfjkNibxgHxnWb2e/u992bee7tCa00YFsffekFY+nUzFtjW0LrvjRXrCDIAaPLlW0nHL0SsZtVoaF98mLrx3pdhOqLtYPHChahZcYYO7KvPFxvRl5XPp1sN3adWiD1ZAqD6XYK1b/dvE5IWryTt2udLFedwc1+9kLp+vbbpoDh+6TklxBeAi9TL0taeWpdmZzQDry0AcO+jQ12RyohqqoYoo8RDwJrU+qXkjWtfi8Xxt58BdQuwQs9qC/afLwCw8tnQbqYAPsgxE1S6F3EAIXux2oQFKm0ihMsOF71dHYx+f3NND68ghCu1YIoePPQN1pGRABkJ6Bus96CutRZMydTl+TvuiRW1m3n0eDl0vRPcEysqdXn+jsQPsrHMquGeXEaY4Yk4wxWcY5V/9scqOMOVUFthatyTy8QyqwZ+kDURKoMWxNKr2EeqVKcTNOajqKoBgOE28U4tdQl5p5bwCw7BWquaZSzAPlwjlithJtp3pTImSqQRrb2Z8PHGigD4RZuNX6JYj6wj7O4TFLbCO/Mn/m8R+h6rYSUb3ekokRY6f/YukArN979jcW+V/S8g0eT/N3VN3kTqWbQ428m9/8k0P/1aIhF36PccEl6EhOcAUCrXKZXXWS3XKd2vc/TRBG9O5ELC17MmWubD2nKhUKZa26Ba2+D3P+4/MNCFwg59oWVeYhkzgN/JDR8deKBoD7Y+ljEjGZ0sosXVTvbc6RHirr2reNy1OXd6pJsQ+gqjk8VWFYmHrwBzW/n+uMPFiRwHB2I7ih8ciHFxIkd/3Omk5tCDV1t+2nNu5sxxpDFNx+huNhVT3/zMDz8usXC3ddaHBj1GHj/As08fwTS7Kt1HBTmyN29vdwAw+/wbwLVOJ3uAD1wi/dUH7Qei66PfyuRj4Ik9is+hglfbkbfR3cnZm7chlUWLdwmprtCohX4HUtlOcQjLYCu+fzGJH2QRKvP3UNz8bWk1qMxjGTOMThZ3kvgLI5AzFfo379UAAAAASUVORK5CYII=")
	require.Nil(t, err)

	tag, err := getTag("delete-tag-4")
	assert.Nil(t, err)

	tagOldImage, imagePresent := tag["image"]
	assert.True(t, imagePresent)
	assert.NotEmpty(t, tagOldImage)

	err = models.DeleteTag("delete-tag-4")
	assert.Nil(t, err)

	_, err = repository.GetImage(context.Background(), tagOldImage.(primitive.ObjectID).Hex())
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "Nothing matches this request")
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

func createTagWithImage(slug, image string) *u.Error {
	_, err := models.CreateEntity(
		u.TAG,
		map[string]any{
			"slug":  slug,
			"name":  slug,
			"color": "aaaaaa",
			"image": image,
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
			"domain":      integration.TestDBName,
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

func getTag(slug string) (map[string]interface{}, *u.Error) {
	return models.GetObject(
		bson.M{"slug": slug},
		u.EntityToString(u.TAG),
		u.RequestFilters{},
		userRoles,
	)
}
