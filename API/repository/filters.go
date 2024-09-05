package repository

import (
	u "p3/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetDateFilters(req bson.M, startDate string, endDate string) error {
	if len(startDate) > 0 || len(endDate) > 0 {
		lastUpdateReq := bson.M{}
		if len(startDate) > 0 {
			startDate, e := time.Parse("2006-01-02", startDate)
			if e != nil {
				return e
			}
			lastUpdateReq["$gte"] = primitive.NewDateTimeFromTime(startDate)
		}

		if len(endDate) > 0 {
			parsedEndDate, e := time.Parse("2006-01-02", endDate)
			parsedEndDate = parsedEndDate.Add(time.Hour * 24)
			if e != nil {
				return e
			}
			lastUpdateReq["$lte"] = primitive.NewDateTimeFromTime(parsedEndDate)
		}
		req["lastUpdated"] = lastUpdateReq
	}
	return nil
}

func GroupContentToOrFilter(content []any, parentId string) primitive.M {
	orReq := bson.A{}
	for _, objectName := range content {
		orReq = append(orReq, bson.M{"id": parentId + u.HN_DELIMETER + objectName.(string)})
	}
	filter := bson.M{"$or": orReq}
	return filter
}

func GetFieldsToShowFilter(fieldsToShow []string) *options.FindOptions {
	var opts *options.FindOptions
	if len(fieldsToShow) > 0 {
		opts = options.Find().SetProjection(GetFieldsToShowCompoundIndex(fieldsToShow))
	}
	return opts
}

func GetFieldsToShowOneFilter(fieldsToShow []string) *options.FindOneOptions {
	var opts *options.FindOneOptions
	if len(fieldsToShow) > 0 {
		opts = options.FindOne().SetProjection(GetFieldsToShowCompoundIndex(fieldsToShow))
	}
	return opts
}

func GetFieldsToShowCompoundIndex(fieldsToShow []string) primitive.D {
	compoundIndex := bson.D{bson.E{Key: "domain", Value: 1}, bson.E{Key: "id", Value: 1}}
	for _, field := range fieldsToShow {
		if field != "domain" && field != "id" {
			compoundIndex = append(compoundIndex, bson.E{Key: field, Value: 1})
		}
	}
	return compoundIndex
}
