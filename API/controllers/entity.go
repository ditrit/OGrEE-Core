package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"p3/models"
	u "p3/utils"
	"strconv"
	"strings"
	"time"

	"github.com/elliotchance/pie/v2"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// This function is useful for debugging
// purposes. It displays any JSON
func viewJson(r *http.Request) {
	var updateData map[string]interface{}
	json.NewDecoder(r.Body).Decode(&updateData)
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "    ")

	if err := enc.Encode(updateData); err != nil {
		log.Fatal(err)
	}
}

func Disp(x map[string]interface{}) {
	jx, _ := json.Marshal(x)
	println("JSON: ", string(jx))
}

func DispRequestMetaData(r *http.Request) {
	fmt.Println("URL:", r.URL.String())
	fmt.Println("IP-ADDR: ", r.RemoteAddr)
	fmt.Println(time.Now().Format("2006-Jan-02 Monday 03:04:05 PM MST -07:00"))
}

const ErrDecodingBodyMsg = "Error while decoding request body"

func decodeRequestBody(w http.ResponseWriter, r *http.Request, dataObj any) error {
	err := json.NewDecoder(r.Body).Decode(dataObj)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message(ErrDecodingBodyMsg))
		u.ErrLog(ErrDecodingBodyMsg, "decodeRequestBody", "", r)
		return err
	}
	return nil
}

var decoder = schema.NewDecoder()

func getFiltersFromQueryParams(r *http.Request) u.RequestFilters {
	var filters u.RequestFilters
	decoder.Decode(&filters, r.URL.Query())
	return filters
}

func getUserFromToken(w http.ResponseWriter, r *http.Request) *models.Account {
	userData := r.Context().Value("user")
	if userData == nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message("Error while parsing path params"))
		u.ErrLog("Error while parsing path params", "GET GENERIC", "", r)
		return nil
	}
	userId := userData.(map[string]interface{})["userID"].(primitive.ObjectID)
	user := models.GetUser(userId)
	if user == nil || len(user.Roles) <= 0 {
		w.WriteHeader(http.StatusUnauthorized)
		u.Respond(w, u.Message("Invalid token: no valid user found"))
		u.ErrLog("Unable to find user associated to token", "GET GENERIC", "", r)
		return nil
	}
	return user
}

// swagger:operation POST /api/{entity} Objects CreateObject
// Creates an object of the given entity in the system.
// ---
// security:
// - bearer: []
// produces:
// - application/json
// parameters:
//   - name: entity
//     in: path
//     description: 'Entity (same as category) of the object. Accepted values: sites, domains,
//     buildings, rooms, racks, devices, acs, panels,
//     cabinets, groups, corridors, virtual_objs
//     room_templates, obj_templates, bldg_templates, tags,
//     stray_objects, hierarchy_objects.'
//     required: true
//     type: string
//     default: "sites"
//   - name: body
//     in: body
//     required: true
//     default: {}
// responses:
//     '201':
//         description: 'Created. A response body will be returned with
//         a meaningful message.'
//     '400':
//         description: 'Bad request. A response body with an error
//         message will be returned.'

func CreateEntity(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 CreateEntity ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)
	// Get entity
	entStr := mux.Vars(r)["entity"]
	entInt := u.EntityStrToInt(entStr)
	println("ENT: ", entStr)

	// Prevents Mongo from creating a new unidentified collection
	if entInt < 0 && entStr != u.HIERARCHYOBJS_ENT {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message("Invalid entity in URL: '"+mux.Vars(r)["entity"]+"' Please provide a valid object"))
		u.ErrLog("Cannot create invalid object", "CREATE "+mux.Vars(r)["entity"], "", r)
		return
	}

	// Get request body
	object := map[string]interface{}{}
	if err := decodeRequestBody(w, r, &object); err != nil {
		return
	}

	// Get user roles for permissions
	user := getUserFromToken(w, r)
	if user == nil {
		return
	}

	if entStr == u.HIERARCHYOBJS_ENT {
		// Get entity from object's category
		entStr = object["category"].(string)
		entInt = u.EntityStrToInt(entStr)
		if entInt < u.SITE || entInt > u.GROUP {
			w.WriteHeader(http.StatusBadRequest)
			u.Respond(w, u.Message("Invalid category for a hierarchy object"))
			u.ErrLog("Cannot create invalid hierarchy object", "CREATE "+mux.Vars(r)["entity"], "", r)
			return
		}
	} else if u.IsEntityHierarchical(entInt) && entInt != u.STRAYOBJ {
		// Check if category and endpoint match, except for non hierarchal entities and strays
		if object["category"] != entStr {
			w.WriteHeader(http.StatusBadRequest)
			u.Respond(w, u.Message("Category in request body does not correspond with desired object in endpoint"))
			u.ErrLog("Cannot create invalid object", "CREATE "+mux.Vars(r)["entity"], "", r)
			return
		}
	}

	// Clean the data of 'id' attribute if present
	delete(object, "_id")

	// Try create and respond
	resp, e := models.CreateEntity(entInt, object, user.Roles)
	if e != nil {
		u.ErrLog("Error creating "+entStr, "CREATE", e.Message, r)
		u.RespondWithError(w, e)
	} else {
		w.WriteHeader(http.StatusCreated)
		u.Respond(w, u.RespDataWrapper("successfully created "+entStr, resp))
		if entInt == u.LAYER {
			eventNotifier <- u.FormatNotifyData("create", entStr, resp)
		}
	}
}

// swagger:operation POST /api/domains/bulk Organization CreateBulkDomain
// Create multiple domains in a single request.
// An array of domains should be provided in the body.
// ---
// security:
// - bearer: []
// produces:
// - application/json
// parameters:
//   - name: body
//     in: body
//     required: true
//     default: [{}]
// responses:
//     '200':
//         description: 'Request processed. Check the response body
//         for individual results for each of the sent domains'
//     '400':
//         description: 'Bad format: body is not a valid list of domains.'

func CreateBulkDomain(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 CreateBulkDomain ")
	fmt.Println("******************************************************")

	// Get user roles for permissions
	user := getUserFromToken(w, r)
	if user == nil {
		return
	}

	// Get domains to create from request body
	listDomains := []map[string]interface{}{}
	if err := decodeRequestBody(w, r, &listDomains); err != nil {
		return
	}
	domainsToCreate, e := getBulkDomainsRecursively("", listDomains)
	if e != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message(e.Error()))
		u.ErrLog(e.Error(), "CREATE BULK DOMAIN", "", r)
		return
	}
	fmt.Println(domainsToCreate)

	// Try create and repond
	resp := map[string]interface{}{}
	for _, domain := range domainsToCreate {
		// Convert back to json to avoid invalid types in json schema validation
		bytes, _ := json.Marshal(domain)
		json.Unmarshal(bytes, &domain)
		// Create and save response
		_, err := models.CreateEntity(u.DOMAIN, domain, user.Roles)
		var name string
		if v, ok := domain["parentId"].(string); ok && v != "" {
			name = v + "." + domain["name"].(string)
		} else {
			name = domain["name"].(string)
		}
		if err != nil {
			resp[name] = err.Message
		} else {
			resp[name] = "successfully created domain"
		}
	}
	w.WriteHeader(http.StatusOK)
	u.Respond(w, resp)
}

func getBulkDomainsRecursively(parent string, listDomains []map[string]interface{}) ([]map[string]interface{}, error) {
	domainsToCreate := []map[string]interface{}{}
	for _, domain := range listDomains {
		domainObj, err := setDomainAttributes(parent, domain)
		if err != nil {
			return nil, err
		}

		domainsToCreate = append(domainsToCreate, domainObj)

		// Add children domain, if any
		if children, ok := domain["domains"].([]interface{}); ok {
			if len(children) > 0 {
				// Convert from interface to map
				dChildren := listAnyTolistMap(children)

				// Set parentId for children
				var parentId string
				if parent == "" {
					parentId = domain["name"].(string)
				} else {
					parentId = parent + "." + domain["name"].(string)
				}

				// Add children
				childDomains, e := getBulkDomainsRecursively(parentId, dChildren)
				if e != nil {
					return nil, e
				}
				domainsToCreate = append(domainsToCreate, childDomains...)
			}
		}
	}
	return domainsToCreate, nil
}

func setDomainAttributes(parent string, domain map[string]any) (map[string]any, error) {
	domainObj := map[string]any{}
	// Name is the only required attribute
	name, ok := domain["name"].(string)
	if !ok {
		return nil, errors.New("Invalid format: Name is required for all domains")
	}
	domainObj["name"] = name

	// Default attributes
	if parent != "" {
		domainObj["parentId"] = parent
	}
	domainObj["category"] = "domain"
	if desc, ok := domain["description"].(string); ok {
		domainObj["description"] = desc
	} else {
		domainObj["description"] = name
	}

	// Optional attributes
	domainObj["attributes"] = map[string]string{}
	if color, ok := domain["color"].(string); ok {
		domainObj["attributes"].(map[string]string)["color"] = color
	} else {
		domainObj["attributes"].(map[string]string)["color"] = "ffffff"
	}
	return domainObj, nil
}

func listAnyTolistMap(data []any) []map[string]interface{} {
	converted := []map[string]interface{}{}
	for _, d := range data {
		converted = append(converted, d.(map[string]interface{}))
	}
	return converted
}

// swagger:operation GET /api/objects Objects GetGenericObject
// Get all objects from any entity. Return as a list.
// Wildcards can be used on any of the parameters present in query.
//
// | Special Terms | Meaning                                     |
// |-------------  | --------------------------------------------|
// | `*`           | matches any sequence of non-path-separators |
// | `.**.`        | matches zero or more directories            |
// | `.**{m,M}.`   | matches from m to M directories             |
//
// A doublestar (`**`) should appear surrounded by id separators such as `.**.`.
// A mid-pattern doublestar (`**`) behaves like star: a pattern
// such as `path.to.**` would return the same results as `path.to.*`. To apply recursion, the
// id you're looking for is `path.to.**.*`.
// Examples:
// id=path.to.a* will return all the children of path.to which name starts with a.
// id=path.to.`**`.a* will return all the descendant hierarchy of path.to which name starts with a.
// id=path.to.`**`{1,3}.a* will return all the grandchildren to great-great-grandchildren of path.to which name starts with a.
// ---
// security:
// - bearer: []
// produces:
// - application/json
// parameters:
//   - name: id
//     in: path
//     description: 'id of the object to obtain.
//     If none provided, all objects of the namespace will be obtained'
//   - name: namespace
//     in: query
//     description: 'One of the values: physical, physical.stray, physical.hierarchy,
//     logical, logical.objtemplate, logical.bldgtemplate, logical.roomtemplate, logical.tag,
//     organisational.
//     If none provided, all namespaces are used by default.'
//   - name: fieldOnly
//     in: query
//     description: 'specify which object field to show in response.
//     Multiple fieldOnly can be added. An invalid field is simply ignored.'
//   - name: startDate
//     in: query
//     description: 'filter objects by lastUpdated >= startDate.
//     Format: yyyy-mm-dd'
//   - name: endDate
//     in: query
//     description: 'filter objects by lastUpdated <= endDate.
//     Format: yyyy-mm-dd'
//   - name: limit
//     in: query
//     description: 'Get limit level of hierarchy for objects in the response.
//     It must be specified alongside id.
//     Example: ?limit=1&id=siteA.B.R1 will return the object R1 with its children nested.
//     ?limit=2&id=siteA.B.R1.* will return all objects one level above R1 with
//     its up to two levels children nested.'
//     required: false
//     type: string
//   - name: attributes
//     in: query
//     description: 'Any other object attributes can be queried.
//     Replace attributes here by the name of the attribute followed by its value.'
//     required: false
//     type: string
//     default: domain=DemoDomain
//     example: vendor=ibm ; name=siteA ; orientation=front
// responses:
//		'200':
//		    description: 'Found. A response body will be returned with
//	        a meaningful message.'
//		'500':
//		    description: Internal Error. A system error stopped the request.

// swagger:operation DELETE /api/objects Objects DeleteGenericObject
// Deletes an object in the system from any of the entities with no need to specify it.
// Wildcards can be used on any of the parameters present in query.
//
// | Special Terms | Meaning                                     |
// |-------------  | --------------------------------------------|
// | `*`           | matches any sequence of non-path-separators |
// | `.**.`        | matches zero or more directories            |
// | `.**{m,M}.`   | matches from m to M directories             |
//
// A doublestar (`**`) should appear surrounded by id separators such as `.**.`.
// A mid-pattern doublestar (`**`) behaves like star: a pattern
// such as `path.to.**` would return the same results as `path.to.*`. To apply recursion, the
// id you're looking for is `path.to.**.*`.
// Examples:
// id=path.to.a* will delete all the children of path.to which name starts with a.
// id=path.to.**.a* will delete all the descendant hierarchy of path.to which name starts with a.
// id=path.to.**{1,3}.a* will delete all the grandchildren to great-great-grandchildren of path.to which name starts with a.
// ---
// security:
// - bearer: []
// produces:
// - application/json
// parameters:
//   - name: id
//     in: path
//     description: ID type hierarchyName of the object
//     required: true
//   - name: fieldOnly
//     in: query
//     description: 'specify which object field to show in response.
//     Multiple fieldOnly can be added. An invalid field is simply ignored.'
//   - name: startDate
//     in: query
//     description: 'filter objects by lastUpdated >= startDate.
//     Format: yyyy-mm-dd'
//   - name: endDate
//     in: query
//     description: 'filter objects by lastUpdated <= endDate.
//     Format: yyyy-mm-dd'
//   - name: namespace
//     in: query
//     description: 'One of the values: physical, physical.stray, physical.hierarchy,
//     logical, logical.objtemplate, logical.bldgtemplate, logical.roomtemplate, logical.tag,
//     organisational. If none provided, all namespaces are used by default.'
// responses:
//		'204':
//			description: Successfully deleted object
//		'404':
//			description: Not found. An error message will be returned

func HandleGenericObjects(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 HandleGenericObjects ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)
	matchingObjects := []map[string]interface{}{}

	// Get user roles for permissions
	user := getUserFromToken(w, r)
	if user == nil {
		return
	}

	// Get objects
	filters := getFiltersFromQueryParams(r)
	req := u.FilteredReqFromQueryParams(r.URL)
	entities := u.GetEntitiesByNamespace(filters.Namespace, filters.Id)

	for _, entStr := range entities {
		// Get objects
		entData, err := models.GetManyObjects(entStr, req, filters, "", user.Roles)
		if err != nil {
			u.ErrLog("Error while looking for objects at  "+entStr, "HandleGenericObjects", err.Message, r)
			u.RespondWithError(w, err)
			return
		}

		// Save entity to help delete and respond
		for _, obj := range entData {
			obj["entity"] = entStr
		}

		if nLimit, e := strconv.Atoi(filters.Limit); e == nil && nLimit > 0 && req["id"] != nil {
			// Get children until limit level (only for GET)
			for _, obj := range entData {
				obj["children"], err = models.GetHierarchyByName(entStr, obj["id"].(string), nLimit, filters)
				if err != nil {
					u.ErrLog("Error while getting "+entStr, "GET "+entStr, err.Message, r)
					u.RespondWithError(w, err)
				}
			}
		}
		matchingObjects = append(matchingObjects, entData...)
	}

	// Respond
	if r.Method == "DELETE" {
		for _, obj := range matchingObjects {
			entStr := obj["entity"].(string)

			var objStr string

			if u.IsEntityNonHierarchical(u.EntityStrToInt(entStr)) {
				objStr = obj["slug"].(string)
			} else {
				objStr = obj["id"].(string)
			}

			modelErr := models.DeleteObject(entStr, objStr, user.Roles)
			if modelErr != nil {
				u.ErrLog("Error while deleting object: "+objStr, "DELETE GetGenericObjectById", modelErr.Message, r)
				u.RespondWithError(w, modelErr)
				return
			}
			eventNotifier <- u.FormatNotifyData("delete", entStr, objStr)
		}
		u.Respond(w, u.RespDataWrapper("successfully deleted objects", matchingObjects))
	} else if r.Method == "OPTIONS" {
		u.WriteOptionsHeader(w, "GET")
	} else {
		matchingObjects = pie.Map(matchingObjects, func(object map[string]any) map[string]any {
			entityStr := object["entity"].(string)
			delete(object, "entity")

			return imageIDToUrl(u.EntityStrToInt(entityStr), object)
		})
		u.Respond(w, u.RespDataWrapper("successfully processed request", matchingObjects))
	}
}

// swagger:operation POST /api/objects/search Objects HandleComplexFilters
// Get all objects from any entity that match the complex filter. Return as a list.
// Wildcards can be used on any of the parameters present in query with equality and inequality operations.
// Check endpoint `HandleGenericObjects` for more information on wildcards
// ---
// security:
// - bearer: []
// produces:
// - application/json
// parameters:
//   - name: id
//     in: path
//     description: 'id of the object to obtain.
//     If none provided, all objects of the namespace will be obtained'
//   - name: namespace
//     in: query
//     description: 'One of the values: physical, physical.stray, physical.hierarchy,
//     logical, logical.objtemplate, logical.bldgtemplate, logical.roomtemplate, logical.tag,
//     organisational.
//     If none provided, all namespaces are used by default.'
//   - name: fieldOnly
//     in: query
//     description: 'specify which object field to show in response.
//     Multiple fieldOnly can be added. An invalid field is simply ignored.'
//   - name: startDate
//     in: query
//     description: 'filter objects by lastUpdated >= startDate.
//     Format: yyyy-mm-dd'
//   - name: endDate
//     in: query
//     description: 'filter objects by lastUpdated <= endDate.
//     Format: yyyy-mm-dd'
//   - name: attributes
//     in: query
//     description: 'Any other object attributes can be queried.
//     Replace attributes here by the name of the attribute followed by its value.'
//     required: false
//     type: string
//     default: domain=DemoDomain
//     example: vendor=ibm ; name=siteA ; orientation=front
//   - name: body
//     in: body
//     description: 'A JSON containing a mongoDB query to select and filter the desired objects.
//     Operators can be `$not`, `$lt`, `$lte`, `$gt`, `$gte`, `$and` and `$or`.
//     For equality, the syntax is: `[field]: value`.
//     Objects can be filtered by any of their properties and attributes.'
//     required: true
//     default: {}
//     example: '{"$and": [{"domain": "DemoDomain"}, {"attributes.height": {"$lt": "3"}}]}'
// responses:
//      '200':
//          description: 'Found. A response body will be returned with
//          a meaningful message.'
//      '400':
//         description: 'Bad request. Request has wrong format.'
//      '500':
//          description: Internal Error. A system error stopped the request.

// swagger:operation DELETE /api/objects Objects HandleDeleteComplexFilters
// Deletes an object that matches the complex filter in the system from any of the entities with no need to specify it.
// Wildcards can be used on any of the parameters present in query.
// Check endpoint `HandleGenericObjects` for more information on wildcards
// ---
// security:
// - bearer: []
// produces:
// - application/json
// parameters:
//   - name: id
//     in: path
//     description: ID type hierarchyName of the object
//     required: true
//   - name: fieldOnly
//     in: query
//     description: 'specify which object field to show in response.
//     Multiple fieldOnly can be added. An invalid field is simply ignored.'
//   - name: startDate
//     in: query
//     description: 'filter objects by lastUpdated >= startDate.
//     Format: yyyy-mm-dd'
//   - name: endDate
//     in: query
//     description: 'filter objects by lastUpdated <= endDate.
//     Format: yyyy-mm-dd'
//   - name: namespace
//     in: query
//     description: 'One of the values: physical, physical.stray, physical.hierarchy,
//     logical, logical.objtemplate, logical.bldgtemplate, logical.roomtemplate, logical.tag,
//     organisational. If none provided, all namespaces are used by default.'
//   - name: attributes
//     in: query
//     description: 'Any other object attributes can be queried.
//     Replace attributes here by the name of the attribute followed by its value.'
//     required: false
//     type: string
//     default: domain=DemoDomain
//     example: vendor=ibm ; name=siteA ; orientation=front
//   - name: body
//     in: body
//     description: 'A JSON containing a mongoDB query to select and filter the desired objects.
//     Operators can be `$not`, `$lt`, `$lte`, `$gt`, `$gte`, `$and` and `$or`.
//     For equality, the syntax is: `[field]: value`.
//     Objects can be filtered by any of their properties and attributes.'
//     required: true
//     default: {}
//     example: '{"$and": [{"domain": "DemoDomain"}, {"attributes.height": {"$lt": "3"}}]}'
// responses:
//		'204':
//			description: Successfully deleted object
//		'404':
//			description: Not found. An error message will be returned

func HandleComplexFilters(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 HandleComplexFilters ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)
	var complexFilters map[string]interface{}
	var complexFilterExp string
	var ok bool
	matchingObjects := []map[string]interface{}{}

	// Get user roles for permissions
	user := getUserFromToken(w, r)
	if user == nil {
		return
	}

	if err := decodeRequestBody(w, r, &complexFilters); err != nil {
		return
	}

	if complexFilterExp, ok = complexFilters["filter"].(string); !ok || len(complexFilterExp) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message("Invalid body format: must contain a filter key with a not empty string as value"))
		u.ErrLog(ErrDecodingBodyMsg, "HANDLE COMPLEX FILTERS", "", r)
		return
	}
	println(complexFilterExp)

	// Get objects
	filters := getFiltersFromQueryParams(r)
	req := u.FilteredReqFromQueryParams(r.URL)
	entities := u.GetEntitiesByNamespace(filters.Namespace, filters.Id)

	for _, entStr := range entities {
		// Get objects
		entData, err := models.GetManyObjects(entStr, req, filters, complexFilterExp, user.Roles)
		if err != nil {
			u.ErrLog("Error while looking for objects at "+entStr, "HandleComplexFilters", err.Message, r)
			u.RespondWithError(w, err)
			return
		}

		// Save entity to help delete and respond
		for _, obj := range entData {
			obj["entity"] = entStr
			if entStr == "device" && strings.Contains(complexFilterExp, "virtual_config.type=node") {
				// add namespace prefix to device nodes
				obj["id"] = "Physical." + obj["id"].(string)
			}
		}

		matchingObjects = append(matchingObjects, entData...)
	}

	if r.Method == "DELETE" {
		for _, obj := range matchingObjects {
			entStr := obj["entity"].(string)

			var objStr string

			if u.IsEntityNonHierarchical(u.EntityStrToInt(entStr)) {
				objStr = obj["slug"].(string)
			} else {
				objStr = obj["id"].(string)
			}

			modelErr := models.DeleteObject(entStr, objStr, user.Roles)
			if modelErr != nil {
				u.ErrLog("Error while deleting object: "+objStr, "DELETE GetGenericObjectById", modelErr.Message, r)
				u.RespondWithError(w, modelErr)
				return
			}
		}
		u.Respond(w, u.RespDataWrapper("successfully deleted objects", matchingObjects))
	} else if r.Method == "OPTIONS" {
		u.WriteOptionsHeader(w, "POST")
	} else {
		u.Respond(w, u.RespDataWrapper("successfully processed request", matchingObjects))
	}
}

// swagger:operation GET /api/{entity}/{id} Objects GetEntity
// Gets an Object of the given entity.
// The ID or hierarchy name must be provided in the URL parameter.
// ---
// security:
// - bearer: []
// produces:
// - application/json
// parameters:
//   - name: entity
//     in: path
//     description: 'Entity (same as category) of the object. Accepted values: sites, domains,
//     buildings, rooms, racks, devices, acs, panels,
//     cabinets, groups, corridors, virtual_objs
//     room_templates, obj_templates, bldg_templates, tags,
//     stray_objects, hierarchy_objects.'
//     required: true
//     type: string
//     default: "sites"
//   - name: id
//     in: path
//     description: 'ID of desired object.
//     For templates and tags the slug is the ID.'
//     required: true
//     type: string
//     default: "siteA"
//   - name: fieldOnly
//     in: query
//     description: 'specify which object field to show in response.
//     Multiple fieldOnly can be added. An invalid field is simply ignored.'
//   - name: startDate
//     in: query
//     description: 'filter objects by lastUpdated >= startDate.
//     Format: yyyy-mm-dd'
//   - name: endDate
//     in: query
//     description: 'filter objects by lastUpdated <= endDate.
//     Format: yyyy-mm-dd'
// responses:
// 	'200':
// 	  description: 'Found. A response body will be returned with
// 	  a meaningful message.'
// 	'400':
// 	  description: Bad request. An error message will be returned.
// 	'404':
// 	  description: Not Found. An error message will be returned.

func GetEntity(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 GetEntity ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)
	var data map[string]interface{}
	var id string
	var canParse bool
	var modelErr *u.Error

	// Get user roles for permissions
	user := getUserFromToken(w, r)
	if user == nil {
		return
	}

	// Get entity type and filters
	entityStr := mux.Vars(r)["entity"]
	filters := getFiltersFromQueryParams(r)

	// Get entity
	if id, canParse = mux.Vars(r)["id"]; canParse {
		var req primitive.M
		if entityStr == u.HIERARCHYOBJS_ENT {
			data, modelErr = models.GetHierarchyObjectById(id, filters, user.Roles)
		} else {
			if u.IsEntityNonHierarchical(u.EntityStrToInt(entityStr)) {
				// Get by slug
				req = bson.M{"slug": id}

			} else {
				req = bson.M{"id": id}
			}

			data, modelErr = models.GetObject(req, entityStr, filters, user.Roles)
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message("Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET ENTITY", "", r)
		return
	}

	// Respond
	if r.Method == "OPTIONS" && data != nil {
		u.WriteOptionsHeader(w, "GET, DELETE, OPTIONS, PATCH, PUT")
	} else {
		if modelErr != nil {
			u.ErrLog("Error while getting "+entityStr, "GET "+strings.ToUpper(entityStr),
				modelErr.Message, r)
			u.RespondWithError(w, modelErr)
		} else {
			imageIDToUrl(u.EntityStrToInt(entityStr), data)

			u.Respond(w, u.RespDataWrapper("successfully got "+entityStr, data))
		}
	}
}

// swagger:operation GET /api/layers/{slug}/objects Objects GetLayerObjects
// Gets the object of a given layer.
// Apply the layer filters to get children objects of a given root query param.
// ---
// security:
// - bearer: []
// produces:
// - application/json
// parameters:
//   - name: slug
//     in: path
//     description: 'ID of desired layer.'
//     required: true
//     type: string
//     default: "layer_slug"
//   - name: root
//     in: query
//     description: 'Mandatory, accepts IDs. The root object from where to apply the layer'
//     required: true
//   - name: recursive
//     in: query
//     description: 'Accepts true or false. If true, get objects
//     from all levels beneath root. If false, get objects directly under root.'
// responses:
// 	'200':
// 	  description: 'Found. A response body will be returned with
// 	  a meaningful message.'
// 	'400':
// 	  description: Bad request. An error message will be returned.
// 	'404':
// 	  description: Not Found. An error message will be returned.

func GetLayerObjects(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 GetLayerObjects ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)
	var data map[string]interface{}
	var id string
	var canParse bool
	var modelErr *u.Error

	// Get user roles for permissions
	user := getUserFromToken(w, r)
	if user == nil {
		return
	}

	// Get query params
	var filters u.LayerObjsFilters
	decoder.Decode(&filters, r.URL.Query())
	if filters.Root == "" {
		//error
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message("Query param root is mandatory"))
		return
	}

	if id, canParse = mux.Vars(r)["slug"]; canParse {
		// Get layer
		data, modelErr = models.GetObject(bson.M{"slug": id}, u.EntityToString(u.LAYER), u.RequestFilters{}, user.Roles)
		if modelErr != nil {
			u.RespondWithError(w, modelErr)
			return
		} else if len(data) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Apply layer to get objects request
		req := bson.M{}
		var searchId string
		if filters.IsRecursive {
			searchId = filters.Root + ".**.*"
		} else {
			searchId = filters.Root + ".*"
		}
		u.AddFilterToReq(req, "id", searchId)

		// Get objects
		matchingObjects := []map[string]interface{}{}
		entities := u.GetEntitiesByNamespace(u.Any, searchId)
		fmt.Println(req)
		fmt.Println(entities)
		for _, entStr := range entities {
			entData, err := models.GetManyObjects(entStr, req, u.RequestFilters{}, data["filter"].(string), user.Roles)
			if err != nil {
				u.RespondWithError(w, err)
				return
			}
			matchingObjects = append(matchingObjects, entData...)
		}

		// Respond
		if r.Method == "OPTIONS" {
			u.WriteOptionsHeader(w, "GET, DELETE, PATCH, PUT")
		} else {
			u.Respond(w, u.RespDataWrapper("successfully processed request", matchingObjects))
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message("Error while parsing path parameters"))
		return
	}
}

// swagger:operation GET /api/{entity} Objects GetAllEntities
// Gets all present objects for specified entity (category).
// Returns JSON body with all specified objects of type.
// ---
// security:
// - bearer: []
// produces:
// - application/json
// parameters:
//   - name: entity
//     in: path
//     description: 'Entity (same as category) of the object. Accepted values: sites, domains,
//     buildings, rooms, racks, devices, acs, panels,
//     cabinets, groups, corridors, virtual_objs
//     room_templates, obj_templates, bldg_templates, stray_objects, tags'
//     required: true
//     type: string
//     default: "sites"
//   - name: fieldOnly
//     in: query
//     description: 'specify which object field to show in response.
//     Multiple fieldOnly can be added. An invalid field is simply ignored.'
//   - name: startDate
//     in: query
//     description: 'filter objects by lastUpdated >= startDate.
//     Format: yyyy-mm-dd'
//   - name: endDate
//     in: query
//     description: 'filter objects by lastUpdated <= endDate.
//     Format: yyyy-mm-dd'
//
// responses:
//		'200':
//			description: 'Found. A response body will be returned with
//			a meaningful message.'
//		'404':
//			description: Nothing Found. An error message will be returned.

func GetAllEntities(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 GetAllEntities ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)
	var entStr string

	// Get user roles for permissions
	user := getUserFromToken(w, r)
	if user == nil {
		return
	}

	// Get entity
	entStr = mux.Vars(r)["entity"]
	println("ENTSTR: ", entStr)

	// Check if entity is valid
	entity := u.EntityStrToInt(entStr)
	if entity < 0 {
		w.WriteHeader(http.StatusNotFound)
		u.Respond(w, u.Message("Invalid entity in URL: '"+mux.Vars(r)["entity"]+
			"' Please provide a valid entity"))
		u.ErrLog("Cannot get invalid entity", "GET "+mux.Vars(r)["entity"], "", r)
		return
	}

	// Get entities
	req := bson.M{}
	data, e := models.GetManyObjects(entStr, req, u.RequestFilters{}, "", user.Roles)

	queryValues, _ := url.ParseQuery(r.URL.RawQuery)
	if entity == u.VIRTUALOBJ && queryValues.Get("limit") == "1" {
		// limit=1 used to get only root nodes of virtual objs
		data = getVirtualRootObjects(data)
	}

	// Respond
	if e != nil {
		u.ErrLog("Error while getting "+entStr+"s", "GET ALL "+strings.ToUpper(entStr),
			e.Message, r)
		u.RespondWithError(w, e)
	} else {
		if entity == u.TAG {
			data = pie.Map(data, func(tagData map[string]any) map[string]any {
				return imageIDToUrl(entity, tagData)
			})
		}

		u.Respond(w, u.RespDataWrapper("successfully got "+entStr+"s", data))
	}
}

func getVirtualRootObjects(data []map[string]any) []map[string]any {
	objects := []map[string]any{}
	fmt.Println(data)
	for _, comparingObj := range data {
		shouldAdd := true
		comparingObjName := comparingObj["id"].(string)
		for _, obj := range data {
			objName := obj["id"].(string)
			if comparingObjName != objName && strings.HasPrefix(comparingObjName, objName) {
				// already has its parent, no need for this one
				shouldAdd = false
				break
			}
		}
		if shouldAdd {
			objects = append(objects, comparingObj)
		}
	}
	return objects
}

// swagger:operation DELETE /api/{entity}/{id} Objects DeleteObject
// Deletes an Object in the system.
// ---
// security:
// - bearer: []
// produces:
// - application/json
// parameters:
//   - name: entity
//     in: path
//     description: 'Entity (same as category) of the object. Accepted values: sites, domains,
//     buildings, rooms, racks, devices, acs, panels,
//     cabinets, groups, corridors, virtual_objs
//     room_templates, obj_templates, bldg_templates, tags,
//     stray_objects, hierarchy_objects.'
//     required: true
//     type: string
//     default: "sites"
//   - name: id
//     in: path
//     description: 'ID of desired object.
//     For templates and tags the slug is the ID.'
//     required: true
//     type: string
//     default: "siteA"
//
// responses:
//		'204':
//			description: 'Successfully deleted object.
//			No response body will be returned'
//		'404':
//			description: Not found. An error message will be returned

func DeleteEntity(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 DeleteEntity ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)

	// Get user roles for permissions
	user := getUserFromToken(w, r)
	if user == nil {
		return
	}

	// Get entityStr from URL
	entityStr := mux.Vars(r)["entity"]

	// Check unidentified collection
	if u.EntityStrToInt(entityStr) < 0 && entityStr != u.HIERARCHYOBJS_ENT {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message("Invalid object in URL: '"+mux.Vars(r)["entity"]+
			"' Please provide a valid object"))
		u.ErrLog("Cannot delete invalid object", "DELETE "+mux.Vars(r)["entity"], "", r)
		return
	}

	// Check id and try delete
	id := mux.Vars(r)["id"]
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message("Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "DELETE ENTITY", "", r)
	} else {
		if entityStr == u.HIERARCHYOBJS_ENT {
			obj, err := models.GetHierarchyObjectById(id, u.RequestFilters{}, user.Roles)
			if err != nil {
				u.ErrLog("Error finding hierarchy obj to delete", "DELETE ENTITY", err.Message, r)
				u.RespondWithError(w, err)
				return
			} else {
				entityStr = obj["category"].(string)
			}
		}

		modelErr := models.DeleteObject(entityStr, id, user.Roles)
		if modelErr != nil {
			u.ErrLog("Error while deleting entity", "DELETE ENTITY", modelErr.Message, r)
			u.RespondWithError(w, modelErr)
		} else {
			w.WriteHeader(http.StatusNoContent)
			u.Respond(w, u.Message("successfully deleted"))
			eventNotifier <- u.FormatNotifyData("delete", entityStr, id)
		}
	}
}

// swagger:operation PATCH /api/{entity}/{id} Objects PartialUpdateObject
// Partially update object.
// This is the preferred method for modifying data in the system.
// If you want to do a full data replace, please use PUT instead.
// If no data is effectively changed, an OK will still be returned.
// ---
// security:
// - bearer: []
// produces:
// - application/json
// parameters:
//   - name: entity
//     in: path
//     description: 'Entity (same as category) of the object. Accepted values: sites, domains,
//     buildings, rooms, racks, devices, acs, panels,
//     cabinets, groups, corridors, virtual_objs
//     room_templates, obj_templates, bldg_templates, stray_objects, tags.'
//     required: true
//     type: string
//     default: "sites"
//   - name: id
//     in: path
//     description: 'ID of desired object.
//     For templates and tags the slug is the ID.'
//     required: true
//     type: string
//     default: "siteA"
//   - name: body
//     in: body
//     description: An object with the attributes to be changed
//     type: json
//     required: true
//     example: '{"domain": "mynewdomain"}'
//
// responses:
//     '200':
//         description: 'Updated. A response body will be returned with
//         a meaningful message.'
//     '400':
//         description: Bad request. An error message will be returned.
//     '404':
//         description: Not Found. An error message will be returned.

// swagger:operation PUT /api/{entity}/{id} Objects UpdateObject
// Completely update object.
// This method will replace the existing data with the JSON
// received, thus fully replacing the data. If you do not
// want to do this, please use PATCH.
// If no data is effectively changed, an OK will still be returned.
// ---
// security:
// - bearer: []
// produces:
// - application/json
// parameters:
//   - name: entity
//     in: path
//     description: 'Entity (same as category) of the object. Accepted values: sites, domains,
//     buildings, rooms, racks, devices, acs, panels,
//     cabinets, groups, corridors, virtual_objs
//     room_templates, obj_templates, bldg_templates, tags,
//     stray_objects, hierarchy_objects.'
//     required: true
//     type: string
//     default: "sites"
//   - name: id
//     in: path
//     description: 'ID of desired object.
//     For templates and tags the slug is the ID.'
//     required: true
//     type: string
//     default: "siteA"
//
// responses:
//     '200':
//         description: 'Updated. A response body will be returned with
//         a meaningful message.'
//     '400':
//         description: Bad request. An error message will be returned.
//     '404':
//         description: Not Found. An error message will be returned.

func UpdateEntity(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 UpdateEntity ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)
	var data map[string]interface{}
	var modelErr *u.Error
	var entity string

	// Get user roles for permissions
	user := getUserFromToken(w, r)
	if user == nil {
		return
	}

	// Patch or put
	isPatch := false
	if r.Method == http.MethodPatch {
		isPatch = true
	}

	// Get request body
	updateData := map[string]interface{}{}
	if err := decodeRequestBody(w, r, &updateData); err != nil {
		return
	}

	//Get entity from URL
	entity = mux.Vars(r)["entity"]

	// Check unidentified collection
	if u.EntityStrToInt(entity) < 0 && entity != u.HIERARCHYOBJS_ENT {
		w.WriteHeader(http.StatusNotFound)
		u.Respond(w, u.Message("Invalid object in URL: '"+mux.Vars(r)["entity"]+"' Please provide a valid object"))
		u.ErrLog("Cannot update invalid object", "UPDATE "+mux.Vars(r)["entity"], "", r)
		return
	}

	// Get query params
	queryValues, _ := url.ParseQuery(r.URL.RawQuery)
	isRecursiveUpdate := queryValues.Get("recursive") == "true"

	// Check id and try update
	id := mux.Vars(r)["id"]
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message("Error while extracting from path parameters"))
		u.ErrLog("Error while extracting from path parameters", "UPDATE ENTITY", "", r)
	} else {
		data, modelErr = models.UpdateObject(entity, id, updateData, isPatch, user.Roles, isRecursiveUpdate)
		if modelErr != nil {
			u.RespondWithError(w, modelErr)
		} else {
			u.Respond(w, u.RespDataWrapper("successfully updated "+entity, data))
			if entity == "tag" || entity == "layer" {
				data = map[string]any{
					"old-slug": id,
					entity:     data,
				}
			}
			eventNotifier <- u.FormatNotifyData("modify", entity, data)
		}
	}
}

// swagger:operation GET /api/{entity}? Objects GetEntityByQuery
// Gets an object filtering by attribute.
// Gets an Object using any attribute (with the exception of description)
// via query in the system
// The attributes are in the form {attr}=xyz&{attr1}=abc
// And any combination can be used given that at least 1 is provided.
// ---
// security:
// - bearer: []
// produces:
// - application/json
// parameters:
//   - name: entity
//     in: path
//     description: 'Entity (same as category) of the object. Accepted values: sites, domains,
//     buildings, rooms, racks, devices, acs, panels,
//     cabinets, groups, corridors, virtual_objs
//     room_templates, obj_templates, bldg_templates, stray_objects, tags.'
//     required: true
//     type: string
//     default: "sites"
//   - name: fieldOnly
//     in: query
//     description: 'specify which object field to show in response.
//     Multiple fieldOnly can be added. An invalid field is simply ignored.'
//   - name: startDate
//     in: query
//     description: 'filter objects by lastUpdated >= startDate.
//     Format: yyyy-mm-dd'
//   - name: endDate
//     in: query
//     description: 'filter objects by lastUpdated <= endDate.
//     Format: yyyy-mm-dd'
//   - name: attributes
//     in: query
//     description: 'Any other object attributes can be queried.
//     Replace attributes here by the name of the attribute followed by its value.'
//     required: false
//     type: string
//     default: domain=DemoDomain
//     example: vendor=ibm ; name=siteA ; orientation=front
//
// responses:
//     '204':
//         description: 'Found. A response body will be returned with
//         a meaningful message.'
//     '404':
//         description: Not found. An error message will be returned.

func GetEntityByQuery(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 GetEntityByQuery ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)
	var data []map[string]interface{}
	var entStr string
	var modelErr *u.Error

	// Get user roles for permissions
	user := getUserFromToken(w, r)
	if user == nil {
		return
	}

	// Get entity
	entStr = r.URL.Path[5 : len(r.URL.Path)-1]

	// Check unidentified collection
	entInt := u.EntityStrToInt(entStr)
	if entInt < 0 {
		w.WriteHeader(http.StatusNotFound)
		u.Respond(w, u.Message("Invalid object in URL: '"+entStr+"' Please provide a valid object"))
		u.ErrLog("Cannot get invalid object", "GET ENTITYQUERY"+entStr, "", r)
		return
	}

	// Get query params
	filters := getFiltersFromQueryParams(r)
	bsonMap := u.FilteredReqFromQueryParams(r.URL)
	// Limit filter
	if entInt == u.DOMAIN || entInt == u.DEVICE || entInt == u.STRAYOBJ {
		if nLimit, e := strconv.Atoi(filters.Limit); e == nil {
			startLimit := "0"
			endLimit := filters.Limit
			if entInt == u.DEVICE {
				// always at least 4 levels (site.bldg.room.rack.dev)
				startLimit = "4"
				endLimit = strconv.Itoa(nLimit + 4)
			}
			pattern := primitive.Regex{Pattern: "^" + u.NAME_REGEX + "(." + u.NAME_REGEX +
				"){" + startLimit + "," + endLimit + "}$", Options: ""}
			bsonMap = bson.M{"id": pattern}
		}
	}

	data, modelErr = models.GetManyObjects(entStr, bsonMap, filters, "", user.Roles)

	if modelErr != nil {
		u.ErrLog("Error while getting "+entStr, "GET ENTITYQUERY", modelErr.Message, r)
		u.RespondWithError(w, modelErr)
	} else {
		u.Respond(w, u.RespDataWrapper("successfully got query for "+entStr, data))
	}
}

// swagger:operation GET /api/tempunits/{id} Objects GetTempUnit
// Gets the temperatureUnit attribute of the parent site of given object.
// ---
// security:
// - bearer: []
// produces:
// - application/json
// parameters:
//   - name: id
//     in: path
//     description: 'ID of desired object'
//     required: true
//     type: string
//     default: "siteA"
// responses:
//  '200':
//     description: 'Found. A response body will be returned with
//     a meaningful message.'
//  '404':
//     description: 'Nothing Found. An error message will be returned.'

// swagger:operation GET /api/sitecolors/{id} Objects GetSiteColors
// Gets the colors attributes of the parent site of given object.
// Returned attributes are always: reservedColor, usableColor and technicalColor.
// ---
// security:
// - bearer: []
// produces:
// - application/json
// parameters:
//   - name: id
//     in: path
//     description: 'ID of desired object'
//     required: true
//     type: string
//     default: "siteA"
// responses:
//  '200':
//     description: 'Found. A response body will be returned with
//     a meaningful message.'
//  '404':
//     description: 'Nothing Found. An error message will be returned.'

func GetSiteAttr(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 GetSiteAttr ")
	fmt.Println("******************************************************")

	// Check id
	id := mux.Vars(r)["id"]
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message("Error while extracting id from URL"))
	}
	siteAttr := mux.Vars(r)["siteAttr"]
	if siteAttr == "tempunits" {
		siteAttr = "temperatureUnit"
	}

	// Try get tempUnit and respond
	data, err := models.GetSiteParentAttribute(id, siteAttr)
	if err != nil {
		u.RespondWithError(w, err)
	} else {
		if r.Method == "OPTIONS" {
			u.WriteOptionsHeader(w, "GET, HEAD")
		} else {
			resp := u.RespDataWrapper(
				"successfully got attribute from object's parent site",
				data)
			u.Respond(w, resp)
		}
	}
}

// swagger:operation GET /api/{entity}/{id}/{subent} Objects GetEntitiesOfAncestor
// Obtain all children object of given id that belong to subent entity.
// Subent must be lower than entity in the hierarchy.
// Examples:
// /api/sites/{id}/rooms will return all rooms of a site
// /api/room/{id}/devices will return all devices of a room
// Returns a JSON body with all children objects under the parent object.
// ---
// security:
// - bearer: []
// produces:
// - application/json
// parameters:
// - name: entity
//   in: path
//   description: 'Indicates the entity.'
//   required: true
//   type: string
//   default: sites
// - name: ID
//   in: path
//   description: ID of object
//   required: true
//   type: string
//   default: siteA
// - name: subent
//   in: path
//   description: 'Indicates the subentity to search for children.'
//   required: true
//   type: string
//   default: buildings
// responses:
//     '200':
//         description: 'Found. A response body will be returned with
//         a meaningful message.'
//     '404':
//         description: Nothing Found. An error message will be returned.

func GetEntitiesOfAncestor(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 GetEntitiesOfAncestor ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)
	var id string
	var e bool

	// Get user roles for permissions
	user := getUserFromToken(w, r)
	if user == nil {
		return
	}

	// Get entities
	entStr := mux.Vars(r)["ancestor"]
	subEnt := mux.Vars(r)["sub"]
	entInt := u.EntityStrToInt(entStr)
	subInt := u.EntityStrToInt(subEnt)

	// Check unidentified collection
	if entInt < 0 || subInt < 0 {
		w.WriteHeader(http.StatusNotFound)
		u.Respond(w, u.Message("Invalid object in URL: '"+entStr+"' Please provide a valid object"))
		u.ErrLog("Cannot get invalid object", "GET CHILDRENOFPARENT"+entStr, "", r)
		return
	} else if entInt >= subInt {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message("Invalid set of entities in URL: first entity should be parent of the second entity"))
		u.ErrLog("Invalid set of entities", "GET CHILDRENOFPARENT"+entStr, "", r)
		return
	}

	// Get id from URL
	id, e = mux.Vars(r)["id"]
	if !e {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message("Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET CHILDRENOFPARENT", "", r)
		return
	}

	data, modelErr := models.GetEntitiesOfAncestor(id, entStr, subEnt, user.Roles)
	if modelErr != nil {
		u.ErrLog("Error while getting children of "+entStr,
			"GET CHILDRENOFPARENT", modelErr.Message, r)
		u.RespondWithError(w, modelErr)
	} else if r.Method == "OPTIONS" {
		u.WriteOptionsHeader(w, "GET")
	} else {
		u.Respond(w, u.RespDataWrapper("successfully got object", data))
	}
}

// swagger:operation GET /api/{entity}/{id}/all Objects GetEntityHierarchy
// Get object and all its children.
// Returns JSON body with all subobjects under the Object.
// ---
// security:
// - bearer: []
// produces:
// - application/json
// parameters:
// - name: entity
//   in: path
//   description: 'Entity (same as category) of the object. Accepted values: sites, domains,
//   buildings, rooms, racks, devices, acs, panels,
//   cabinets, groups, corridors, virtual_objs
//   stray_objects, hierarchy_objects.'
//   required: true
//   type: string
//   default: "sites"
// - name: id
//   in: path
//   description: 'ID of desired object.'
//   required: true
//   type: string
//   default: "siteA"
// - name: limit
//   in: query
//   description: 'Limits the level of hierarchy for retrieval. if not
//   specified for devices then the default value is maximum.
//   Example: /api/devices/{id}/all?limit=2'
//   required: false
//   type: string
//   default: 1
// - name: fieldOnly
//   in: query
//   description: 'specify which object field to show in response.
//   Multiple fieldOnly can be added. An invalid field is simply ignored.'
// - name: startDate
//   in: query
//   description: 'filter objects by lastUpdated >= startDate.
//   Format: yyyy-mm-dd'
// - name: endDate
//   in: query
//   description: 'filter objects by lastUpdated <= endDate.
//   Format: yyyy-mm-dd'
// responses:
//     '200':
//         description: 'Found. A response body will be returned with
//         a meaningful message.'
//     '404':
//         description: Nothing Found. An error message will be returned.

func GetHierarchyByName(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 GetHierarchyByName ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)
	var limit int

	// Get user roles for permissions
	user := getUserFromToken(w, r)
	if user == nil {
		return
	}

	// Get id and entity
	id, e := mux.Vars(r)["id"]
	entity, e2 := mux.Vars(r)["entity"]
	if !e || !e2 {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message("Error while parsing URL"))
		u.ErrLog("Error while parsing path parameters", "GetHierarchyByName", "", r)
		return
	}

	// Check if the request is a ranged hierarchy
	filters := getFiltersFromQueryParams(r)
	if len(filters.Limit) > 0 {
		//limit={number} was provided
		limit, _ = strconv.Atoi(filters.Limit)
	} else {
		limit = 999
	}

	println("The limit is: ", limit)

	// Get object and its family
	var modelErr *u.Error
	var data map[string]interface{}
	if entity == u.HIERARCHYOBJS_ENT {
		// Generic endpoint only for physical objs
		data, modelErr = models.GetHierarchyObjectById(id, filters, user.Roles)
		if modelErr == nil {
			entity = data["category"].(string)
		}
	} else {
		// Entity already known
		data, modelErr = models.GetObject(bson.M{"id": id}, entity, filters, user.Roles)
	}
	if limit >= 1 && modelErr == nil {
		if entity == u.EntityToString(u.STRAYOBJ) {
			// use stray's category as entity
			entity = data["category"].(string)
		}

		if vconfig, ok := data["attributes"].(map[string]any)["virtual_config"].(map[string]any); ok && entity == u.EntityToString(u.VIRTUALOBJ) && vconfig["type"] == "cluster" {
			data["children"], modelErr = models.GetHierarchyByCluster(id, limit, filters)
		} else {
			data["children"], modelErr = models.GetHierarchyByName(entity, id, limit, filters)
		}
	}

	// Respond
	if modelErr != nil {
		u.ErrLog("Error while getting "+entity, "GET "+entity, modelErr.Message, r)
		u.RespondWithError(w, modelErr)
	} else if r.Method == "OPTIONS" {
		u.WriteOptionsHeader(w, "GET")
	} else {
		u.Respond(w, u.RespDataWrapper("successfully got object's hierarchy", data))
	}
}

// swagger:operation GET /api/hierarchy Objects GetCompleteHierarchy
// Returns system complete hierarchy.
// Return is arranged by relationship (father:[children]), starting with "\*":[sites].
// The "\*" indicates root.
// User permissions apply.
// ---
// security:
// - bearer: []
// produces:
// - application/json
// parameters:
//   - name: namespace
//     in: query
//     description: 'One of the values: physical, logical or organisational.
//     If none provided, all namespaces are used by default.'
//   - name: withcategories
//     in: query
//     description: 'besides the hierarchy, returns also an structure with
//     the objects organized by category.'
//   - name: startDate
//     in: query
//     description: 'filter objects by lastUpdated >= startDate.
//     Format: yyyy-mm-dd'
//   - name: endDate
//     in: query
//     description: 'filter objects by lastUpdated <= endDate.
//     Format: yyyy-mm-dd'
// responses:
//		'200':
//			description: 'Request is valid.'
//		'500':
//			description: Server error.

func GetCompleteHierarchy(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 GetCompleteHierarchy ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)

	// Get user roles for permissions
	user := getUserFromToken(w, r)
	if user == nil {
		return
	}

	var filters u.HierarchyFilters
	decoder.Decode(&filters, r.URL.Query())

	data, err := models.GetCompleteHierarchy(user.Roles, filters)
	if err != nil {
		u.RespondWithError(w, err)
	} else {
		if r.Method == "OPTIONS" {
			u.WriteOptionsHeader(w, "GET, HEAD")
		} else {
			u.Respond(w, u.RespDataWrapper("successfully got hierarchy", data))
		}
	}
}

// swagger:operation GET /api/hierarchy/attributes Objects GetCompleteHierarchyAttrs
// Returns attributes of all objects.
// Return is arranged by hierarchyName (objHierarchyName:{attributes}).
// User permissions apply.
// ---
// security:
// - bearer: []
// produces:
// - application/json
// responses:
//		'200':
//			description: 'Request is valid.'
//		'500':
//			description: Server error.

func GetCompleteHierarchyAttributes(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 GetCompleteHierarchyAttributes ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)

	// Get user roles for permissions
	user := getUserFromToken(w, r)
	if user == nil {
		return
	}

	data, err := models.GetCompleteHierarchyAttributes(user.Roles)
	if err != nil {
		u.RespondWithError(w, err)
	} else {
		if r.Method == "OPTIONS" {
			u.WriteOptionsHeader(w, "GET, HEAD")
		} else {
			u.Respond(w, u.RespDataWrapper("successfully got attrs hierarchy", data))
		}
	}
}

// swagger:operation PATCH /api/{entity}/{id}/unlink Objects UnlinkObject
// Removes the object from its original entity and hierarchy tree to make it stray.
// The object will no longer have a parent, its id will change as well as the id of all its children.
// The object will then belong to the stray_objects entity.
// ---
// security:
// - bearer: []
// produces:
// - application/json
// parameters:
//   - name: id
//     in: path
//     description: 'ID of desired object.'
//     required: true
//     type: string
//     default: "Site.Building.Room.RackB"
//   - name: entity
//     in: path
//     description: 'Entity (same as category) of the object. Accepted values:
//     buildings, rooms, racks, devices, acs, panels,
//     cabinets, groups, corridors, virtual_objs.'
//     required: true
//     type: string
//     default: "sites"
//   - name: body
//     in: body
//     required: false
//     description: 'Name is optional to change the name of the object when turning stray.'
//     default: {"name": "MyNewStrayObjectName"}
// responses:
//     '200':
//         description: 'Unlinked. The object is now a stray.'
//     '400':
//         description: 'Bad request. Request has wrong format.'
//     '500':
//         description: 'Internal error. Unable to remove object from entity and create it as stray.'

// swagger:operation PATCH /api/stray_objects/{id}/link Objects LinkObject
// Removes the object from stray and add it to the entity of its category attribute.
// The object will again have a parent, its id will change as well as the id of all its children.
// The object will then belong to the given entity.
// ---
// security:
// - bearer: []
// produces:
// - application/json
// parameters:
//   - name: id
//     in: path
//     description: 'ID of desired object.'
//     required: true
//     type: string
//     default: "StrayRackB"
//   - name: body
//     in: body
//     required: true
//     description: 'ParentId is mandatory. Name is optional.'
//     default: {"parentId": "MyNewParent", "name": "MyNewObjectName"}
// responses:
//     '200':
//         description: 'Linked. The object is no longer a stray.'
//     '400':
//         description: 'Bad request. Request has wrong format.'
//     '500':
//         description: 'Internal error. Unable to remove object from stray and create it in an entity.'

func LinkEntity(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 LinkEntity ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)
	var data map[string]interface{}
	var id string
	var canParse bool
	var modelErr *u.Error
	var body map[string]string
	var newName string

	// Get user roles for permissions
	user := getUserFromToken(w, r)
	if user == nil {
		return
	}
	entityStr, isUnlink := mux.Vars(r)["entity"]

	body = map[string]string{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if isUnlink {
		if err == nil && len(body) > 0 {
			if newName = body["name"]; newName == "" || len(body) > 1 {
				w.WriteHeader(http.StatusBadRequest)
				u.Respond(w, u.Message("Body must be empty or only contain valid name"))
				return
			}
		}
	} else {
		// It's link, get parentId from body
		if err != nil || body["parentId"] == "" {
			w.WriteHeader(http.StatusBadRequest)
			u.Respond(w, u.Message("Error while decoding request body: must contain parentId"))
			return
		}
		newName = body["name"]
		entityStr = "stray_object"
	}

	// Get entity
	if id, canParse = mux.Vars(r)["id"]; canParse {
		if entityStr == u.HIERARCHYOBJS_ENT {
			data, modelErr = models.GetHierarchyObjectById(id, u.RequestFilters{}, user.Roles)
		} else {
			data, modelErr = models.GetObject(bson.M{"id": id}, entityStr, u.RequestFilters{}, user.Roles)
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message("Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET ENTITY", "", r)
		return
	}
	if modelErr != nil {
		u.ErrLog("Error while getting "+entityStr, "GET "+strings.ToUpper(entityStr),
			modelErr.Message, r)
		u.RespondWithError(w, modelErr)
		return
	}

	// Adjust retrieved object to recreate it
	if isUnlink {
		delete(data, "parentId")
		if entityStr == "device" {
			delete(data, "slot")
			delete(data, "posU")
		}
		entityStr = "stray_object"
	} else {
		data["parentId"] = body["parentId"]
		entityStr = data["category"].(string)
		delete(body, "parentId")
		delete(body, "name")
		for attr, value := range body {
			println("add " + attr)
			println(value)
			data["attributes"].(map[string]any)[attr] = value
		}
	}
	if newName != "" {
		data["name"] = newName
	}
	// Remove api fields
	delete(data, "id")
	delete(data, "createdDate")
	delete(data, "lastUpdated")
	// Convert primitive.A and similar types
	bytes, _ := json.Marshal(data)
	json.Unmarshal(bytes, &data)

	createEnt := entityStr
	var deleteEnt string
	if isUnlink {
		deleteEnt = data["category"].(string)
	} else {
		deleteEnt = "stray_object"
	}

	if modelErr := models.SwapEntity(createEnt, deleteEnt, id, data, user.Roles); modelErr != nil {
		u.RespondWithError(w, modelErr)
		return
	} else {
		if isUnlink {
			u.Respond(w, u.Message("successfully unlinked"))
		} else {
			u.Respond(w, u.Message("successfully linked"))
		}
	}
}

func BaseOption(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 BaseOption ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)
	entity, e1 := mux.Vars(r)["entity"]
	if !e1 || u.EntityStrToInt(entity) == -1 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	u.WriteOptionsHeader(w, "GET, DELETE, PATCH, PUT, POST")
}

// swagger:operation GET /api/stats About GetStats
// Displays DB statistics.
// ---
// security:
// - bearer: []
// produces:
// - application/json
// responses:
//		'200':
//			description: 'Request is valid.'
//		'504':
//			description: Server error.

func GetStats(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 GetStats ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)
	if r.Method == "OPTIONS" {
		w.Header().Add("Allow", "GET, HEAD, OPTIONS")
	} else {
		r := models.GetStats()
		u.Respond(w, r)
	}
}

// swagger:operation POST /api/validate/{entity} Objects ValidateObject
// Checks the received data and verifies if the object can be created in the system.
// ---
// security:
// - bearer: []
// produces:
// - application/json
// parameters:
//   - name: entity
//     in: path
//     description: 'Entity (same as category) of the object. Accepted values: sites, domains,
//     buildings, rooms, racks, devices, acs, panels,
//     cabinets, groups, corridors, virtual_objs
//     room_templates, obj_templates, bldg_templates, stray_objects, tags.'
//     required: true
//     type: string
//     default: "sites"
//   - name: body
//     in: body
//     required: true
//     default: {}
// responses:
//     '200':
//         description: 'Createable. A response body will be returned with
//         a meaningful message.'
//     '400':
//         description: 'Bad request. A response body with an error
//         message will be returned.'
//     '404':
//         description: Not Found. An error message will be returned.

func ValidateEntity(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 ValidateEntity ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)
	var obj map[string]interface{}
	entity, e1 := mux.Vars(r)["entity"]

	// Get user roles for permissions
	user := getUserFromToken(w, r)
	if user == nil {
		return
	}

	entInt := u.EntityStrToInt(entity)

	if !e1 || entInt == -1 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if r.Method == "OPTIONS" {
		u.WriteOptionsHeader(w, "POST")
		return
	}

	if err := decodeRequestBody(w, r, &obj); err != nil {
		return
	}

	if u.IsEntityHierarchical(entInt) {
		var domain string
		if entInt == u.DOMAIN {
			domain = obj["parentId"].(string) + obj["name"].(string)
		} else {
			domain = obj["domain"].(string)
		}
		if permission := models.CheckUserPermissions(user.Roles, entInt, domain); permission < models.WRITE {
			w.WriteHeader(http.StatusUnauthorized)
			u.Respond(w, u.Message("This user"+
				" does not have sufficient permissions to create"+
				" this object under this domain "))
			u.ErrLog("Cannot validate object creation due to limited user privilege",
				"Validate CREATE "+entity, "", r)
			return
		}
	}

	// ok, err := models.ValidateEntity(entInt, obj)
	if ok, err := models.ValidateJsonSchema(entInt, obj); !ok {
		u.RespondWithError(w, err)
	} else {
		u.Respond(w, u.Message("This object can be created"))
	}
}

// swagger:operation GET /api/version About GetAPIVersion
// Gets the API version.
// ---
// security:
// - bearer: []
// produces:
// - application/json
// responses:
//     '200':
//         description: 'OK. A response body will be returned with
//         version details.'

func GetVersion(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{}
	if r.Method == "OPTIONS" {
		u.WriteOptionsHeader(w, "GET, HEAD")
		return
	} else {
		data["status"] = true
		data["data"] = map[string]interface{}{
			"BuildDate":  u.GetBuildDate(),
			"BuildHash":  u.GetBuildHash(),
			"CommitDate": u.GetCommitDate(),
			"BuildTree":  u.GetBuildTree(),
			"Customer":   models.GetDBName(),
		}
	}
	u.Respond(w, data)
}
