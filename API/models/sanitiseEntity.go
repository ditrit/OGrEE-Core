package models

import (
	u "p3/utils"
	"strings"

	"github.com/elliotchance/pie/v2"
)

func shouldFillTags(entity int, filters u.RequestFilters) bool {
	return u.EntityHasTags(entity) && (len(filters.FieldsToShow) == 0 || pie.Contains(filters.FieldsToShow, "tags"))
}

// Adds empty list of tags if not present
func fillTags(object map[string]any) map[string]any {
	tags, tagsPresent := object["tags"]

	if !tagsPresent || tags == nil {
		object["tags"] = []any{}
	}

	return object
}

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
