package repository

import (
	"fmt"

	"p3/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetTagBySlug(slug string) (map[string]any, *utils.Error) {
	return GetObject(bson.M{"slug": slug}, utils.EntityToString(utils.TAG), utils.RequestFilters{})
}

func DeleteTagFromEntity(ctx mongo.SessionContext, slug string, entity int) *utils.Error {
	_, err := GetDB().Collection(utils.EntityToString(entity)).UpdateMany(
		ctx, bson.D{},
		bson.M{"$pull": bson.M{"tags": bson.M{"$eq": slug}}},
	)
	if err != nil {
		return &utils.Error{
			Type:    utils.ErrDBError,
			Message: fmt.Sprintf("Could not delete tag from %s", utils.EntityToString(entity)),
		}
	}

	return nil
}

func UpdateTagSlugInEntities(ctx mongo.SessionContext, slug, newSlug string) *utils.Error {
	for _, entity := range utils.EntitiesWithTags {
		err := updateTagSlugInEntity(ctx, slug, newSlug, entity)
		if err != nil {
			return err
		}
	}

	return nil
}

func updateTagSlugInEntity(ctx mongo.SessionContext, slug, newSlug string, entity int) *utils.Error {
	_, err := GetDB().Collection(utils.EntityToString(entity)).UpdateMany(
		ctx,
		bson.M{"tags": bson.M{"$eq": slug}},
		bson.M{"$set": bson.M{"tags.$": newSlug}},
	)
	if err != nil {
		return &utils.Error{
			Type:    utils.ErrDBError,
			Message: fmt.Sprintf("Could not edit tag slug in %s", utils.EntityToString(entity)),
		}
	}

	return nil
}
