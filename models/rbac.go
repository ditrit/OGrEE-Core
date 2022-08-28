package models

import (
	u "p3/utils"

	"go.mongodb.org/mongo-driver/bson"
)

func CheckDomainExists(domain string) bool {
	x, e := GetEntity(bson.M{"name": domain}, "domain")
	if e != "" || x == nil {
		return false
	}
	return true
}

//Checks if domain is in RBAC Table and gets
//the hierarchy of the Domain
func GetUserDomainSpace(domain string) []string {
	ans := []string{domain}
	raw, e := GetHierarchyByName("domain", domain, nil, u.DOMAIN, 99)
	if e != "" {
		return nil
	}

	for i := range raw["children"].([]map[string]interface{}) {
		domain := raw["children"].([]map[string]interface{})[i]
		ans = append(ans, domain["name"].(string))
	}

	return ans
}

func RequestGen(x map[string]interface{}, role, domain string) {
	switch role {
	case "super":

	case "issuer":

	case "manager":
		RBACTable := GetUserDomainSpace(domain)
		if RBACTable == nil {
			x["domain"] = map[string]interface{}{"domain": domain}
		} else {
			x["domain"] = map[string]interface{}{"$in": RBACTable[:]}
		}

	case "user":
		x["domain"] = domain

	}
}

//Check if user can create Domains or accounts
func EnsureUserIsSuper(role string) bool {
	return role == "super" || role == "issuer"
}

//Checks if user can create/delete/update this object
func EnsureObjectPermission(obj map[string]interface{}, domain, role string) (bool, string) {

	switch role {
	case "user":
		return obj["domain"] == domain, ""
	case "manager":
		domains := GetUserDomainSpace(domain)
		/*println("DEBUG view Domain Space")
		for q := range domains {
			println(domains[q])
		}*/
		for i := range domains {
			if obj["domain"] == domains[i] {
				return true, ""
			}
		}
		return false, ""
	default:
		return true, ""
	}

}
