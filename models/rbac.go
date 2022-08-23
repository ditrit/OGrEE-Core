package models

type AccessControl struct {
	Table []map[string]interface{} //Store Table of Domains
}

var RBAC AccessControl

//Need to load domains if they exist into the table
/*func Init() {
	RBAC.Table = []map[string]interface{}{}
	domains, e := GetManyEntities("domain", bson.M{"parentId": nil}, nil)
	if e != "" {
		return
	}

	if len(domains) > 1 {
		return //for now we only support 1 domain tree
	}

	root := domains[0]
	RBAC.Table = append(RBAC.Table, root)
	ID := root["_id"].(primitive.ObjectID)

	hierarchyInf, e := GetEntityHierarchy(ID, "domain", 0, 99)
	if d, ok := hierarchyInf["data"]; ok {
		if dMap, ok := d.(map[string]interface{}); ok {
			if objInf2, ok := dMap["objects"]; ok {
				if objInf, ok := objInf2.(map[string]interface{}); ok {
					if objArr, ok := objInf["objects"]; ok {
						if objs, ok := objArr.([]map[string]interface{}); ok {
							for i := range objs {
								RBAC.Table = append(RBAC.Table, objs[i])
							}
						}
					}
				}
			}
		}
	}

}

//Insert New Domain in RBAC Table
func NewDomain(domain map[string]interface{}, parentID string) {
	if parentID == "" {
		x := []map[string]interface{}{domain}
		RBAC.Table = append(x, RBAC.Table...)
	} else {
		for i, v := range RBAC.Table {
			parent, _ := primitive.ObjectIDFromHex(parentID)
			if v["id"] == parent {
				if i == len(RBAC.Table) {
					RBAC.Table = append(RBAC.Table, domain)
				} else {
					//Middle insertion case
					newTable := RBAC.Table[:i]
					newTable = append(newTable, domain)
					newTable = append(newTable, RBAC.Table[i+1:]...)
				}
			}
		}
	}
}

func DeleteDomain(domain string) {
	for i, v := range RBAC.Table {
		//Found Domain to delete
		if v["name"].(string) == domain {
			if i == 0 { //Head Case
				RBAC.Table = RBAC.Table[i+1:]
			} else if i == len(RBAC.Table) { //Tail Case
				RBAC.Table = RBAC.Table[:i-1]
			} else { //Middle Case
				RBAC.Table = append(RBAC.Table[:i], RBAC.Table[i+1:]...)
			}
		}
	}
}

//TODO: Later
func UpdateDomain(oldName, newName string) {
	for _, v := range RBAC.Table {
		if v["name"].(string) == oldName {
			v["name"] = newName
		}
	}
}*/

//Checks if domain is in RBAC Table and gets
//the hierarchy of the Domain
func GetUserDomainSpace(domain string) []string {
	ans := []string{domain}
	raw, e := GetHierarchyByName("domain", domain, 0, 99)
	if e != "" {
		return nil
	}

	for i := range raw["children"].([]map[string]interface{}) {
		domain := raw["children"].([]map[string]interface{})[i]
		ans = append(ans, domain["name"].(string))
	}

	return ans
	/*for i, v := range RBAC.Table {
		if v["name"].(string) == domain {
			return i
		}
	}
	return -1*/
}

func RequestGen(x map[string]interface{}, role, domain string) {
	switch role {
	case "super":

	case "issuer":

	case "manager":
		RBACTable := GetUserDomainSpace(domain)
		if RBACTable == nil {
			x["domain"] = map[string]interface{}{"$in": domain}
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
