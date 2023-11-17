package models

import (
	"context"
	"fmt"
	"p3/repository"
	u "p3/utils"

	"github.com/elliotchance/pie/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Obtains the list of tags present in a object
func getTags(object map[string]any) ([]any, bool) {
	tags, tagsPresent := object["tags"].([]any)

	if !tagsPresent {
		// when retrieving from database, it has the primitive.A type
		tags, tagsPresent = object["tags"].(primitive.A)
	}

	return tags, tagsPresent
}

// Verifies that a list of tags has not duplicated values and that all the elements exist
func verifyTagList(ctx mongo.SessionContext, tags []any) *u.Error {
	if !pie.AreUnique(tags) {
		return &u.Error{
			Type:    u.ErrBadFormat,
			Message: "Tags has duplicated values",
		}
	}

	for _, tagSlugAny := range tags {
		tagSlug := tagSlugAny.(string)

		_, err := repository.GetTagBySlug(ctx, tagSlug)
		if err != nil {
			if err.Type == u.ErrNotFound {
				return &u.Error{
					Type:    u.ErrBadFormat,
					Message: fmt.Sprintf("Tag %q not found", tagSlug),
				}
			}

			return err
		}
	}

	return nil
}

// Edits object's "tags" list by:
//  1. adding tags in "tags+"
//  2. removing tags in "tags-"
func addAndRemoveFromTags(ctx mongo.SessionContext, entity int, objectID string, object map[string]interface{}) *u.Error {
	if u.EntityHasTags(entity) {
		tags, tagsPresent := getTags(object)
		if !tagsPresent || tags == nil {
			tags = []any{}
		}

		plusTag, plusTagPresent := object["tags+"]
		if plusTagPresent {
			err := verifyTagList(ctx, []any{plusTag})
			if err != nil {
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
		tag, err := repository.GetTagBySlug(ctx, slug)
		if err != nil {
			return nil, err
		}

		err = repository.DeleteObject(ctx, u.EntityToString(u.TAG), bson.M{"slug": slug})
		if err != nil {
			// Unable to delete given id
			return nil, err
		}

		tagImageID, hasImage := tag["image"].(primitive.ObjectID)
		if hasImage {
			err = repository.DeleteImage(ctx, tagImageID)
			if err != nil {
				return nil, err
			}
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
		imageID, err := createImageFromDataURI(ctx, encodedImage)
		if err != nil {
			return err
		}

		data["image"] = imageID
	}

	return nil
}

// Creates an image if updateData has one and different to oldEntity
// and updates updateData with the id of the new image
func updateTagImage(ctx context.Context, oldObject, updateData map[string]any) *u.Error {
	newImage, hasNewImage := updateData["image"].(string)
	oldImage, hasOldImage := oldObject["image"].(primitive.ObjectID)

	if !hasNewImage {
		return nil
	}

	if hasOldImage && newImage != oldImage.Hex() {
		err := repository.DeleteImage(ctx, oldImage)
		if err != nil {
			return err
		}

		return createTagImage(ctx, updateData)
	}

	if !hasOldImage {
		return createTagImage(ctx, updateData)
	}

	return nil
}
