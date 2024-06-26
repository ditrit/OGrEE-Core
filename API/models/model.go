package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"p3/repository"
	u "p3/utils"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Helper functions

func domainHasObjects(domain string) bool {
	data := map[string]interface{}{}
	// Get all collections names
	ctx, cancel := u.Connect()
	db := repository.GetDB()
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
	if parentEntStr == u.EntityToString(u.DOMAIN) {
		return []int{u.DOMAIN}
	} else if parentEntStr == u.EntityToString(u.DEVICE) {
		// device special case (devices can have devices and vobjs)
		return []int{u.DEVICE, u.VIRTUALOBJ}
	} else if parentEntStr == u.EntityToString(u.VIRTUALOBJ) {
		return []int{u.VIRTUALOBJ}
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

func updateOldObjWithPatch(old map[string]interface{}, patch map[string]interface{}) error {
	for k, v := range patch {
		switch patchValueCasted := v.(type) {
		case map[string]interface{}:
			if k == "pillars" || k == "separators" {
				old[k] = v
			} else {
				switch oldValueCasted := old[k].(type) {
				case map[string]interface{}:
					err := updateOldObjWithPatch(oldValueCasted, patchValueCasted)
					if err != nil {
						return err
					}
				default:
					old[k] = v
				}
			}
		default:
			if k == "filter" && strings.HasPrefix(v.(string), "&") {
				v = "(" + old["filter"].(string) + ") " + v.(string)
			}
			old[k] = v
		}
	}

	return nil
}

// Entity handlers

func CreateEntity(entity int, t map[string]interface{}, userRoles map[string]Role) (map[string]interface{}, *u.Error) {
	tags, tagsPresent := getTags(t)
	if tagsPresent {
		err := verifyTagList(tags)
		if err != nil {
			return nil, err
		}
	}

	if err := prepareCreateEntity(entity, t, userRoles); err != nil {
		return nil, err
	}

	return WithTransaction(func(ctx mongo.SessionContext) (map[string]any, error) {
		if entity == u.TAG {
			err := createTagImage(ctx, t)
			if err != nil {
				return nil, err
			}
		}

		entStr := u.EntityToString(entity)

		_, err := repository.CreateObject(ctx, entStr, t)
		if err != nil {
			return nil, err
		}

		fixID(t)
		return t, nil
	})
}

func prepareCreateEntity(entity int, t map[string]interface{}, userRoles map[string]Role) *u.Error {
	if err := ValidateEntity(entity, t); err != nil {
		return err
	}

	// Check user permissions
	if u.IsEntityHierarchical(entity) {
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

	delete(t, "parentId")

	return nil
}

func GetHierarchyObjectById(hierarchyName string, filters u.RequestFilters, userRoles map[string]Role) (map[string]interface{}, *u.Error) {
	// Get possible collections for this name
	rangeEntities := u.GetEntitiesByNamespace(u.PHierarchy, hierarchyName)
	req := bson.M{"id": hierarchyName}

	// Search each collection
	for _, entityStr := range rangeEntities {
		data, _ := GetObject(req, entityStr, filters, userRoles)
		if data != nil {
			return data, nil
		}
	}

	return nil, &u.Error{Type: u.ErrNotFound, Message: "Unable to find object"}
}

func GetObject(req bson.M, entityStr string, filters u.RequestFilters, userRoles map[string]Role) (map[string]interface{}, *u.Error) {
	object, err := repository.GetObject(req, entityStr, filters)

	if err != nil {
		return nil, err
	}

	//Remove _id
	object = fixID(object)

	entity := u.EntityStrToInt(entityStr)

	if shouldFillTags(entity, filters) {
		object = fillTags(object)
	}

	// Check permissions
	if u.IsEntityHierarchical(entity) {
		var domain string
		if entity == u.DOMAIN {
			domain = object["id"].(string)
		} else {
			domain = object["domain"].(string)
		}
		if userRoles != nil {
			if permission := CheckUserPermissions(userRoles, u.EntityStrToInt(entityStr), domain); permission == NONE {
				return nil, &u.Error{Type: u.ErrUnauthorized,
					Message: "User does not have permission to see this object"}
			} else if permission == READONLYNAME {
				object = FixReadOnlyName(object)
			}
		}
	}

	return object, nil
}

func GetManyObjects(entityStr string, req bson.M, filters u.RequestFilters, complexFilterExp string, userRoles map[string]Role) ([]map[string]interface{}, *u.Error) {
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
	err = repository.GetDateFilters(req, filters.StartDate, filters.EndDate)
	if err != nil {
		return nil, &u.Error{Type: u.ErrBadFormat, Message: err.Error()}
	}

	if complexFilterExp != "" {
		if complexFilters, err := ComplexFilterToMap(complexFilterExp); err != nil {
			return nil, &u.Error{Type: u.ErrBadFormat, Message: err.Error()}
		} else {
			err = getDatesFromComplexFilters(complexFilters)
			if err != nil {
				return nil, &u.Error{Type: u.ErrBadFormat, Message: err.Error()}
			}
			u.ApplyWildcardsOnComplexFilter(complexFilters)
			maps.Copy(req, complexFilters)
		}
	}

	if opts != nil {
		c, err = repository.GetDB().Collection(entityStr).Find(ctx, req, opts)
	} else {
		c, err = repository.GetDB().Collection(entityStr).Find(ctx, req)
	}
	if err != nil {
		fmt.Println(err)
		return nil, &u.Error{Type: u.ErrDBError, Message: err.Error()}
	}
	defer cancel()

	entity := u.EntityStrToInt(entityStr)
	data, e1 := ExtractCursor(c, ctx, entity, userRoles)
	if e1 != nil {
		fmt.Println(e1)
		return nil, &u.Error{Type: u.ErrInternal, Message: e1.Error()}
	}

	if shouldFillTags(entity, filters) {
		for i := range data {
			fillTags(data[i])
		}
	}

	return data, nil
}

func ComplexFilterToMap(complexFilter string) (map[string]any, error) {
	// Split the input string into individual filter expressions
	chars := []string{"(", ")", "&", "|"}
	for _, char := range chars {
		complexFilter = strings.ReplaceAll(complexFilter, char, " "+char+" ")
	}
	return complexExpressionToMap(strings.Fields(complexFilter))
}

func complexExpressionToMap(expressions []string) (map[string]any, error) {
	// Find the rightmost operator (AND, OR) outside of parentheses
	parenCount := 0
	for i := len(expressions) - 1; i >= 0; i-- {
		switch expressions[i] {
		case "(":
			parenCount++
		case ")":
			parenCount--
		case "&":
			if parenCount == 0 {
				first, _ := complexExpressionToMap(expressions[:i])
				second, _ := complexExpressionToMap(expressions[i+1:])
				return map[string]any{"$and": []map[string]any{
					first,
					second,
				}}, nil
			}
		case "|":
			if parenCount == 0 {
				first, _ := complexExpressionToMap(expressions[:i])
				second, _ := complexExpressionToMap(expressions[i+1:])
				return map[string]any{"$or": []map[string]any{
					first,
					second,
				}}, nil
			}
		}
	}

	// If there are no operators outside of parentheses, look for the innermost pair of parentheses
	for i := 0; i < len(expressions); i++ {
		if expressions[i] == "(" {
			start, end := i+1, i+1
			for parenCount := 1; end < len(expressions) && parenCount > 0; end++ {
				switch expressions[end] {
				case "(":
					parenCount++
				case ")":
					parenCount--
				}
			}
			return complexExpressionToMap(append(expressions[:start-1], expressions[start:end-1]...))
		}
	}

	// Base case: single filter expression
	re := regexp.MustCompile(`^([\w-.]+)\s*(<=|>=|<|>|!=|=)\s*([\w-.*]+)$`)

	ops := map[string]string{"<=": "$lte", ">=": "$gte", "<": "$lt", ">": "$gt", "!=": "$not"}

	if len(expressions) <= 3 {
		expression := strings.Join(expressions[:], "")

		if match := re.FindStringSubmatch(expression); match != nil {
			switch match[1] {
			case "startDate":
				return map[string]any{"lastUpdated": map[string]any{"$gte": match[3]}}, nil
			case "endDate":
				return map[string]any{"lastUpdated": map[string]any{"$lte": match[3]}}, nil
			case "id", "name", "category", "description", "domain", "createdDate", "lastUpdated", "slug":
				if match[2] == "=" {
					return map[string]any{match[1]: match[3]}, nil
				}
				return map[string]any{match[1]: map[string]any{ops[match[2]]: match[3]}}, nil
			default:
				if match[2] == "=" {
					return map[string]any{"attributes." + match[1]: match[3]}, nil
				}
				return map[string]any{"attributes." + match[1]: map[string]any{ops[match[2]]: match[3]}}, nil
			}
		}
	}

	fmt.Println("Error: Invalid filter expression")
	return nil, errors.New("invalid filter expression")
}

func getDatesFromComplexFilters(req map[string]any) error {
	for k, v := range req {
		if k == "$and" || k == "$or" {
			for _, complexFilter := range v.([]map[string]any) {
				err := getDatesFromComplexFilters(complexFilter)
				if err != nil {
					return err
				}
			}
		} else if k == "lastUpdated" {
			for op, date := range v.(map[string]any) {
				parsedDate, err := time.Parse("2006-01-02", date.(string))
				if err != nil {
					return err
				}
				if op == "$lte" {
					parsedDate = parsedDate.Add(time.Hour * 24)
				}
				req[k] = map[string]any{op: parsedDate}
			}
		}
	}
	return nil
}

// GetCompleteHierarchy: gets all objects in db using hierachyName and returns:
//   - tree: map with parents as key and their children as an array value
//     tree: {parent:[children]}
//   - categories: map with category name as key and corresponding objects
//     as an array value
//     categories: {categoryName:[children]}
func GetCompleteHierarchy(userRoles map[string]Role, filters u.HierarchyFilters) (map[string]interface{}, *u.Error) {
	response := make(map[string]interface{})
	categories := make(map[string][]string)
	hierarchy := make(map[string]interface{})

	switch filters.Namespace {
	case u.Any:
		for _, ns := range []u.Namespace{u.Physical, u.Logical, u.Organisational} {
			data, err := getHierarchyWithNamespace(ns, userRoles, filters, categories)
			if err != nil {
				return nil, err
			}
			hierarchy[u.NamespaceToString(ns)] = data
		}
	default:
		data, err := getHierarchyWithNamespace(filters.Namespace, userRoles, filters, categories)
		if err != nil {
			return nil, err
		}
		hierarchy[u.NamespaceToString(filters.Namespace)] = data

	}

	response["tree"] = hierarchy
	if filters.WithCategories {
		categories["KeysOrder"] = []string{"site", "building", "room", "rack"}
		response["categories"] = categories
	}
	return response, nil
}

func getHierarchyWithNamespace(namespace u.Namespace, userRoles map[string]Role, filters u.HierarchyFilters,
	categories map[string][]string) (map[string][]string, *u.Error) {
	hierarchy := make(map[string][]string)
	rootIdx := "*"

	ctx, cancel := u.Connect()
	db := repository.GetDB()
	dbFilter := bson.M{}

	// Depth of hierarchy defined by user
	if filters.Limit != "" && namespace != u.PStray &&
		!strings.Contains(u.NamespaceToString(namespace), string(u.Logical)) {
		if _, e := strconv.Atoi(filters.Limit); e == nil {
			pattern := primitive.Regex{Pattern: "^" + u.NAME_REGEX + "(." + u.NAME_REGEX + "){0," +
				filters.Limit + "}$", Options: ""}
			dbFilter = bson.M{"id": pattern}
		}
	}
	// User date filters
	err := repository.GetDateFilters(dbFilter, filters.StartDate, filters.EndDate)
	if err != nil {
		return nil, &u.Error{Type: u.ErrBadFormat, Message: err.Error()}
	}

	// Search collections according to namespace
	entities := u.GetEntitiesByNamespace(namespace, "")

	for _, entityName := range entities {
		// Get data
		opts := options.Find().SetProjection(bson.D{{Key: "domain", Value: 1}, {Key: "id", Value: 1}, {Key: "category", Value: 1}})

		if u.IsEntityNonHierarchical(u.EntityStrToInt(entityName)) {
			opts = options.Find().SetProjection(bson.D{{Key: "slug", Value: 1}})
		}

		c, err := db.Collection(entityName).Find(ctx, dbFilter, opts)
		if err != nil {
			println(err.Error())
			return nil, &u.Error{Type: u.ErrDBError, Message: err.Error()}
		}

		data, e := ExtractCursor(c, ctx, u.EntityStrToInt(entityName), userRoles)
		if e != nil {
			return nil, &u.Error{Type: u.ErrInternal, Message: e.Error()}
		}

		// Format data
		for _, obj := range data {
			if strings.Contains(u.NamespaceToString(namespace), string(u.Logical)) {
				// Logical
				var objId string
				if u.IsEntityNonHierarchical(u.EntityStrToInt(entityName)) {
					objId = obj["slug"].(string)
					hierarchy[rootIdx+entityName] = append(hierarchy[rootIdx+entityName], objId)
				} else {
					objId = obj["id"].(string)
					categories[entityName] = append(categories[entityName], objId)
					if strings.Contains(objId, ".") && obj["category"] != "group" {
						// Physical or Org Children
						fillHierarchyMap(objId, hierarchy)
					} else {
						hierarchy[rootIdx+entityName] = append(hierarchy[rootIdx+entityName], objId)
					}
				}
			} else if strings.Contains(obj["id"].(string), ".") {
				// Physical or Org Children
				categories[entityName] = append(categories[entityName], obj["id"].(string))
				fillHierarchyMap(obj["id"].(string), hierarchy)
			} else {
				// Physical or Org Roots
				objId := obj["id"].(string)
				categories[entityName] = append(categories[entityName], objId)
				if u.EntityStrToInt(entityName) == u.STRAYOBJ {
					hierarchy[rootIdx+entityName] = append(hierarchy[rootIdx+entityName], objId)
				} else if u.EntityStrToInt(entityName) != u.VIRTUALOBJ {
					hierarchy[rootIdx] = append(hierarchy[rootIdx], objId)
				}
			}
		}
	}

	// For the root of VIRTUAL objects we also need to check for devices
	for _, vobj := range hierarchy[rootIdx+u.EntityToString(u.VIRTUALOBJ)] {
		if !strings.ContainsAny(vobj, u.HN_DELIMETER) { // is stray virtual
			// Get device linked to this vobj through its virtual config
			entityName := u.EntityToString(u.DEVICE)
			dbFilter := bson.M{"attributes.virtual_config.clusterId": vobj}
			opts := options.Find().SetProjection(bson.D{{Key: "domain", Value: 1}, {Key: "id", Value: 1}, {Key: "category", Value: 1}})
			c, err := db.Collection(entityName).Find(ctx, dbFilter, opts)
			if err != nil {
				println(err.Error())
				return nil, &u.Error{Type: u.ErrDBError, Message: err.Error()}
			}
			data, e := ExtractCursor(c, ctx, u.EntityStrToInt(entityName), userRoles)
			if e != nil {
				return nil, &u.Error{Type: u.ErrInternal, Message: e.Error()}
			}
			// Add data
			for _, obj := range data {
				objId := obj["id"].(string)
				categories[entityName] = append(categories[entityName], objId)
				hierarchy[vobj] = append(hierarchy[vobj], objId)
			}
		}
	}

	defer cancel()
	return hierarchy, nil
}

func GetCompleteHierarchyAttributes(userRoles map[string]Role) (map[string]interface{}, *u.Error) {
	response := make(map[string]interface{})
	// Get all collections names
	ctx, cancel := u.Connect()
	db := repository.GetDB()
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

// GetSiteParentAttribute: search for the object of given ID,
// then search for its site parent and return its requested attribute
func GetSiteParentAttribute(id string, attribute string) (map[string]any, *u.Error) {
	data := map[string]interface{}{}

	// Get all collections names
	ctx, cancel := u.Connect()
	db := repository.GetDB()
	collNames, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		fmt.Println(err.Error())
		return nil, &u.Error{Type: u.ErrDBError, Message: err.Error()}
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
					return nil, &u.Error{Type: u.ErrNotFound,
						Message: "Could not find parent site for given object"}
				}
			}
		}
	}

	defer cancel()

	if len(data) == 0 {
		return nil, &u.Error{Type: u.ErrNotFound, Message: "No object found with given id"}
	} else if attribute == "sitecolors" {
		resp := map[string]any{}
		for _, colorName := range []string{"reservedColor", "technicalColor", "usableColor"} {
			if color := data["attributes"].(map[string]interface{})[colorName]; color != nil {
				resp[colorName] = color
			} else {
				resp[colorName] = ""
			}
		}
		return resp, nil
	} else if attrValue := data["attributes"].(map[string]interface{})[attribute]; attrValue == nil {
		return nil, &u.Error{Type: u.ErrNotFound,
			Message: "Parent site has no temperatureUnit in attributes"}
	} else {
		return map[string]any{attribute: attrValue}, nil
	}
}

func GetEntityCount(entity int) int64 {
	ent := u.EntityToString(entity)
	ctx, cancel := u.Connect()
	ans, e := repository.GetDB().Collection(ent).CountDocuments(ctx, bson.M{}, nil)
	if e != nil {
		println(e.Error())
		return -1
	}
	defer cancel()
	return ans
}

func CommandRunner(cmd interface{}) *mongo.SingleResult {
	ctx, cancel := u.Connect()
	result := repository.GetDB().RunCommand(ctx, cmd, nil)
	defer cancel()
	return result
}

func GetStats() map[string]interface{} {
	ans := map[string]interface{}{}
	t := map[string]interface{}{}
	latestDocArr := []map[string]interface{}{}
	var latestTime interface{}

	for entity := range u.Entities {
		num := GetEntityCount(entity)
		if num == -1 {
			num = 0
		}

		ans["Number of "+u.EntityToString(entity)+"s:"] = num

		//Retrieve the latest updated document in each collection
		//and store into the latestDocArr array
		obj := map[string]interface{}{}
		filter := options.FindOne().SetSort(bson.M{"lastUpdated": -1})
		ctx, cancel := u.Connect()

		e := repository.GetDB().Collection(u.EntityToString(entity)).FindOne(ctx, bson.M{}, filter).Decode(&obj)
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
	name := repository.GetDB().Name()

	//Remove the preceding 'ogree' at beginning of name
	if strings.Index(name, "ogree") == 0 {
		name = name[5:] //5=len('ogree')
	}
	return name
}

// DeleteHierarchicalObject: delete object of given hierarchyName
// search for all its children and delete them too, return:
// - success or fail message map
func DeleteHierarchicalObject(entity string, id string, userRoles map[string]Role) *u.Error {
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

	_, err := WithTransaction(func(ctx mongo.SessionContext) (any, error) {
		err := repository.DeleteObject(ctx, entity, req)
		if err != nil {
			// Unable to delete given id
			return nil, err
		}

		// Delete possible children
		rangeEntities := getChildrenCollections(u.GROUP, entity)
		for _, childEnt := range rangeEntities {
			childEntName := u.EntityToString(childEnt)
			pattern := primitive.Regex{Pattern: "^" + id + u.HN_DELIMETER, Options: ""}

			repository.GetDB().Collection(childEntName).DeleteMany(ctx,
				bson.M{"id": pattern})
		}

		return nil, nil
	})

	return err
}

func DeleteNonHierarchicalObject(entity, slug string) *u.Error {
	req := bson.M{"slug": slug}
	ctx, cancel := u.Connect()
	defer cancel()
	return repository.DeleteObject(ctx, entity, req)
}

func prepareUpdateObject(ctx mongo.SessionContext, entity int, id string, updateData, oldObject map[string]any, userRoles map[string]Role) *u.Error {
	// Check user permissions in case domain is being updated
	if entity != u.DOMAIN && u.IsEntityHierarchical(entity) && (oldObject["domain"] != updateData["domain"]) {
		if perm := CheckUserPermissions(userRoles, entity, updateData["domain"].(string)); perm < WRITE {
			return &u.Error{Type: u.ErrUnauthorized, Message: "User does not have permission to change this object"}
		}
	}

	// tag list edition support
	err := addAndRemoveFromTags(entity, id, updateData)
	if err != nil {
		return err
	}

	// Ensure the update is valid
	err = ValidateEntity(entity, updateData)
	if err != nil {
		return err
	}

	updateData["lastUpdated"] = primitive.NewDateTimeFromTime(time.Now())
	updateData["createdDate"] = oldObject["createdDate"]
	delete(updateData, "parentId")

	if entity == u.TAG {
		// tag slug edition support
		if updateData["slug"].(string) != oldObject["slug"].(string) {
			err := repository.UpdateTagSlugInEntities(ctx, oldObject["slug"].(string), updateData["slug"].(string))
			if err != nil {
				return err
			}
		}

		err := updateTagImage(ctx, oldObject, updateData)
		if err != nil {
			return err
		}
	}

	return nil
}

func UpdateObject(entityStr string, id string, updateData map[string]interface{}, isPatch bool, userRoles map[string]Role, isRecursive bool) (map[string]interface{}, *u.Error) {
	var idFilter bson.M
	if u.IsEntityNonHierarchical(u.EntityStrToInt(entityStr)) {
		idFilter = bson.M{"slug": id}
	} else {
		idFilter = bson.M{"id": id}
	}

	//Update timestamp requires first obj retrieval
	//there isn't any way for mongoDB to make a field
	//immutable in a document
	var oldObj map[string]any
	var err *u.Error
	if entityStr == u.HIERARCHYOBJS_ENT {
		oldObj, err = GetHierarchyObjectById(id, u.RequestFilters{}, userRoles)
		if err == nil {
			entityStr = oldObj["category"].(string)
		}
	} else {
		oldObj, err = GetObject(idFilter, entityStr, u.RequestFilters{}, userRoles)
	}
	if err != nil {
		return nil, err
	}

	entity := u.EntityStrToInt(entityStr)

	// Check if permission is only readonly
	if u.IsEntityHierarchical(entity) && oldObj["description"] == nil {
		// Description is always present, unless GetEntity was called with readonly permission
		return nil, &u.Error{Type: u.ErrUnauthorized,
			Message: "User does not have permission to change this object"}
	}

	tags, tagsPresent := getTags(updateData)

	// Update old object data with patch data
	if isPatch {
		if tagsPresent {
			return nil, &u.Error{
				Type:    u.ErrBadFormat,
				Message: "Tags cannot be modified in this way, use tags+ and tags-",
			}
		}

		var formattedOldObj map[string]interface{}
		// Convert primitive.A and similar types
		bytes, _ := json.Marshal(oldObj)
		json.Unmarshal(bytes, &formattedOldObj)
		// Update old with new
		err := updateOldObjWithPatch(formattedOldObj, updateData)
		if err != nil {
			return nil, &u.Error{Type: u.ErrBadFormat, Message: err.Error()}
		}

		updateData = formattedOldObj
		// Remove API set fields
		delete(updateData, "id")
		delete(updateData, "lastUpdated")
		delete(updateData, "createdDate")
	} else if tagsPresent {
		err := verifyTagList(tags)
		if err != nil {
			return nil, err
		}
	}

	result, err := WithTransaction(func(ctx mongo.SessionContext) (interface{}, error) {
		err = prepareUpdateObject(ctx, entity, id, updateData, oldObj, userRoles)
		if err != nil {
			return nil, err
		}

		mongoRes := repository.GetDB().Collection(entityStr).FindOneAndReplace(
			ctx,
			idFilter, updateData,
			options.FindOneAndReplace().SetReturnDocument(options.After),
		)
		if mongoRes.Err() != nil {
			return nil, mongoRes.Err()
		}

		if oldObj["id"] != updateData["id"] {
			// Changes to id should be propagated
			if err := repository.PropagateParentIdChange(
				ctx,
				oldObj["id"].(string),
				updateData["id"].(string),
				entity,
			); err != nil {
				return nil, err
			} else if entity == u.DOMAIN {
				if err := repository.PropagateDomainChange(ctx,
					oldObj["id"].(string),
					updateData["id"].(string),
				); err != nil {
					return nil, err
				}
			}
		}
		if u.IsEntityHierarchical(entity) && (oldObj["domain"] != updateData["domain"]) {
			if isRecursive {
				// Change domain of all children too
				if err := repository.PropagateDomainChangeToChildren(
					ctx,
					updateData["id"].(string),
					updateData["domain"].(string),
				); err != nil {
					return nil, err
				}
			} else {
				// Check if children domains are compatible
				if err := repository.CheckParentDomainChange(entity, updateData["id"].(string),
					updateData["domain"].(string)); err != nil {
					return nil, err
				}
			}
		}

		return mongoRes, nil
	})

	if err != nil {
		return nil, err
	}

	var updatedDoc map[string]interface{}
	result.(*mongo.SingleResult).Decode(&updatedDoc)

	return fixID(updatedDoc), nil
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
			"(." + u.NAME_REGEX + "){1," + strconv.Itoa(limit) + "}$", Options: ""}
		children, e1 := GetManyObjects(checkEntName, bson.M{"id": pattern}, filters, "", nil)
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

// GetHierarchyByCluster: get children devices and vobjs of given cluster
func GetHierarchyByCluster(clusterName string, limit int, filters u.RequestFilters) ([]map[string]interface{}, *u.Error) {
	allChildren := map[string]interface{}{}
	hierarchy := make(map[string][]string)

	// Get children from device and vobjs
	for _, checkEnt := range []int{u.DEVICE, u.VIRTUALOBJ} {
		checkEntName := u.EntityToString(checkEnt)
		var dbFilter primitive.M
		if checkEnt == u.VIRTUALOBJ {
			// linked by Id
			pattern := primitive.Regex{Pattern: "^" + clusterName +
				"(." + u.NAME_REGEX + "){1," + strconv.Itoa(limit) + "}$", Options: ""}
			dbFilter = bson.M{"id": pattern}
		} else {
			// DEVICE links to vobj via virtual config
			dbFilter = bson.M{"attributes.virtual_config.clusterId": clusterName}
		}
		children, e1 := GetManyObjects(checkEntName, dbFilter, filters, "", nil)
		if e1 != nil {
			return nil, e1
		}
		for _, child := range children {
			if checkEnt == u.VIRTUALOBJ {
				allChildren[child["id"].(string)] = child
				fillHierarchyMap(child["id"].(string), hierarchy)
			} else {
				// add namespace prefix to devices
				child["id"] = "Physical." + child["id"].(string)
				allChildren[child["id"].(string)] = child
				// add as direct child to cluster
				hierarchy[clusterName] = append(hierarchy[clusterName], child["id"].(string))
			}
		}
	}

	// Organize vobj family according to relations (nest children)
	return recursivelyGetChildrenFromMaps(clusterName, hierarchy, allChildren), nil
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
	_, e := GetObject(req, entStr, u.RequestFilters{}, userRoles)
	if e != nil {
		return nil, e
	}

	// Get sub entity objects
	pattern := primitive.Regex{Pattern: "^" + id + u.HN_DELIMETER, Options: ""}
	req = bson.M{"id": pattern}
	sub, e1 := GetManyObjects(wantedEnt, req, u.RequestFilters{}, "", userRoles)
	if e1 != nil {
		return nil, e1
	}

	return sub, nil
}

// SwapEntity: use id to remove object from deleteEnt and then use data to create it in createEnt.
// Propagates id changes to children objects. For atomicity, all is done in a Mongo transaction.
func SwapEntity(createEnt, deleteEnt, id string, data map[string]interface{}, userRoles map[string]Role) *u.Error {
	if err := prepareCreateEntity(u.EntityStrToInt(createEnt), data, userRoles); err != nil {
		return err
	}

	_, err := WithTransaction(func(ctx mongo.SessionContext) (any, error) {
		// Create
		if _, err := repository.CreateObject(ctx, createEnt, data); err != nil {
			return nil, err
		}

		// Propagate
		if err := repository.PropagateParentIdChange(ctx, id, data["id"].(string),
			u.EntityStrToInt(data["category"].(string))); err != nil {
			return nil, err
		}

		// Delete
		if c, err := repository.GetDB().Collection(deleteEnt).DeleteOne(ctx, bson.M{"id": id}); err != nil {
			return nil, err
		} else if c.DeletedCount == 0 {
			return nil, errors.New("Error deleting object: not found")
		}

		return nil, nil
	})

	return err
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
		if u.IsEntityHierarchical(entity) && userRoles != nil {
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
