package models

import (
	"encoding/hex"
	"encoding/json"
	u "p3/utils"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func validateParent(ent string, entNum int, t map[string]interface{}) (map[string]interface{}, bool) {

	if entNum == u.SITE {
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
	case u.DEVICE:
		x, _ := GetEntity(bson.M{"_id": objID}, "rack", []string{})
		if x != nil {
			parent["parent"] = "rack"
			return parent, true
		}

		y, _ := GetEntity(bson.M{"_id": objID}, "device", []string{})
		if y != nil {
			parent["parent"] = "device"
			return parent, true
		}

		return u.Message(false,
			"ParentID should be correspond to Existing ID"), false

	case u.SENSOR, u.GROUP:
		w, _ := GetEntity(bson.M{"_id": objID}, "device", []string{})
		if w != nil {
			parent["parent"] = "device"
			return parent, true
		}

		x, _ := GetEntity(bson.M{"_id": objID}, "rack", []string{})
		if x != nil {
			parent["parent"] = "rack"
			return parent, true
		}

		y, _ := GetEntity(bson.M{"_id": objID}, "room", []string{})
		if y != nil {
			parent["parent"] = "room"
			return parent, true
		}

		z, _ := GetEntity(bson.M{"_id": objID}, "building", []string{})
		if z != nil {
			parent["parent"] = "building"
			return parent, true
		}

		return u.Message(false,
			"ParentID should be correspond to Existing ID"), false

	case u.STRAYDEV, u.STRAYSENSOR:
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
			//Only for Entities until u.GROUP
			//And u.OBJTMPL
			if ent < u.GROUP+1 || ent == u.OBJTMPL {
				if v, _ := t[k]; v == nil {
					return u.Message(false,
						"Field: "+k+" cannot be nullified!"), false
				}
			}

		case "parentId":
			if ent < u.ROOMTMPL && ent > u.SITE {
				x, ok := validateParent(u.EntityToString(ent), ent, t)
				if !ok {
					return x, ok
				}
			}
			//u.STRAYDEV's schema is very loose
			//thus we can safely invoke validateEntity
			if ent == u.STRAYDEV {
				x, ok := ValidateEntity(ent, t)
				if !ok {
					return x, ok
				}
			}

		case "attributes.orientation": //SITE, ROOM, RACK, DEVICE
			if ent >= u.SITE && ent <= u.DEVICE {
				if v, _ := t[k]; v == nil {
					return u.Message(false,
						"Field: "+k+" cannot nullified!"), false
				}
			}

		case "attributes.usableColor",
			"attributes.reservedColor",
			"attributes.technicalColor": //u.SITE
			if ent == u.SITE {
				if v, _ := t[k]; v == nil {
					return u.Message(false,
						"Field: "+k+" cannot nullified!"), false
				}
			}

		case "attributes.posXY", "attributes.posXYUnit": // u.BLDG, u.ROOM, u.RACK
			if ent >= u.BLDG && ent <= u.RACK {
				if v, _ := t[k]; v == nil {
					return u.Message(false,
						"Field: "+k+" cannot nullified!"), false
				}
			}

		case "attributes": //u.SITE ... u.SENSOR, u.OBJTMPL
			if (ent >= u.SITE && ent < u.ROOMTMPL) || ent == u.OBJTMPL {
				if v, _ := t[k]; v == nil {
					return u.Message(false,
						"Field: "+k+" cannot nullified!"), false
				}
			}

		case "attributes.size", "attributes.sizeUnit",
			"attributes.height", "attributes.heightUnit":
			//u.BLDG ... u.DEVICE
			if ent >= u.BLDG && ent <= u.DEVICE {
				if v, _ := t[k]; v == nil {
					return u.Message(false,
						"Field: "+k+" cannot nullified!"), false
				}
			}

		case "attributes.floorUnit": //u.ROOM
			if ent == u.ROOM {
				if v, _ := t[k]; v == nil {
					return u.Message(false,
						"Field: "+k+" cannot nullified!"), false
				}
			}

		case "slug", "colors": //TEMPLATES
			if ent == u.OBJTMPL || ent == u.ROOMTMPL {
				if v, _ := t[k]; v == nil {
					return u.Message(false,
						"Field: "+k+" cannot nullified!"), false
				}
			}

		case "orientation", "sizeWDHm", "reservedArea",
			"technicalArea", "separators", "tiles": //u.ROOMTMPL
			if ent == u.ROOMTMPL {
				if v, _ := t[k]; v == nil {
					return u.Message(false,
						"Field: "+k+" cannot nullified!"), false
				}
			}

		case "description", "slots",
			"sizeWDHmm", "fbxModel": //u.OBJTMPL
			if ent == u.OBJTMPL {
				if v, _ := t[k]; v == nil {
					return u.Message(false,
						"Field: "+k+" cannot nullified!"), false
				}
			}

			/*case "type":
			if ent == u.SENSOR {
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
	case u.SITE, u.BLDG, u.ROOM, u.RACK, u.DEVICE, u.AC,
		u.PWRPNL, u.CABINET, u.CORIDOR, u.SENSOR, u.GROUP:
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
			return u.Message(false,
				"Description should be on the payload as an array"), false
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

		if entity < u.AC || entity == u.PWRPNL ||
			entity == u.GROUP || entity == u.ROOMTMPL ||
			entity == u.OBJTMPL || entity == u.CORIDOR {
			if _, ok := t["attributes"]; !ok {
				return u.Message(false, "Attributes should be on the payload"), false
			} else {
				if v, ok := t["attributes"].(map[string]interface{}); !ok {
					return u.Message(false, "Attributes should be a JSON Dictionary"), false
				} else {
					switch entity {

					case u.SITE:
						if !IsOrientation(v["orientation"], entity) {
							if v["orientation"] == nil || v["orientation"] == "" {
								return u.Message(false, "Orientation should be on the payload"), false
							}
							return u.Message(false, "Orientation is invalid!"), false
						}

					case u.BLDG:
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

					case u.ROOM:
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

						//Check Orientation
						if !IsNonStdOrientation(v["orientation"], entity) {
							if v["orientation"] == nil || v["orientation"] == "" {
								return u.Message(false, "Orientation should be on the payload"), false
							}
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

					case u.RACK:
						if v["posXYZ"] == "" || v["posXYZ"] == nil {
							return u.Message(false, "XYZ coordinates should be on payload"), false
						} else {
							//check if format is Vector3, example {\"x\":25,\"y\":29.4,\"z\":0}"
							var posXYZ map[string]float32
							err := json.Unmarshal([]byte(v["posXYZ"].(string)), &posXYZ)
							if err != nil {
								return u.Message(false, "Invalid posXYZ on payload: "+err.Error()), false
							}

							if len(posXYZ) != 3 {
								return u.Message(false, "Invalid posXYZ on payload: should be Vector3 "), false
							}

							for _, key := range []string{"x", "y", "z"} {
								if _, ok = posXYZ[key]; !ok {
									return u.Message(false, "Invalid posXYZ on payload: missing "+key), false
								}
							}
						}

						//Check Orientation
						if !IsOrientation(v["orientation"], entity) {
							if v["orientation"] == nil || v["orientation"] == "" {
								return u.Message(false, "Orientation should be on the payload"), false
							}
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

					case u.DEVICE:
						if !IsOrientation(v["orientation"], entity) {
							if v["orientation"] == nil || v["orientation"] == "" {
								return u.Message(false, "Orientation should be on the payload"), false
							}
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

					case u.CORIDOR:
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

						//Trim Spaces because they mess up
						//the retrieval of objects from DB
						racks[0] = strings.TrimSpace(racks[0])
						racks[1] = strings.TrimSpace(racks[1])

						//Ensure the name is also unique among racks
						req := bson.M{"name": t["name"].(string)}
						nameCheck, _ := GetManyEntities("rack", req, nil)
						if nameCheck != nil {
							if len(nameCheck) != 0 {
								msg := "Corridor name must be unique among corridors and racks"
								return u.Message(false, msg), false
							}

						}

						//Fetch the 2 racks and ensure they exist
						filter := bson.M{"_id": t["parentId"], "name": racks[0]}
						orReq := bson.A{bson.D{{"name", racks[0]}}, bson.D{{"name", racks[1]}}}

						filter = bson.M{"parentId": t["parentId"], "$or": orReq}
						ans, e := GetManyEntities("rack", filter, nil)
						if e != "" {
							msg := "The racks you specified were not found." +
								" Please verify your input and try again"
							println(e)
							return u.Message(false, msg), false
						}

						if len(ans) != 2 {
							//Request possibly mentioned same racks
							//thus giving length of 1
							if !(len(ans) == 1 && racks[0] == racks[1]) {

								//Figure out the rack name that wasn't found
								var notFound string
								if racks[0] != ans[0]["name"].(string) {
									notFound = racks[0]
								} else {
									notFound = racks[1]
								}
								msg := "Unable to get the rack: " + notFound + ". Please check your inventory and try again"
								println("LENGTH OF u.RACK CHECK:", len(ans))
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

					case u.GROUP:
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
	case u.ROOMTMPL, u.OBJTMPL:
		if t["slug"] == "" || t["slug"] == nil {
			return u.Message(false, "Slug should be on payload"), false
		}

		if _, ok := t["colors"]; !ok {
			return u.Message(false,
				"Colors should be on payload"), false
		}

		if entity == u.OBJTMPL {
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

		} else { //u.ROOMTMPL
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

	case u.BLDGTMPL:
		if t["slug"] == "" || t["slug"] == nil {
			return u.Message(false, "Slug should be on payload"), false
		}

		if t["orientation"] == "" || t["orientation"] == nil {
			return u.Message(false, "Orientation should be on payload"), false
		}

		if !IsOrientation(t["orientation"], u.BLDGTMPL) {
			return u.Message(false, "Orientation is invalid!"), false
		}

		if !IsInfArr(t["sizeWDHm"]) {
			return u.Message(false,
				"Please provide the size as a 3 numerical element array"), false
		}

		if len(t["sizeWDHm"].([]interface{})) != 3 {
			return u.Message(false,
				"The size can only have 3 numerical elements"), false
		}

		for _, elt := range t["sizeWDHm"].([]interface{}) {
			if !IsInt(elt) && !IsFloat(elt) {
				return u.Message(false,
					"An element in the size was not numerical!"), false
			}
		}

		if !IsInfArr(t["sizeWDHm"]) {
			return u.Message(false,
				"Please provide the size as a 3 numerical element array"), false
		}

		if len(t["sizeWDHm"].([]interface{})) != 3 {
			return u.Message(false,
				"The size can only have 3 numerical elements"), false
		}

		if !IsInfArr(t["vertices"]) {
			return u.Message(false,
				"Please provide an array of 2 element (int) array(s) for the vertices"), false
		}

		for _, arr := range t["vertices"].([]interface{}) {
			if !IsInfArr(arr) {
				return u.Message(false,
					"An element in the vertices array was not an array!"), false
			}

			if len(arr.([]interface{})) != 2 {
				return u.Message(false,
					"An element in the vertices array does not have length of 2"), false
			}

			for i, element := range arr.([]interface{}) {
				if !IsFloat(element) {
					return u.Message(false,
						"All arrays in the vertices must have integers and be of length 2"), false
				}
				//Since ints are being interpreted as floats
				//we can just truncate all floats for now as a compromise
				arr.([]interface{})[i] = int(element.(float64))
			}
		}

	case u.STRAYDEV, u.STRAYSENSOR:
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

// Auxillary Functions
// Auxillary Functions
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

func IsInt(x interface{}) bool {
	if _, ok := x.(int); ok {
		return true
	}
	return false
}

func IsFloat(x interface{}) bool {
	if _, ok := x.(float64); ok {
		return true
	}
	return false
}

func IsFloatArr(x interface{}) bool {
	if _, ok := x.([]float64); ok {
		return true
	}
	return false
}

func IsInfArr(x interface{}) bool {
	if _, ok := x.([]interface{}); ok {
		return true
	}
	return false
}

func IsIntArr(x interface{}) bool {
	if _, ok := x.([]int); ok {
		return true
	}
	return false
}

func IsHexString(s string) bool {
	//Eliminate 'odd length' errors
	if len(s)%2 != 0 {
		s = "0" + s
	}

	_, err := hex.DecodeString(s)
	return err == nil
}

func IsOrientation(x interface{}, ent int) bool {
	if ent == u.SITE {
		switch x {
		case "EN", "NW", "WS", "SE", "NE", "SW":
			return true

		default:
			return false
		}
	}

	if ent == u.BLDGTMPL || ent == u.ROOMTMPL {
		switch x {
		case "EN", "NW", "WS", "SE", "NE", "SW",
			"-E-N", "-E+N", "+E-N", "+E+N", "+N+E",
			"+N-E", "-N-E", "-N+E",
			"-N-W", "-N+W", "+N-W", "+N+W",
			"-W-S", "-W+S", "+W-S", "+W+S",
			"-S-E", "-S+E", "+S-E", "+S+E":
			return true
		default:
			if !IsString(x) {
				return false
			}
			_, e := strconv.Atoi(x.(string))
			_, e1 := strconv.ParseFloat(x.(string), 64)
			return e1 == nil || e == nil
		}
	}

	if ent == u.ROOM {
		switch x {
		case "EN", "NW", "WS", "SE", "NE", "SW",
			"-E-N", "-E+N", "+E-N", "+E+N", "+N+E",
			"+N-E", "-N-E", "-N+E",
			"-N-W", "-N+W", "+N-W", "+N+W",
			"-W-S", "-W+S", "+W-S", "+W+S",
			"-S-E", "-S+E", "+S-E", "+S+E":
			return true
		default:
			return false
		}
	}

	if ent == u.RACK {
		switch x {
		case "front", "rear", "left", "right":
			return true
		default:
			return false
		}
	}

	if ent == u.DEVICE {
		switch x {
		case "front", "rear", "frontflipped", "rearflipped":
			return true
		default:
			return false
		}

	}
	//Control should not reach here
	return false
}

func IsNonStdOrientation(x interface{}, ent int) bool {
	switch x.(type) {
	case string:
		if !IsOrientation(x, ent) {
			//Check if it is a numerical string
			_, intErr := strconv.Atoi(x.(string))
			_, floatErr := strconv.ParseFloat(x.(string), 64)
			return intErr == nil || floatErr == nil
		}
		return true
	case float64, float32, int:
		return true
	default:
		return false
	}
}
