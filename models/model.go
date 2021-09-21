package models

import (
	u "p3/utils"
)

const (
	TENANT = iota
	SITE
	BLDG
	ROOM
	RACK
	DEVICE
	SUBDEV
	SUBDEV1
)

func ValidateEntity(entity int, t map[string]interface{}) (map[string]interface{}, bool) {
	switch entity {
	case TENANT:
		if _, ok := t["name"]; !ok {
			return u.Message(false, "Tenant Name should be on payload"), false
		}

		if _, ok := t["category"]; !ok {
			return u.Message(false, "Category should be on the payload"), false
		}

		if _, ok := t["domain"]; !ok {
			return u.Message(false, "Domain should be on the payload"), false
		}

		if _, ok := t["attributes"]; !ok {
			return u.Message(false, "Color Attribute should be on the payload"), false
		} else {
			if v, ok := t["attributes"].(map[string]interface{}); !ok {
				return u.Message(false, "Color Attribute should be on the payload"), false
			} else {
				if _, ok := v["color"]; !ok {
					return u.Message(false,
						"Color Attribute must be specified on the payload"), false
				}
			}
		}

	}

	//Successfully validated the Object
	return u.Message(true, "success"), true
}

func CreateEntity(entity int, t map[string]interface{}) (map[string]interface{}, string) {

	if resp, ok := ValidateEntity(entity, t); !ok {
		return resp, "validate"
	}

	ctx, cancel := u.Connect()

	/*switch entity {
	case TENANT:

		if _, e := GetDB().Collection("tenant").InsertOne(ctx, t); e != nil {
			return u.Message(false, "Internal error while creating Tenant: "+e.Error()),
				e.Error()
		}

	}*/

	entStr := u.EntityToString(entity)
	res, e := GetDB().Collection(entStr).InsertOne(ctx, t)
	if e != nil {
		return u.Message(false,
				"Internal error while creating "+entStr+": "+e.Error()),
			e.Error()
	}
	defer cancel()

	t["id"] = res.InsertedID

	resp := u.Message(true, "success")
	resp["data"] = t
	return resp, ""
}
