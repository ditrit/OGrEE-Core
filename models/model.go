package models

import (
	"context"
	"fmt"
	u "p3/utils"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateEntity(entity int, t map[string]interface{}) (map[string]interface{}, string) {
	message := ""
	if resp, ok := ValidateEntity(entity, t); !ok {
		return resp, "validate"
	}

	//Set timestamp
	t["createdDate"] = primitive.NewDateTimeFromTime(time.Now())
	t["lastUpdated"] = t["createdDate"]

	//Last modifications before insert
	FixAttributesBeforeInsert(entity, t)

	ctx, cancel := u.Connect()
	entStr := u.EntityToString(entity)
	res, e := GetDB().Collection(entStr).InsertOne(ctx, t)
	if e != nil {
		if strings.Contains(e.Error(), "E11000") {
			return u.Message(false,
					"Error while creating "+entStr+": Duplicates not allowed"),
				"duplicate"
		}
		return u.Message(false,
				"Internal error while creating "+entStr+": "+e.Error()),
			e.Error()
	}
	defer cancel()

	t["id"] = res.InsertedID

	switch entity {
	case u.ROOMTMPL:
		message = "successfully created room_template"
	case u.OBJTMPL:
		message = "successfully created obj_template"
	case u.BLDGTMPL:
		message = "successfully created bldg_template"
	default:
		message = "successfully created object"
	}

	resp := u.Message(true, message)
	resp["data"] = t
	return resp, ""
}

// GetObjectByName: search for hierarchyName in all possible collections
func GetObjectByName(hierarchyName string, filters u.RequestFilters) (map[string]interface{}, string) {
	var resp map[string]interface{}
	// Get possible collections for this name
	rangeEntities := u.HierachyNameToEntity(hierarchyName)

	// Search each collection
	for _, entity := range rangeEntities {
		req := bson.M{"hierarchyName": hierarchyName}
		if entity == u.SITE {
			req = bson.M{"name": hierarchyName}
		}
		entityStr := u.EntityToString(entity)
		data, _ := GetEntity(req, entityStr, filters)
		if data != nil {
			resp = data
			break
		}
	}

	if resp != nil {
		return resp, ""
	} else {
		return nil, "Unable to find object"
	}
}

func GetEntity(req bson.M, ent string, filters u.RequestFilters) (map[string]interface{}, string) {
	t := map[string]interface{}{}
	ctx, cancel := u.Connect()
	var e error

	var opts *options.FindOneOptions
	if len(filters.FieldsToShow) > 0 {
		var compoundIndex bson.D
		for _, field := range filters.FieldsToShow {
			compoundIndex = append(compoundIndex, bson.E{Key: field, Value: 1})
		}
		opts = options.FindOne().SetProjection(compoundIndex)
	}
	e = getDateFilters(req, filters)
	if e != nil {
		return nil, e.Error()
	}

	if opts != nil {
		e = GetDB().Collection(ent).FindOne(ctx, req, opts).Decode(&t)
	} else {
		e = GetDB().Collection(ent).FindOne(ctx, req).Decode(&t)
	}
	if e != nil {
		return nil, e.Error()
	}
	defer cancel()
	//Remove _id
	t = fixID(t)

	//If entity has '_' remove it
	if strings.Contains(ent, "_") {
		FixUnderScore(t)
	}
	return t, ""
}

func getDateFilters(req bson.M, filters u.RequestFilters) error {
	if len(filters.StartDate) > 0 || len(filters.EndDate) > 0 {
		lastUpdateReq := bson.M{}
		if len(filters.StartDate) > 0 {
			startDate, e := time.Parse("2006-01-02", filters.StartDate)
			if e != nil {
				return e
			}
			lastUpdateReq["$gte"] = primitive.NewDateTimeFromTime(startDate)
		}

		if len(filters.EndDate) > 0 {
			endDate, e := time.Parse("2006-01-02", filters.EndDate)
			endDate = endDate.Add(time.Hour * 24)
			if e != nil {
				return e
			}
			lastUpdateReq["$lte"] = primitive.NewDateTimeFromTime(endDate)
		}
		req["lastUpdated"] = lastUpdateReq
	}
	return nil
}

func GetManyEntities(ent string, req bson.M, filters u.RequestFilters) ([]map[string]interface{}, string) {
	ctx, cancel := u.Connect()
	var err error
	var c *mongo.Cursor

	var opts *options.FindOptions
	if len(filters.FieldsToShow) > 0 {
		var compoundIndex bson.D
		for _, field := range filters.FieldsToShow {
			compoundIndex = append(compoundIndex, bson.E{Key: field, Value: 1})
		}
		opts = options.Find().SetProjection(compoundIndex)
	}
	err = getDateFilters(req, filters)
	if err != nil {
		return nil, err.Error()
	}

	if opts != nil {
		c, err = GetDB().Collection(ent).Find(ctx, req, opts)
	} else {
		c, err = GetDB().Collection(ent).Find(ctx, req)
	}
	if err != nil {
		fmt.Println(err)
		return nil, err.Error()
	}
	defer cancel()

	data, e1 := ExtractCursor(c, ctx)
	if e1 != "" {
		fmt.Println(e1)
		return nil, e1
	}

	//Remove underscore If the entity has '_'
	if strings.Contains(ent, "_") {
		for i := range data {
			FixUnderScore(data[i])
		}
	}

	return data, ""
}

// GetCompleteHierarchy: gets all objects in db using hierachyName and returns:
//   - tree: map with parents as key and their children as an array value
//     tree: {parent:[children]}
//   - categories: map with category name as key and corresponding objects
//     as an array value
//     categories: {categoryName:[children]}
func GetCompleteHierarchy() (map[string]interface{}, string) {
	response := make(map[string]interface{})
	categories := make(map[string][]string)
	hierarchy := make(map[string][]string)
	rootCollectionName := "site"

	// Get all collections names
	ctx, cancel := u.Connect()
	db := GetDB()
	collNames, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		fmt.Println(err.Error())
		return nil, err.Error()
	}

	// Get all objects hierarchyNames for each collection
	for _, collName := range collNames {
		opts := options.Find().SetProjection(bson.D{{Key: "hierarchyName", Value: 1}})
		if collName == rootCollectionName {
			opts = options.Find().SetProjection(bson.D{{Key: "name", Value: 1}})
		}

		c, err := db.Collection(collName).Find(ctx, bson.M{}, opts)
		if err != nil {
			println(err.Error())
		}
		data, error := ExtractCursor(c, ctx)
		if error != "" {
			fmt.Println(error)
			return nil, error
		}

		for _, obj := range data {
			if obj["hierarchyName"] != nil {
				categories[collName] = append(categories[collName], obj["hierarchyName"].(string))
				fillHierarchyMap(obj["hierarchyName"].(string), hierarchy)
			} else if obj["name"] != nil {
				categories[rootCollectionName] = append(categories[rootCollectionName], obj["name"].(string))
				hierarchy["Root"] = append(hierarchy["Root"], obj["name"].(string))
			}
		}
	}

	response["tree"] = hierarchy
	response["categories"] = categories
	defer cancel()
	return response, ""
}

// fillHierarchyMap: add hierarchyName to the children array of its parent
func fillHierarchyMap(hierarchyName string, data map[string][]string) {
	i := strings.LastIndex(hierarchyName, ".")
	if i > 0 {
		parent := hierarchyName[:i]
		data[parent] = append(data[parent], hierarchyName)
	}
}

// GetSiteParentTempUnit: search for the object of given ID,
// then search for is site parent and return its attributes.temperatureUnit
func GetSiteParentTempUnit(id string) (string, string) {
	data := map[string]interface{}{}

	// Get all collections names
	ctx, cancel := u.Connect()
	db := GetDB()
	collNames, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		fmt.Println(err.Error())
		return "", err.Error()
	}
	// Find object
	for _, collName := range collNames {
		var filter primitive.M
		objID, e := primitive.ObjectIDFromHex(id)
		if e == nil {
			filter = bson.M{"_id": objID}
		} else {
			filter = bson.M{"hierarchyName": id}
		}
		err := db.Collection(collName).FindOne(ctx, filter).Decode(&data)
		if err == nil {
			// Found object with given id
			if data["category"].(string) == "site" {
				// it's a site
				break
			} else {
				// Find its parent site
				nameSlice := strings.Split(data["hierarchyName"].(string), ".")
				//if len(nameSlice) < 2 { // REMOVE IT FOR DBFORTENANTS
				//	return "", "Could not find parent site for given object"
				//}
				siteName := nameSlice[1] // CONSIDER SITE AS 0
				err := db.Collection("site").FindOne(ctx, bson.M{"hierarchyName": siteName}).Decode(&data)
				if err != nil {
					// id not found in any collection
					return "", "Could not find parent site for given object"
				}
			}
		}
	}

	defer cancel()

	if len(data) == 0 {
		return "", "No object found with given id"
	} else if tempUnit := data["attributes"].(map[string]interface{})["temperatureUnit"]; tempUnit == nil {
		return "", "Parent site has no temperatureUnit in attributes"
	} else {
		return tempUnit.(string), ""
	}
}

func GetEntityCount(entity int) int64 {
	ent := u.EntityToString(entity)
	ctx, cancel := u.Connect()
	ans, e := GetDB().Collection(ent).CountDocuments(ctx, bson.M{}, nil)
	if e != nil {
		println(e.Error())
		return -1
	}
	defer cancel()
	return ans
}

func CommandRunner(cmd interface{}) *mongo.SingleResult {
	ctx, cancel := u.Connect()
	result := GetDB().RunCommand(ctx, cmd, nil)
	defer cancel()
	return result
}

func GetStats() map[string]interface{} {
	ans := map[string]interface{}{}
	t := map[string]interface{}{}
	latestDocArr := []map[string]interface{}{}
	var latestTime interface{}

	for i := 0; i <= u.STRAYSENSOR; i++ {
		num := GetEntityCount(i)
		if num == -1 {
			num = 0
		}

		ans["Number of "+u.EntityToString(i)+"s:"] = num

		//Retrieve the latest updated document in each collection
		//and store into the latestDocArr array
		obj := map[string]interface{}{}
		filter := options.FindOne().SetSort(bson.M{"lastUpdated": -1})
		ctx, cancel := u.Connect()

		e := GetDB().Collection(u.EntityToString(i)).FindOne(ctx, bson.M{}, filter).Decode(&obj)
		if e == nil {
			latestDocArr = append(latestDocArr, obj)
		}
		defer cancel()
	}

	//Get the latest update out of latestDocArr
	value := -1
	for _, obj := range latestDocArr {
		if int(obj["lastUpdated"].(primitive.DateTime)) > value {
			value = int(obj["lastUpdated"].(primitive.DateTime))
			latestTime = obj["lastUpdated"]
		}
	}

	if latestTime == nil {
		latestTime = "N/A"
	}

	cmd := bson.D{{"dbStats", 1}, {"scale", 1024}}

	if e := CommandRunner(cmd).Decode(&t); e != nil {
		println(e.Error())
		return nil
	}

	ans["Number of Hierarchal Objects"] = t["collections"]
	ans["Last Job Timestamp"] = latestTime

	return ans
}

func GetDBName() string {
	name := GetDB().Name()

	//Remove the preceding 'ogree' at beginning of name
	if strings.Index(name, "ogree") == 0 {
		name = name[5:] //5=len('ogree')
	}
	return name
}

// DeleteEntityByName: delete object of given hierarchyName
// search for all its children and delete them too, return:
// - success or fail message map
func DeleteEntityByName(entity string, name string) map[string]interface{} {
	var req primitive.M
	if entity == "site" {
		req = bson.M{"name": name}
	} else {
		req = bson.M{"hierarchyName": name}
	}
	resp, err := DeleteEntityManual(entity, req)
	if err != "" {
		// Unable to delete given object
		return resp
	} else {
		// Delete possible children
		rangeEntities := getChildrenCollections(u.STRAYSENSOR, entity)
		for _, childEnt := range rangeEntities {
			childEntName := u.EntityToString(childEnt)
			pattern := primitive.Regex{Pattern: name, Options: ""}

			ctx, cancel := u.Connect()
			GetDB().Collection(childEntName).DeleteMany(ctx,
				bson.M{"hierarchyName": pattern})
			defer cancel()
		}
	}

	return u.Message(true, "success")
}

func DeleteEntityManual(entity string, req bson.M) (map[string]interface{}, string) {
	//Finally delete the Entity
	ctx, cancel := u.Connect()
	c, _ := GetDB().Collection(entity).DeleteOne(ctx, req)
	if c.DeletedCount == 0 {
		return u.Message(false, "There was an error in deleting the entity"), "not found"
	}
	defer cancel()

	return u.Message(true, "success"), ""
}

func DeleteEntity(entity string, id primitive.ObjectID) (map[string]interface{}, string) {
	var t map[string]interface{}
	var e string
	eNum := u.EntityStrToInt(entity)
	if eNum > u.DEVICE {
		//Delete the non hierarchal objects
		t, e = GetEntityHierarchy(id, entity, eNum, eNum+eNum, u.RequestFilters{})
	} else {
		t, e = GetEntityHierarchy(id, entity, eNum, u.AC, u.RequestFilters{})
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
					if ent == u.STRAYDEV || ent == u.DOMAIN {
						deleteHelper(x[i], ent)
					} else {
						deleteHelper(x[i], ent+1)
					}

				}
			}
		}

		println("So we got: ", ent)

		if ent == u.RACK {
			ctx, cancel := u.Connect()
			GetDB().Collection("sensor").DeleteMany(ctx,
				bson.M{"parentId": t["id"].(primitive.ObjectID).Hex()})

			GetDB().Collection("group").DeleteMany(ctx,
				bson.M{"parentId": t["id"].(primitive.ObjectID).Hex()})
			defer cancel()
		}

		//Delete associated non hierarchal objs
		if ent == u.ROOM {
			//ITER Through all nonhierarchal objs
			ctx, cancel := u.Connect()
			for i := u.AC; i < u.GROUP+1; i++ {
				ent := u.EntityToString(i)
				GetDB().Collection(ent).DeleteMany(ctx, bson.M{"parentId": t["id"].(primitive.ObjectID).Hex()})
			}
			defer cancel()
		}

		//Delete hierarchy under stray-device
		if ent == u.STRAYDEV {
			ctx, cancel := u.Connect()
			entity := u.EntityToString(u.STRAYSENSOR)
			GetDB().Collection(entity).DeleteMany(ctx, bson.M{"parentId": t["id"].(primitive.ObjectID).Hex()})
			defer cancel()
		}

		if ent == u.DEVICE {
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

func UpdateEntity(ent string, req bson.M, t *map[string]interface{}, isPatch bool) (map[string]interface{}, string) {
	var e *mongo.SingleResult
	updatedDoc := bson.M{}
	retDoc := options.ReturnDocument(options.After)

	//Update timestamp requires first obj retrieval
	//there isn't any way for mongoDB to make a field
	//immutable in a document
	oldObj, e1 := GetEntity(req, ent, u.RequestFilters{})
	if e1 != "" {
		return u.Message(false, "Error: "+e1), e1
	}
	(*t)["lastUpdated"] = primitive.NewDateTimeFromTime(time.Now())
	(*t)["createdDate"] = oldObj["createdDate"]

	// Ensure the update is valid and apply it
	ctx, cancel := u.Connect()
	if isPatch {
		msg, ok := ValidatePatch(u.EntityStrToInt(ent), *t)
		if !ok {
			return msg, "invalid"
		}
		e = GetDB().Collection(ent).FindOneAndUpdate(ctx,
			req, bson.M{"$set": *t},
			&options.FindOneAndUpdateOptions{ReturnDocument: &retDoc})
		if e.Err() != nil {
			return u.Message(false, "failure: "+e.Err().Error()), e.Err().Error()
		}
	} else {
		println("NOT A PATCH")
		msg, ok := ValidateEntity(u.EntityStrToInt(ent), *t)
		if !ok {
			return msg, "invalid"
		}
		e = GetDB().Collection(ent).FindOneAndReplace(ctx,
			req, *t,
			&options.FindOneAndReplaceOptions{ReturnDocument: &retDoc})
		if e.Err() != nil {
			return u.Message(false, "failure: "+e.Err().Error()), e.Err().Error()
		}
	}

	// Changes to hierarchyName should be propagated to its children
	if ent == "site" && oldObj["name"] != (*t)["name"] {
		propagateParentNameChange(ctx, oldObj["name"].(string),
			(*t)["name"].(string), u.EntityStrToInt(ent))
	}
	if oldObj["hierarchyName"] != (*t)["hierarchyName"] {
		propagateParentNameChange(ctx, oldObj["hierarchyName"].(string),
			(*t)["hierarchyName"].(string), u.EntityStrToInt(ent))
	}

	//Obtain new document then
	//Fix the _id / id discrepancy
	e.Decode(&updatedDoc)
	updatedDoc = fixID(updatedDoc)

	//Response Message
	message := ""
	switch u.EntityStrToInt(ent) {
	case u.ROOMTMPL:
		message = "successfully updated room_template"
	case u.OBJTMPL:
		message = "successfully updated obj_template"
	case u.BLDGTMPL:
		message = "successfully created bldg_template"
	default:
		message = "successfully updated object"
	}

	defer cancel()
	resp := u.Message(true, message)
	resp["data"] = updatedDoc
	return resp, ""
}

// propagateParentNameChange: search for given parent children and
// update their hierarchyName with new parent name
func propagateParentNameChange(ctx context.Context, oldParentName, newName string, entityInt int) {
	// Find all objects containing parent name
	req := bson.M{"hierarchyName": primitive.Regex{Pattern: oldParentName + ".", Options: ""}}
	// For each object found, replace old name by new
	update := bson.D{{
		Key: "$set", Value: bson.M{
			"hierarchyName": bson.M{
				"$replaceOne": bson.M{
					"input":       "$hierarchyName",
					"find":        oldParentName,
					"replacement": newName}}}}}

	if entityInt == u.DEVICE {
		_, e := GetDB().Collection(u.EntityToString(u.DEVICE)).UpdateMany(ctx,
			req, mongo.Pipeline{update})
		if e != nil {
			println(e.Error())
		}
	} else if entityInt == u.STRAYDEV {
		_, e := GetDB().Collection(u.EntityToString(u.STRAYDEV)).UpdateMany(ctx,
			req, mongo.Pipeline{update})
		if e != nil {
			println(e.Error())
		}
	} else if entityInt >= u.SITE && entityInt <= u.RACK {
		for i := entityInt + 1; i <= u.GROUP; i++ {
			_, e := GetDB().Collection(u.EntityToString(i)).UpdateMany(ctx,
				req, mongo.Pipeline{update})
			if e != nil {
				println(e.Error())
			}
		}
	}
}

func GetEntityHierarchy(ID primitive.ObjectID, ent string, start, end int, filters u.RequestFilters) (map[string]interface{}, string) {
	var childEnt string

	if start < end {
		top, e := GetEntity(bson.M{"_id": ID}, ent, filters)
		if top == nil {
			return nil, e
		}
		top = fixID(top)

		children := []map[string]interface{}{}
		pid := ID.Hex()
		//Get sensors & groups
		if ent == "rack" || ent == "device" {
			//Get sensors
			sensors, _ := GetManyEntities("sensor", bson.M{"parentId": pid}, filters)

			//Get groups
			groups, _ := GetManyEntities("group", bson.M{"parentId": pid}, filters)

			if sensors != nil {
				children = append(children, sensors...)
			}
			if groups != nil {
				children = append(children, groups...)
			}
		}

		if ent == "device" || ent == "stray_device" || ent == "domain" {
			childEnt = ent
		} else {
			childEnt = u.EntityToString(start + 1)
		}

		subEnts, _ := GetManyEntities(childEnt, bson.M{"parentId": pid}, filters)

		for idx := range subEnts {
			tmp, _ := GetEntityHierarchy(subEnts[idx]["id"].(primitive.ObjectID), childEnt, start+1, end, filters)
			if tmp != nil {
				subEnts[idx] = tmp
			}
		}

		if subEnts != nil {
			children = append(children, subEnts...)
		}

		if ent == "room" {
			for i := u.AC; i < u.CABINET+1; i++ {
				roomEnts, _ := GetManyEntities(u.EntityToString(i), bson.M{"parentId": pid}, filters)
				if roomEnts != nil {
					children = append(children, roomEnts...)
				}
			}
			for i := u.PWRPNL; i < u.SENSOR+1; i++ {
				roomEnts, _ := GetManyEntities(u.EntityToString(i), bson.M{"parentId": pid}, filters)
				if roomEnts != nil {
					children = append(children, roomEnts...)
				}
			}
			roomEnts, _ := GetManyEntities(u.EntityToString(u.CORRIDOR), bson.M{"parentId": pid}, filters)
			if roomEnts != nil {
				children = append(children, roomEnts...)
			}
			roomEnts, _ = GetManyEntities(u.EntityToString(u.GROUP), bson.M{"parentId": pid}, filters)
			if roomEnts != nil {
				children = append(children, roomEnts...)
			}
		}

		if ent == "stray_device" {
			sSensors, _ := GetManyEntities("stray_sensor", bson.M{"parentId": pid}, filters)
			if sSensors != nil {
				children = append(children, sSensors...)
			}
		}

		if len(children) > 0 {
			top["children"] = children
		}

		return top, ""
	}
	return nil, ""
}

func GetEntitiesUsingAncestorNames(ent string, id primitive.ObjectID, ancestry []map[string]string) ([]map[string]interface{}, string) {
	top, e := GetEntity(bson.M{"_id": id}, ent, u.RequestFilters{})
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
				return GetManyEntities(k, bson.M{"parentId": pid}, u.RequestFilters{})
			}

			x, e1 = GetEntity(bson.M{"parentId": pid, "name": v}, k, u.RequestFilters{})
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
	top, e := GetEntity(bson.M{"_id": id}, ent, u.RequestFilters{})
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

			x, e1 = GetEntity(bson.M{"parentId": pid, "name": v}, k, u.RequestFilters{})
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

// GetHierarchyByName: get children objects of given parent.
// - Param limit: max relationship distance between parent and child, example:
// limit=1 only direct children, limit=2 includes nested children of children
func GetHierarchyByName(entity, hierarchyName string, limit int, filters u.RequestFilters) ([]map[string]interface{}, string) {
	allChildren := map[string]interface{}{}
	hierarchy := make(map[string][]string)

	// Define in which collections we can find children
	rangeEntities := getChildrenCollections(limit, entity)

	// Guarantee hierarchyName is present even with filters
	if len(filters.FieldsToShow) > 0 && !u.StrSliceContains(filters.FieldsToShow, "hierarchyName") {
		filters.FieldsToShow = append(filters.FieldsToShow, "hierarchyName")
	}

	// Get children from all given collections
	for _, checkEnt := range rangeEntities {
		checkEntName := u.EntityToString(checkEnt)
		// Obj should include parentName and not surpass limit range
		pattern := primitive.Regex{Pattern: hierarchyName +
			"(.[A-Za-z0-9_\" \"]+){1," + strconv.Itoa(limit) + "}$", Options: ""}
		children, e1 := GetManyEntities(checkEntName, bson.M{"hierarchyName": pattern}, filters)
		if e1 != "" {
			println("SUBENT: ", checkEntName)
			println("ERR: ", e1)
			return nil, e1
		}
		for _, child := range children {
			// store child data
			allChildren[child["hierarchyName"].(string)] = child
			// create hierarchy map
			fillHierarchyMap(child["hierarchyName"].(string), hierarchy)
		}
	}

	// Organize the family
	return recursivelyGetChildrenFromMaps(hierarchyName, hierarchy, allChildren), ""
}

// recursivelyGetChildrenFromMaps: nest children data as the array value of
// its parents "children" key
func recursivelyGetChildrenFromMaps(parentHierarchyName string, hierarchy map[string][]string,
	allChildrenData map[string]interface{}) []map[string]interface{} {
	var children []map[string]interface{}
	for _, childName := range hierarchy[parentHierarchyName] {
		// Get the child data and get its own children
		child := allChildrenData[childName].(map[string]interface{})
		child["children"] = recursivelyGetChildrenFromMaps(childName, hierarchy, allChildrenData)
		children = append(children, child)
	}
	return children
}

// getChildrenCollections: get a list of entites where children of given parentEntStr
// may be found, considering limit as the max possible distance of child to parent
func getChildrenCollections(limit int, parentEntStr string) []int {
	rangeEntities := []int{}
	startEnt := u.EntityStrToInt(parentEntStr) + 1
	endEnt := startEnt + limit
	if parentEntStr == "device" {
		// device special case (devices can have devices)
		startEnt = u.DEVICE
		endEnt = u.DEVICE
	} else if parentEntStr == "stray_device" {
		// stray device special case
		startEnt = u.STRAYDEV
		endEnt = u.STRAYDEV
	} else if endEnt >= u.DEVICE {
		// include AC, CABINET, CORRIDOR, PWRPNL and GROUP
		// beacause of ROOM and RACK possible children
		// but no need to search further than group
		endEnt = u.GROUP
	}
	for i := startEnt; i <= endEnt; i++ {
		rangeEntities = append(rangeEntities, i)
	}
	if startEnt == u.ROOM && endEnt == u.RACK {
		// ROOM limit=1 special case should include extra
		// ROOM children but avoiding DEVICE (big collection)
		rangeEntities = append(rangeEntities, u.CORRIDOR, u.CABINET, u.PWRPNL, u.GROUP)
	}

	return rangeEntities
}

func GetEntitiesUsingSiteAsAncestor(ent, id string, ancestry []map[string]string) ([]map[string]interface{}, string) {
	top, e := GetEntity(bson.M{"name": id}, ent, u.RequestFilters{})
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
				return GetManyEntities(k, bson.M{"parentId": pid}, u.RequestFilters{})
			}

			x, e1 = GetEntity(bson.M{"parentId": pid, "name": v}, k, u.RequestFilters{})
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

func GetEntityUsingSiteAsAncestor(ent, id string, ancestry []map[string]string) (map[string]interface{}, string) {
	top, e := GetEntity(bson.M{"name": id}, ent, u.RequestFilters{})
	if e != "" {
		return nil, e
	}

	pid := (top["id"].(primitive.ObjectID)).Hex()

	var x map[string]interface{}
	var e1 string
	for i := range ancestry {
		for k, v := range ancestry[i] {

			println("KEY:", k, " VAL:", v)

			x, e1 = GetEntity(bson.M{"parentId": pid, "name": v}, k, u.RequestFilters{})
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
	if ent == u.SITE {

		t, e = GetEntity(bson.M{"name": id}, "site", u.RequestFilters{})
		if e != "" {
			return nil, e
		}

	} else {
		ID, _ := primitive.ObjectIDFromHex(id.(string))
		t, e = GetEntity(bson.M{"_id": ID}, entStr, u.RequestFilters{})
		if e != "" {
			return nil, e
		}
	}

	sub, e1 := GetManyEntities(u.EntityToString(ent+1),
		bson.M{"parentId": t["id"].(primitive.ObjectID).Hex()}, u.RequestFilters{})
	if e1 != "" {
		return nil, e1
	}

	if wantedEnt == "" {
		wantedEnt = u.EntityToString(ent + 2)
	}

	for i := range sub {
		x, _ := GetManyEntities(wantedEnt,
			bson.M{"parentId": sub[i]["id"].(primitive.ObjectID).Hex()}, u.RequestFilters{})
		ans = append(ans, x...)
	}
	return ans, ""
}

//DEV FAMILY FUNCS

func DeleteDeviceF(entityID primitive.ObjectID) (map[string]interface{}, string) {
	//var deviceType string

	t, e := GetEntityHierarchy(entityID, "device", 0, 999, u.RequestFilters{})
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
			}
		}

		ctx, cancel := u.Connect()
		//Delete relevant non hierarchal objects
		GetDB().Collection("sensor").DeleteMany(ctx,
			bson.M{"parentId": t["id"].(primitive.ObjectID).Hex()})

		GetDB().Collection("group").DeleteMany(ctx,
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
	ans := []map[string]interface{}{}
	for c.Next(ctx) {
		x := map[string]interface{}{}
		err := c.Decode(x)
		if err != nil {
			fmt.Println(err)
			return nil, err.Error()
		}
		//Remove _id
		x = fixID(x)
		ans = append(ans, x)
	}
	return ans, ""
}

// DEAD CODE
// Function will recursively iterate through nested obj
// and accumulate whatever is found into category arrays
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
