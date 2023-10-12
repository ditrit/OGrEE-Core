package repository

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	u "p3/utils"
)

func GetEntity(ctx mongo.SessionContext, req bson.M, ent string, filters u.RequestFilters) (map[string]interface{}, *u.Error) {
	t := map[string]interface{}{}

	var opts *options.FindOneOptions
	if len(filters.FieldsToShow) > 0 {
		compoundIndex := bson.D{bson.E{Key: "domain", Value: 1}, bson.E{Key: "id", Value: 1}}
		for _, field := range filters.FieldsToShow {
			if field != "domain" && field != "id" {
				compoundIndex = append(compoundIndex, bson.E{Key: field, Value: 1})
			}
		}
		opts = options.FindOne().SetProjection(compoundIndex)
	}

	err := GetDateFilters(req, filters.StartDate, filters.EndDate)
	if err != nil {
		return nil, &u.Error{Type: u.ErrBadFormat, Message: err.Error()}
	}

	if opts != nil {
		err = GetDB().Collection(ent).FindOne(ctx, req, opts).Decode(&t)
	} else {
		err = GetDB().Collection(ent).FindOne(ctx, req).Decode(&t)
	}
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, &u.Error{Type: u.ErrNotFound,
				Message: "Nothing matches this request"}
		}
		return nil, &u.Error{Type: u.ErrBadFormat, Message: err.Error()}
	}

	return t, nil
}

func DeleteEntity(ctx mongo.SessionContext, entityType string, filter bson.M) *u.Error {
	result, err := GetDB().Collection(entityType).DeleteOne(ctx, filter)
	if err != nil {
		return &u.Error{Type: u.ErrDBError, Message: err.Error()}
	}

	if result.DeletedCount == 0 {
		return &u.Error{Type: u.ErrNotFound, Message: "Error deleting object: not found"}
	}

	return nil
}
