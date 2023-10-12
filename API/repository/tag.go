package repository

import (
	"fmt"

	"p3/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

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
