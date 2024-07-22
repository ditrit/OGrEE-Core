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

const rootIdx = "*"

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

// getHierarchyByNamespace: get complete hierarchy of a given namespace
func getHierarchyByNamespace(namespace u.Namespace, userRoles map[string]Role, filters u.HierarchyFilters,
	categories map[string][]string) (map[string][]string, *u.Error) {
	hierarchy := make(map[string][]string)

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

		// Fill hierarchy with formatted data
		fillHierarchyData(data, namespace, entityName, hierarchy, categories)
	}

	// For the root of VIRTUAL objects we also need to check for devices
	if err := fillVirtualHierarchyByVconfig(userRoles, hierarchy, categories); err != nil {
		return nil, err
	}

	defer cancel()
	return hierarchy, nil
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
			data, err := getHierarchyByNamespace(ns, userRoles, filters, categories)
			if err != nil {
				return nil, err
			}
			hierarchy[u.NamespaceToString(ns)] = data
		}
	default:
		data, err := getHierarchyByNamespace(filters.Namespace, userRoles, filters, categories)
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

// GetCompleteHierarchyAttributes: get all objects with all its attributes
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
		var entInt int
		if entInt = u.EntityStrToInt(collName); entInt == -1 {
			continue
		}

		// Get attributes
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

		// Add to response
		formatCompleteHierarchyAttributes(data, response)
	}
	defer cancel()
	return response, nil
}

func formatCompleteHierarchyAttributes(data []map[string]any, response map[string]any) {
	for _, obj := range data {
		if obj["attributes"] == nil {
			continue
		}
		if id, isStr := obj["id"].(string); isStr && id != "" {
			response[id] = obj["attributes"]
		} else if obj["name"] != nil {
			response[obj["name"].(string)] = obj["attributes"]
		}
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

func fillHierarchyData(data []map[string]any, namespace u.Namespace, entityName string, hierarchy, categories map[string][]string) {
	for _, obj := range data {
		if strings.Contains(u.NamespaceToString(namespace), string(u.Logical)) {
			// Logical
			if u.IsEntityNonHierarchical(u.EntityStrToInt(entityName)) {
				objId := obj["slug"].(string)
				hierarchy[rootIdx+entityName] = append(hierarchy[rootIdx+entityName], objId)
			} else {
				objId := obj["id"].(string)
				categories[entityName] = append(categories[entityName], objId)
				if strings.Contains(objId, ".") && obj["category"] != "group" {
					// Physical or Org Children
					fillHierarchyMap(objId, hierarchy)
				} else {
					hierarchy[rootIdx+entityName] = append(hierarchy[rootIdx+entityName], objId)
				}
			}
		} else {
			fillHierarchyDataWithHierarchicalObj(entityName, obj, hierarchy, categories)
		}
	}
}

func fillHierarchyDataWithHierarchicalObj(entityName string, obj map[string]any, hierarchy, categories map[string][]string) {
	objId := obj["id"].(string)
	categories[entityName] = append(categories[entityName], objId)
	if strings.Contains(objId, ".") {
		// Physical or Org Children
		fillHierarchyMap(objId, hierarchy)
	} else {
		// Physical or Org Roots
		if u.EntityStrToInt(entityName) == u.STRAYOBJ {
			hierarchy[rootIdx+entityName] = append(hierarchy[rootIdx+entityName], objId)
		} else if u.EntityStrToInt(entityName) != u.VIRTUALOBJ {
			hierarchy[rootIdx] = append(hierarchy[rootIdx], objId)
		}
	}
}

func fillVirtualHierarchyByVconfig(userRoles map[string]Role, hierarchy, categories map[string][]string) *u.Error {
	ctx, cancel := u.Connect()
	db := repository.GetDB()
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
				return &u.Error{Type: u.ErrDBError, Message: err.Error()}
			}
			data, e := ExtractCursor(c, ctx, u.EntityStrToInt(entityName), userRoles)
			if e != nil {
				return &u.Error{Type: u.ErrInternal, Message: e.Error()}
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
	return nil
}
