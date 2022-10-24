package models

import (
	u "p3/utils"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func validateParent(ent string, entNum int, t map[string]interface{}) (map[string]interface{}, bool) {

	if entNum == TENANT {
		return nil, true
	}

	//Check ParentID is valid
	if t["parentId"] == nil {
		return u.Message(false, "ParentID is not valid"), false
	}

	objID, err := primitive.ObjectIDFromHex(t["parentId"].(string))
	if err != nil {
		return u.Message(false, "ParentID is not valid"), false
	}

	parent := map[string]interface{}{"parent": ""}
	switch entNum {
	case DEVICE:
		x, _ := GetEntity(bson.M{"_id": objID}, "rack")
		if x != nil {
			parent["parent"] = "rack"
			return parent, true
		}

		y, _ := GetEntity(bson.M{"_id": objID}, "device")
		if y != nil {
			parent["parent"] = "device"
			return parent, true
		}

		return u.Message(false,
			"ParentID should be correspond to Existing ID"), false

	case SENSOR, GROUP:
		w, _ := GetEntity(bson.M{"_id": objID}, "device")
		if w != nil {
			parent["parent"] = "device"
			return parent, true
		}

		x, _ := GetEntity(bson.M{"_id": objID}, "rack")
		if x != nil {
			parent["parent"] = "rack"
			return parent, true
		}

		y, _ := GetEntity(bson.M{"_id": objID}, "room")
		if y != nil {
			parent["parent"] = "room"
			return parent, true
		}

		z, _ := GetEntity(bson.M{"_id": objID}, "building")
		if z != nil {
			parent["parent"] = "building"
			return parent, true
		}

		return u.Message(false,
			"ParentID should be correspond to Existing ID"), false

	case STRAYDEV, STRAYSENSOR:
		if t["parentId"] != nil && t["parentId"] != "" {
			if pid, ok := t["parentId"].(string); ok {
				ID, _ := primitive.ObjectIDFromHex(pid)

				ctx, cancel := u.Connect()
				if GetDB().Collection("stray_device").FindOne(ctx,
					bson.M{"_id": ID}).Err() != nil {
					return u.Message(false,
						"ParentID should be an Existing ID or null"), false
				}
				defer cancel()
			} else {
				return u.Message(false,
					"ParentID should be an Existing ID or null"), false
			}
		}

	default:
		parentInt := u.GetParentOfEntityByInt(entNum)
		parent := u.EntityToString(parentInt)

		ctx, cancel := u.Connect()
		if GetDB().Collection(parent).
			FindOne(ctx, bson.M{"_id": objID}).Err() != nil {
			println("ENTITY VALUE: ", ent)
			println("We got Parent: ", parent, " with ID:", t["parentId"].(string))
			return u.Message(false,
				"ParentID should correspond to Existing ID"), false

		}
		defer cancel()
	}
	return nil, true
}

func ValidatePatch(ent int, t map[string]interface{}) (map[string]interface{}, bool) {
	for k := range t {
		switch k {
		case "name", "category", "domain":
			//Only for Entities until GROUP
			//And OBJTMPL
			if ent < GROUP+1 || ent == OBJTMPL {
				if v, _ := t[k]; v == nil {
					return u.Message(false,
						"Field: "+k+" cannot be nullified!"), false
				}
			}

		case "parentId":
			if ent < ROOMTMPL && ent > TENANT {
				x, ok := validateParent(u.EntityToString(ent), ent, t)
				if !ok {
					return x, ok
				}
			}
			//STRAYDEV's schema is very loose
			//thus we can safely invoke validateEntity
			if ent == STRAYDEV {
				x, ok := ValidateEntity(ent, t)
				if !ok {
					return x, ok
				}
			}

		case "attributes.color": // TENANT
			if ent == TENANT {
				if v, _ := t[k]; v == nil {
					return u.Message(false,
						"Field: "+k+" cannot nullified!"), false
				}
			}

		case "attributes.orientation": //SITE, ROOM, RACK, DEVICE
			if ent >= SITE && ent <= DEVICE {
				if v, _ := t[k]; v == nil {
					return u.Message(false,
						"Field: "+k+" cannot nullified!"), false
				}
			}

		case "attributes.usableColor",
			"attributes.reservedColor",
			"attributes.technicalColor": //SITE
			if ent == SITE {
				if v, _ := t[k]; v == nil {
					return u.Message(false,
						"Field: "+k+" cannot nullified!"), false
				}
			}

		case "attributes.posXY", "attributes.posXYUnit": // BLDG, ROOM, RACK
			if ent >= BLDG && ent <= RACK {
				if v, _ := t[k]; v == nil {
					return u.Message(false,
						"Field: "+k+" cannot nullified!"), false
				}
			}

		case "attributes": //TENANT ... SENSOR, OBJTMPL
			if (ent >= TENANT && ent < ROOMTMPL) || ent == OBJTMPL {
				if v, _ := t[k]; v == nil {
					return u.Message(false,
						"Field: "+k+" cannot nullified!"), false
				}
			}

		case "attributes.size", "attributes.sizeUnit",
			"attributes.height", "attributes.heightUnit":
			//BLDG ... DEVICE
			if ent >= BLDG && ent <= DEVICE {
				if v, _ := t[k]; v == nil {
					return u.Message(false,
						"Field: "+k+" cannot nullified!"), false
				}
			}

		case "attributes.floorUnit": //ROOM
			if ent == ROOM {
				if v, _ := t[k]; v == nil {
					return u.Message(false,
						"Field: "+k+" cannot nullified!"), false
				}
			}

		case "slug", "colors": //TEMPLATES
			if ent == OBJTMPL || ent == ROOMTMPL {
				if v, _ := t[k]; v == nil {
					return u.Message(false,
						"Field: "+k+" cannot nullified!"), false
				}
			}

		case "orientation", "sizeWDHm", "reservedArea",
			"technicalArea", "separators", "tiles": //ROOMTMPL
			if ent == ROOMTMPL {
				if v, _ := t[k]; v == nil {
					return u.Message(false,
						"Field: "+k+" cannot nullified!"), false
				}
			}

		case "description", "slots",
			"sizeWDHmm", "fbxModel": //OBJTMPL
			if ent == OBJTMPL {
				if v, _ := t[k]; v == nil {
					return u.Message(false,
						"Field: "+k+" cannot nullified!"), false
				}
			}

			/*case "type":
			if ent == SENSOR {
				if v, _ := t[k]; v == nil {
					return u.Message(false,
						"Field: "+k+" cannot nullified!"), false
				}

				if t[k] != "rack" &&
					t[k] != "device" && t[k] != "room" {
					return u.Message(false,
						"Incorrect values given for: "+k+"!"+
							"Please provide rack or device or room"), false
				}
			}*/

		}
	}
	return nil, true

}

func ValidateEntity(entity int, t map[string]interface{}) (map[string]interface{}, bool) {

	//parentObj := nil
	/*
		TODO:
		Need to capture device if it is a parent
		and check that the device parent has a slot
		attribute
	*/
	switch entity {
	case TENANT, SITE, BLDG, ROOM, RACK, DEVICE, AC,
		PWRPNL, CABINET, CORIDOR, SENSOR, GROUP:
		if t["name"] == nil || t["name"] == "" {
			return u.Message(false, "Name should be on payload"), false
		}

		if t["category"] == nil || t["category"] == "" {
			return u.Message(false, "Category should be on the payload"), false
		}

		if t["domain"] == nil || t["domain"] == "" {
			return u.Message(false, "Domain should be on the payload"), false
		}

		if t["description"] == nil || t["description"] == "" {
			return u.Message(false, "Description should be on the payload"), false
		}

		if _, ok := t["description"].([]interface{}); !ok {
			return u.Message(false, "Description should be an array type"), false
		}

		//Check if Parent ID is valid
		//returns a map[string]interface{} to hold parent entity
		//if parent found
		r, ok := validateParent(u.EntityToString(entity), entity, t)
		if !ok {
			return r, ok
		}

		if entity < AC || entity == PWRPNL ||
			entity == GROUP || entity == ROOMTMPL ||
			entity == OBJTMPL || entity == CORIDOR {
			if _, ok := t["attributes"]; !ok {
				return u.Message(false, "Attributes should be on the payload"), false
			} else {
				if v, ok := t["attributes"].(map[string]interface{}); !ok {
					return u.Message(false, "Attributes should be a JSON Dictionary"), false
				} else {
					switch entity {
					case TENANT:
						if _, ok := v["color"]; !ok || v["color"] == "" {
							return u.Message(false,
								"Color Attribute must be specified on the payload"), false
						}

					case SITE:
						switch v["orientation"] {
						case "EN", "NW", "WS", "SE", "NE", "SW":
						case "", nil:
							return u.Message(false, "Orientation should be on the payload"), false

						default:
							return u.Message(false, "Orientation is invalid!"), false
						}

					case BLDG:
						if v["posXY"] == "" || v["posXY"] == nil {
							return u.Message(false, "XY coordinates should be on payload"), false
						}

						if v["posXYUnit"] == "" || v["posXYUnit"] == nil {
							return u.Message(false, "PositionXYUnit should be on the payload"), false
						}

						if v["size"] == "" || v["size"] == nil {
							return u.Message(false, "Invalid building size on the payload"), false
						}

						if v["sizeUnit"] == "" || v["sizeUnit"] == nil {
							return u.Message(false, "Building size unit should be on the payload"), false
						}

						if v["height"] == "" || v["height"] == nil {
							return u.Message(false, "Invalid Height on payload"), false
						}

						if v["heightUnit"] == "" || v["heightUnit"] == nil {
							return u.Message(false, "Building Height unit should be on the payload"), false
						}

					case ROOM:
						if v["posXY"] == "" || v["posXY"] == nil {
							return u.Message(false, "XY coordinates should be on payload"), false
						}

						if v["posXYUnit"] == "" || v["posXYUnit"] == nil {
							return u.Message(false, "PositionXYUnit should be on the payload"), false
						}

						switch v["floorUnit"] {
						case "f", "m", "t":
						case "", nil:
							return u.Message(false, "floorUnit should be on the payload"), false
						default:
							return u.Message(false, "floorUnit is invalid!"), false

						}

						switch v["orientation"] {
						case "-E-N", "-E+N", "+E-N", "+E+N", "+N+E":
						case "-N-W", "-N+W", "+N-W", "+N+W":
						case "-W-S", "-W+S", "+W-S", "+W+S":
						case "-S-E", "-S+E", "+S-E", "+S+E":
						case "", nil:
							return u.Message(false, "Orientation should be on the payload"), false

						default:
							return u.Message(false, "Orientation is invalid!"), false
						}

						if v["size"] == "" || v["size"] == nil {
							return u.Message(false, "Invalid size on the payload"), false
						}

						if v["sizeUnit"] == "" || v["sizeUnit"] == nil {
							return u.Message(false, "Room size unit should be on the payload"), false
						}

						if v["height"] == "" || v["height"] == nil {
							return u.Message(false, "Invalid Height on payload"), false
						}

						if v["heightUnit"] == "" || v["heightUnit"] == nil {
							return u.Message(false, "Room Height unit should be on the payload"), false
						}

					case RACK:
						if v["posXY"] == "" || v["posXY"] == nil {
							return u.Message(false, "XY coordinates should be on payload"), false
						}

						if v["posXYUnit"] == "" || v["posXYUnit"] == nil {
							return u.Message(false, "PositionXYUnit should be on the payload"), false
						}

						switch v["orientation"] {
						case "front", "rear", "left", "right":
						case "", nil:
							return u.Message(false, "Orientation should be on the payload"), false

						default:
							return u.Message(false, "Orientation is invalid!"), false
						}

						if v["size"] == "" || v["size"] == nil {
							return u.Message(false, "Invalid size on the payload"), false
						}

						if v["sizeUnit"] == "" || v["sizeUnit"] == nil {
							return u.Message(false, "Rack size unit should be on the payload"), false
						}

						if v["height"] == "" || v["height"] == nil {
							return u.Message(false, "Invalid Height on payload"), false
						}

						if v["heightUnit"] == "" || v["heightUnit"] == nil {
							return u.Message(false, "Rack Height unit should be on the payload"), false
						}

						//Ensure the name is also unique among corridors
						req := bson.M{"name": t["name"].(string)}
						nameCheck, _ := GetManyEntities("corridor", req, nil)
						if nameCheck != nil {
							if len(nameCheck) != 0 {
								msg := "Rack name name must be unique among corridors and racks"
								if nameCheck != nil {
									println(nameCheck[0]["name"].(string))
								}
								return u.Message(false, msg), false
							}

						}

					case DEVICE:
						switch v["orientation"] {
						case "front", "rear", "frontflipped", "rearflipped":
						case "", nil:
							return u.Message(false, "Orientation should be on the payload"), false

						default:
							return u.Message(false, "Orientation is invalid!"), false
						}

						if v["size"] == "" || v["size"] == nil {
							return u.Message(false, "Invalid size on the payload"), false
						}

						if v["sizeUnit"] == "" || v["sizeUnit"] == nil {
							return u.Message(false, "Device size unit should be on the payload"), false
						}

						if v["height"] == "" || v["height"] == nil {
							return u.Message(false, "Invalid Height on payload"), false
						}

						if v["heightUnit"] == "" || v["heightUnit"] == nil {
							return u.Message(false, "Device Height unit should be on the payload"), false
						}

						if side, ok := v["side"]; ok {
							switch side {
							case "front", "rear", "frontflipped", "rearflipped":
							default:
								msg := "The 'Side' value (if given) must be one of" +
									"the given values: front, rear, frontflipped, rearflipped"
								return u.Message(false, msg), false
							}
						}

					case CORIDOR:
						//Ensure the temperature and 2 racks are valid
						if !IsString(v["temperature"]) {
							msg := "The temperature must be on the " +
								"payload and can only be a string value 'cold' or 'warm'"
							return u.Message(false, msg), false
						}
						if !StringIsAmongValues(v["temperature"].(string), []string{"warm", "cold"}) {
							msg := "The temperature must be on the " +
								"payload and can only be 'cold' or 'warm'"
							return u.Message(false, msg), false
						}

						if !IsString(v["content"]) {
							msg := "The racks must be on the payload and have the key:" +
								"'content' "
							return u.Message(false, msg), false
						}

						racks := strings.Split(v["content"].(string), ",")
						if len(racks) != 2 {
							msg := "2 racks separated by a comma must be on the payload"
							return u.Message(false, msg), false
						}

						//Ensure the name is also unique among racks
						req := bson.M{"name": t["name"].(string)}
						nameCheck, _ := GetManyEntities("rack", req, nil)
						if nameCheck != nil {
							if len(nameCheck) != 0 {
								msg := "Corridor name name must be unique among corridors and racks"
								return u.Message(false, msg), false
							}

						}

						//Fetch the 2 racks and ensure they exist
						ctx, cancel := u.Connect()
						filter := bson.M{"_id": t["parentId"], "name": racks[0]}
						orReq := bson.A{bson.D{{"name", racks[0]}}, bson.D{{"name", racks[1]}}}
						ans := bson.D{}

						filter = bson.M{"parentId": t["parentId"], "$or": orReq}
						res, e := GetDB().Collection("rack").Find(ctx, filter)
						defer cancel()
						if e != nil {
							msg := "The racks you specified were not found." +
								" Please verify your input and try again"
							println(e.Error())
							return u.Message(false, msg), false
						}

						//No need to remove '_id' so we can skip calling
						//ExtractCursor auxillary func
						res.All(ctx, &ans)

						if len(ans) != 2 {
							//Request possibly mentioned same racks
							//thus giving length of 1
							if !(len(ans) == 1 && racks[0] == racks[1]) {
								msg := "Unable to get the racks. Please check your inventory and try again"
								println("LENGTH OF RACK CHECK:", len(ans))
								println("CORRIDOR PARENTID: ", t["parentId"].(string))
								return u.Message(false, msg), false
							}

						}

						//Set the color manually based on temp. as specified by client
						if v["temperature"] == "warm" {
							v["color"] = "990000"
						} else if v["temperature"] == "cold" {
							v["color"] = "000099"
						}

					case GROUP:
						if !IsString(v["content"]) {
							msg := "The objects to be grouped must be on the payload," +
								" separated by a comma and have the key:" +
								"'content' "
							return u.Message(false, msg), false
						}

						objects := strings.Split(v["content"].(string), ",")
						if len(objects) <= 1 {
							if objects[0] == "" {
								msg := "objects separated by a comma must be" +
									" on the payload"
								return u.Message(false, msg), false
							}

						}

						//Ensure objects are all unique
						if _, ok := EnsureUnique(objects); !ok {
							msg := "The group cannot have duplicate objects"
							return u.Message(false, msg), false
						}

						//Ensure objects all exist
						orReq := bson.A{}
						for i := range objects {
							orReq = append(orReq, bson.D{{"name", objects[i]}})
						}
						filter := bson.M{"parentId": t["parentId"], "$or": orReq}

						//If parent is rack, retrieve devices
						if r["parent"].(string) == "rack" {
							ans, ok := GetManyEntities("device", filter, nil)
							if ok != "" {
								return u.Message(false, ok), false
							}
							if len(ans) != len(objects) {
								msg := "Unable to verify objects in specified group" +
									" please check and try again"
								return u.Message(false, msg), false
							}

						} else if r["parent"].(string) == "room" {

							//If parent is room, retrieve corridors and racks
							corridors, e1 := GetManyEntities("corridor", filter, nil)
							if e1 != "" {
								return u.Message(false, e1), false
							}

							racks, e2 := GetManyEntities("rack", filter, nil)
							if e2 != "" {
								return u.Message(false, e1), false
							}
							if len(racks)+len(corridors) != len(objects) {
								msg := "Some object(s) could be not be found. " +
									"Please check and try again"
								return u.Message(false, msg), false
							}
						}

					}
				}
			}
		}
	case ROOMTMPL, OBJTMPL:
		if t["slug"] == "" || t["slug"] == nil {
			return u.Message(false, "Slug should be on payload"), false
		}

		if _, ok := t["colors"]; !ok {
			return u.Message(false,
				"Colors should be on payload"), false
		}

		if entity == OBJTMPL {
			if _, ok := t["description"]; !ok {
				return u.Message(false,
					"Description should be on payload"), false
			}

			if _, ok := t["description"].([]interface{}); !ok {
				return u.Message(false,
					"Description should an array type"), false
			}

			/*if _, ok := t["category"]; !ok {
				return u.Message(false,
					"Category should be on payload"), false
			}*/

			if _, ok := t["sizeWDHmm"]; !ok {
				return u.Message(false,
					"Size,Width,Depth (mm) Array should be on payload"), false
			}

			if t["fbxModel"] == nil {
				return u.Message(false,
					"fbxModel should be on payload"), false
			}

			if _, ok := t["attributes"]; !ok {
				return u.Message(false,
					"Attributes should be on payload"), false
			}

			if _, ok := t["slots"]; !ok {
				return u.Message(false,
					"slots should be on payload"), false
			}

		} else { //ROOMTMPL
			if _, ok := t["orientation"]; !ok {
				return u.Message(false,
					"Orientation should be on payload"), false
			}

			if _, ok := t["sizeWDHm"]; !ok {
				return u.Message(false,
					"Size,Width,Depth Array should be on payload"), false
			}

			if _, ok := t["technicalArea"]; !ok {
				return u.Message(false,
					"TechnicalArea should be on payload"), false
			}

			if _, ok := t["reservedArea"]; !ok {
				return u.Message(false,
					"ReservedArea should be on payload"), false
			}

			if _, ok := t["separators"]; !ok {
				return u.Message(false,
					"Separators should be on payload"), false
			}

			if _, ok := t["tiles"]; !ok {
				return u.Message(false,
					"Tiles should be on payload"), false
			}
		}

	case STRAYDEV, STRAYSENSOR:
		//Check for parent if PID provided

		if t["name"] == nil || t["name"] == "" {
			return u.Message(false, "Please provide a valid name"), false
		}

		//Need to check for uniqueness before inserting
		//this is helpful for the validation endpoints
		ctx, cancel := u.Connect()
		entStr := u.EntityToString(entity)

		if c, _ := GetDB().Collection(entStr).CountDocuments(ctx,
			bson.M{"name": t["name"]}); c != 0 {
			msg := "Error a " + entStr + " with the name provided already exists." +
				"Please provide a unique name"
			return u.Message(false, msg), false
		}
		defer cancel()

	}

	//Successfully validated the Object
	return u.Message(true, "success"), true
}

//Auxillary Functions
func EnsureUnique(x []string) (string, bool) {
	dict := map[string]int{}
	for _, item := range x {
		dict[item]++
		if dict[item] > 1 {
			return item, false
		}
	}
	return "", true
}

func StringIsAmongValues(x string, values []string) bool {
	for i := range values {
		if x == values[i] {
			return true
		}
	}
	return false
}

func IsString(x interface{}) bool {
	if _, ok := x.(string); ok {
		return true
	}
	return false
}
