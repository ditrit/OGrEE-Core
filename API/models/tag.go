package models

import (
	"p3/repository"
	u "p3/utils"

	"github.com/elliotchance/pie/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Updates the list of entities of the tag "tagSlug" using the function updateFunc
func updateTagEntities(ctx mongo.SessionContext, tagSlug string, updateFunc func(entities []any) []any) *u.Error {
	tag, err := repository.GetEntity(
		ctx,
		bson.M{"slug": tagSlug},
		u.EntityToString(u.TAG),
		u.RequestFilters{},
	)
	if err != nil {
		if err.Type == u.ErrNotFound {
			return &u.Error{Type: u.ErrNotFound, Message: "Tag to add not found"}
		}

		return err
	}

	entities, ok := tag["entities"].(bson.A)
	if !ok {
		entities = []any{}
	}

	_, err = UpdateEntity(
		u.EntityToString(u.TAG),
		tagSlug,
		map[string]any{
			"entities": updateFunc(entities),
		},
		true,
		nil,
	)

	return err
}

// Adds the entity "entityID" to list of entities of the tag
func addEntityToTag(ctx mongo.SessionContext, tagSlug, entityID string) *u.Error {
	return updateTagEntities(ctx, tagSlug, func(entities []any) []any {
		return append(entities, entityID)
	})
}

// Removes the entity "entityID" to list of entities of the tag
func removeEntityFromTag(ctx mongo.SessionContext, tagSlug, entityID string) *u.Error {
	return updateTagEntities(ctx, tagSlug, func(entities []any) []any {
		return u.SliceRemove[any](entities, entityID)
	})
}

// Edits entityMap's "tags" list by:
//  1. adding tags in "tags+"
//  2. removing tags in "tags-"
func addAndRemoveFromTags(ctx mongo.SessionContext, entityType int, entityID string, entityMap map[string]interface{}) *u.Error {
	if u.EntityHasTags(entityType) {
		tags, tagsPresent := entityMap["tags"].([]any)
		if !tagsPresent || tags == nil {
			tags = []any{}
		}

		plusTag, plusTagPresent := entityMap["tags+"]
		if plusTagPresent {
			if !pie.Contains(tags, plusTag) {
				err := addEntityToTag(ctx, plusTag.(string), entityID)
				if err != nil {
					return err
				}

				tags = append(tags, plusTag.(string))
				entityMap["tags"] = tags
			}

			delete(entityMap, "tags+")
		}

		minusTag, minusTagPresent := entityMap["tags-"]
		if minusTagPresent {
			if pie.Contains(tags, minusTag) {
				tagSlug := minusTag.(string)

				err := removeEntityFromTag(ctx, tagSlug, entityID)
				if err != nil {
					return err
				}

				tags = u.SliceRemove[any](tags, tagSlug)
				entityMap["tags"] = tags
			}

			delete(entityMap, "tags-")
		}
	}

	return nil
}

// Deletes tag with slug "slug"
func DeleteTag(slug string) *u.Error {
	_, err := WithTransaction(func(ctx mongo.SessionContext) (interface{}, error) {
		err := repository.DeleteTag(ctx, bson.M{"slug": slug})
		if err != nil {
			// Unable to delete given id
			return nil, err
		}

		// Delete tag from all tags lists
		for _, entityType := range u.EntitiesWithTags {
			err := repository.DeleteTagFromEntity(ctx, slug, entityType)
			if err != nil {
				return nil, err
			}
		}

		return nil, nil
	})

	return err
}
