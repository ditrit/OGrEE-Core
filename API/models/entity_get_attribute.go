package models

import (
	"fmt"
	"p3/repository"
	u "p3/utils"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

// GetSiteParentAttribute: search for the object of given ID,
// then search for its site parent and return its requested attribute
func GetSiteParentAttribute(id string, attribute string) (map[string]any, *u.Error) {
	data := map[string]interface{}{}

	// Get all collections names
	ctx, cancel := u.Connect()
	db := repository.GetDB()
	collNames, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		fmt.Println(err.Error())
		return nil, &u.Error{Type: u.ErrDBError, Message: err.Error()}
	}
	// Find object
	for _, collName := range collNames {
		if err := db.Collection(collName).FindOne(ctx, bson.M{"id": id}).Decode(&data); err != nil {
			continue
		}
		// Found object with given id
		if data["category"].(string) == "site" {
			// it's a site
			break
		} else {
			// Find its parent site
			nameSlice := strings.Split(data["id"].(string), u.HN_DELIMETER)
			siteName := nameSlice[0] // CONSIDER SITE AS 0
			if err := db.Collection("site").FindOne(ctx, bson.M{"id": siteName}).Decode(&data); err != nil {
				return nil, &u.Error{Type: u.ErrNotFound,
					Message: "Could not find parent site for given object"}
			}
		}
	}

	defer cancel()

	if len(data) == 0 {
		return nil, &u.Error{Type: u.ErrNotFound, Message: "No object found with given id"}
	} else if attribute == "sitecolors" {
		resp := map[string]any{}
		for _, colorName := range []string{"reservedColor", "technicalColor", "usableColor"} {
			if color := data["attributes"].(map[string]interface{})[colorName]; color != nil {
				resp[colorName] = color
			} else {
				resp[colorName] = ""
			}
		}
		return resp, nil
	} else if attrValue := data["attributes"].(map[string]interface{})[attribute]; attrValue == nil {
		return nil, &u.Error{Type: u.ErrNotFound,
			Message: "Parent site has no temperatureUnit in attributes"}
	} else {
		return map[string]any{attribute: attrValue}, nil
	}
}
