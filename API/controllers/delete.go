package controllers

import (
	"log"
	"p3/models"
	u "p3/utils"
)

func deleteObject(entityStr string, id string, userRoles map[string]models.Role) *u.Error {
	log.Println("-----------delete-------------")
	log.Println(entityStr)
	log.Println(id)
	entity := u.EntityStrToInt(entityStr)
	if entity == u.TAG {
		log.Println("delete tag")
		return models.DeleteTag(id)
	} else if u.IsEntityNonHierarchical(entity) {
		return models.DeleteNonHierarchicalObject(entityStr, id)
	} else {
		return models.DeleteHierarchicalObject(entityStr, id, userRoles)
	}
}
