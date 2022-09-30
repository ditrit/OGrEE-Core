package models

import (
	u "p3/utils"

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

	switch entNum {
	case DEVICE:
		x, _ := GetEntity(bson.M{"_id": objID}, "rack")
		y, _ := GetEntity(bson.M{"_id": objID}, "device")
		if x == nil && y == nil {
			return u.Message(false,
				"ParentID should be correspond to Existing ID"), false
		}
	case SENSOR, GROUP:
		w, _ := GetEntity(bson.M{"_id": objID}, "device")
		x, _ := GetEntity(bson.M{"_id": objID}, "rack")
		y, _ := GetEntity(bson.M{"_id": objID}, "room")
		z, _ := GetEntity(bson.M{"_id": objID}, "building")
		if w == nil && x == nil && y == nil && z == nil {
			return u.Message(false,
				"ParentID should be correspond to Existing ID"), false
		}
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
		PWRPNL, CABINET, CORIDOR, SENSOR:
		if t["name"] == nil || t["name"] == "" {
			return u.Message(false, "Name should be on payload"), false
		}

		/*if t["category"] == nil || t["category"] == "" {
			return u.Message(false, "Category should be on the payload"), false
		}*/

		if t["domain"] == nil || t["domain"] == "" {
			return u.Message(false, "Domain should be on the payload"), false
		}

		//Check if Parent ID is valid
		r, ok := validateParent(u.EntityToString(entity), entity, t)
		if !ok {
			return r, ok
		}

		if entity < AC || entity == PWRPNL ||
			entity == GROUP || entity == ROOMTMPL ||
			entity == OBJTMPL {
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
	case GROUP:
		if t["name"] == "" || t["name"] == nil {
			return u.Message(false, "Name should be on payload"), false
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
