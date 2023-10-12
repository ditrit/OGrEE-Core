package repository

import (
	"fmt"

	"p3/utils"
	u "p3/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// TODO docs
func DeleteTag(ctx mongo.SessionContext, filter primitive.M) *utils.Error {
	return DeleteEntity(ctx, u.EntityToString(u.TAG), filter)
}

func DeleteTagFromEntity(ctx mongo.SessionContext, slug string, entityType int) *utils.Error {
	_, err := GetDB().Collection(utils.EntityToString(entityType)).UpdateMany(
		ctx, bson.D{},
		bson.M{"$pull": bson.M{"tags": bson.M{"$eq": slug}}},
	)
	if err != nil {
		return &utils.Error{
			Type:    utils.ErrDBError,
			Message: fmt.Sprintf("Could not delete tag from %s", utils.EntityToString(entityType)),
		}
	}

	return nil
}

func UpdateTagSlugInEntities(ctx mongo.SessionContext, slug, newSlug string) *utils.Error {
	for _, entityType := range utils.EntitiesWithTags {
		err := updateTagSlugInEntity(ctx, slug, newSlug, entityType)
		if err != nil {
			return err
		}
	}

	return nil
}

func updateTagSlugInEntity(ctx mongo.SessionContext, slug, newSlug string, entityType int) *utils.Error {
	_, err := GetDB().Collection(utils.EntityToString(entityType)).UpdateMany(
		ctx,
		bson.M{"tags": bson.M{"$eq": slug}},
		bson.M{"$set": bson.M{"tags.$": newSlug}},
	)
	if err != nil {
		return &utils.Error{
			Type:    utils.ErrDBError,
			Message: fmt.Sprintf("Could not edit tag slug in %s", utils.EntityToString(entityType)),
		}
	}

	return nil
}
