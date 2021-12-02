package models

import (
	"context"
	"crypto/rand"
	"fmt"
	u "p3/utils"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	TENANT = iota
	SITE
	BLDG
	ROOM
	RACK
	DEVICE
	AC
	PWRPNL
	WALL
	CABINET
	AISLE
	TILE

	CORIDOR
	ROOMSENSOR
	RACKSENSOR
	DEVICESENSOR
	ROOMTMPL
	OBJTMPL
	GROUP
)

//Function will recursively iterate through nested obj
//and accumulate whatever is found into category arrays
func parseDataForNonStdResult(ent string, eNum, end int, data map[string]interface{}) map[string][]map[string]interface{} {
	var nxt string
	ans := map[string][]map[string]interface{}{}
	add := data[u.EntityToString(eNum+1)+"s"].([]map[string]interface{})

	//NEW REWRITE
	for i := eNum; i+2 < end; i++ {
		idx := u.EntityToString(i + 1)
		//println("trying IDX: ", idx)
		firstArr := add

		ans[idx+"s"] = firstArr

		for q := range firstArr {
			nxt = u.EntityToString(i + 2)
			println("NXT: ", nxt)
			ans[nxt+"s"] = append(ans[nxt+"s"],
				ans[idx+"s"][q][nxt+"s"].([]map[string]interface{})...)
		}
		add = ans[nxt+"s"]

	}

	return ans
}

func genID(length int) string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-"
	ll := len(chars)
	b := make([]byte, length)
	rand.Read(b) // generates len(b) random bytes
	for i := 0; i < length; i++ {
		b[i] = chars[int(b[i])%ll]
	}
	return string(b)
}

//Mongo returns '_id' instead of id
func fixID(data map[string]interface{}) map[string]interface{} {
	if v, ok := data["_id"]; ok {
		data["id"] = v
		delete(data, "_id")
	}
	return data
}

func ValidateEntity(entity int, t map[string]interface{}) (map[string]interface{}, bool) {
	var objID primitive.ObjectID
	var err error
	switch entity {
	case TENANT, SITE, BLDG, ROOM, RACK, DEVICE, AC,
		PWRPNL, WALL, CABINET, AISLE,
		TILE, CORIDOR, RACKSENSOR, DEVICESENSOR:
		if t["name"] == "" {
			return u.Message(false, "Name should be on payload"), false
		}

		if t["category"] == "" {
			return u.Message(false, "Category should be on the payload"), false
		}

		if t["domain"] == "" {
			return u.Message(false, "Domain should be on the payload"), false
		}

		//Check if Parent ID is valid
		//Tenants do not have Parents
		if entity == DEVICE {
			objID, err = primitive.ObjectIDFromHex(t["parentId"].(string))
			if err != nil {
				return u.Message(false, "ParentID is not valid"), false
			}

			ctx, cancel := u.Connect()
			if GetDB().Collection("rack").FindOne(ctx,
				bson.M{"_id": objID}).Err() != nil &&
				GetDB().Collection("device").FindOne(ctx,
					bson.M{"_id": objID}).Err() != nil {
				return u.Message(false, "ParentID should be correspond to Existing ID"), false
			}
			defer cancel()

		} else if entity > TENANT && entity <= DEVICESENSOR {
			_, ok := t["parentId"].(string)
			if !ok {
				return u.Message(false, "ParentID is not valid"), false
			}
			objID, err = primitive.ObjectIDFromHex(t["parentId"].(string))
			if err != nil {
				return u.Message(false, "ParentID is not valid"), false
			}
			parent := u.EntityToString(u.GetParentOfEntityByInt(entity))

			ctx, cancel := u.Connect()
			if GetDB().Collection(parent).
				FindOne(ctx, bson.M{"_id": objID}).Err() != nil {
				println("ENTITY VALUE: ", entity)
				println("We got Parent: ", parent, " with ID:", t["parentId"].(string))
				return u.Message(false, "ParentID should correspond to Existing ID"), false

			}
			defer cancel()
		}

		if entity < CABINET || entity > DEVICESENSOR {
			if _, ok := t["attributes"]; !ok {
				return u.Message(false, "Attributes should be on the payload"), false
			} else {
				if v, ok := t["attributes"].(map[string]interface{}); !ok {
					return u.Message(false, "Attributes should be on the payload"), false
				} else {
					switch entity {
					case TENANT:
						if _, ok := v["color"]; !ok {
							return u.Message(false,
								"Color Attribute must be specified on the payload"), false
						}

					case SITE:
						switch v["orientation"] {
						case "EN", "NW", "WS", "SE":
						case "", nil:
							return u.Message(false, "Orientation should be on the payload"), false

						default:
							return u.Message(false, "Orientation is invalid!"), false
						}

						if v["usableColor"] == "" || v["usableColor"] == nil {
							return u.Message(false, "Usable Color should be on the payload"), false
						}

						if v["reservedColor"] == "" || v["reservedColor"] == nil {
							return u.Message(false, "Reserved Color should be on the payload"), false
						}

						if v["technicalColor"] == "" || v["technicalColor"] == nil {
							return u.Message(false, "Technical Color should be on the payload"), false
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
						case "-E-N", "-E+N", "+E-N", "+E+N":
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

			if _, ok := t["category"]; !ok {
				return u.Message(false,
					"Category should be on payload"), false
			}

			if _, ok := t["sizeWDHmm"]; !ok {
				return u.Message(false,
					"Size,Width,Depth (mm) Array should be on payload"), false
			}

			if _, ok := t["fbxModel"]; !ok {
				return u.Message(false,
					"fbxModel should be on payload"), false
			}

			if _, ok := t["attributes"]; !ok {
				return u.Message(false,
					"Attributes should be on payload"), false
			}

			if _, ok := t["slots"]; !ok {
				return u.Message(false,
					"fbxModel should be on payload"), false
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

		switch t["type"] {
		case "rack", "device":
		default:
			return u.Message(false, "Group type (rack or device) should be specified"), false
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
	entStr := u.EntityToString(entity)
	res, e := GetDB().Collection(entStr).InsertOne(ctx, t)
	if e != nil {
		return u.Message(false,
				"Internal error while creating "+entStr+": "+e.Error()),
			e.Error()
	}
	defer cancel()

	//Remove _id
	t["id"] = res.InsertedID
	//t = fixID(t)

	resp := u.Message(true, "success")
	resp["data"] = t
	return resp, ""
}

func GetEntity(entityID primitive.ObjectID, ent string) (map[string]interface{}, string) {
	t := map[string]interface{}{}

	ctx, cancel := u.Connect()
	e := GetDB().Collection(ent).FindOne(ctx, bson.M{"_id": entityID}).Decode(&t)
	if e != nil {
		return nil, e.Error()
	}
	defer cancel()
	//Remove _id
	t = fixID(t)
	return t, ""
}

//Only useful for tenant since tenants are unique in the DB
func GetEntityByName(name, ent string) (map[string]interface{}, string) {
	t := map[string]interface{}{}

	ctx, cancel := u.Connect()
	e := GetDB().Collection(ent).FindOne(ctx, bson.M{"name": name}).Decode(&t)
	if e != nil {
		return nil, e.Error()
	}
	defer cancel()
	//Remove _id
	t = fixID(t)
	return t, ""
}

//Only useful for tenant since tenants are unique in the DB
func GetEntityBySlug(name, ent string) (map[string]interface{}, string) {
	t := map[string]interface{}{}

	ctx, cancel := u.Connect()
	e := GetDB().Collection(ent).FindOne(ctx, bson.M{"slug": name}).Decode(&t)
	if e != nil {
		return nil, e.Error()
	}
	defer cancel()
	//Remove _id
	t = fixID(t)
	return t, ""
}

func GetAllEntities(ent string) ([]map[string]interface{}, string) {
	data := make([]map[string]interface{}, 0)
	ctx, cancel := u.Connect()
	c, err := GetDB().Collection(ent).Find(ctx, bson.D{{}})
	if err != nil {
		fmt.Println(err)
		return nil, err.Error()
	}
	defer cancel()

	for c.Next(ctx) {
		x := map[string]interface{}{}
		e := c.Decode(x)
		if e != nil {
			fmt.Println(err)
			return nil, err.Error()
		}
		//Remove _id
		x = fixID(x)
		data = append(data, x)
	}

	return data, ""
}

func DeleteEntity(entity string, id primitive.ObjectID) (map[string]interface{}, string) {
	var t map[string]interface{}
	var e string
	eNum := u.EntityStrToInt(entity)
	if eNum > DEVICE && eNum < ROOMTMPL {
		//Delete the non hierarchal objects
		t, e = GetEntityHierarchy(entity, id, eNum, eNum+1)

	} else {
		t, e = GetEntityHierarchy(entity, id, eNum, AC)
	}

	if e != "" {
		return u.Message(false,
			"There was an error in deleting the entity: "+e), "not found"
	}

	return deleteHelper(t, eNum)
}

func deleteHelper(t map[string]interface{}, ent int) (map[string]interface{}, string) {
	if t != nil {

		if v, ok := t["children"]; ok {
			if x, ok := v.([]map[string]interface{}); ok {
				for i := range x {
					deleteHelper(x[i], ent+1)
				}
			} else {
				println("JSON not formatted as expected")
				return u.Message(false,
					"There was an error in deleting the entity"), "not found"
			}
		}

		println("So we got: ", ent)

		if ent == RACK {
			ctx, cancel := u.Connect()
			GetDB().Collection("rack_sensor").DeleteMany(ctx,
				bson.M{"parentId": t["id"].(primitive.ObjectID).Hex()})
			defer cancel()
		}

		//Delete associated non hierarchal objs
		if ent == ROOM {
			//ITER Through all nonhierarchal objs
			ctx, cancel := u.Connect()
			for i := AC; i < RACKSENSOR; i++ {
				ent := u.EntityToString(i)
				GetDB().Collection(ent).DeleteMany(ctx, bson.M{"parentId": t["id"].(primitive.ObjectID).Hex()})
			}
			defer cancel()
		}

		if ent == DEVICE {
			DeleteDeviceF(t["id"].(primitive.ObjectID))
		} else {
			ctx, cancel := u.Connect()
			entity := u.EntityToString(ent)
			c, _ := GetDB().Collection(entity).DeleteOne(ctx, bson.M{"_id": t["id"].(primitive.ObjectID)})
			if c.DeletedCount == 0 {
				return u.Message(false, "There was an error in deleting the entity"), "not found"
			}
			defer cancel()

		}
	}
	return nil, ""
}

func UpdateEntity(ent string, id primitive.ObjectID, t *map[string]interface{}, isPatch bool) (map[string]interface{}, string) {
	var e *mongo.SingleResult
	updatedDoc := bson.M{}
	retDoc := options.ReturnDocument(options.After)

	ctx, cancel := u.Connect()
	if isPatch == true {
		e = GetDB().Collection(ent).FindOneAndUpdate(ctx,
			bson.M{"_id": id}, bson.M{"$set": *t},
			&options.FindOneAndUpdateOptions{ReturnDocument: &retDoc})
		if e.Err() != nil {
			return u.Message(false, "failure: "+e.Err().Error()), e.Err().Error()
		}
	} else {

		e = GetDB().Collection(ent).FindOneAndReplace(ctx,
			bson.M{"_id": id}, *t,
			&options.FindOneAndReplaceOptions{ReturnDocument: &retDoc})

		if e.Err() != nil {
			return u.Message(false, "failure: "+e.Err().Error()), e.Err().Error()
		}
	}

	//Obtain new document then
	//Fix the _id / id discrepancy
	e.Decode(&updatedDoc)
	updatedDoc = fixID(updatedDoc)

	defer cancel()
	resp := u.Message(true, "success")
	resp["data"] = updatedDoc
	return resp, ""
}

func GetEntityByQuery(ent string, query bson.M) ([]map[string]interface{}, string) {
	results := make([]map[string]interface{}, 0)
	ctx, cancel := u.Connect()
	c, err := GetDB().Collection(ent).Find(ctx, query)
	if err != nil {
		fmt.Println(err)
		return nil, err.Error()
	}
	defer cancel()

	for c.Next(ctx) {
		x := map[string]interface{}{}
		e := c.Decode(x)
		if e != nil {
			fmt.Println(err)
			return nil, err.Error()
		}
		//Remove _id
		x = fixID(x)
		results = append(results, x)
	}

	return results, ""
}

//Gets children of an entity
//Example: /api/buildings/{id}/rooms
//will return all rooms associated with
//the BldgID
//Be sure to pass the Child Entity and NOT Parent Entity
func GetEntitiesOfParent(ent, id string) ([]map[string]interface{}, string) {
	var c *mongo.Cursor
	var err error
	enfants := make([]map[string]interface{}, 0)
	ctx, cancel := u.Connect()
	c, err = GetDB().Collection(ent).Find(ctx, bson.M{"parentId": id})
	if err != nil {
		fmt.Println(err)
		return nil, err.Error()
	}
	defer cancel()

	for c.Next(ctx) {
		s := map[string]interface{}{}
		e := c.Decode(&s)
		if e != nil {
			fmt.Println(err)
			return nil, err.Error()
		}
		//Remove _id
		s = fixID(s)
		enfants = append(enfants, s)
	}

	//println("The length of children is: ", len(enfants))

	return enfants, ""
}

func GetEntityHierarchy(entity string, ID primitive.ObjectID, entnum, end int) (map[string]interface{}, string) {

	//Check if at the end of the hierarchy
	if entnum != end {

		//Get the top entity
		top, e := GetEntity(ID, entity)
		if e != "" {
			return nil, e
		}

		//Remove _id
		top = fixID(top)

		//Retrieve associated nonhierarchal objects
		if entity == "device" {
			x, e := GetEntitiesOfParent("device_sensor",
				top["id"].(primitive.ObjectID).Hex())
			if e == "" {
				top["device_sensors"] = x
			}
		}

		if entity == "rack" {
			x, e := GetEntitiesOfParent("rack_sensor",
				top["id"].(primitive.ObjectID).Hex())
			if e == "" {
				top["rack_sensors"] = x
			}
		}

		if entity == "room" {
			//ITER Through all nonhierarchal objs
			for i := AC; i < RACKSENSOR; i++ {
				ent := u.EntityToString(i)
				x, e := GetEntitiesOfParent(ent, top["id"].(primitive.ObjectID).Hex())
				if e == "" {
					top[ent+"s"] = x
				}
			}
		}

		subEnt := u.EntityToString(entnum + 1)

		//Get immediate children
		children, e1 := GetEntitiesOfParent(subEnt, ID.Hex())
		if e1 != "" {
			println("Are we here")
			println("SUBENT: ", subEnt)
			println("PID: ", ID.Hex())
			return nil, e1
		}

		top["children"] = children

		//Get the rest of hierarchy for children
		for i := range children {
			var x map[string]interface{}
			subIdx := u.EntityToString(entnum + 1)
			subID := (children[i]["id"].(primitive.ObjectID))
			x, _ =
				GetEntityHierarchy(subIdx, subID, entnum+1, end)

			//So that output JSON will not have
			// "children": [null]
			if x != nil {
				children[i] = x
			}
		}

		return top, ""
	}
	return nil, ""
}

func GetEntityByNameAndParentID(ent, id, name string) (map[string]interface{}, string) {
	t := map[string]interface{}{}

	ctx, cancel := u.Connect()
	e := GetDB().Collection(ent).FindOne(ctx, bson.M{"name": name, "parentId": id}).Decode(&t)
	if e != nil {
		return nil, e.Error()
	}
	defer cancel()
	//Remove _id
	t = fixID(t)
	return t, ""
}

func GetEntitiesUsingAncestorNames(ent string, id primitive.ObjectID, ancestry []map[string]string) ([]map[string]interface{}, string) {
	top, e := GetEntity(id, ent)
	if e != "" {
		return nil, e
	}

	//Remove _id
	top = fixID(top)

	pid := (top["id"].(primitive.ObjectID)).Hex()

	var x map[string]interface{}
	var e1 string
	for i := range ancestry {
		for k, v := range ancestry[i] {

			println("KEY:", k, " VAL:", v)

			if v == "all" {
				println("K:", k)
				//println("ID", x["_id"].(primitive.ObjectID).String())
				/*if k == "device" {
					return GetDeviceFByParentID(pid) nil, ""
				}*/
				return GetEntitiesOfParent(k, pid)
			}

			x, e1 = GetEntityByNameAndParentID(k, pid, v)
			if e1 != "" {
				println("Failing here")
				return nil, ""
			}
			pid = (x["id"].(primitive.ObjectID)).Hex()
		}
	}

	return nil, ""
}

func GetEntityUsingAncestorNames(ent string, id primitive.ObjectID, ancestry []map[string]string) (map[string]interface{}, string) {
	top, e := GetEntity(id, ent)
	if e != "" {
		return nil, e
	}

	//Remove _id
	top = fixID(top)

	pid := (top["id"].(primitive.ObjectID)).Hex()

	var x map[string]interface{}
	var e1 string
	for i := range ancestry {
		for k, v := range ancestry[i] {

			println("KEY:", k, " VAL:", v)

			x, e1 = GetEntityByNameAndParentID(k, pid, v)
			if e1 != "" {
				println("Failing here")
				return nil, ""
			}
			pid = (x["id"].(primitive.ObjectID)).Hex()
		}
	}

	//Remove _id
	x = fixID(x)
	return x, ""
}

func GetTenantHierarchy(entity, name string, entnum, end int) (map[string]interface{}, string) {

	t, e := GetEntityByName(name, "tenant")
	if e != "" {
		fmt.Println(e)
		return nil, e
	}

	//Remove _id
	t = fixID(t)

	subEnt := u.EntityToString(entnum + 1)
	tid := t["id"].(primitive.ObjectID).Hex()

	//Get immediate children
	children, e1 := GetEntitiesOfParent(subEnt, tid)
	if e1 != "" {
		println("Are we here")
		println("SUBENT: ", subEnt)
		println("PID: ", tid)
		return nil, e1
	}
	t["children"] = children

	//Get the rest of hierarchy for children
	for i := range children {
		var x map[string]interface{}
		subIdx := u.EntityToString(entnum + 1)
		subID := (children[i]["id"].(primitive.ObjectID))
		x, _ =
			GetEntityHierarchy(subIdx, subID, entnum+1, end)
		if x != nil {
			children[i] = x
		}
	}

	return t, ""

}

func GetEntitiesUsingTenantAsAncestor(ent, id string, ancestry []map[string]string) ([]map[string]interface{}, string) {
	top, e := GetEntityByName(id, ent)
	if e != "" {
		return nil, e
	}

	//Remove _id
	top = fixID(top)

	pid := (top["id"].(primitive.ObjectID)).Hex()

	var x map[string]interface{}
	var e1 string
	println("ANCS-LEN: ", len(ancestry))
	for i := range ancestry {
		for k, v := range ancestry[i] {

			println("KEY:", k, " VAL:", v)

			if v == "all" {
				println("K:", k)
				return GetEntitiesOfParent(k, pid)
			}

			x, e1 = GetEntityByNameAndParentID(k, pid, v)
			if e1 != "" {
				println("Failing here")
				println("E1: ", e1)
				return nil, ""
			}
			pid = (x["id"].(primitive.ObjectID)).Hex()
		}
	}

	return nil, ""
}

func GetEntityUsingTenantAsAncestor(ent, id string, ancestry []map[string]string) (map[string]interface{}, string) {
	top, e := GetEntityByName(id, ent)
	if e != "" {
		return nil, e
	}

	pid := (top["id"].(primitive.ObjectID)).Hex()

	var x map[string]interface{}
	var e1 string
	for i := range ancestry {
		for k, v := range ancestry[i] {

			println("KEY:", k, " VAL:", v)

			x, e1 = GetEntityByNameAndParentID(k, pid, v)
			if e1 != "" {
				println("Failing here")
				return nil, ""
			}
			pid = (x["id"].(primitive.ObjectID)).Hex()
		}
	}

	return x, ""
}

func GetEntitiesOfAncestor(id interface{}, ent int, entStr, wantedEnt string) ([]map[string]interface{}, string) {
	var ans []map[string]interface{}
	var t map[string]interface{}
	var e, e1 string
	if ent == TENANT {

		t, e = GetEntityByName(id.(string), "tenant")
		if e != "" {
			return nil, e
		}

	} else {
		ID, _ := primitive.ObjectIDFromHex(id.(string))
		t, e = GetEntity(ID, entStr)
		if e != "" {
			return nil, e
		}
	}

	sub, e1 := GetEntitiesOfParent(u.EntityToString(ent+1),
		t["id"].(primitive.ObjectID).Hex())
	if e1 != "" {
		return nil, e1
	}

	if wantedEnt == "" {
		wantedEnt = u.EntityToString(ent + 2)
	}

	for i := range sub {
		x, _ := GetEntitiesOfParent(wantedEnt, sub[i]["id"].(primitive.ObjectID).Hex())
		ans = append(ans, x...)
	}
	return ans, ""
}

//ent string, ID primitive.ObjectID, nestID string
func GetNestedEntity(ID primitive.ObjectID, ent, nestID string) (map[string]interface{}, string) {
	t := map[string]interface{}{}

	ctx, cancel := u.Connect()
	parent := u.EntityToString(u.GetParentOfEntityByInt(u.EntityStrToInt(ent)))
	criteria := bson.M{"_id": ID, ent + "s.id": nestID}
	e := GetDB().Collection(parent).FindOne(ctx, criteria).Decode(&t)
	if e != nil {
		return nil, e.Error()
	}

	//Because applying filters to the Mongo Request is a hassle
	//for now
	for _, entry := range t[ent+"s"].(primitive.A) {
		if entry.(map[string]interface{})["id"] == nestID {
			t = entry.(map[string]interface{})
			break
		}
	}
	defer cancel()
	return t, ""
}

func CreateNestedEntity(entity int, eStr string, t map[string]interface{}) (map[string]interface{}, string) {
	var check map[string]interface{}
	if resp, ok := ValidateEntity(entity, t); !ok {
		return resp, "validate"
	}

	ctx, cancel := u.Connect()

	parent := u.EntityToString(u.GetParentOfEntityByInt(entity))
	pid, _ := primitive.ObjectIDFromHex(t["parentId"].(string))

	//CHECK FOR DUPLICATE SECTION
	e1 := GetDB().Collection(parent).FindOne(ctx, bson.M{"_id": pid}).Decode(&check)
	if e1 != nil {
		return u.Message(false,
				"Internal error while creating "+eStr+": "+e1.Error()),
			e1.Error()
	}

	if v, ok := check[eStr+"s"].(primitive.A); ok {
		for i := range v {
			if elt, ok := v[i].(map[string]interface{}); ok {
				if elt["name"].(string) == t["name"] { //DUPLICATE FOUND
					return u.Message(false,
							"Error: Cannot create duplicate object. Please use a different name"),
						"duplicate"
				}
			}
		}
	}
	//CHECK FOR DUPLICATE SECTION FINISH

	_, e := GetDB().Collection(parent).UpdateOne(ctx, bson.M{"_id": pid}, bson.M{"$addToSet": bson.M{eStr + "s": t}})
	if e != nil {
		return u.Message(false,
				"Internal error while creating "+eStr+": "+e.Error()),
			e.Error()
	}
	defer cancel()

	resp := u.Message(true, "success")
	resp["data"] = t
	return resp, ""
}

func GetAllNestedEntities(ID primitive.ObjectID, ent string) ([]map[string]interface{}, string) {
	t := map[string]interface{}{}
	ans := []map[string]interface{}{}
	data := []interface{}{}

	ctx, cancel := u.Connect()
	parent := u.EntityToString(u.GetParentOfEntityByInt(u.EntityStrToInt(ent)))
	e := GetDB().Collection(parent).FindOne(ctx, bson.M{"_id": ID}).Decode(&t)
	if e != nil {
		return nil, e.Error()
	}

	//Because applying filters to the Mongo Request is a hassle
	//for now
	if v, ok := t[ent+"s"].(primitive.A); ok {
		data = v
	}

	for i := range data {
		if x, ok := data[i].(map[string]interface{}); ok {
			ans = append(ans, x)
		}
	}

	println("LENANS: ", len(ans))
	defer cancel()
	return ans, ""
}

func DeleteNestedEntity(ent string, ID primitive.ObjectID, nestID string) (map[string]interface{}, string) {
	t := map[string]interface{}{}
	newSubEnts := []interface{}{}

	ctx, cancel := u.Connect()
	parent := u.EntityToString(u.GetParentOfEntityByInt(u.EntityStrToInt(ent)))
	criteria := bson.M{"_id": ID, ent + "s.id": nestID}
	e := GetDB().Collection(parent).FindOne(ctx, criteria).Decode(&t)
	if e != nil {
		return u.Message(false,
			"There was an error in deleting the entity: "+e.Error()), "parent not found"
	}
	defer cancel()

	if v, ok := t[ent+"s"].(primitive.A); ok {
		for i := range v {
			if v[i].(map[string]interface{})["id"] != nestID {
				newSubEnts = append(newSubEnts, v[i])
			}
		}
	}

	t[ent+"s"] = newSubEnts

	c1, cancel2 := u.Connect()
	_, e1 := GetDB().Collection(parent).UpdateOne(c1, criteria, bson.M{"$set": t})
	if e1 != nil {
		return u.Message(false,
			"There was an error in deleting the entity2: "+e.Error()), "unable update"
	}
	defer cancel2()

	return t, ""
}

func UpdateNestedEntity(ent string, ID primitive.ObjectID,
	nestID string, t map[string]interface{}, isPatch bool) (map[string]interface{}, string) {
	foundParent := map[string]interface{}{}
	updateDoc := map[string]interface{}{}
	var idx int

	//OG WAY///////////////////////////////////////
	ctx, cancel := u.Connect()
	parent := u.EntityToString(u.GetParentOfEntityByInt(u.EntityStrToInt(ent)))
	criteria := bson.M{"_id": ID, ent + "s.id": nestID}
	e := GetDB().Collection(parent).FindOne(ctx, criteria).Decode(&foundParent)
	if e != nil {
		return u.Message(false,
			"There was an error in updating the entity: "+e.Error()), "parent not found"
	}
	defer cancel()
	delete(t, "id")

	//Once parent is found, search the nested array
	//for matching NestID and change the attributes
	if v, ok := foundParent[ent+"s"].(primitive.A); ok {
		for i := range v {
			if v[i].(map[string]interface{})["id"] == nestID {
				old := v[i].(map[string]interface{})
				idx = i

				//Ensure the ID & PID are
				//preserved
				t["id"] = nestID
				t["parentId"] = old["parentId"]

				if isPatch == true {
					for key := range t {
						old[key] = t[key]
					}
				} else {
					foundParent[ent+"s"].(primitive.A)[i] = t
				}

				updateDoc = old
				break
			}
		}
	}

	retDoc := options.After

	c1, cancel2 := u.Connect()
	e1 := GetDB().Collection(parent).FindOneAndUpdate(c1,
		criteria, bson.M{"$set": foundParent},
		&options.FindOneAndUpdateOptions{ReturnDocument: &retDoc})
	if e1.Err() != nil {
		return u.Message(false,
			"There was an error in deleting the entity2: "+e.Error()), "unable update"
	}
	defer cancel2()
	e1.Decode(&updateDoc)

	resp := u.Message(true, "success")
	resp["data"] = updateDoc[ent+"s"].(primitive.A)[idx]
	return resp, ""
}

func GetNestedEntityByQuery(parent, entity string, query bson.M) ([]map[string]interface{}, string) {
	ans := make([]map[string]interface{}, 0)
	parents, e := GetAllEntities(parent)
	if e != "" {
		return nil, e
	}

	//Now get all subentities from parents
	for i := range parents {
		pid := parents[i]["id"].(primitive.ObjectID)
		nestedEnts, e1 := GetAllNestedEntities(pid, entity)
		if e1 != "" {
			return nil, e1
		}

		//Iterate over the nestedEntities to see if they match the query
		for k := range nestedEnts {

			nestedEntity := nestedEnts[k]

			match := true
			//Check if the ent matches
			for q := range query {
				if nestedEntity[q] != query[q] {
					match = false
					break
				}
			}

			//The entity matches
			if match == true {
				ans = append(ans, nestedEntity)
			}

		}
	}
	return ans, ""
}

func DeleteEntityBySlug(entity, id string) (map[string]interface{}, string) {
	//Finally delete the Entity
	ctx, cancel := u.Connect()
	c, _ := GetDB().Collection(entity).DeleteOne(ctx, bson.M{"slug": id})
	if c.DeletedCount == 0 {
		return u.Message(false, "There was an error in deleting the entity"), "not found"
	}
	defer cancel()

	return u.Message(true, "success"), ""
}

func UpdateEntityBySlug(ent, id string, t *map[string]interface{},
	isPatch bool) (map[string]interface{}, string) {
	var e *mongo.SingleResult
	updatedDoc := map[string]interface{}{}
	retDoc := options.After

	ctx, cancel := u.Connect()
	if isPatch == true {
		e = GetDB().Collection(ent).FindOneAndUpdate(ctx,
			bson.M{"slug": id}, bson.M{"$set": *t},
			&options.FindOneAndUpdateOptions{ReturnDocument: &retDoc})
		if e.Err() != nil {
			return u.Message(false, "failure: "+e.Err().Error()), e.Err().Error()
		}
	} else {
		//Preserve the slug if not
		//provided
		if _, ok := (*t)["slug"]; !ok {
			(*t)["slug"] = id
		}
		e = GetDB().Collection(ent).FindOneAndReplace(ctx,
			bson.M{"slug": id}, *t,
			&options.FindOneAndReplaceOptions{ReturnDocument: &retDoc})
		if e.Err() != nil {
			return u.Message(false, "failure: "+e.Err().Error()), e.Err().Error()
		}
	}

	//Retrieve the result
	e.Decode(&updatedDoc)
	updatedDoc = fixID(updatedDoc)

	defer cancel()
	resp := u.Message(true, "success")
	resp["data"] = updatedDoc
	return resp, ""
}

//DEV FAMILY FUNCS

func GetDevEntitiesOfParent(ent, id string) ([]map[string]interface{}, string) {
	var c *mongo.Cursor
	var err error
	enfants := make([]map[string]interface{}, 0)
	ctx, cancel := u.Connect()
	c, err = GetDB().Collection(ent).Find(ctx, bson.M{"parentId": id})
	if err != nil {
		fmt.Println(err)
		return nil, err.Error()
	}
	defer cancel()

	for c.Next(ctx) {
		s := map[string]interface{}{}
		e := c.Decode(&s)
		if e != nil {
			fmt.Println(err)
			return nil, err.Error()
		}
		enfants = append(enfants, s)
	}

	//println("The length of children is: ", len(enfants))

	return enfants, ""
}

func RetrieveDeviceHierarch(ID primitive.ObjectID, start, end int) (map[string]interface{}, string) {
	if start < end {
		//Get the top entity
		top, e := GetEntity(ID, "device")
		if e != "" {
			return nil, e
		}

		//Retrieve sensors
		ctx, cancel := u.Connect()
		x, err := GetDB().Collection("device_sensor").Find(ctx,
			bson.M{"parentId": top["id"].(primitive.ObjectID).Hex()})
		if err == nil {
			data := []map[string]interface{}{}
			for x.Next(ctx) {
				v := map[string]interface{}{}
				e := x.Decode(v)
				if e != nil {
					fmt.Println(err)
					return nil, err.Error()
				}
				//Remove _id
				v = fixID(v)
				data = append(data, v)
			}
			top["device_sensors"] = data
		}
		defer cancel()

		children, e1 := GetEntitiesOfParent("device", ID.Hex())
		if e1 != "" {
			return top, ""
		}

		for i := range children {
			children[i], _ = RetrieveDeviceHierarch(
				children[i]["id"].(primitive.ObjectID), start+1, end)
		}

		top["children"] = children

		return top, ""
	}
	return nil, ""
}

func DeleteDeviceF(entityID primitive.ObjectID) (map[string]interface{}, string) {
	//var deviceType string

	t, e := RetrieveDeviceHierarch(entityID, 0, 999)
	if e != "" {
		return u.Message(false,
			"There was an error in deleting the entity"), "not found"
	}

	return deleteDeviceHelper(t)
}

func deleteDeviceHelper(t map[string]interface{}) (map[string]interface{}, string) {
	println("entered ddH")
	if t != nil {

		if v, ok := t["children"]; ok {
			if x, ok := v.([]map[string]interface{}); ok {
				for i := range x {
					deleteDeviceHelper(x[i])
				}
			} else {
				println("JSON not formatted as expected")
				return u.Message(false,
					"There was an error in deleting the entity"), "not found"
			}
		}

		ctx, cancel := u.Connect()
		//Delete relevant non hierarchal objects
		GetDB().Collection("device_sensor").DeleteMany(ctx,
			bson.M{"parentId": t["id"].(primitive.ObjectID).Hex()})

		c, _ := GetDB().Collection("device").DeleteOne(ctx, bson.M{"_id": t["id"].(primitive.ObjectID)})
		if c.DeletedCount == 0 {
			return u.Message(false, "There was an error in deleting the entity"), "not found"
		}
		defer cancel()

	}
	return nil, ""
}

func ExtractCursor(c *mongo.Cursor, ctx context.Context) ([]map[string]interface{}, string) {
	var ans []map[string]interface{}
	for c.Next(ctx) {
		x := map[string]interface{}{}
		err := c.Decode(x)
		if err != nil {
			fmt.Println(err)
			return nil, err.Error()
		}
		ans = append(ans, x)
	}
	return ans, ""
}
