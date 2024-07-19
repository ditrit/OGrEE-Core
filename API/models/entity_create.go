package models

import (
	"p3/repository"
	u "p3/utils"

	"go.mongodb.org/mongo-driver/mongo"
)

func CreateEntity(entity int, t map[string]interface{}, userRoles map[string]Role) (map[string]interface{}, *u.Error) {
	tags, tagsPresent := getTags(t)
	if tagsPresent {
		err := verifyTagList(tags)
		if err != nil {
			return nil, err
		}
	}

	if err := prepareCreateEntity(entity, t, userRoles); err != nil {
		return nil, err
	}

	return WithTransaction(func(ctx mongo.SessionContext) (map[string]any, error) {
		if entity == u.TAG {
			err := createTagImage(ctx, t)
			if err != nil {
				return nil, err
			}
		}

		entStr := u.EntityToString(entity)

		_, err := repository.CreateObject(ctx, entStr, t)
		if err != nil {
			return nil, err
		}

		fixID(t)
		return t, nil
	})
}

func prepareCreateEntity(entity int, t map[string]interface{}, userRoles map[string]Role) *u.Error {
	if err := ValidateEntity(entity, t); err != nil {
		return err
	}

	// Check user permissions
	if u.IsEntityHierarchical(entity) {
		var domain string
		if entity == u.DOMAIN {
			domain = t["id"].(string)
		} else {
			domain = t["domain"].(string)
		}
		if permission := CheckUserPermissions(userRoles, entity, domain); permission < WRITE {
			return &u.Error{Type: u.ErrUnauthorized,
				Message: "User does not have permission to create this object"}
		}
	}

	delete(t, "parentId")

	return nil
}
