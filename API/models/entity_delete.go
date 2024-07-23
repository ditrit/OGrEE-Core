package models

import (
	"os"
	"p3/repository"
	u "p3/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func DeleteObject(entityStr string, id string, userRoles map[string]Role) *u.Error {
	entity := u.EntityStrToInt(entityStr)
	if entity == u.TAG {
		return DeleteTag(id)
	} else if u.IsEntityNonHierarchical(entity) {
		return DeleteNonHierarchicalObject(entityStr, id)
	} else {
		return DeleteHierarchicalObject(entityStr, id, userRoles)
	}
}

// DeleteHierarchicalObject: delete object of given hierarchyName
// search for all its children and delete them too, return:
// - success or fail message map
func DeleteHierarchicalObject(entity string, id string, userRoles map[string]Role) *u.Error {
	// Special check for delete domain
	if entity == "domain" {
		if id == os.Getenv("db") {
			return &u.Error{Type: u.ErrForbidden, Message: "Cannot delete tenant's default domain"}
		}
		if domainHasObjects(id) {
			return &u.Error{Type: u.ErrForbidden, Message: "Cannot delete domain if it has at least one object"}
		}
	}

	// Delete with given id
	req, ok := GetRequestFilterByDomain(userRoles)
	if !ok {
		return &u.Error{Type: u.ErrUnauthorized, Message: "User does not have permission to delete"}
	}

	req["id"] = id

	_, err := WithTransaction(func(ctx mongo.SessionContext) (any, error) {
		err := repository.DeleteObject(ctx, entity, req)
		if err != nil {
			// Unable to delete given id
			return nil, err
		}

		// Delete possible children
		rangeEntities := getChildrenCollections(u.GROUP, entity)
		for _, childEnt := range rangeEntities {
			childEntName := u.EntityToString(childEnt)
			pattern := primitive.Regex{Pattern: "^" + id + u.HN_DELIMETER, Options: ""}

			repository.GetDB().Collection(childEntName).DeleteMany(ctx,
				bson.M{"id": pattern})
		}

		return nil, nil
	})

	return err
}

func DeleteNonHierarchicalObject(entity, slug string) *u.Error {
	req := bson.M{"slug": slug}
	ctx, cancel := u.Connect()
	defer cancel()
	return repository.DeleteObject(ctx, entity, req)
}

// Helper functions

func domainHasObjects(domain string) bool {
	data := map[string]interface{}{}
	// Get all collections names
	ctx, cancel := u.Connect()
	db := repository.GetDB()
	collNames, _ := db.ListCollectionNames(ctx, bson.D{})

	// Check if at least one object belongs to domain
	for _, collName := range collNames {
		pattern := primitive.Regex{Pattern: "^" + domain, Options: ""}
		e := db.Collection(collName).FindOne(ctx, bson.M{"domain": pattern}).Decode(&data)
		if e == nil {
			// Found one!
			return true
		}
	}

	defer cancel()
	return false
}
