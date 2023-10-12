package controllers

import (
	"p3/models"
	u "p3/utils"
)

func deleteObject(entityStr string, id string, userRoles map[string]models.Role) *u.Error {
	entity := u.EntityStrToInt(entityStr)
	if entity == u.TAG {
		return models.DeleteTag(id)
	} else if u.IsEntityNonHierarchical(entity) {
		return models.DeleteNonHierarchicalObject(entityStr, id)
	} else {
		return models.DeleteHierarchicalObject(entityStr, id, userRoles)
	}
}
