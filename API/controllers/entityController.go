package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"p3/models"
	u "p3/utils"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func getObjID(x string) (primitive.ObjectID, error) {
	objID, err := primitive.ObjectIDFromHex(x)
	if err != nil {
		return objID, err
	}
	return objID, nil
}

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
//     cabinets, groups, corridors,
//     room-templates, obj-templates, bldg-templates, stray-devices.'
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

	object := map[string]interface{}{}
	err := json.NewDecoder(r.Body).Decode(&object)

	entStr, _ := mux.Vars(r)["entity"]

	// Get user roles for permissions
	user := getUserFromToken(w, r)
	if user == nil {
		return
	}

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message("Error while decoding request body"))
		u.ErrLog("Error while decoding request body", "CREATE "+entStr, "", r)
		return
	}

	//If creating templates, format them
	entStr = strings.Replace(entStr, "-", "_", 1)

	entInt := u.EntityStrToInt(entStr)
	println("ENT: ", entStr)

	//Prevents Mongo from creating a new unidentified collection
	if entInt < 0 {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message("Invalid entity in URL: '"+mux.Vars(r)["entity"]+"' Please provide a valid object"))
		u.ErrLog("Cannot create invalid object", "CREATE "+mux.Vars(r)["entity"], "", r)
		return
	}

	//Check if category and endpoint match, except for templates and strays
	if entInt < u.ROOMTMPL {
		if object["category"] != entStr {
			w.WriteHeader(http.StatusBadRequest)
			u.Respond(w, u.Message("Category in request body does not correspond with desired object in endpoint"))
			u.ErrLog("Cannot create invalid object", "CREATE "+mux.Vars(r)["entity"], "", r)
			return
		}
	}

	//Clean the data of 'id' attribute if present
	delete(object, "id")

	resp, e := models.CreateEntity(entInt, object, user.Roles)
	if e != nil {
		u.ErrLog("Error creating "+entStr, "CREATE", e.Message, r)
		u.RespondWithError(w, e)
	} else {
		w.WriteHeader(http.StatusCreated)
		u.Respond(w, u.RespDataWrapper("successfully created "+entStr, resp))
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

	listDomains := []map[string]interface{}{}
	err := json.NewDecoder(r.Body).Decode(&listDomains)
	if err != nil || len(listDomains) < 0 {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message("Error while decoding request body"))
		u.ErrLog("Error while decoding request body", "CREATE BULK DOMAIN", "", r)
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
		domainObj := map[string]interface{}{}
		// Name is the only required attribute
		name, ok := domain["name"].(string)
		if !ok {
			return nil, errors.New("Invalid format: Name is required for all domains")
		}
		domainObj["name"] = name

		// Optional/default attributes
		if parent != "" {
			domainObj["parentId"] = parent
		}
		domainObj["category"] = "domain"
		if desc, ok := domain["description"].(string); ok {
			domainObj["description"] = []string{desc}
		} else {
			domainObj["description"] = []string{name}
		}
		domainObj["attributes"] = map[string]string{}
		if color, ok := domain["color"].(string); ok {
			domainObj["attributes"].(map[string]string)["color"] = color
		} else {
			domainObj["attributes"].(map[string]string)["color"] = "ffffff"
		}

		domainsToCreate = append(domainsToCreate, domainObj)

		// Add children domain, if any
		if children, ok := domain["domains"].([]interface{}); ok {
			if len(children) > 0 {
				// Convert from interface to map
				dChildren := []map[string]interface{}{}
				for _, d := range children {
					dChildren = append(dChildren, d.(map[string]interface{}))
				}
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

// swagger:operation GET /api/objects/{hierarchyName} Objects GetGenericObject
// Get an object from any entity.
// Gets an object from any of the physical entities with no need to specify it.
// The hierarchyName must be provided in the URL as a parameter.
// ---
// security:
// - bearer: []
// produces:
// - application/json
// parameters:
//   - name: hierarchyName
//     in: path
//     description: hierarchyName of the object
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
// responses:
//		'200':
//		    description: 'Found. A response body will be returned with
//	        a meaningful message.'
//		'404':
//		    description: Not Found. An error message will be returned.

func GetGenericObject(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 GetGenericObject ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)
	var data map[string]interface{}
	var err *u.Error

	// Get user roles for permissions
	user := getUserFromToken(w, r)
	if user == nil {
		return
	}

	name, e := mux.Vars(r)["name"]
	filters := getFiltersFromQueryParams(r)
	if e {
		data, err = models.GetObjectByName(name, filters, user.Roles)
	} else {
		u.Respond(w, u.Message("Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET ENTITY", "", r)
		return
	}

	if r.Method == "OPTIONS" && data != nil {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Allow", "GET, DELETE, OPTIONS, PATCH, PUT")
	} else {
		if err != nil {
			u.ErrLog("Error while getting "+name, "GET GENERIC", err.Message, r)
			u.RespondWithError(w, err)
		} else {
			u.Respond(w, u.RespDataWrapper("successfully got object", data))
		}
	}

}

// swagger:operation GET /api/{entity}/{IdOrHierarchyName} Objects GetEntity
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
//     cabinets, groups, corridors,
//     room-templates, obj-templates, bldg-templates, stray-devices.'
//     required: true
//     type: string
//     default: "sites"
//   - name: IdOrHierarchyName
//     in: path
//     description: 'ID or hierarchy name of desired object.
//     For templates the slug is the ID.'
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

	//Get entity type and strip trailing 'entityStr'
	entityStr := mux.Vars(r)["entity"]
	filters := getFiltersFromQueryParams(r)

	//If templates, format them
	entityStr = strings.Replace(entityStr, "-", "_", 1)

	//GET By ID
	if id, canParse = mux.Vars(r)["id"]; canParse {
		objId, e := getObjID(id)
		if e != nil {
			u.Respond(w, u.Message("Error while converting ID to ObjectID"))
			u.ErrLog("Error while converting ID to ObjectID", "GET ENTITY", "", r)
			return
		}

		//Prevents API from creating a new unidentified collection
		if i := u.EntityStrToInt(entityStr); i < 0 {
			w.WriteHeader(http.StatusNotFound)
			u.Respond(w, u.Message("Invalid object in URL: '"+mux.Vars(r)["entity"]+"' Please provide a valid object"))
			u.ErrLog("Cannot get invalid object", "GET "+mux.Vars(r)["entity"], "", r)
			return
		}

		req := bson.M{"_id": objId}
		data, modelErr = models.GetEntity(req, entityStr, filters, user.Roles)

	} else if id, canParse = mux.Vars(r)["name"]; canParse { //GET By String
		if strings.Contains(entityStr, "template") { //GET By Slug (template)
			req := bson.M{"slug": id}
			data, modelErr = models.GetEntity(req, entityStr, filters, user.Roles)
		} else {
			println(id)
			req := bson.M{"hierarchyName": id}
			data, modelErr = models.GetEntity(req, entityStr, filters, user.Roles) // GET By hierarchyName
		}
	}

	if !canParse {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message("Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET ENTITY", "", r)
		return
	}

	if r.Method == "OPTIONS" && data != nil {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Allow", "GET, DELETE, OPTIONS, PATCH, PUT")
	} else {
		if modelErr != nil {
			u.ErrLog("Error while getting "+entityStr, "GET "+strings.ToUpper(entityStr),
				modelErr.Message, r)
			u.RespondWithError(w, modelErr)
		} else {
			u.Respond(w, u.RespDataWrapper("successfully got "+entityStr, data))
		}
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
//     cabinets, groups, corridors,
//     room-templates, obj-templates, bldg-templates, stray-devices'
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

	entStr = mux.Vars(r)["entity"]
	println("ENTSTR: ", entStr)

	//If templates, format them
	entStr = strings.Replace(entStr, "-", "_", 1)

	//Prevents Mongo from creating a new unidentified collection
	if i := u.EntityStrToInt(entStr); i < 0 {
		w.WriteHeader(http.StatusNotFound)
		u.Respond(w, u.Message("Invalid object in URL: '"+mux.Vars(r)["entity"]+
			"' Please provide a valid object"))
		u.ErrLog("Cannot get invalid object", "GET "+mux.Vars(r)["entity"], "", r)
		return
	}

	req := bson.M{}
	data, e := models.GetManyEntities(entStr, req, u.RequestFilters{}, user.Roles)

	if e != nil {
		u.ErrLog("Error while getting "+entStr+"s", "GET ALL "+strings.ToUpper(entStr),
			e.Message, r)
		u.RespondWithError(w, e)
	} else {
		u.Respond(w, u.RespDataWrapper("successfully got "+entStr+"s",
			map[string]interface{}{"objects": data}))
	}

}

// swagger:operation DELETE /api/{entity}/{IdOrHierarchyName} Objects DeleteObject
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
//     cabinets, groups, corridors,
//     room-templates, obj-templates, bldg-templates, stray-devices.'
//     required: true
//     type: string
//     default: "sites"
//   - name: IdOrHierarchyName
//     in: path
//     description: 'ID or hierarchy name of desired object.
//     For templates the slug is the ID.'
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

	id, e := mux.Vars(r)["id"]
	name, e2 := mux.Vars(r)["name"]

	// Get user roles for permissions
	user := getUserFromToken(w, r)
	if user == nil {
		return
	}

	//Get entity from URL
	entity := mux.Vars(r)["entity"]

	//If templates, format them
	entity = strings.Replace(entity, "-", "_", 1)

	//Prevents Mongo from creating a new unidentified collection
	if u.EntityStrToInt(entity) < 0 {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message("Invalid object in URL: '"+mux.Vars(r)["entity"]+
			"' Please provide a valid object"))
		u.ErrLog("Cannot delete invalid object", "DELETE "+mux.Vars(r)["entity"], "", r)
		return
	}

	var modelErr *u.Error
	switch {
	case e2 && !e: // DELETE by name
		if strings.Contains(entity, "template") {
			modelErr = models.DeleteSingleEntity(entity, bson.M{"slug": name})
		} else {
			//use hierarchyName
			modelErr = models.DeleteEntityByName(entity, name, user.Roles)
		}

	case e && !e2: // DELETE by id
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			u.Respond(w, u.Message("Error while converting ID to ObjectID"))
			u.ErrLog("Error while converting ID to ObjectID", "DELETE ENTITY", "", r)
			return
		}

		req, ok := models.GetRequestFilterByDomain(user.Roles)
		if !ok {
			modelErr = &u.Error{Type: u.ErrUnauthorized,
				Message: "User does not have permission to delete"}
		} else {
			if entity == "device" {
				_, modelErr = models.DeleteDeviceF(objID, req)
			} else {
				modelErr = models.DeleteEntity(entity, objID, req)
			}
		}

	default:
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message("Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "DELETE ENTITY", "", r)
		return
	}

	if modelErr != nil {
		u.ErrLog("Error while deleting entity", "DELETE ENTITY", modelErr.Message, r)
		u.RespondWithError(w, modelErr)
	} else {
		w.WriteHeader(http.StatusNoContent)
		u.Respond(w, u.Message("successfully deleted"))
	}
}

// swagger:operation PATCH /api/{entity}/{IdOrHierarchyName} Objects PartialUpdateObject
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
//     cabinets, groups, corridors,
//     room-templates, obj-templates, bldg-templates, stray-devices.'
//     required: true
//     type: string
//     default: "sites"
//   - name: IdOrHierarchyName
//     in: path
//     description: 'ID or hierarchy name of desired object.
//     For templates the slug is the ID.'
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

// swagger:operation PUT /api/{objs}/{IdOrHierarchyName} Objects UpdateObject
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
//     cabinets, groups, corridors,
//     room-templates, obj-templates, bldg-templates, stray-devices.'
//     required: true
//     type: string
//     default: "sites"
//   - name: IdOrHierarchyName
//     in: path
//     description: 'ID or hierarchy name of desired object.
//     For templates the slug is the ID.'
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

	updateData := map[string]interface{}{}
	id, e := mux.Vars(r)["id"]
	name, e2 := mux.Vars(r)["name"]
	isPatch := false
	if r.Method == "PATCH" {
		isPatch = true
	}
	println(r.Method)

	err := json.NewDecoder(r.Body).Decode(&updateData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message("Error while decoding request body"))
		u.ErrLog("Error while decoding request body", "UPDATE ENTITY", "", r)
		return
	}

	//Get entity from URL
	entity = mux.Vars(r)["entity"]

	//If templates, format them
	entity = strings.Replace(entity, "-", "_", 1)

	//Prevents Mongo from creating a new unidentified collection
	if u.EntityStrToInt(entity) < 0 {
		w.WriteHeader(http.StatusNotFound)
		u.Respond(w, u.Message("Invalid object in URL: '"+mux.Vars(r)["entity"]+"' Please provide a valid object"))
		u.ErrLog("Cannot update invalid object", "UPDATE "+mux.Vars(r)["entity"], "", r)
		return
	}

	switch {
	case e2: // Update with slug/hierarchyName
		var req bson.M
		if strings.Contains(entity, "template") {
			req = bson.M{"slug": name}
		} else {
			req = bson.M{"hierarchyName": name}
		}

		data, modelErr = models.UpdateEntity(entity, req, updateData, isPatch, user.Roles)

	case e: // Update with id
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			u.Respond(w, u.Message("Error while converting ID to ObjectID"))
			u.ErrLog("Error while converting ID to ObjectID", "UPDATE ENTITY", "", r)
			return
		}

		req := bson.M{"_id": objID}
		data, modelErr = models.UpdateEntity(entity, req, updateData, isPatch, user.Roles)

	default:
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message("Error while extracting from path parameters"))
		u.ErrLog("Error while extracting from path parameters", "UPDATE ENTITY", "", r)
		return
	}

	if modelErr != nil {
		u.RespondWithError(w, modelErr)
	} else {
		u.Respond(w, u.RespDataWrapper("successfully updated "+entity, data))
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
//     cabinets, groups, corridors,
//     room-templates, obj-templates, bldg-templates, stray-devices.'
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
//		'204':
//			description: 'Found. A response body will be returned with
//			a meaningful message.'
//		'404':
//			description: Not found. An error message will be returned.

func GetEntityByQuery(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 GetEntityByQuery ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)
	var data []map[string]interface{}
	var bsonMap bson.M
	var entStr string
	var modelErr *u.Error

	// Get user roles for permissions
	user := getUserFromToken(w, r)
	if user == nil {
		return
	}

	entStr = r.URL.Path[5 : len(r.URL.Path)-1]
	filters := getFiltersFromQueryParams(r)

	//If templates, format them
	entStr = strings.Replace(entStr, "-", "_", 1)

	query := u.ParamsParse(r.URL, u.EntityStrToInt(entStr))
	js, _ := json.Marshal(query)
	json.Unmarshal(js, &bsonMap)

	//Prevents Mongo from creating a new unidentified collection
	if u.EntityStrToInt(entStr) < 0 {
		w.WriteHeader(http.StatusNotFound)
		u.Respond(w, u.Message("Invalid object in URL: '"+entStr+"' Please provide a valid object"))
		u.ErrLog("Cannot get invalid object", "GET ENTITYQUERY"+entStr, "", r)
		return
	}

	data, modelErr = models.GetManyEntities(entStr, bsonMap, filters, user.Roles)

	if modelErr != nil {
		u.ErrLog("Error while getting "+entStr, "GET ENTITYQUERY", modelErr.Message, r)
		u.RespondWithError(w, modelErr)
	} else {
		u.Respond(w, u.RespDataWrapper("successfully got query for "+entStr,
			map[string]interface{}{"objects": data}))
	}
}

// swagger:operation GET /api/tempunits/{IdOrHierarchyName} Objects GetTempUnit
// Gets the temperatureUnit attribute of the parent site of given object.
// ---
// security:
// - bearer: []
// produces:
// - application/json
// parameters:
//   - name: IdOrHierarchyName
//     in: path
//     description: 'ID or hierarchy name of desired object.
//     For templates the slug is the ID.'
//     required: true
//     type: string
//     default: "siteA"
// responses:
//  '200':
//     description: 'Found. A response body will be returned with
//     a meaningful message.'
//  '404':
//     description: 'Nothing Found. An error message will be returned.'

func GetTempUnit(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 GetTempUnit ")
	fmt.Println("******************************************************")

	data, err := models.GetSiteParentTempUnit(mux.Vars(r)["id"])
	if err != nil {
		u.RespondWithError(w, err)
	} else {
		if r.Method == "OPTIONS" {
			w.Header().Add("Content-Type", "application/json")
			w.Header().Add("Allow", "GET, OPTIONS, HEAD")
		} else {
			resp := u.RespDataWrapper(
				"successfully got temperatureUnit from object's parent site",
				map[string]interface{}{"temperatureUnit": data})
			u.Respond(w, resp)
		}
	}
}

// swagger:operation GET /api/{entity}/{id}/{subent} Objects GetEntitiesOfAncestor
// Obtain all objects 2 levels lower in the system.
// For Example: /api/sites/{id}/buildings
// Will return all buildings of a site
// Returns JSON body with all subobjects under the Object
// ---
// security:
// - bearer: []
// produces:
// - application/json
// parameters:
// - name: entity
//   in: query
//   description: 'Indicates the entity. Only values of "sites",
//   "buildings", "rooms" are acceptable'
//   required: true
//   type: string
//   default: "sites"
// - name: ID
//   in: query
//   description: ID of object
//   required: true
//   type: int
//   default: 999
// - name: subent
//   in: query
//   description: Objects which 2 are levels lower in the hierarchy.
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
	//Extract string between /api and /{id}
	entStr := mux.Vars(r)["ancestor"]
	entStr = entStr[:len(entStr)-1] // remove s
	enum := u.EntityStrToInt(entStr)

	// Get user roles for permissions
	user := getUserFromToken(w, r)
	if user == nil {
		return
	}

	//Prevents Mongo from creating a new unidentified collection
	if enum < 0 {
		w.WriteHeader(http.StatusNotFound)
		u.Respond(w, u.Message("Invalid object in URL: '"+entStr+"' Please provide a valid object"))
		u.ErrLog("Cannot get invalid object", "GET CHILDRENOFPARENT"+entStr, "", r)
		return
	}

	if enum == u.SITE {
		id, e = mux.Vars(r)["site_name"]
	} else {
		id, e = mux.Vars(r)["id"]
	}

	if !e {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message("Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET CHILDRENOFPARENT", "", r)
		return
	}

	//Could be: "ac", "panel", "corridor", "cabinet", "sensor"
	indicator := mux.Vars(r)["sub"]

	req := bson.M{}
	data, modelErr := models.GetEntitiesOfAncestor(id, req, enum, entStr, indicator)
	if modelErr != nil {
		u.ErrLog("Error while getting children of "+entStr,
			"GET CHILDRENOFPARENT", modelErr.Message, r)
		u.RespondWithError(w, modelErr)
	} else if r.Method == "OPTIONS" {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Allow", "GET, OPTIONS")
	} else {
		u.Respond(w, u.RespDataWrapper("successfully got object", map[string]interface{}{"objects": data}))
	}
}

// swagger:operation GET /api/{entity}/{IdOrHierarchyName}/all Objects GetEntityHierarchy
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
//   cabinets, groups, corridors,
//   room-templates, obj-templates, bldg-templates, stray-devices.'
//   required: true
//   type: string
//   default: "sites"
// - name: IdOrHierarchyName
//   in: path
//   description: 'ID or hierarchy name of desired object.
//   For templates the slug is the ID.'
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

func GetEntityHierarchy(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 GetEntityHierarchy ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)
	entity := mux.Vars(r)["entity"]
	var limit int
	var end int
	var data map[string]interface{}
	var modelErr *u.Error

	// Get user roles for permissions
	user := getUserFromToken(w, r)
	if user == nil {
		return
	}

	req := bson.M{}

	//If template or stray convert '-' -> '_'
	entity = strings.Replace(entity, "-", "_", 1)

	id, e := mux.Vars(r)["id"]
	if !e {
		u.Respond(w, u.Message("Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET ENTITYHIERARCHY", "", r)
		return
	}

	//Check if the request is a ranged hierarchy
	filters := getFiltersFromQueryParams(r)
	if len(filters.Limit) > 0 { //limit={number} was provided
		end, _ = strconv.Atoi(filters.Limit)
		limit = u.EntityStrToInt(entity) + end
		if end == 0 {
			// It's a GetEntity, treat it here
			objID, _ := primitive.ObjectIDFromHex(id)
			newReq := req
			newReq["_id"] = objID
			data, modelErr := models.GetEntity(newReq, entity, filters, user.Roles)
			if modelErr != nil {
				u.ErrLog("Error while getting "+entity, "GET "+entity, modelErr.Message, r)
				u.RespondWithError(w, modelErr)
			} else {
				u.Respond(w, u.RespDataWrapper("successfully got object",
					map[string]interface{}{"data": data}))
			}
			return
		}
	} else {
		//arbitrarily set value to 999
		limit = 999
	}

	println("The limit is: ", limit)
	oID, _ := getObjID(id)
	entNum := u.EntityStrToInt(entity)
	println("EntNum:", entNum)

	// Prevents Mongo from creating a new unidentified collection
	if entNum < 0 {
		w.WriteHeader(http.StatusNotFound)
		u.Respond(w, u.Message("Invalid object in URL:"+entity+" Please provide a valid object"))
		u.ErrLog("Cannot get invalid object", "GET ENTITYHIERARCHY "+entity, "", r)
		return
	}

	// Get hierarchy
	println("Entity: ", entity, " & OID: ", oID.Hex())
	data, modelErr = models.GetEntityHierarchy(oID, req, entity, entNum, limit, filters, user.Roles)

	if modelErr != nil {
		u.ErrLog("Error while getting "+entity, "GET "+entity, modelErr.Message, r)
		u.RespondWithError(w, modelErr)
	} else if r.Method == "OPTIONS" {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Allow", "GET, OPTIONS")
	} else {
		u.Respond(w, u.RespDataWrapper("successfully got object", data))
	}
}

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

	name, e := mux.Vars(r)["name"]
	if !e {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message("Error while parsing name"))
		u.ErrLog("Error while parsing path parameters", "GetHierarchyByName", "", r)
		return
	}

	entity, e2 := mux.Vars(r)["entity"]
	if !e2 {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message("Error while parsing entity"))
		u.ErrLog("Error while parsing path parameters", "GetHierarchyByName", "", r)
		return
	}

	// If template or stray convert '-' -> '_'
	entity = strings.Replace(entity, "-", "_", 1)

	// Check if the request is a ranged hierarchy
	filters := getFiltersFromQueryParams(r)
	if len(filters.Limit) > 0 {
		//limit={number} was provided
		limit, _ = strconv.Atoi(filters.Limit)
	} else {
		limit = 999
	}

	println("The limit is: ", limit)

	data, e1 := models.GetEntity(bson.M{"hierarchyName": name}, entity, filters, user.Roles)
	if limit >= 1 && e1 == nil {
		data["children"], e1 = models.GetHierarchyByName(entity, name, limit, filters)
	}

	if data == nil {
		u.ErrLog("Error while getting "+entity, "GET "+entity, e1.Message, r)
		u.RespondWithError(w, e1)
	} else if r.Method == "OPTIONS" {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Allow", "GET, OPTIONS")
	} else {
		u.Respond(w, u.RespDataWrapper("successfully got object's hierarchy", data))
	}
}

// swagger:operation GET /api/hierarchy Objects GetCompleteHierarchy
// Returns system complete hierarchy.
// Return is arranged by relationship (father:[children])
// and category (category:[objects]), starting with "Root":[sites].
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

	data, err := models.GetCompleteHierarchy(user.Roles)
	if err != nil {
		u.RespondWithError(w, err)
	} else {
		if r.Method == "OPTIONS" {
			w.Header().Add("Content-Type", "application/json")
			w.Header().Add("Allow", "GET, OPTIONS, HEAD")
		} else {
			u.Respond(w, u.RespDataWrapper("successfully got hierarchy", data))
		}
	}
}

// swagger:operation GET /api/hierarchy/domains Organization GetCompleteDomainHierarchy
// Returns domain complete hierarchy.
// Return is arranged by relationship (father:[children]),
// starting with "Root":[root domains].
// ---
// security:
// - bearer: []
// produces:
// - application/json
// responses:
//     '200':
//          description: 'Request is valid.'
//     '500':
//          description: Server error.

func GetCompleteDomainHierarchy(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 GetCompleteHierarchy ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)

	// Get user roles for permissions
	user := getUserFromToken(w, r)
	if user == nil {
		return
	}

	data, err := models.GetCompleteDomainHierarchy(user.Roles)
	if err != nil {
		u.RespondWithError(w, err)
	} else {
		if r.Method == "OPTIONS" {
			w.Header().Add("Content-Type", "application/json")
			w.Header().Add("Allow", "GET, OPTIONS, HEAD")
		} else {
			u.Respond(w, u.RespDataWrapper("successfully got domain hierarchy", data))
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
			w.Header().Add("Content-Type", "application/json")
			w.Header().Add("Allow", "GET, OPTIONS, HEAD")
		} else {
			u.Respond(w, u.RespDataWrapper("successfully got attrs hierarchy", data))
		}
	}
}

// swagger:operation GET /api/{firstEntity}/{id}/{HierarchalPath} Objects GetEntitiesUsingNamesOfParents
// Get an object with its full hierarchal path.
// The path should begin with an entity name (firstEntity) and the ID of an object of this entity
// followed by a hierarchal path until the desired objet, that is,
// a sequence of entity names (category) and object names.
// ---
// security:
// - bearer: []
// produces:
// - application/json
// parameters:
// - name: firstEntity
//   in: query
//   description: 'Root entity followed by an id. Can be: sites, buildings, rooms, racks or devices'
//   required: true
//   type: string
//   default: "sites"
// - name: id
//   in: path
//   description: 'id of object of firstEntity'
//   required: true
//   type: string
//   default: "123"
// - name: HierarchalPath
//   in: path
//   description: 'Hierarchal path to desired object.'
//   required: true
//   type: string
//   example: '/api/sites/{id}/buildings/{building_name} ;
//   /api/sites/{id}/buildings/{building_name}/rooms/{room_name} ;
//   /api/sites/{id}/buildings/{building_name}/rooms/{room_name}/acs/{ac_name} ;
//   /api/sites/{id}/buildings/{building_name}/rooms/{room_name}/corridors/{corridor_name} ;
//   /api/sites/{id}/buildings/{building_name}/rooms/{room_name}/panels/{panel_name} ;
//   /api/sites/{id}/buildings/{building_name}/rooms/{room_name}/groups/{group_name} ;
//   /api/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices/{device_name} ;
//   /api/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name} ;
//   /api/buildings/{id}/rooms/{room_name} ;
//   /api/buildings/{id}/rooms/{room_name}/acs/{ac_name} ;
//   /api/buildings/{id}/rooms/{room_name}/corridors/{corridor_name} ;
//   /api/buildings/{id}/rooms/{room_name}/panels/{panel_name} ;
//   /api/buildings/{id}/rooms/{room_name}/groups/{group_name} ;
//   /api/buildings/{id}/rooms/{room_name}/rack/{rack_name} ;
//   /api/buildings/{id}/rooms/{room_name}/racks/{rack_name}/devices/{device_name} ;
//   /api/rooms/{id}/acs/{ac_name} ;
//   /api/rooms/{id}/corridors/{corridor_name} ;
//   /api/rooms/{id}/panels/{panel_name} ;
//   /api/rooms/{id}/groups/{group_name} ;
//   /api/rooms/{id}/racks/{rack_name}/devices/{device_name} ;
//   /api/racks/{id}/devices/{device_name} ;
//   /api/devices/{id}/devices/{device_name} ;'
//   default: "/buildings/BuildingB/rooms/RoomA"
// responses:
//     '200':
//         description: 'Found. A response body will be returned with
//         a meaningful message.'
//     '404':
//         description: Not Found. An error message will be returned.

func GetEntitiesUsingNamesOfParents(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 GetEntitiesUsingNamesOfParents ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)
	entity := mux.Vars(r)["entity"]

	//If template or stray convert '-' -> '_'
	entity = strings.Replace(entity, "-", "_", 1)

	// Get user roles for permissions
	user := getUserFromToken(w, r)
	if user == nil {
		return
	}

	id, e := mux.Vars(r)["id"]
	tname, e1 := mux.Vars(r)["site_name"]
	if !e && !e1 {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message("Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET ENTITIESUSINGANCESTORNAMES", "", r)
		return
	}

	//Prevents Mongo from creating a new unidentified collection
	if u.EntityStrToInt(entity) < 0 {
		w.WriteHeader(http.StatusNotFound)
		u.Respond(w, u.Message("Invalid object in URL:"+entity+" Please provide a valid object"))
		u.ErrLog("Cannot get invalid object", "GET ENTITIESUSINGANCESTORNAMES "+entity, "", r)
		return
	}

	arr := (strings.Split(r.URL.Path, "/")[4:])
	ancestry := make([]map[string]string, 0)

	for i, k := range arr {

		key := k[:len(k)-1]

		//If templates, format them
		key = strings.Replace(key, "-", "_", 1)

		if i%2 == 0 { //The keys (entities) are at the even indexes
			if i+1 >= len(arr) {
				//Small front end hack since client wants stray-device URLs
				//to be like: URL/stray-devices/ID/devices
				if key == "device" && entity == "stray_device" {
					key = "stray_device"
				}
				ancestry = append(ancestry,
					map[string]string{key: "all"})
			} else {
				//Prevents Mongo from creating a new unidentified collection
				if u.EntityStrToInt(key) < 0 {
					w.WriteHeader(http.StatusNotFound)
					u.Respond(w, u.Message("Invalid object in URL:"+key+" Please provide a valid object"))
					u.ErrLog("Cannot get invalid object", "GET "+key, "", r)
					return
				}
				//Small front end hack since client wants stray-device URLs
				//to be like: URL/stray-devices/ID/devices/ID/devices
				if key == "device" && entity == "stray_device" {
					key = "stray_device"
				}
				ancestry = append(ancestry,
					map[string]string{key: arr[i+1]})
			}
		}
	}

	oID, _ := getObjID(id)

	if len(arr)%2 != 0 { //This means we are getting entities
		var data []map[string]interface{}
		var e3 *u.Error
		req := bson.M{}
		if e1 {
			data, e3 = models.GetEntitiesUsingSiteAsAncestor(entity, tname, req, ancestry, user.Roles)

		} else {
			data, e3 = models.GetEntitiesUsingAncestorNames(entity, oID, req, ancestry, user.Roles)
		}

		if len(data) == 0 {
			u.ErrLog("Error while getting "+entity, "GET "+entity, e3.Message, r)
			u.RespondWithError(w, e3)
		} else {
			if r.Method == "OPTIONS" {
				w.Header().Add("Content-Type", "application/json")
				w.Header().Add("Allow", "GET, OPTIONS")
				return
			}
			u.Respond(w, u.RespDataWrapper("successfully got object",
				map[string]interface{}{"objects": data}))
		}
	} else { //We are only retrieving an entity
		var data map[string]interface{}
		var e3 *u.Error
		if e1 {
			req := bson.M{"name": tname}
			data, e3 = models.GetEntityUsingSiteAsAncestor(req, entity, ancestry)
		} else {
			req := bson.M{"_id": oID}
			data, e3 = models.GetEntityUsingAncestorNames(req, entity, ancestry)
		}

		if len(data) == 0 {
			u.ErrLog("Error while getting "+entity, "GET "+entity, e3.Message, r)
			u.RespondWithError(w, e3)
		} else if r.Method == "OPTIONS" && data != nil {
			w.Header().Add("Content-Type", "application/json")
			w.Header().Add("Allow", "GET, OPTIONS")
		} else {
			u.Respond(w, u.RespDataWrapper("successfully got object", data))
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
	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Allow", "GET, DELETE, OPTIONS, PATCH, POST, PUT")

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
		//w.WriteHeader(http.StatusOK)
	} else {
		r := models.GetStats()
		u.Respond(w, r)
	}
	//w.Header().Add("Content-Type", "application/json")

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
//     cabinets, groups, corridors,
//     room-templates, obj-templates, bldg-templates, stray-devices.'
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

	//If templates or stray-devices, format them
	if idx := strings.Index(entity, "-"); idx != -1 {
		//entStr[idx] = '_'
		entity = entity[:idx] + "_" + entity[idx+1:]
	}
	entInt := u.EntityStrToInt(entity)

	if !e1 || entInt == -1 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if r.Method == "OPTIONS" {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Allow", "POST, OPTIONS")
		return
	}

	err := json.NewDecoder(r.Body).Decode(&obj)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message("Error while decoding request body"))
		u.ErrLog("Error while decoding request body", "VALIDATE "+entity, "", r)
		return
	}

	if entInt != u.BLDGTMPL && entInt != u.ROOMTMPL && entInt != u.OBJTMPL {
		if permission := models.CheckUserPermissions(user.Roles, entInt, obj["domain"].(string)); permission < models.WRITE {
			w.WriteHeader(http.StatusUnauthorized)
			u.Respond(w, u.Message("This user"+
				" does not have sufficient permissions to create"+
				" this object under this domain "))
			u.ErrLog("Cannot validate object creation due to limited user privilege",
				"Validate CREATE "+entity, "", r)
			return
		}
	}

	ok, e := models.ValidateEntity(entInt, obj)
	if ok {
		u.Respond(w, u.Message("This object can be created"))
		return
	} else {
		u.RespondWithError(w, e)
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
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Allow", "GET, OPTIONS, HEAD")
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
