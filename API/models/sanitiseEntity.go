package models

import (
	u "p3/utils"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

// Remove mongos _id and add parentId
func fixID(data map[string]interface{}) map[string]interface{} {
	delete(data, "_id")
	if id, ok := data["id"].(string); ok {
		lastInd := strings.LastIndex(id, u.HN_DELIMETER)
		if lastInd > 0 {
			data["parentId"] = id[:lastInd]
		}
	}
	return data
}

// Removes underscore in object category if present
func FixUnderScore(x map[string]interface{}) {
	if catInf, ok := x["category"]; ok {
		if cat, _ := catInf.(string); strings.Contains(cat, "_") {
			x["category"] = strings.Replace(cat, "_", "-", 1)
		}
	}
}

// Perform any neccessary adjustments to objects before insertion into DB
func FixAttributesBeforeInsert(entity int, data map[string]interface{}) {
	if entity == u.RACK {
		pid, _ := data["parentId"].(string)
		req := bson.M{"id": pid}
		parent, _ := GetEntity(req, "room", u.RequestFilters{}, nil)
		parentUnit := parent["attributes"].(map[string]interface{})["posXYUnit"]
		data["attributes"].(map[string]interface{})["posXYUnit"] = parentUnit
	}
}

func FixReadOnlyName(data map[string]interface{}) map[string]interface{} {
	cleanData := map[string]interface{}{}
	cleanData["id"] = data["id"]
	cleanData["category"] = data["category"]
	cleanData["name"] = data["name"]
	if _, ok := data["children"]; ok {
		cleanData["children"] = data["children"]
	}
	return cleanData
}
