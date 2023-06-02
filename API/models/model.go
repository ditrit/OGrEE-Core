package models

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	u "p3/utils"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateEntity(entity int, t map[string]interface{}, userRoles map[string]Role) (map[string]interface{}, string) {
	message := ""
	if resp, ok := ValidateEntity(entity, t); !ok {
		return resp, "validate"
	}

	// Check user permissions
	if entity != u.BLDGTMPL && entity != u.ROOMTMPL && entity != u.OBJTMPL {
		var domain string
		if entity == u.DOMAIN {
			domain = t["hierarchyName"].(string)
		} else {
			domain = t["domain"].(string)
		}
		if permission := CheckUserPermissions(userRoles, entity, domain); permission < WRITE {
			return u.Message(false,
					"User does not have permission to create this object"),
				"permission"
		}
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
func GetObjectByName(hierarchyName string, filters u.RequestFilters, userRoles map[string]Role) (map[string]interface{}, string) {
	var resp map[string]interface{}
	// Get possible collections for this name
	rangeEntities := u.HierachyNameToEntity(hierarchyName)

	// Search each collection
	for _, entity := range rangeEntities {
		req := bson.M{"hierarchyName": hierarchyName}
		entityStr := u.EntityToString(entity)
		data, _ := GetEntity(req, entityStr, filters, userRoles)
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

func GetEntity(req bson.M, ent string, filters u.RequestFilters, userRoles map[string]Role) (map[string]interface{}, string) {
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

	// Check permissions
	if !strings.Contains(ent, "template") {
		var domain string
		if ent == "domain" {
			domain = t["name"].(string)
		} else {
			domain = t["domain"].(string)
		}
		if userRoles != nil {
			if permission := CheckUserPermissions(userRoles, u.EntityStrToInt(ent), domain); permission == NONE {
				return u.Message(false,
						"User does not have permission to see this object"),
					"permission"
			} else if permission == READONLYNAME {
				t = FixReadOnlyName(t)
			}
		}
	}

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

func GetManyEntities(ent string, req bson.M, filters u.RequestFilters, userRoles map[string]Role) ([]map[string]interface{}, string) {
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

	data, e1 := ExtractCursor(c, ctx, u.EntityStrToInt(ent), userRoles)
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
func GetCompleteDomainHierarchy(userRoles map[string]Role) (map[string]interface{}, string) {
	response := make(map[string]interface{})
	hierarchy := make(map[string][]string)

	// Get all collections names
	ctx, cancel := u.Connect()
	db := GetDB()
	collName := "domain"

	// Get all objects hierarchyNames for each collection
	opts := options.Find().SetProjection(bson.D{{Key: "hierarchyName", Value: 1}, {Key: "domain", Value: 1}})

	c, err := db.Collection(collName).Find(ctx, bson.M{}, opts)
	if err != nil {
		println(err.Error())
	}
	data, error := ExtractCursor(c, ctx, u.EntityStrToInt(collName), userRoles)
	if error != "" {
		fmt.Println(error)
		return nil, error
	}

	for _, obj := range data {
		if strings.Contains(obj["hierarchyName"].(string), ".") {
			fillHierarchyMap(obj["hierarchyName"].(string), hierarchy)
		} else {
			hierarchy["Root"] = append(hierarchy["Root"], obj["hierarchyName"].(string))
		}
	}

	response["tree"] = hierarchy
	defer cancel()
	return response, ""
}

// GetCompleteHierarchy: gets all objects in db using hierachyName and returns:
//   - tree: map with parents as key and their children as an array value
//     tree: {parent:[children]}
//   - categories: map with category name as key and corresponding objects
//     as an array value
//     categories: {categoryName:[children]}
func GetCompleteHierarchy(userRoles map[string]Role) (map[string]interface{}, string) {
	response := make(map[string]interface{})
	categories := make(map[string][]string)
	hierarchy := make(map[string][]string)
	rootCollectionName := "site"

	// Get all collections names
	var collNames []string
	for i := u.SITE; i <= u.GROUP; i++ {
		collNames = append(collNames, u.EntityToString(i))
	}

	ctx, cancel := u.Connect()
	db := GetDB()

	// Get all objects hierarchyNames for each collection
	for _, collName := range collNames {
		opts := options.Find().SetProjection(bson.D{{Key: "hierarchyName", Value: 1}, {Key: "domain", Value: 1}})

		c, err := db.Collection(collName).Find(ctx, bson.M{}, opts)
		if err != nil {
			println(err.Error())
		}
		data, error := ExtractCursor(c, ctx, u.EntityStrToInt(collName), userRoles)
		if error != "" {
			fmt.Println(error)
			return nil, error
		}

		for _, obj := range data {
			if collName == rootCollectionName {
				categories[rootCollectionName] = append(categories[rootCollectionName], obj["hierarchyName"].(string))
				hierarchy["Root"] = append(hierarchy["Root"], obj["hierarchyName"].(string))

			} else if obj["hierarchyName"] != nil {
				categories[collName] = append(categories[collName], obj["hierarchyName"].(string))
				fillHierarchyMap(obj["hierarchyName"].(string), hierarchy)
			}
		}
	}

	categories["KeysOrder"] = []string{"site", "building", "room", "rack"}
	response["tree"] = hierarchy
	response["categories"] = categories
	defer cancel()
	return response, ""
}

// fillHierarchyMap: add hierarchyName to the children array of its parent
func fillHierarchyMap(hierarchyName string, data map[string][]string) {
	i := strings.LastIndex(hierarchyName, u.HN_DELIMETER)
	if i > 0 {
		parent := hierarchyName[:i]
		data[parent] = append(data[parent], hierarchyName)
	}
}

func GetCompleteHierarchyAttributes(userRoles map[string]Role) (map[string]interface{}, string) {
	response := make(map[string]interface{})
	// Get all collections names
	ctx, cancel := u.Connect()
	db := GetDB()
	collNames, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		fmt.Println(err.Error())
		return nil, err.Error()
	}

	for _, collName := range collNames {
		var projection primitive.D
		if collName == "site" {
			projection = bson.D{{Key: "name", Value: 1}, {Key: "attributes", Value: 1}}
		} else {
			projection = bson.D{{Key: "hierarchyName", Value: 1}, {Key: "attributes", Value: 1}}
		}
		opts := options.Find().SetProjection(projection)

		c, err := db.Collection(collName).Find(ctx, bson.M{}, opts)
		if err != nil {
			println(err.Error())
		}
		data, error := ExtractCursor(c, ctx, u.EntityStrToInt(collName), userRoles)
		if error != "" {
			fmt.Println(error)
			return nil, error
		}

		for _, obj := range data {
			if obj["attributes"] != nil {
				if obj["hierarchyName"] != nil {
					response[obj["hierarchyName"].(string)] = obj["attributes"]
				} else if obj["name"] != nil {
					response[obj["name"].(string)] = obj["attributes"]
				}
			}
		}
	}
	defer cancel()
	return response, ""
}

func domainHasObjects(domain string) bool {
	data := map[string]interface{}{}
	// Get all collections names
	ctx, cancel := u.Connect()
	db := GetDB()
	collNames, _ := db.ListCollectionNames(ctx, bson.D{})

	// Check if at least one object belongs to domain
	for _, collName := range collNames {
		pattern := primitive.Regex{Pattern: "^" + domain, Options: ""}
		e := db.Collection(collName).FindOne(ctx, bson.M{"domain": pattern}).Decode(&data)
		if e == nil {
			// Found one!
			return true
		}
	}

	defer cancel()
	return false
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
				nameSlice := strings.Split(data["hierarchyName"].(string), u.HN_DELIMETER)
				siteName := nameSlice[0] // CONSIDER SITE AS 0
				err := db.Collection("site").FindOne(ctx, bson.M{"hierarchyName": siteName}).Decode(&data)
				if err != nil {
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
func DeleteEntityByName(entity string, name string, userRoles map[string]Role) (map[string]interface{}, string) {
	if entity == "domain" {
		if name == os.Getenv("db") {
			return u.Message(false, "Cannot delete tenant's default domain"), "domain"
		}
		if domainHasObjects(name) {
			return u.Message(false, "Cannot delete domain if it has at least one object"), "domain"
		}
	}

	req, ok := GetRequestFilterByDomain(userRoles)
	if !ok {
		return u.Message(false, "User does not have permission to delete"), "permission"
	}
	req["hierarchyName"] = name
	resp, err := DeleteSingleEntity(entity, req)

	if err != "" {
		// Unable to delete given object
		return resp, err
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

	return u.Message(true, "success"), ""
}

func DeleteSingleEntity(entity string, req bson.M) (map[string]interface{}, string) {
	//Finally delete the Entity
	ctx, cancel := u.Connect()
	c, _ := GetDB().Collection(entity).DeleteOne(ctx, req)
	if c.DeletedCount == 0 {
		return u.Message(false, "There was an error in deleting the entity"), "not found"
	}
	defer cancel()

	return u.Message(true, "success"), ""
}

func DeleteEntity(entity string, id primitive.ObjectID, rnd map[string]interface{}) (map[string]interface{}, string) {
	var t map[string]interface{}
	var e string
	eNum := u.EntityStrToInt(entity)
	if eNum > u.DEVICE {
		//Delete the non hierarchal objects
		t, e = GetEntityHierarchy(id, rnd, entity, eNum, eNum+eNum, u.RequestFilters{}, nil)
	} else {
		t, e = GetEntityHierarchy(id, rnd, entity, eNum, u.AC, u.RequestFilters{}, nil)
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
			DeleteDeviceF(t["id"].(primitive.ObjectID), nil)
		} else {
			if ent == u.DOMAIN {
				if t["name"] == os.Getenv("db") {
					return u.Message(false, "Cannot delete tenant's default domain"), "domain"
				}
				if domainHasObjects(t["hierarchyName"].(string)) {
					return u.Message(false, "Cannot delete domain if it has at least one object"), "domain"
				}
			}
			ctx, cancel := u.Connect()
			entity := u.EntityToString(ent)
			c, _ := GetDB().Collection(entity).DeleteOne(ctx, bson.M{"_id": t["id"].(primitive.ObjectID)})
			if c.DeletedCount == 0 {
				return u.Message(false, "No Records Found!"), "not found"
			}
			defer cancel()

		}
	}
	return nil, ""
}

func updateOldObjWithPatch(old map[string]interface{}, patch map[string]interface{}) string {
	for k, v := range patch {
		switch child := v.(type) {
		case map[string]interface{}:
			switch oldChild := old[k].(type) {
			case map[string]interface{}:
				updateOldObjWithPatch(oldChild, child)
			default:
				return "Wrong format for property " + k
			}
		default:
			old[k] = v
		}
	}
	return ""
}

func UpdateEntity(ent string, req bson.M, t map[string]interface{}, isPatch bool, userRoles map[string]Role) (map[string]interface{}, string) {
	var e *mongo.SingleResult
	updatedDoc := bson.M{}
	retDoc := options.ReturnDocument(options.After)
	entInt := u.EntityStrToInt(ent)

	//Update timestamp requires first obj retrieval
	//there isn't any way for mongoDB to make a field
	//immutable in a document
	oldObj, e1 := GetEntity(req, ent, u.RequestFilters{}, userRoles)
	if e1 != "" {
		if e1 == "permission" {
			return oldObj, e1
		}
		return u.Message(false, "Error: "+e1), e1
	}

	//Check if permission is only readonly
	if entInt != u.BLDGTMPL && entInt != u.ROOMTMPL && entInt != u.OBJTMPL &&
		(oldObj["description"] == nil) {
		// Description is always present, unless GetEntity was called with readonly permission
		return u.Message(false,
				"User does not have permission to change this object"),
			"permission"
	}

	t["lastUpdated"] = primitive.NewDateTimeFromTime(time.Now())
	t["createdDate"] = oldObj["createdDate"]

	// Update old object data with patch data
	if isPatch {
		var formattedOldObj map[string]interface{}
		// Convert primitive.A and similar types
		bytes, _ := json.Marshal(oldObj)
		json.Unmarshal(bytes, &formattedOldObj)
		// Update old with new
		e1 = updateOldObjWithPatch(formattedOldObj, t)
		if e1 != "" {
			return u.Message(false, "Error: "+e1), e1
		}
		t = formattedOldObj
		// Remove API set fields
		delete(t, "id")
		delete(t, "hierarchyName")
	}

	// Ensure the update is valid and apply it
	ctx, cancel := u.Connect()
	msg, ok := ValidateEntity(u.EntityStrToInt(ent), t)
	if !ok {
		return msg, "invalid"
	}

	// Check user permissions in case domain is being updated
	if entInt != u.DOMAIN && entInt != u.BLDGTMPL && entInt != u.ROOMTMPL && entInt != u.OBJTMPL &&
		(oldObj["domain"] != t["domain"]) {
		if permission := CheckUserPermissions(userRoles, entInt, t["domain"].(string)); permission < WRITE {
			return u.Message(false,
					"User does not have permission to change this object"),
				"permission"
		}
	}

	// Update database
	e = GetDB().Collection(ent).FindOneAndReplace(ctx,
		req, t,
		&options.FindOneAndReplaceOptions{ReturnDocument: &retDoc})
	if e.Err() != nil {
		return u.Message(false, "failure: "+e.Err().Error()), e.Err().Error()
	}

	// Changes to hierarchyName should be propagated to its children
	if oldObj["hierarchyName"] != t["hierarchyName"] {
		propagateParentNameChange(ctx, oldObj["hierarchyName"].(string),
			t["hierarchyName"].(string), u.EntityStrToInt(ent))
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
	req := bson.M{"hierarchyName": primitive.Regex{Pattern: oldParentName + u.HN_DELIMETER, Options: ""}}
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

func GetEntityHierarchy(ID primitive.ObjectID, req bson.M, ent string, start, end int, filters u.RequestFilters, userRoles map[string]Role) (map[string]interface{}, string) {
	var childEnt string

	if start < end {
		//We want to filter using RBAC requirements and the ID
		//The RBAC requirements are included in req
		newReq := req
		if req == nil {
			newReq = bson.M{"_id": ID}
		} else {
			newReq["_id"] = ID
		}

		top, e := GetEntity(newReq, ent, filters, userRoles)
		if top == nil {
			return nil, e
		}
		top = fixID(top)

		children := []map[string]interface{}{}
		pid := ID.Hex()
		//Get sensors & groups
		if ent == "rack" || ent == "device" {
			//Get sensors
			sensors, _ := GetManyEntities("sensor", bson.M{"parentId": pid}, filters, userRoles)

			//Get groups
			groups, _ := GetManyEntities("group", bson.M{"parentId": pid}, filters, userRoles)

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

		subEnts, _ := GetManyEntities(childEnt, bson.M{"parentId": pid}, filters, userRoles)

		for idx := range subEnts {
			tmp, _ := GetEntityHierarchy(subEnts[idx]["id"].(primitive.ObjectID), req, childEnt, start+1, end, filters, userRoles)
			if tmp != nil {
				subEnts[idx] = tmp
			}
		}

		if subEnts != nil {
			children = append(children, subEnts...)
		}

		if ent == "room" {
			for i := u.AC; i < u.CABINET+1; i++ {
				roomEnts, _ := GetManyEntities(u.EntityToString(i), bson.M{"parentId": pid}, filters, userRoles)
				if roomEnts != nil {
					children = append(children, roomEnts...)
				}
			}
			for i := u.PWRPNL; i < u.SENSOR+1; i++ {
				roomEnts, _ := GetManyEntities(u.EntityToString(i), bson.M{"parentId": pid}, filters, userRoles)
				if roomEnts != nil {
					children = append(children, roomEnts...)
				}
			}
			roomEnts, _ := GetManyEntities(u.EntityToString(u.CORRIDOR), bson.M{"parentId": pid}, filters, userRoles)
			if roomEnts != nil {
				children = append(children, roomEnts...)
			}
			roomEnts, _ = GetManyEntities(u.EntityToString(u.GROUP), bson.M{"parentId": pid}, filters, userRoles)
			if roomEnts != nil {
				children = append(children, roomEnts...)
			}
		}

		if ent == "stray_device" {
			sSensors, _ := GetManyEntities("stray_sensor", bson.M{"parentId": pid}, filters, userRoles)
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

func GetEntitiesUsingAncestorNames(ent string, id primitive.ObjectID, req map[string]interface{}, ancestry []map[string]string, userRoles map[string]Role) ([]map[string]interface{}, string) {

	newReq := req
	if newReq == nil {
		newReq = bson.M{"_id": id}
	} else {
		newReq["_id"] = id
	}
	top, e := GetEntity(newReq, ent, u.RequestFilters{}, nil)
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
				return GetManyEntities(k, bson.M{"parentId": pid}, u.RequestFilters{}, userRoles)
			}

			x, e1 = GetEntity(bson.M{"parentId": pid, "name": v}, k, u.RequestFilters{}, userRoles)
			if e1 != "" {
				println("Failing here")
				return nil, ""
			}
			pid = (x["id"].(primitive.ObjectID)).Hex()
		}
	}

	return nil, ""
}

func GetEntityUsingAncestorNames(req map[string]interface{}, ent string, ancestry []map[string]string) (map[string]interface{}, string) {
	top, e := GetEntity(req, ent, u.RequestFilters{}, nil)
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

			x, e1 = GetEntity(bson.M{"parentId": pid, "name": v}, k, u.RequestFilters{}, nil)
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
		children, e1 := GetManyEntities(checkEntName, bson.M{"hierarchyName": pattern}, filters, nil)
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
	if parentEntStr == "domain" {
		// device special case (devices can have devices)
		startEnt = u.DOMAIN
		endEnt = u.DOMAIN
	} else if parentEntStr == "device" {
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

func GetEntitiesUsingSiteAsAncestor(ent, id string, req map[string]interface{}, ancestry []map[string]string, userRoles map[string]Role) ([]map[string]interface{}, string) {

	newReq := req
	if newReq == nil {
		newReq = bson.M{"name": id}
	} else {
		newReq["name"] = id
	}
	top, e := GetEntity(newReq, ent, u.RequestFilters{}, nil)
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
				return GetManyEntities(k, bson.M{"parentId": pid}, u.RequestFilters{}, userRoles)
			}

			x, e1 = GetEntity(bson.M{"parentId": pid, "name": v}, k, u.RequestFilters{}, userRoles)
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

func GetEntityUsingSiteAsAncestor(req map[string]interface{}, ent string, ancestry []map[string]string) (map[string]interface{}, string) {
	top, e := GetEntity(req, ent, u.RequestFilters{}, nil)
	if e != "" {
		return nil, e
	}

	pid := (top["id"].(primitive.ObjectID)).Hex()

	var x map[string]interface{}
	var e1 string
	for i := range ancestry {
		for k, v := range ancestry[i] {

			println("KEY:", k, " VAL:", v)

			x, e1 = GetEntity(bson.M{"parentId": pid, "name": v}, k, u.RequestFilters{}, nil)
			if e1 != "" {
				println("Failing here")
				return nil, ""
			}
			pid = (x["id"].(primitive.ObjectID)).Hex()
		}
	}

	return x, ""
}

func GetEntitiesOfAncestor(id interface{}, req bson.M, ent int, entStr, wantedEnt string) ([]map[string]interface{}, string) {
	var ans []map[string]interface{}
	var t map[string]interface{}
	var e, e1 string
	newReq := req
	if ent == u.SITE {

		if newReq == nil {
			newReq = bson.M{"name": id}
		} else {
			newReq["name"] = id
		}

		t, e = GetEntity(newReq, "site", u.RequestFilters{}, nil)
		if e != "" {
			return nil, e
		}

	} else {
		ID, _ := primitive.ObjectIDFromHex(id.(string))

		//Apply the RBAC filter
		if newReq == nil {
			newReq = bson.M{"_id": ID}
		} else {
			newReq["_id"] = ID
		}

		t, e = GetEntity(newReq, entStr, u.RequestFilters{}, nil)
		if e != "" {
			return nil, e
		}
	}

	sub, e1 := GetManyEntities(u.EntityToString(ent+1),
		bson.M{"parentId": t["id"].(primitive.ObjectID).Hex()}, u.RequestFilters{}, nil)
	if e1 != "" {
		return nil, e1
	}

	if wantedEnt == "" {
		wantedEnt = u.EntityToString(ent + 2)
	}

	for i := range sub {
		x, _ := GetManyEntities(wantedEnt,
			bson.M{"parentId": sub[i]["id"].(primitive.ObjectID).Hex()}, u.RequestFilters{}, nil)
		ans = append(ans, x...)
	}
	return ans, ""
}

//DEV FAMILY FUNCS

func DeleteDeviceF(entityID primitive.ObjectID, req bson.M) (map[string]interface{}, string) {
	t, e := GetEntityHierarchy(entityID, req, "device", 0, 999, u.RequestFilters{}, nil)
	if e != "" {
		return u.Message(false,
			"There was an error in deleting the entity"), "not found"
	}

	return deleteDeviceHelper(t)
}

func deleteDeviceHelper(t map[string]interface{}) (map[string]interface{}, string) {
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

func ExtractCursor(c *mongo.Cursor, ctx context.Context, entity int, userRoles map[string]Role) ([]map[string]interface{}, string) {
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
		if entity != u.BLDGTMPL && entity != u.ROOMTMPL && entity != u.OBJTMPL && userRoles != nil {
			//Check permissions
			var domain string
			if entity == u.DOMAIN {
				fmt.Println(x)
				domain = x["hierarchyName"].(string)
			} else {
				domain = x["domain"].(string)
			}
			if permission := CheckUserPermissions(userRoles, entity, domain); permission >= READONLYNAME {
				if permission == READONLYNAME {
					x = FixReadOnlyName(x)
				}
				ans = append(ans, x)
			}
		} else {
			ans = append(ans, x)
		}

	}
	return ans, ""
}
