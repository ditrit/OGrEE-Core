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
