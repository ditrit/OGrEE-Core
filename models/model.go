package models

import (
	"context"
	"crypto/rand"
	"fmt"
	u "p3/utils"

	"go.mongodb.org/mongo-driver/mongo"

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
	SUBDEV
	SUBDEV1
	AC
	PWRPNL
	WALL
	ROOMTMPL
	OBJTMPL
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

func delSubEnts(eNum int, data map[string][]map[string]interface{}) (map[string]interface{}, string) {
	//Delete the Subentities
	for i := SUBDEV1; i > eNum; i-- {
		eStr := u.EntityToString(i)
		if arr, ok := data[eStr+"s"]; ok {

			for idx := range arr {
				println("LEN: ", len(arr[idx]))

				if len(arr[idx]) > 0 {

					locID := arr[idx]["_id"].(primitive.ObjectID)
					ctx, cancel := u.Connect()
					println("Now deleting: ", eStr)
					c, _ := GetDB().Collection(eStr).DeleteOne(ctx, bson.M{"_id": locID})
					if c.DeletedCount == 0 {
						return u.Message(false, "There was an error in deleting the entity"), "not found"
					}
					defer cancel()
				}

			}
		}
	}
	return u.Message(true, "success"), ""
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

func CreateParentPlaceHolder(ent, pid string) (map[string]interface{}, string) {
	name := "Device-" + genID(4)
	category := "device"
	domain := "placeholder"
	t := map[string]interface{}{"name": name,
		"category": category, "domain": domain, "parentId": pid}
	x, e := CreateEntity(DEVICE, t)

	if e == "" {
		switch ent {
		case "subdevice":
			return x, ""
		case "subdevice1":
			parent := x["parentId"]
			category = "subdevice"
			subName := "Subdevice-" + genID(4)
			t1 := map[string]interface{}{"name": subName,
				"category": category, "domain": domain, "parentId": parent}
			y, e1 := CreateEntity(SUBDEV, t1)
			if e1 != "" {
				return nil, e1
			}
			x["subdevice"] = y
			return x, ""
		}
	}
	return nil, "error"
}

func ValidateEntity(entity int, t map[string]interface{}) (map[string]interface{}, bool) {
	var objID primitive.ObjectID
	var err error
	switch entity {
	case TENANT, SITE, BLDG, ROOM, RACK, DEVICE, SUBDEV,
		SUBDEV1, AC, PWRPNL, WALL:
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

		} else if entity > TENANT && entity <= WALL {
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
				return u.Message(false, "ParentID should be correspond to Existing ID"), false

			}
			defer cancel()
		}

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
					case "":
						return u.Message(false, "Orientation should be on the payload"), false

					default:
						return u.Message(false, "Orientation is invalid!"), false
					}

					if v["usableColor"] == "" {
						return u.Message(false, "Usable Color should be on the payload"), false
					}

					if v["reservedColor"] == "" {
						return u.Message(false, "Reserved Color should be on the payload"), false
					}

					if v["technicalColor"] == "" {
						return u.Message(false, "Technical Color should be on the payload"), false
					}

				case BLDG:
					if v["posXY"] == "" {
						return u.Message(false, "XY coordinates should be on payload"), false
					}

					if v["posXYU"] == "" {
						return u.Message(false, "PositionXYU string should be on the payload"), false
					}

					if v["size"] == "" {
						return u.Message(false, "Invalid building size on the payload"), false
					}

					if v["sizeU"] == "" {
						return u.Message(false, "Building size string should be on the payload"), false
					}

					if v["height"] == "" {
						return u.Message(false, "Invalid Height on payload"), false
					}

					if v["heightU"] == "" {
						return u.Message(false, "Building Height string should be on the payload"), false
					}

				case ROOM:
					if v["posXY"] == "" {
						return u.Message(false, "XY coordinates should be on payload"), false
					}

					if v["posXYU"] == "" {
						return u.Message(false, "PositionXYU string should be on the payload"), false
					}

					switch v["orientation"] {
					case "-E-N", "-E+N", "+E-N", "+E+N":
					case "-N-W", "-N+W", "+N-W", "+N+W":
					case "-W-S", "-W+S", "+W-S", "+W+S":
					case "-S-E", "-S+E", "+S-E", "+S+E":
					case "":
						return u.Message(false, "Orientation should be on the payload"), false

					default:
						return u.Message(false, "Orientation is invalid!"), false
					}

					if v["size"] == "" {
						return u.Message(false, "Invalid size on the payload"), false
					}

					if v["sizeU"] == "" {
						return u.Message(false, "Room size string should be on the payload"), false
					}

					if v["height"] == "" {
						return u.Message(false, "Invalid Height on payload"), false
					}

					if v["heightU"] == "" {
						return u.Message(false, "Room Height string should be on the payload"), false
					}
				case RACK:
					if v["posXY"] == "" {
						return u.Message(false, "XY coordinates should be on payload"), false
					}

					if v["posXYU"] == "" {
						return u.Message(false, "PositionXYU string should be on the payload"), false
					}

					switch v["orientation"] {
					case "front", "rear", "left", "right":
					case "":
						return u.Message(false, "Orientation should be on the payload"), false

					default:
						return u.Message(false, "Orientation is invalid!"), false
					}

					if v["size"] == "" {
						return u.Message(false, "Invalid size on the payload"), false
					}

					if v["sizeU"] == "" {
						return u.Message(false, "Rack size string should be on the payload"), false
					}

					if v["height"] == "" {
						return u.Message(false, "Invalid Height on payload"), false
					}

					if v["heightU"] == "" {
						return u.Message(false, "Rack Height string should be on the payload"), false
					}
				case DEVICE:
					switch v["orientation"] {
					case "front", "rear", "frontflipped", "rearflipped":
					case "":
						return u.Message(false, "Orientation should be on the payload"), false

					default:
						return u.Message(false, "Orientation is invalid!"), false
					}

					if v["size"] == "" {
						return u.Message(false, "Invalid size on the payload"), false
					}

					if v["sizeUnit"] == "" {
						return u.Message(false, "Rack size string should be on the payload"), false
					}

					if v["height"] == "" {
						return u.Message(false, "Invalid Height on payload"), false
					}

					if v["heightU"] == "" {
						return u.Message(false, "Rack Height string should be on the payload"), false
					}
				case SUBDEV, SUBDEV1:

					switch v["orientation"] {
					case "front", "rear", "frontflipped", "rearflipped":
					case "":
						return u.Message(false, "Orientation should be on the payload"), false

					default:
						return u.Message(false, "Orientation is invalid!"), false
					}

					if v["size"] == "" {
						return u.Message(false, "Invalid size on the payload"), false
					}

					if v["sizeUnit"] == "" {
						return u.Message(false, "Subdevice size string should be on the payload"), false
					}

					if v["height"] == "" {
						return u.Message(false, "Invalid Height on payload"), false
					}

					if v["heightU"] == "" {
						return u.Message(false, "Subdevice Height string should be on the payload"), false
					}
				}
			}
		}
	case ROOMTMPL, OBJTMPL:
		if t["slug"] == "" {
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

	t["id"] = res.InsertedID

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

	for c.Next(GetCtx()) {
		x := map[string]interface{}{}
		e := c.Decode(x)
		if e != nil {
			fmt.Println(err)
			return nil, err.Error()
		}
		data = append(data, x)
	}

	return data, ""
}

func DeleteEntity(entity string, id primitive.ObjectID) (map[string]interface{}, string) {
	eNum := u.EntityStrToInt(entity)
	t, e := GetEntityHierarchy(entity, id, eNum, SUBDEV1)
	if e != "" {
		return u.Message(false,
			"There was an error in deleting the entity: "+e), "not found"
	}

	data := parseDataForNonStdResult(entity, eNum, AC, t)

	//Delete the Subentities
	for i := SUBDEV1; i > eNum; i-- {
		eStr := u.EntityToString(i)
		if arr, ok := data[eStr+"s"]; ok {

			for idx := range arr {
				println("LEN: ", len(arr[idx]))

				if len(arr[idx]) > 0 {

					locID := arr[idx]["_id"].(primitive.ObjectID)
					ctx, cancel := u.Connect()
					c, _ := GetDB().Collection(eStr).DeleteOne(ctx, bson.M{"_id": locID})
					if c.DeletedCount == 0 {
						return u.Message(false, "There was an error in deleting the entity"), "not found"
					}
					defer cancel()
				}

			}
		}
	}
	//Finally delete the Entity
	ctx, cancel := u.Connect()
	c, _ := GetDB().Collection(entity).DeleteOne(ctx, bson.M{"_id": id})
	if c.DeletedCount == 0 {
		return u.Message(false, "There was an error in deleting the entity"), "not found"
	}
	defer cancel()

	return u.Message(true, "success"), ""
}

func UpdateEntity(ent string, id primitive.ObjectID, t *map[string]interface{}) (map[string]interface{}, string) {
	ctx, cancel := u.Connect()
	e := GetDB().Collection(ent).FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$set": *t}).Err()
	if e != nil {
		return u.Message(false, "failure: "+e.Error()), e.Error()
	}
	defer cancel()
	return u.Message(true, "success"), ""
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

	for c.Next(GetCtx()) {
		x := map[string]interface{}{}
		e := c.Decode(x)
		if e != nil {
			fmt.Println(err)
			return nil, err.Error()
		}
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

	for c.Next(GetCtx()) {
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

func GetEntityHierarchy(entity string, ID primitive.ObjectID, entnum, end int) (map[string]interface{}, string) {

	//Check if at the end of the hierarchy
	if entnum != end {

		//Get the top entity
		top, e := GetEntity(ID, entity)
		if e != "" {
			return nil, e
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
		top[subEnt+"s"] = children

		//Get the rest of hierarchy for children
		for i := range children {
			subIdx := u.EntityToString(entnum + 1)
			subID := (children[i]["_id"].(primitive.ObjectID))
			children[i], _ =
				GetEntityHierarchy(subIdx, subID, entnum+1, end)
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
	return t, ""
}

func GetEntitiesUsingAncestorNames(ent string, id primitive.ObjectID, ancestry map[string]string) ([]map[string]interface{}, string) {
	top, e := GetEntity(id, ent)
	if e != "" {
		return nil, e
	}

	pid := (top["_id"].(primitive.ObjectID)).Hex()

	var x map[string]interface{}
	var e1 string
	for k, v := range ancestry {

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
		pid = (x["_id"].(primitive.ObjectID)).Hex()
	}
	return nil, ""
}

func GetEntityUsingAncestorNames(ent string, id primitive.ObjectID, ancestry map[string]string) (map[string]interface{}, string) {
	top, e := GetEntity(id, ent)
	if e != "" {
		return nil, e
	}

	pid := (top["_id"].(primitive.ObjectID)).Hex()

	var x map[string]interface{}
	var e1 string
	for k, v := range ancestry {

		println("KEY:", k, " VAL:", v)

		/*if k == "device" {
			println("entered")
			q, e2 := GetDeviceFByNameAndParentID(pid, v)
			if e2 != "" {
				println("Failing here")
				return nil, ""
			}

			if len(q) == 1 {
				x = q[0]
			} else {
				x = map[string]interface{}{"devices": q}
			}
			return x, ""
		}*/

		x, e1 = GetEntityByNameAndParentID(k, pid, v)
		if e1 != "" {
			println("Failing here")
			return nil, ""
		}
		pid = (x["_id"].(primitive.ObjectID)).Hex()
	}
	return x, ""
}

func GetTenantHierarchy(entity, name string, entnum, end int) (map[string]interface{}, string) {

	t, e := GetEntityByName(name, "tenant")
	if e != "" {
		fmt.Println(e)
		return nil, e
	}

	subEnt := u.EntityToString(entnum + 1)
	tid := t["_id"].(primitive.ObjectID).Hex()

	//Get immediate children
	children, e1 := GetEntitiesOfParent(subEnt, tid)
	if e1 != "" {
		println("Are we here")
		println("SUBENT: ", subEnt)
		println("PID: ", tid)
		return nil, e1
	}
	t[subEnt+"s"] = children

	//Get the rest of hierarchy for children
	for i := range children {
		subIdx := u.EntityToString(entnum + 1)
		subID := (children[i]["_id"].(primitive.ObjectID))
		children[i], _ =
			GetEntityHierarchy(subIdx, subID, entnum+1, end)
	}

	return t, ""

}

func GetEntitiesUsingTenantAsAncestor(ent, id string, ancestry map[string]string) ([]map[string]interface{}, string) {
	top, e := GetEntityByName(id, ent)
	if e != "" {
		return nil, e
	}

	pid := (top["_id"].(primitive.ObjectID)).Hex()

	var x map[string]interface{}
	var e1 string
	for k, v := range ancestry {

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
		pid = (x["_id"].(primitive.ObjectID)).Hex()
	}
	return nil, ""
}

func GetEntityUsingTenantAsAncestor(ent, id string, ancestry map[string]string) (map[string]interface{}, string) {
	top, e := GetEntityByName(id, ent)
	if e != "" {
		return nil, e
	}

	pid := (top["_id"].(primitive.ObjectID)).Hex()

	var x map[string]interface{}
	var e1 string
	for k, v := range ancestry {

		println("KEY:", k, " VAL:", v)

		/*if k == "device" {
			println("entered")
			q, e2 := GetDeviceFByNameAndParentID(pid, v)
			if e2 != "" {
				println("Failing here")
				return nil, ""
			}

			if len(q) == 1 {
				x = q[0]
			} else {
				x = map[string]interface{}{"devices": q}
			}
			return x, ""
		}*/

		x, e1 = GetEntityByNameAndParentID(k, pid, v)
		if e1 != "" {
			println("Failing here")
			return nil, ""
		}
		pid = (x["_id"].(primitive.ObjectID)).Hex()
	}
	return x, ""
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
	if resp, ok := ValidateEntity(entity, t); !ok {
		return resp, "validate"
	}

	ctx, cancel := u.Connect()

	parent := u.EntityToString(u.GetParentOfEntityByInt(entity))
	pid, _ := primitive.ObjectIDFromHex(t["parentId"].(string))
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
	nestID string, t map[string]interface{}) (map[string]interface{}, string) {
	foundParent := map[string]interface{}{}

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

	if v, ok := foundParent[ent+"s"].(primitive.A); ok {
		for i := range v {
			if v[i].(map[string]interface{})["id"] == nestID {
				old := v[i].(map[string]interface{})
				for key := range t {
					if _, ok := old[key]; ok {
						old[key] = t[key]
					}
				}
				break
			}
		}
	}

	c1, cancel2 := u.Connect()
	_, e1 := GetDB().Collection(parent).UpdateOne(c1, criteria, bson.M{"$set": foundParent})
	if e1 != nil {
		return u.Message(false,
			"There was an error in deleting the entity2: "+e.Error()), "unable update"
	}
	defer cancel2()

	return u.Message(true, "success"), ""
}

func GetNestedEntityByQuery(parent, entity string, query bson.M) ([]map[string]interface{}, string) {
	ans := make([]map[string]interface{}, 0)
	parents, e := GetAllEntities(parent)
	if e != "" {
		return nil, e
	}

	//Now get all subentities from parents
	for i := range parents {
		pid := parents[i]["_id"].(primitive.ObjectID)
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

func UpdateEntityBySlug(ent, id string, t *map[string]interface{}) (map[string]interface{}, string) {
	ctx, cancel := u.Connect()
	e := GetDB().Collection(ent).FindOneAndUpdate(ctx, bson.M{"slug": id}, bson.M{"$set": *t}).Err()
	if e != nil {
		return u.Message(false, "failure: "+e.Error()), e.Error()
	}
	defer cancel()
	return u.Message(true, "success"), ""
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

	for c.Next(GetCtx()) {
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

func RetrieveDeviceHierarch(ID primitive.ObjectID) (map[string]interface{}, string) {
	//Get the top entity
	top, e := GetEntity(ID, "device")
	if e != "" {
		return nil, e
	}

	children, e1 := GetEntitiesOfParent("device", ID.Hex())
	if e1 != "" {
		return top, ""
	}

	for i := range children {
		children[i], _ = RetrieveDeviceHierarch(
			children[i]["_id"].(primitive.ObjectID))
	}

	top["children"] = children

	return top, ""
}

func DeleteDeviceF(entityID primitive.ObjectID) (map[string]interface{}, string) {
	var deviceType string

	t, e := RetrieveDeviceHierarch(entityID)
	if e != "" {
		return nil, e
	}
	del := parseDataForNonStdResult(deviceType, DEVICE, SUBDEV1+1, t)
	println("LEN OF DEL: ", len(del))
	delSubEnts(DEVICE, del)

	ctx, cancel := u.Connect()
	c, _ := GetDB().Collection("device").DeleteOne(ctx, bson.M{"_id": entityID})
	if c.DeletedCount == 0 {
		return u.Message(false, "There was an error in deleting the entity"), "not found"
	}
	defer cancel()

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
