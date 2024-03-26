package models

import (
	u "p3/utils"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestGetRequestFilterByDomainRootRoles(t *testing.T) {
	roles := map[string]Role{
		"*": Manager,
	}
	_, ok := GetRequestFilterByDomain(roles)
	assert.True(t, ok)

	roles["*"] = User
	_, ok = GetRequestFilterByDomain(roles)
	assert.True(t, ok)
}

func TestGetRequestFilterByDomain(t *testing.T) {
	domain := "domain1"
	subdomain := domain + ".subdomain1"
	roles := map[string]Role{
		"*":       Viewer,
		domain:    Manager,
		subdomain: User,
	}
	filter, ok := GetRequestFilterByDomain(roles)
	assert.True(t, ok)
	regex := filter["domain"].(primitive.Regex)
	// the pattern only has the manager domains
	assert.Equal(t, domain, regex.Pattern)

	// we change subdomain to manager role
	roles[subdomain] = Manager
	filter, ok = GetRequestFilterByDomain(roles)
	assert.True(t, ok)
	regex = filter["domain"].(primitive.Regex)
	assert.Equal(t, domain+"|"+subdomain, regex.Pattern)

	// only viewer roles
	roles[subdomain] = Viewer
	roles[domain] = Viewer
	_, ok = GetRequestFilterByDomain(roles)
	assert.False(t, ok)
}

func TestCheckUserPermissionsDomain(t *testing.T) {
	entity := u.DOMAIN
	roles := map[string]Role{
		"*":       Manager,
		"domain2": Viewer,
		"domain1": Viewer,
	}
	domain := "domain1.subdomain1"

	// root manager
	permission := CheckUserPermissions(roles, entity, domain)
	if permission != WRITE {
		t.Error("Root manager should have write permission")
	}

	// domain1 manager
	roles["*"] = User
	roles["domain1"] = Manager
	permission = CheckUserPermissions(roles, entity, domain)
	if permission != WRITE {
		t.Error("Parent domain manager should have write permission")
	}

	// domain1.subdomain1 manager
	roles["domain1"] = User
	roles[domain] = Manager
	permission = CheckUserPermissions(roles, entity, domain)
	if permission != WRITE {
		t.Error("Domain manager should have write permission")
	}

	// domain1.subdomain1 user
	roles["domain1"] = User
	roles[domain] = User
	permission = CheckUserPermissions(roles, entity, domain)
	if permission != NONE {
		t.Error("User should not have permission")
	}
}

func TestCheckUserPermissions(t *testing.T) {
	entity := u.ROOM
	rootRoles := map[string]Role{
		"*": Manager,
	}
	domain := "domain1"
	subdomain := domain + ".subdomain1"
	childSubdomain := subdomain + ".child"

	// root manager
	permission := CheckUserPermissions(rootRoles, entity, subdomain)
	if permission != WRITE {
		t.Error("Root manager should have write permission")
	}

	// root user
	rootRoles["*"] = User
	permission = CheckUserPermissions(rootRoles, entity, subdomain)
	if permission != WRITE {
		t.Error("Root user should have write permission")
	}

	// root viewer
	rootRoles["*"] = Viewer
	permission = CheckUserPermissions(rootRoles, entity, subdomain)
	if permission != READ {
		t.Error("Root viewer should have read permission")
	}

	roles := map[string]Role{
		domain: Manager,
	}

	// domain1 manager
	permission = CheckUserPermissions(roles, entity, subdomain)
	if permission != WRITE {
		t.Error("Parent domain manager should have write permission")
	}

	// domain1 viewer
	roles[domain] = Viewer
	permission = CheckUserPermissions(roles, entity, subdomain)
	if permission != READ {
		t.Error("Parent domain viewer should have read permission")
	}

	// domain1.subdomain1 manager
	delete(roles, domain)
	roles[subdomain] = Manager
	permission = CheckUserPermissions(roles, entity, subdomain)
	if permission != WRITE {
		t.Error("Domain manager should have write permission")
	}

	// domain1.subdomain1 viewer
	roles[subdomain] = Viewer
	permission = CheckUserPermissions(roles, entity, subdomain)
	if permission != READ {
		t.Error("Domain viewer should should have read permission")
	}

	// domain1.subdomain1.child manager
	delete(roles, subdomain)
	roles[childSubdomain] = Manager
	permission = CheckUserPermissions(roles, entity, subdomain)
	if permission != READONLYNAME {
		t.Error("Child manager should should have read only name permission")
	}

	// domain1.subdomain1.child viewer
	delete(roles, subdomain)
	roles[childSubdomain] = Viewer
	permission = CheckUserPermissions(roles, entity, subdomain)
	if permission != READONLYNAME {
		t.Error("Child viewer should should have read only name permission")
	}

	// no roles
	delete(roles, childSubdomain)
	permission = CheckUserPermissions(roles, entity, subdomain)
	if permission != NONE {
		t.Error("User with no roles should not have any permission")
	}
}
