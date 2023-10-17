package models

import (
	"context"
	"p3/repository"
	u "p3/utils"

	"github.com/elliotchance/pie/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Edits object's "tags" list by:
//  1. adding tags in "tags+"
//  2. removing tags in "tags-"
func addAndRemoveFromTags(ctx mongo.SessionContext, entity int, objectID string, object map[string]interface{}) *u.Error {
	if u.EntityHasTags(entity) {
		tags, tagsPresent := object["tags"].([]any)
		if !tagsPresent || tags == nil {
			tags = []any{}
		}

		plusTag, plusTagPresent := object["tags+"]
		if plusTagPresent {
			_, err := GetObject(
				bson.M{"slug": plusTag.(string)},
				u.EntityToString(u.TAG),
				u.RequestFilters{},
				nil,
			)
			if err != nil {
				if err.Type == u.ErrNotFound {
					return &u.Error{Type: u.ErrNotFound, Message: "Tag to add not found"}
				}

				return err
			}

			if !pie.Contains(tags, plusTag) {
				tags = append(tags, plusTag.(string))
				object["tags"] = tags
			}

			delete(object, "tags+")
		}

		minusTag, minusTagPresent := object["tags-"]
		if minusTagPresent {
			if pie.Contains(tags, minusTag) {
				tagSlug := minusTag.(string)
				tags = u.SliceRemove[any](tags, tagSlug)
				object["tags"] = tags
			}

			delete(object, "tags-")
		}
	}

	return nil
}

// Deletes tag with slug "slug"
func DeleteTag(slug string) *u.Error {
	_, err := WithTransaction(func(ctx mongo.SessionContext) (interface{}, error) {
		err := repository.DeleteObject(ctx, u.EntityToString(u.TAG), bson.M{"slug": slug})
		if err != nil {
			// Unable to delete given id
			return nil, err
		}

		// Delete tag from all tags lists
		for _, entity := range u.EntitiesWithTags {
			err := repository.DeleteTagFromEntity(ctx, slug, entity)
			if err != nil {
				return nil, err
			}
		}

		return nil, nil
	})

	return err
}

// Creates an image if data has one
// and updates data with the id of the new image
func createTagImage(ctx context.Context, data map[string]any) *u.Error {
	encodedImage, hasImage := data["image"].(string)
	if hasImage && encodedImage != "" {
		imageID, err := createImageFromBase64(ctx, encodedImage)
		if err != nil {
			return err
		}

		data["image"] = imageID
	}

	return nil
}

// Creates an image if updateData has one and different to oldEntity
// and updates updateData with the id of the new image
func updateTagImage(ctx context.Context, oldEntity, updateData map[string]any) *u.Error {
	newImage, hasImage := updateData["image"].(string)
	oldImage, hasOldImage := oldEntity["image"].(string)
	if hasImage && (!hasOldImage || newImage != oldImage) {
		return createTagImage(ctx, updateData)
	}

	return nil
}
