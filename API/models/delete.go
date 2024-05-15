package models

import (
	u "p3/utils"
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
