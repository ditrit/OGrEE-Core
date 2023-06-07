package models

import (
	"context"
	"encoding/json"
	"errors"
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

func CreateEntity(entity int, t map[string]interface{}, userRoles map[string]Role) (map[string]interface{}, *u.Error) {
	message := ""
	if ok, err := ValidateEntity(entity, t); !ok {
		return nil, err
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
			return nil, &u.Error{Type: u.ErrUnauthorized,
				Message: "User does not have permission to create this object"}
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
			return nil, &u.Error{Type: u.ErrDuplicate,
				Message: "Error while creating " + entStr + ": Duplicates not allowed"}
		}
		return nil, &u.Error{Type: u.ErrDBError,
			Message: "Internal error while creating " + entStr + ": " + e.Error()}
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

	resp := u.Message(message)
	resp["data"] = t
	return resp, nil
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

func GetEntity(req bson.M, ent string, filters u.RequestFilters, userRoles map[string]Role) (map[string]interface{}, *u.Error) {
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
		return nil, &u.Error{Type: u.ErrBadFormat, Message: e.Error()}
	}

	if opts != nil {
		e = GetDB().Collection(ent).FindOne(ctx, req, opts).Decode(&t)
	} else {
		e = GetDB().Collection(ent).FindOne(ctx, req).Decode(&t)
	}
	if e != nil {
		return nil, &u.Error{Type: u.ErrBadFormat, Message: e.Error()}
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
				return nil, &u.Error{Type: u.ErrUnauthorized,
					Message: "User does not have permission to see this object"}
			} else if permission == READONLYNAME {
				t = FixReadOnlyName(t)
			}
		}
	}

	//If entity has '_' remove it
	if strings.Contains(ent, "_") {
		FixUnderScore(t)
	}
	return t, nil
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

func GetManyEntities(ent string, req bson.M, filters u.RequestFilters, userRoles map[string]Role) ([]map[string]interface{}, *u.Error) {
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
		return nil, &u.Error{Type: u.ErrBadFormat, Message: err.Error()}
	}

	if opts != nil {
		c, err = GetDB().Collection(ent).Find(ctx, req, opts)
	} else {
		c, err = GetDB().Collection(ent).Find(ctx, req)
	}
	if err != nil {
		fmt.Println(err)
		return nil, &u.Error{Type: u.ErrDBError, Message: err.Error()}
	}
	defer cancel()

	data, e1 := ExtractCursor(c, ctx, u.EntityStrToInt(ent), userRoles)
	if e1 != nil {
		fmt.Println(e1)
		return nil, &u.Error{Type: u.ErrInternal, Message: e1.Error()}
	}

	//Remove underscore If the entity has '_'
	if strings.Contains(ent, "_") {
		for i := range data {
			FixUnderScore(data[i])
		}
	}

	return data, nil
}

// GetCompleteHierarchy: gets all objects in db using hierachyName and returns:
//   - tree: map with parents as key and their children as an array value
//     tree: {parent:[children]}
//   - categories: map with category name as key and corresponding objects
//     as an array value
//     categories: {categoryName:[children]}
func GetCompleteDomainHierarchy(userRoles map[string]Role) (map[string]interface{}, *u.Error) {
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
		return nil, &u.Error{Type: u.ErrDBError, Message: err.Error()}
	}
	data, e := ExtractCursor(c, ctx, u.EntityStrToInt(collName), userRoles)
	if e != nil {
		return nil, &u.Error{Type: u.ErrInternal, Message: e.Error()}
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
	return response, nil
}

// GetCompleteHierarchy: gets all objects in db using hierachyName and returns:
//   - tree: map with parents as key and their children as an array value
//     tree: {parent:[children]}
//   - categories: map with category name as key and corresponding objects
//     as an array value
//     categories: {categoryName:[children]}
func GetCompleteHierarchy(userRoles map[string]Role) (map[string]interface{}, *u.Error) {
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
			return nil, &u.Error{Type: u.ErrDBError, Message: err.Error()}
		}
		data, e := ExtractCursor(c, ctx, u.EntityStrToInt(collName), userRoles)
		if e != nil {
			return nil, &u.Error{Type: u.ErrInternal, Message: e.Error()}
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
	return response, nil
}

// fillHierarchyMap: add hierarchyName to the children array of its parent
func fillHierarchyMap(hierarchyName string, data map[string][]string) {
	i := strings.LastIndex(hierarchyName, u.HN_DELIMETER)
	if i > 0 {
		parent := hierarchyName[:i]
		data[parent] = append(data[parent], hierarchyName)
	}
}

func GetCompleteHierarchyAttributes(userRoles map[string]Role) (map[string]interface{}, *u.Error) {
	response := make(map[string]interface{})
	// Get all collections names
	ctx, cancel := u.Connect()
	db := GetDB()
	collNames, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		fmt.Println(err.Error())
		return nil, &u.Error{Type: u.ErrDBError, Message: err.Error()}

	}

	for _, collName := range collNames {
		if entInt := u.EntityStrToInt(collName); entInt > -1 {
			projection := bson.D{{Key: "hierarchyName", Value: 1}, {Key: "attributes", Value: 1},
				{Key: "domain", Value: 1}}

			opts := options.Find().SetProjection(projection)

			c, err := db.Collection(collName).Find(ctx, bson.M{}, opts)
			if err != nil {
				println(err.Error())
				return nil, &u.Error{Type: u.ErrDBError, Message: err.Error()}
			}
			data, e := ExtractCursor(c, ctx, entInt, userRoles)
			if e != nil {
				return nil, &u.Error{Type: u.ErrInternal, Message: e.Error()}
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
	}
	defer cancel()
	return response, nil
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
func DeleteEntityByName(entity string, name string, userRoles map[string]Role) *u.Error {
	if entity == "domain" {
		if name == os.Getenv("db") {
			return &u.Error{Type: u.ErrForbidden, Message: "Cannot delete tenant's default domain"}
		}
		if domainHasObjects(name) {
			return &u.Error{Type: u.ErrForbidden, Message: "Cannot delete domain if it has at least one object"}
		}
	}

	req, ok := GetRequestFilterByDomain(userRoles)
	if !ok {
		return &u.Error{Type: u.ErrUnauthorized, Message: "User does not have permission to delete"}
	}
	req["hierarchyName"] = name
	err := DeleteSingleEntity(entity, req)

	if err != nil {
		// Unable to delete given object
		return err
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

	return nil
}

func DeleteSingleEntity(entity string, req bson.M) *u.Error {
	//Finally delete the Entity
	ctx, cancel := u.Connect()
	c, _ := GetDB().Collection(entity).DeleteOne(ctx, req)
	if c.DeletedCount == 0 {
		return &u.Error{Type: u.ErrNotFound, Message: "Error deleting object: not found"}
	}
	defer cancel()

	return nil
}

func DeleteEntity(entity string, id primitive.ObjectID, rnd map[string]interface{}) *u.Error {
	var t map[string]interface{}
	var e *u.Error
	eNum := u.EntityStrToInt(entity)
	if eNum > u.DEVICE {
		//Delete the non hierarchal objects
		t, e = GetEntityHierarchy(id, rnd, entity, eNum, eNum+eNum, u.RequestFilters{}, nil)
	} else {
		t, e = GetEntityHierarchy(id, rnd, entity, eNum, u.AC, u.RequestFilters{}, nil)
	}

	if e != nil {
		return e
	}

	return deleteHelper(t, eNum)
}

func deleteHelper(t map[string]interface{}, ent int) *u.Error {
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
					return &u.Error{Type: u.ErrForbidden,
						Message: "Cannot delete tenant's default domain"}
				}
				if domainHasObjects(t["hierarchyName"].(string)) {
					return &u.Error{Type: u.ErrForbidden,
						Message: "Cannot delete domain if it has at least one object"}
				}
			}
			ctx, cancel := u.Connect()
			entity := u.EntityToString(ent)
			c, _ := GetDB().Collection(entity).DeleteOne(ctx, bson.M{"_id": t["id"].(primitive.ObjectID)})
			if c.DeletedCount == 0 {
				return &u.Error{Type: u.ErrNotFound,
					Message: "No Records Found!"}
			}
			defer cancel()

		}
	}
	return nil
}

func updateOldObjWithPatch(old map[string]interface{}, patch map[string]interface{}) error {
	for k, v := range patch {
		switch child := v.(type) {
		case map[string]interface{}:
			switch oldChild := old[k].(type) {
			case map[string]interface{}:
				updateOldObjWithPatch(oldChild, child)
			default:
				return errors.New("Wrong format for property " + k)
			}
		default:
			old[k] = v
		}
	}
	return nil
}

func UpdateEntity(ent string, req bson.M, t map[string]interface{}, isPatch bool, userRoles map[string]Role) (map[string]interface{}, *u.Error) {
	var mongoRes *mongo.SingleResult
	updatedDoc := bson.M{}
	retDoc := options.ReturnDocument(options.After)
	entInt := u.EntityStrToInt(ent)

	//Update timestamp requires first obj retrieval
	//there isn't any way for mongoDB to make a field
	//immutable in a document
	oldObj, err := GetEntity(req, ent, u.RequestFilters{}, userRoles)
	if err != nil {
		return nil, err
	}

	//Check if permission is only readonly
	if entInt != u.BLDGTMPL && entInt != u.ROOMTMPL && entInt != u.OBJTMPL &&
		(oldObj["description"] == nil) {
		// Description is always present, unless GetEntity was called with readonly permission
		return nil, &u.Error{Type: u.ErrUnauthorized,
			Message: "User does not have permission to change this object"}
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
		err := updateOldObjWithPatch(formattedOldObj, t)
		if err != nil {
			return nil, &u.Error{Type: u.ErrBadFormat, Message: err.Error()}
		}
		t = formattedOldObj
		// Remove API set fields
		delete(t, "id")
		delete(t, "hierarchyName")
	}

	// Ensure the update is valid and apply it
	ctx, cancel := u.Connect()
	if ok, err := ValidateEntity(u.EntityStrToInt(ent), t); !ok {
		return nil, err
	}

	// Check user permissions in case domain is being updated
	if entInt != u.DOMAIN && entInt != u.BLDGTMPL && entInt != u.ROOMTMPL && entInt != u.OBJTMPL &&
		(oldObj["domain"] != t["domain"]) {
		if permission := CheckUserPermissions(userRoles, entInt, t["domain"].(string)); permission < WRITE {
			return nil, &u.Error{Type: u.ErrUnauthorized,
				Message: "User does not have permission to change this object"}
		}
	}

	// Update database
	mongoRes = GetDB().Collection(ent).FindOneAndReplace(ctx,
		req, t,
		&options.FindOneAndReplaceOptions{ReturnDocument: &retDoc})
	if mongoRes.Err() != nil {
		return nil, &u.Error{Type: u.ErrUnauthorized,
			Message: mongoRes.Err().Error()}
	}

	// Changes to hierarchyName should be propagated to its children
	if oldObj["hierarchyName"] != t["hierarchyName"] {
		propagateParentNameChange(ctx, oldObj["hierarchyName"].(string),
			t["hierarchyName"].(string), u.EntityStrToInt(ent))
	}

	//Obtain new document then
	//Fix the _id / id discrepancy
	mongoRes.Decode(&updatedDoc)
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
	resp := u.Message(message)
	resp["data"] = updatedDoc
	return resp, nil
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

func GetEntityHierarchy(ID primitive.ObjectID, req bson.M, ent string, start, end int, filters u.RequestFilters, userRoles map[string]Role) (map[string]interface{}, *u.Error) {
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

		return top, nil
	}
	return nil, nil
}

func GetEntitiesUsingAncestorNames(ent string, id primitive.ObjectID, req map[string]interface{},
	ancestry []map[string]string, userRoles map[string]Role) ([]map[string]interface{}, *u.Error) {

	newReq := req
	if newReq == nil {
		newReq = bson.M{"_id": id}
	} else {
		newReq["_id"] = id
	}
	top, e := GetEntity(newReq, ent, u.RequestFilters{}, nil)
	if e != nil {
		return nil, e
	}

	//Remove _id
	top = fixID(top)

	pid := (top["id"].(primitive.ObjectID)).Hex()

	for i := range ancestry {
		for k, v := range ancestry[i] {
			println("KEY:", k, " VAL:", v)
			if v == "all" {
				println("K:", k)
				return GetManyEntities(k, bson.M{"parentId": pid}, u.RequestFilters{}, userRoles)
			}

			x, e1 := GetEntity(bson.M{"parentId": pid, "name": v}, k, u.RequestFilters{}, userRoles)
			if e1 != nil {
				return nil, e1
			}
			pid = (x["id"].(primitive.ObjectID)).Hex()
		}
	}

	return nil, nil
}

func GetEntityUsingAncestorNames(req map[string]interface{}, ent string,
	ancestry []map[string]string) (map[string]interface{}, *u.Error) {
	top, e := GetEntity(req, ent, u.RequestFilters{}, nil)
	if e != nil {
		return nil, e
	}

	//Remove _id
	top = fixID(top)

	pid := (top["id"].(primitive.ObjectID)).Hex()

	var x map[string]interface{}
	var e1 *u.Error
	for i := range ancestry {
		for k, v := range ancestry[i] {

			println("KEY:", k, " VAL:", v)

			x, e1 = GetEntity(bson.M{"parentId": pid, "name": v}, k, u.RequestFilters{}, nil)
			if e1 != nil {
				return nil, e1
			}
			pid = (x["id"].(primitive.ObjectID)).Hex()
		}
	}

	//Remove _id
	x = fixID(x)
	return x, nil
}

// GetHierarchyByName: get children objects of given parent.
// - Param limit: max relationship distance between parent and child, example:
// limit=1 only direct children, limit=2 includes nested children of children
func GetHierarchyByName(entity, hierarchyName string, limit int, filters u.RequestFilters) ([]map[string]interface{}, *u.Error) {
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
		if e1 != nil {
			println("SUBENT: ", checkEntName)
			println("ERR: ", e1.Message)
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
	return recursivelyGetChildrenFromMaps(hierarchyName, hierarchy, allChildren), nil
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

func GetEntitiesUsingSiteAsAncestor(ent, id string, req map[string]interface{}, ancestry []map[string]string,
	userRoles map[string]Role) ([]map[string]interface{}, *u.Error) {

	newReq := req
	if newReq == nil {
		newReq = bson.M{"name": id}
	} else {
		newReq["name"] = id
	}
	top, e := GetEntity(newReq, ent, u.RequestFilters{}, nil)
	if e != nil {
		return nil, e
	}

	//Remove _id
	top = fixID(top)

	pid := (top["id"].(primitive.ObjectID)).Hex()

	println("ANCS-LEN: ", len(ancestry))
	for i := range ancestry {
		for k, v := range ancestry[i] {

			println("KEY:", k, " VAL:", v)

			if v == "all" {
				println("K:", k)
				return GetManyEntities(k, bson.M{"parentId": pid}, u.RequestFilters{}, userRoles)
			}

			x, e1 := GetEntity(bson.M{"parentId": pid, "name": v}, k, u.RequestFilters{}, userRoles)
			if e1 != nil {
				println("Failing here")
				println("E1: ", e1)
				return nil, e1
			}
			pid = (x["id"].(primitive.ObjectID)).Hex()
		}
	}

	return nil, nil
}

func GetEntityUsingSiteAsAncestor(req map[string]interface{}, ent string, ancestry []map[string]string) (map[string]interface{}, *u.Error) {
	top, e := GetEntity(req, ent, u.RequestFilters{}, nil)
	if e != nil {
		return nil, e
	}

	pid := (top["id"].(primitive.ObjectID)).Hex()

	var x map[string]interface{}
	var e1 *u.Error
	for i := range ancestry {
		for k, v := range ancestry[i] {
			println("KEY:", k, " VAL:", v)
			x, e1 = GetEntity(bson.M{"parentId": pid, "name": v}, k, u.RequestFilters{}, nil)
			if e1 != nil {
				return nil, e1
			}
			pid = (x["id"].(primitive.ObjectID)).Hex()
		}
	}

	return x, nil
}

func GetEntitiesOfAncestor(id interface{}, req bson.M, ent int, entStr, wantedEnt string) ([]map[string]interface{}, *u.Error) {
	var ans []map[string]interface{}
	var t map[string]interface{}
	var e, e1 *u.Error
	newReq := req
	if ent == u.SITE {

		if newReq == nil {
			newReq = bson.M{"name": id}
		} else {
			newReq["name"] = id
		}

		t, e = GetEntity(newReq, "site", u.RequestFilters{}, nil)
		if e != nil {
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
		if e != nil {
			return nil, e
		}
	}

	sub, e1 := GetManyEntities(u.EntityToString(ent+1),
		bson.M{"parentId": t["id"].(primitive.ObjectID).Hex()}, u.RequestFilters{}, nil)
	if e1 != nil {
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
	return ans, nil
}

//DEV FAMILY FUNCS

func DeleteDeviceF(entityID primitive.ObjectID, req bson.M) (map[string]interface{}, *u.Error) {
	t, e := GetEntityHierarchy(entityID, req, "device", 0, 999, u.RequestFilters{}, nil)
	if e != nil {
		return nil, e
	}

	return deleteDeviceHelper(t)
}

func deleteDeviceHelper(t map[string]interface{}) (map[string]interface{}, *u.Error) {
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
			return nil, &u.Error{Type: u.ErrNotFound, Message: "Error deleting object: not found"}
		}
		defer cancel()

	}
	return nil, nil
}

func ExtractCursor(c *mongo.Cursor, ctx context.Context, entity int, userRoles map[string]Role) ([]map[string]interface{}, error) {
	ans := []map[string]interface{}{}
	for c.Next(ctx) {
		x := map[string]interface{}{}
		err := c.Decode(x)
		if err != nil {
			fmt.Println(err.Error())
			return nil, err
		}
		//Remove _id
		x = fixID(x)
		if entity != u.BLDGTMPL && entity != u.ROOMTMPL && entity != u.OBJTMPL && userRoles != nil {
			//Check permissions
			var domain string
			if entity == u.DOMAIN {
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
	return ans, nil
}
