package models

import (
	"fmt"
	"p3/repository"
	u "p3/utils"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetHierarchyObjectById(hierarchyName string, filters u.RequestFilters, userRoles map[string]Role) (map[string]interface{}, *u.Error) {
	// Get possible collections for this name
	rangeEntities := u.GetEntitiesById(u.PHierarchy, hierarchyName)
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
	entities := u.GetEntitiesById(namespace, "")

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
