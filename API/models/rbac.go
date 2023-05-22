package models

import (
	u "p3/utils"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Roles
const (
	Manager = "manager"
	User    = "user"
	Viewer  = "viewer"
)

// Actions
const (
	READ = iota
	WRITE
	READONLYNAME
)

const ROOT_DOMAIN = "*"

func CheckDomainExists(domain string) bool {
	if domain == ROOT_DOMAIN {
		return true
	}
	x, e := GetEntity(bson.M{"hierarchyName": domain}, "domain", u.RequestFilters{}, nil)
	return e == "" && x != nil
}

func GetRequestFilterByDomain(userRoles map[string]string) (bson.M, bool) {
	filter := bson.M{}
	if userRoles[ROOT_DOMAIN] == Manager || userRoles[ROOT_DOMAIN] == User {
		return filter, true
	}
	domainPattern := ""
	for domain, role := range userRoles {
		switch role {
		case User:
		case Manager:
			if domainPattern == "" {
				domainPattern = domain
			} else {
				domainPattern = domainPattern + "|" + domain
			}
		}
	}
	if domainPattern == "" {
		return filter, false
	} else {
		filter["domain"] = primitive.Regex{Pattern: domainPattern, Options: ""}
		return filter, true
	}
}

func CheckUserPermissions(userRoles map[string]string, objEntity, requestType int, objDomain string) (bool, int) {
	if objEntity == u.DOMAIN {
		if userRoles[ROOT_DOMAIN] == Manager {
			return true, WRITE
		}
		for userDomain, role := range userRoles {
			if role == Manager && strings.Contains(objDomain, userDomain) {
				return true, WRITE
			}
		}

	} else {
		if requestType == READ {
			if userRoles[ROOT_DOMAIN] != "" {
				return true, READ
			}
			action := -1
			for userDomain := range userRoles {
				if strings.Contains(objDomain, userDomain) {
					//objDomain is equal or child of userDomain
					action = READ
					break
				} else if strings.Contains(objDomain, userDomain) {
					// objDomain is father of userDomain
					action = READONLYNAME
				}
			}
			if action >= 0 {
				return true, action
			}
		} else {
			// WRITE
			if userRoles[ROOT_DOMAIN] == User || userRoles[ROOT_DOMAIN] == Manager {
				return true, WRITE
			}
			for userDomain, role := range userRoles {
				if role != Viewer && strings.Contains(objDomain, userDomain) {
					return true, WRITE
				}
			}
		}
	}
	return false, -1
}

func CheckCanManageUser(callerRoles map[string]string, newUserRoles map[string]string) bool {
	if callerRoles[ROOT_DOMAIN] != Manager {
		for newUserDomain := range newUserRoles {
			roleValidated := false
			for callerDomain, callerRole := range callerRoles {
				if callerRole == Manager && strings.Contains(newUserDomain, callerDomain) {
					//newUserDomain is equal or child of callerDomain
					roleValidated = true
					break
				}
			}
			if !roleValidated {
				return false
			}
		}
	}
	return true
}
