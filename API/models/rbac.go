package models

import (
	u "p3/utils"
	"regexp"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Roles
type Role string

const (
	Manager Role = "manager"
	User    Role = "user"
	Viewer  Role = "viewer"
)

// Actions
type Permission int

const (
	NONE Permission = iota
	READONLYNAME
	READ
	WRITE
)

const ROOT_DOMAIN = "*"

func CheckDomainExists(domain string) bool {
	if domain == ROOT_DOMAIN {
		return true
	}
	x, e := GetEntity(bson.M{"id": domain}, "domain", u.RequestFilters{}, nil)
	return e == nil && x != nil
}

func GetRequestFilterByDomain(userRoles map[string]Role) (bson.M, bool) {
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
		// the user is only a viewer, return false
		return filter, false
	} else {
		filter["domain"] = primitive.Regex{Pattern: domainPattern, Options: ""}
		return filter, true
	}
}

func CheckUserPermissions(userRoles map[string]Role, objEntity int, objDomain string) Permission {
	permission := NONE
	if objEntity == u.DOMAIN {
		if userRoles[ROOT_DOMAIN] == Manager {
			return WRITE
		}
		for userDomain, role := range userRoles {
			if DomainIsEqualOrChild(userDomain, objDomain) && role == Manager {
				//objDomain is equal or child of userDomain
				return WRITE
			}
		}
	} else {
		if userRoles[ROOT_DOMAIN] == User || userRoles[ROOT_DOMAIN] == Manager {
			return WRITE
		} else if userRoles[ROOT_DOMAIN] == Viewer {
			return READ
		}

		for userDomain, role := range userRoles {
			if DomainIsEqualOrChild(userDomain, objDomain) {
				//objDomain is equal or child of userDomain
				if role == Viewer {
					permission = READ
				} else {
					permission = WRITE
					break // highest possible
				}
			} else if DomainIsEqualOrChild(objDomain, userDomain) {
				// objDomain is father of userDomain
				if permission < READONLYNAME {
					permission = READONLYNAME
				}
			}
		}
	}
	return permission
}

func DomainIsEqualOrChild(refDomain, domainToCheck string) bool {
	match, _ := regexp.MatchString("^"+refDomain+"\\.", domainToCheck)
	return match || refDomain == domainToCheck
}

func CheckCanManageUser(callerRoles map[string]Role, newUserRoles map[string]Role) bool {
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
