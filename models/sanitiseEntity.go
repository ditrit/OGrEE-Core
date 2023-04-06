package models

import (
	u "p3/utils"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Mongo returns '_id' instead of id
func fixID(data map[string]interface{}) map[string]interface{} {
	if v, ok := data["_id"]; ok {
		data["id"] = v
		delete(data, "_id")
	}
	return data
}

// Removes underscore in object category if present
func FixUnderScore(x map[string]interface{}) {
	if catInf, ok := x["category"]; ok {
		if cat, _ := catInf.(string); strings.Contains(cat, "_") == true {
			x["category"] = strings.Replace(cat, "_", "-", 1)
		}
	}
}

// Perform any neccessary adjustments to objects before insertion into DB
func FixAttributesBeforeInsert(entity int, data map[string]interface{}) {
	if entity == u.RACK {
		pid, _ := primitive.ObjectIDFromHex(data["parentId"].(string))
		req := bson.M{"_id": pid}
		parent, _ := GetEntity(req, "room", u.RequestFilters{})
		parentUnit := parent["attributes"].(map[string]interface{})["posXYUnit"]
		data["attributes"].(map[string]interface{})["posXYUnit"] = parentUnit
	}
}
