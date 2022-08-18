package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type AccessControl struct {
	Table []map[string]interface{}
}

var RBAC AccessControl

func Init() {
	RBAC.Table = []map[string]interface{}{}
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
			//Head Case
			if i == 0 {
				RBAC.Table = RBAC.Table[i+1:]
			} else if i == len(RBAC.Table) { //Tail Case
				RBAC.Table = RBAC.Table[:i-1]
			} else { //Middle Case
				x := RBAC.Table[:i]
				x = append(x, RBAC.Table[i+1:]...)
				RBAC.Table = x
			}
		}
	}
}

func GetUserDomainIdxInTable(domain string) int {
	for i, v := range RBAC.Table {
		if v["name"].(string) == domain {
			return i
		}
	}
	return -1
}

func RequestGen(x map[string]interface{}, role, domain string) {
	switch role {
	case "super":
		return
	case "issuer":
		return
	case "manager":
		idx := GetUserDomainIdxInTable(domain)
		if idx == -1 {
			return
		}
		x["domain"] = map[string]interface{}{"$in": RBAC.Table[idx:]}
		return
	case "user":
		x["domain"] = domain
		return

	}
}
