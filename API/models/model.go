package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
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

// Helper functions

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

// fillHierarchyMap: add obj id (hierarchyName) to the children array of its parent
func fillHierarchyMap(hierarchyName string, data map[string][]string) {
	i := strings.LastIndex(hierarchyName, u.HN_DELIMETER)
	if i > 0 {
		parent := hierarchyName[:i]
		data[parent] = append(data[parent], hierarchyName)
	}
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

// PropagateParentIdChange: search for given parent children and
// update their hierarchyName with new parent name
func PropagateParentIdChange(ctx context.Context, oldParentId, newId string, entityInt int) error {
	// Find all objects containing parent name
	req := bson.M{"id": primitive.Regex{Pattern: oldParentId + u.HN_DELIMETER, Options: ""}}
	// For each object found, replace old name by new
	update := bson.D{{
		Key: "$set", Value: bson.M{
			"id": bson.M{
				"$replaceOne": bson.M{
					"input":       "$id",
					"find":        oldParentId,
					"replacement": newId}}}}}
	if entityInt == u.DOMAIN {
		_, e := GetDB().Collection(u.EntityToString(u.DOMAIN)).UpdateMany(ctx,
			req, mongo.Pipeline{update})
		if e != nil {
			println(e.Error())
			return e
		}
	} else if entityInt == u.DEVICE {
		_, e := GetDB().Collection(u.EntityToString(u.DEVICE)).UpdateMany(ctx,
			req, mongo.Pipeline{update})
		if e != nil {
			println(e.Error())
			return e
		}
	} else {
		for i := entityInt + 1; i <= u.GROUP; i++ {
			_, e := GetDB().Collection(u.EntityToString(i)).UpdateMany(ctx,
				req, mongo.Pipeline{update})
			if e != nil {
				println(e.Error())
				return e
			}
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

// Entity handlers

func CreateEntity(entity int, t map[string]interface{}, userRoles map[string]Role) (map[string]interface{}, *u.Error) {
	if err := prepareCreateEntity(entity, t, userRoles); err != nil {
		return nil, err
	}

	ctx, cancel := u.Connect()
	entStr := u.EntityToString(entity)
	_, e := GetDB().Collection(entStr).InsertOne(ctx, t)
	if e != nil {
		if strings.Contains(e.Error(), "E11000") {
			return nil, &u.Error{Type: u.ErrDuplicate,
				Message: "Error while creating " + entStr + ": Duplicates not allowed"}
		}
		return nil, &u.Error{Type: u.ErrDBError,
			Message: "Internal error while creating " + entStr + ": " + e.Error()}
	}
	defer cancel()

	fixID(t)
	return t, nil
}

func prepareCreateEntity(entity int, t map[string]interface{}, userRoles map[string]Role) *u.Error {
	if ok, err := ValidateEntity(entity, t); !ok {
		return err
	}

	// Check user permissions
	if entity != u.BLDGTMPL && entity != u.ROOMTMPL && entity != u.OBJTMPL {
		var domain string
		if entity == u.DOMAIN {
			domain = t["id"].(string)
		} else {
			domain = t["domain"].(string)
		}
		if permission := CheckUserPermissions(userRoles, entity, domain); permission < WRITE {
			return &u.Error{Type: u.ErrUnauthorized,
				Message: "User does not have permission to create this object"}
		}
	}

	//Set timestamp
	t["createdDate"] = primitive.NewDateTimeFromTime(time.Now())
	t["lastUpdated"] = t["createdDate"]

	//Last modifications before insert
	FixAttributesBeforeInsert(entity, t)
	delete(t, "parentId")
	return nil
}

// GetObjectById: search for id (hierarchyName) in all possible collections
func GetObjectById(hierarchyName string, filters u.RequestFilters, userRoles map[string]Role) (map[string]interface{}, *u.Error) {
	var resp map[string]interface{}
	// Get possible collections for this name
	rangeEntities := u.HierachyNameToEntity(hierarchyName)

	// Search each collection
	for _, entity := range rangeEntities {
		req := bson.M{"id": hierarchyName}
		entityStr := u.EntityToString(entity)
		data, _ := GetEntity(req, entityStr, filters, userRoles)
		if data != nil {
			resp = data
			break
		}
	}

	if resp != nil {
		return resp, nil
	} else {
		return nil, &u.Error{Type: u.ErrNotFound, Message: "Unable to find object"}
	}
}

func GetEntity(req bson.M, ent string, filters u.RequestFilters, userRoles map[string]Role) (map[string]interface{}, *u.Error) {
	t := map[string]interface{}{}
	ctx, cancel := u.Connect()
	var e error

	var opts *options.FindOneOptions
	if len(filters.FieldsToShow) > 0 {
		compoundIndex := bson.D{bson.E{Key: "domain", Value: 1}, bson.E{Key: "id", Value: 1}}
		for _, field := range filters.FieldsToShow {
			if field != "domain" && field != "id" {
				compoundIndex = append(compoundIndex, bson.E{Key: field, Value: 1})
			}
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
		if e == mongo.ErrNoDocuments {
			return nil, &u.Error{Type: u.ErrNotFound,
				Message: "Nothing matches this request"}
		}
		return nil, &u.Error{Type: u.ErrBadFormat, Message: e.Error()}
	}
	defer cancel()

	//Remove _id
	t = fixID(t)

	// Check permissions
	if !strings.Contains(ent, "template") {
		var domain string
		if ent == "domain" {
			domain = t["id"].(string)
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

func GetManyEntities(ent string, req bson.M, filters u.RequestFilters, userRoles map[string]Role) ([]map[string]interface{}, *u.Error) {
	ctx, cancel := u.Connect()
	var err error
	var c *mongo.Cursor

	var opts *options.FindOptions
	if len(filters.FieldsToShow) > 0 {
		compoundIndex := bson.D{bson.E{Key: "domain", Value: 1}, bson.E{Key: "id", Value: 1}}
		for _, field := range filters.FieldsToShow {
			if field != "domain" && field != "id" {
				compoundIndex = append(compoundIndex, bson.E{Key: field, Value: 1})
			}
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

	c, err := db.Collection(collName).Find(ctx, bson.M{})
	if err != nil {
		println(err.Error())
		return nil, &u.Error{Type: u.ErrDBError, Message: err.Error()}
	}
	data, e := ExtractCursor(c, ctx, u.EntityStrToInt(collName), userRoles)
	if e != nil {
		return nil, &u.Error{Type: u.ErrInternal, Message: e.Error()}
	}

	for _, obj := range data {
		if strings.Contains(obj["id"].(string), ".") {
			fillHierarchyMap(obj["id"].(string), hierarchy)
		} else {
			hierarchy["Root"] = append(hierarchy["Root"], obj["id"].(string))
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
		opts := options.Find().SetProjection(bson.D{{Key: "domain", Value: 1}, {Key: "id", Value: 1}})

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
				categories[rootCollectionName] = append(categories[rootCollectionName], obj["id"].(string))
				hierarchy["Root"] = append(hierarchy["Root"], obj["id"].(string))

			} else if obj["id"] != nil {
				categories[collName] = append(categories[collName], obj["id"].(string))
				fillHierarchyMap(obj["id"].(string), hierarchy)
			}
		}
	}

	categories["KeysOrder"] = []string{"site", "building", "room", "rack"}
	response["tree"] = hierarchy
	response["categories"] = categories
	defer cancel()
	return response, nil
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
			projection := bson.D{{Key: "attributes", Value: 1},
				{Key: "domain", Value: 1}, {Key: "id", Value: 1}}

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
					if id, isStr := obj["id"].(string); isStr && id != "" {
						response[obj["id"].(string)] = obj["attributes"]
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

// GetSiteParentTempUnit: search for the object of given ID,
// then search for its site parent and return its attributes.temperatureUnit
func GetSiteParentTempUnit(id string) (string, *u.Error) {
	data := map[string]interface{}{}

	// Get all collections names
	ctx, cancel := u.Connect()
	db := GetDB()
	collNames, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		fmt.Println(err.Error())
		return "", &u.Error{Type: u.ErrDBError, Message: err.Error()}
	}
	// Find object
	for _, collName := range collNames {
		err := db.Collection(collName).FindOne(ctx, bson.M{"id": id}).Decode(&data)
		if err == nil {
			// Found object with given id
			if data["category"].(string) == "site" {
				// it's a site
				break
			} else {
				// Find its parent site
				nameSlice := strings.Split(data["id"].(string), u.HN_DELIMETER)
				siteName := nameSlice[0] // CONSIDER SITE AS 0
				err := db.Collection("site").FindOne(ctx, bson.M{"id": siteName}).Decode(&data)
				if err != nil {
					return "", &u.Error{Type: u.ErrNotFound,
						Message: "Could not find parent site for given object"}
				}
			}
		}
	}

	defer cancel()

	if len(data) == 0 {
		return "", &u.Error{Type: u.ErrNotFound, Message: "No object found with given id"}
	} else if tempUnit := data["attributes"].(map[string]interface{})["temperatureUnit"]; tempUnit == nil {
		return "", &u.Error{Type: u.ErrNotFound,
			Message: "Parent site has no temperatureUnit in attributes"}
	} else {
		return tempUnit.(string), nil
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

	for i := 0; i <= u.BLDGTMPL; i++ {
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

// DeleteEntity: delete object of given hierarchyName
// search for all its children and delete them too, return:
// - success or fail message map
func DeleteEntity(entity string, id string, userRoles map[string]Role) *u.Error {
	// Special check for delete domain
	if entity == "domain" {
		if id == os.Getenv("db") {
			return &u.Error{Type: u.ErrForbidden, Message: "Cannot delete tenant's default domain"}
		}
		if domainHasObjects(id) {
			return &u.Error{Type: u.ErrForbidden, Message: "Cannot delete domain if it has at least one object"}
		}
	}

	// Delete with given id
	req, ok := GetRequestFilterByDomain(userRoles)
	if !ok {
		return &u.Error{Type: u.ErrUnauthorized, Message: "User does not have permission to delete"}
	}
	req["id"] = id
	err := DeleteSingleEntity(entity, req)

	if err != nil {
		// Unable to delete given id
		return err
	} else {
		// Delete possible children
		rangeEntities := getChildrenCollections(u.GROUP, entity)
		for _, childEnt := range rangeEntities {
			childEntName := u.EntityToString(childEnt)
			pattern := primitive.Regex{Pattern: "^" + id + u.HN_DELIMETER, Options: ""}

			ctx, cancel := u.Connect()
			GetDB().Collection(childEntName).DeleteMany(ctx,
				bson.M{"id": pattern})
			defer cancel()
		}
	}

	return nil
}

func DeleteSingleEntity(entity string, req bson.M) *u.Error {
	ctx, cancel := u.Connect()
	c, _ := GetDB().Collection(entity).DeleteOne(ctx, req)
	if c.DeletedCount == 0 {
		return &u.Error{Type: u.ErrNotFound, Message: "Error deleting object: not found"}
	}
	defer cancel()
	return nil
}

func UpdateEntity(ent string, req bson.M, t map[string]interface{}, isPatch bool, userRoles map[string]Role) (map[string]interface{}, *u.Error) {
	var mongoRes *mongo.SingleResult
	var updatedDoc map[string]interface{}
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
		delete(t, "lastUpdated")
		delete(t, "createdDate")
	}

	// Ensure the update is valid
	ctx, cancel := u.Connect()
	if ok, err := ValidateEntity(u.EntityStrToInt(ent), t); !ok {
		return nil, err
	}

	t["lastUpdated"] = primitive.NewDateTimeFromTime(time.Now())
	t["createdDate"] = oldObj["createdDate"]
	delete(t, "parentId")

	// Check user permissions in case domain is being updated
	if entInt != u.DOMAIN && entInt < u.ROOMTMPL && (oldObj["domain"] != t["domain"]) {
		if perm := CheckUserPermissions(userRoles, entInt, t["domain"].(string)); perm < WRITE {
			return nil, &u.Error{Type: u.ErrUnauthorized,
				Message: "User does not have permission to change this object"}
		}
	}

	// Update database callback
	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		mongoRes := GetDB().Collection(ent).FindOneAndReplace(ctx,
			req, t,
			&options.FindOneAndReplaceOptions{ReturnDocument: &retDoc})
		if mongoRes.Err() != nil {
			return nil, mongoRes.Err()
		}
		if oldObj["id"] != t["id"] {
			// Changes to id should be propagated to its children
			if err := PropagateParentIdChange(ctx, oldObj["id"].(string),
				t["id"].(string), entInt); err != nil {
				return nil, err
			}
		}

		return mongoRes, nil
	}

	// Start a session and run the callback to update db
	session, e := GetClient().StartSession()
	if e != nil {
		return nil, &u.Error{Type: u.ErrDBError, Message: "Unable to start session: " + e.Error()}
	}
	defer session.EndSession(ctx)
	result, e := session.WithTransaction(ctx, callback)
	if e != nil {
		return nil, &u.Error{Type: u.ErrDBError, Message: "Unable to complete transaction: " + e.Error()}
	}

	mongoRes = result.(*mongo.SingleResult)
	mongoRes.Decode(&updatedDoc)
	updatedDoc = fixID(updatedDoc)

	defer cancel()
	return updatedDoc, nil
}

// GetHierarchyByName: get children objects of given parent.
// - Param limit: max relationship distance between parent and child, example:
// limit=1 only direct children, limit=2 includes nested children of children
func GetHierarchyByName(entity, hierarchyName string, limit int, filters u.RequestFilters) ([]map[string]interface{}, *u.Error) {
	// Get all children and their relations
	allChildren, hierarchy, err := getChildren(entity, hierarchyName, limit, filters)
	if err != nil {
		return nil, err
	}

	// Organize the family according to relations (nest children)
	return recursivelyGetChildrenFromMaps(hierarchyName, hierarchy, allChildren), nil
}

func getChildren(entity, hierarchyName string, limit int, filters u.RequestFilters) (map[string]interface{},
	map[string][]string, *u.Error) {
	allChildren := map[string]interface{}{}
	hierarchy := make(map[string][]string)

	// Define in which collections we can find children
	rangeEntities := getChildrenCollections(limit, entity)

	// Get children from all given collections
	for _, checkEnt := range rangeEntities {
		checkEntName := u.EntityToString(checkEnt)
		// Obj should include parentName and not surpass limit range
		pattern := primitive.Regex{Pattern: "^" + hierarchyName +
			"(.[A-Za-z0-9_\" \"]+){1," + strconv.Itoa(limit) + "}$", Options: ""}
		children, e1 := GetManyEntities(checkEntName, bson.M{"id": pattern}, filters, nil)
		if e1 != nil {
			println("SUBENT: ", checkEntName)
			println("ERR: ", e1.Message)
			return nil, nil, e1
		}
		for _, child := range children {
			// store child data
			allChildren[child["id"].(string)] = child
			// create hierarchy map
			fillHierarchyMap(child["id"].(string), hierarchy)
		}
	}

	return allChildren, hierarchy, nil
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

func GetEntitiesOfAncestor(id string, entStr, wantedEnt string, userRoles map[string]Role) ([]map[string]interface{}, *u.Error) {
	// Get parent object
	req := bson.M{"id": id}
	_, e := GetEntity(req, entStr, u.RequestFilters{}, userRoles)
	if e != nil {
		return nil, e
	}

	// Get sub entity objects
	pattern := primitive.Regex{Pattern: "^" + id + u.HN_DELIMETER, Options: ""}
	req = bson.M{"id": pattern}
	sub, e1 := GetManyEntities(wantedEnt, req, u.RequestFilters{}, userRoles)
	if e1 != nil {
		return nil, e1
	}

	return sub, nil
}

// SwapEntity: use id to remove object from deleteEnt and then use data to create it in createEnt.
// Propagates id changes to children objects. For atomicity, all is done in a Mongo transaction.
func SwapEntity(createEnt, deleteEnt, id string, data map[string]interface{}, userRoles map[string]Role) *u.Error {
	ctx, _ := u.Connect()
	if e := prepareCreateEntity(u.EntityStrToInt(createEnt), data, userRoles); e != nil {
		return e
	}

	// Define the callback that specifies the sequence of operations to perform inside the transaction.
	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Create
		if _, err := GetDB().Collection(createEnt).InsertOne(ctx, data); err != nil {
			return nil, err
		}

		// Propagate
		if err := PropagateParentIdChange(sessCtx, id, data["id"].(string),
			u.EntityStrToInt(data["category"].(string))); err != nil {
			return nil, err
		}

		// Delete
		if c, err := GetDB().Collection(deleteEnt).DeleteOne(ctx, bson.M{"id": id}); err != nil {
			return nil, err
		} else if c.DeletedCount == 0 {
			return nil, errors.New("Error deleting object: not found")
		}

		return nil, nil
	}

	// Start a session and run the callback using WithTransaction.
	session, err := GetClient().StartSession()
	if err != nil {
		return &u.Error{Type: u.ErrDBError, Message: "Unable to start session: " + err.Error()}
	}
	defer session.EndSession(ctx)
	result, err := session.WithTransaction(ctx, callback)
	if err != nil {
		return &u.Error{Type: u.ErrDBError, Message: "Unable to complete transaction: " + err.Error()}
	}
	log.Printf("result: %v\n", result)
	return nil
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
				domain = x["id"].(string)
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
