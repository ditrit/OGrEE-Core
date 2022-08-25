package controllers

import (
	"encoding/json"
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

func parseDataForNonStdResult(ent string, eNum int, data map[string]interface{}) map[string][]map[string]interface{} {

	ans := map[string][]map[string]interface{}{}
	add := []map[string]interface{}{}

	firstIndex := u.EntityToString(eNum + 1)
	firstArr := data[firstIndex+"s"].([]map[string]interface{})

	ans[firstIndex+"s"] = firstArr

	for i := range firstArr {
		nxt := u.EntityToString(eNum + 2)
		add = append(add, firstArr[i][nxt+"s"].([]map[string]interface{})...)
	}

	ans[u.EntityToString(eNum+2)+"s"] = add
	newAdd := []map[string]interface{}{}
	for i := range add {
		nxt := u.EntityToString(eNum + 3)
		newAdd = append(newAdd, add[i][nxt+"s"].([]map[string]interface{})...)
	}

	ans[u.EntityToString(eNum+3)+"s"] = newAdd

	newAdd2 := []map[string]interface{}{}
	for i := range newAdd {
		nxt := u.EntityToString(eNum + 4)
		newAdd2 = append(newAdd2, newAdd[i][nxt+"s"].([]map[string]interface{})...)
	}

	ans[u.EntityToString(eNum+4)+"s"] = newAdd2
	newAdd3 := []map[string]interface{}{}

	for i := range newAdd2 {
		nxt := u.EntityToString(eNum + 5)
		newAdd3 = append(newAdd3, newAdd2[i][nxt+"s"].([]map[string]interface{})...)
	}
	ans[u.EntityToString(eNum+5)+"s"] = newAdd3

	newAdd4 := []map[string]interface{}{}

	for i := range newAdd3 {
		nxt := u.EntityToString(eNum + 6)
		newAdd4 = append(newAdd4, newAdd3[i][nxt+"s"].([]map[string]interface{})...)
	}

	ans[u.EntityToString(eNum+6)+"s"] = newAdd4

	newAdd5 := []map[string]interface{}{}

	for i := range newAdd4 {
		nxt := u.EntityToString(eNum + 7)
		newAdd5 = append(newAdd5, newAdd4[i][nxt+"s"].([]map[string]interface{})...)
	}

	ans[u.EntityToString(eNum+7)+"s"] = newAdd5

	//add := []map[string]interface{}{}

	//Get All first entities
	/*for i := eNum + 1; i < SUBDEV1; i++ {
		add = append(add, firstArr[i])
	}*/
	return ans
}

//This function is useful for debugging
//purposes. It displays any JSON
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

//'Flattens' the map[string]interface{}
//for PATCH requests
func Flatten(prefix string, src map[string]interface{}, dest map[string]interface{}) {
	if len(prefix) > 0 {
		prefix += "."
	}
	for k, v := range src {
		switch child := v.(type) {
		case map[string]interface{}:
			Flatten(prefix+k, child, dest)
		case []interface{}:
			for i := 0; i < len(child); i++ {
				dest[prefix+k+"."+strconv.Itoa(i)] = child[i]
			}
		default:
			dest[prefix+k] = v
		}
	}
}

func DispRequestMetaData(r *http.Request) {
	fmt.Println("URL:", r.URL.String())
	fmt.Println("IP-ADDR: ", r.RemoteAddr)
	fmt.Println(time.Now().Format("2006-Jan-02 Monday 03:04:05 PM MST -07:00"))
}

// swagger:operation POST /api/{obj} objects CreateObject
// Creates an object in the system.
// ---
// produces:
// - application/json
// parameters:
// - name: objs
//   in: query
//   description: 'Indicates the Object. Only values of "sites","domains",
//   "buildings", "rooms", "racks", "devices", "acs", "panels",
//   "cabinets", "groups", "corridors",
//   "room-templates", "obj-templates", "sensors", "stray-devices",
//   "stray-sensors" are acceptable'
//   required: true
//   type: string
//   default: "sites"
// - name: Name
//   in: query
//   description: Name of object
//   required: true
//   type: string
//   default: "Object A"
// - name: Category
//   in: query
//   description: Category of Object (ex. Consumer Electronics, Medical)
//   required: true
//   type: string
//   default: "Research"
// - name: Domain
//   description: 'Domain of Object'
//   required: true
//   type: string
//   default: 999
// - name: ParentID
//   description: 'All objects are linked to a
//   parent with the exception of Site since it has no parent'
//   required: true
//   type: int
//   default: 999
// - name: Description
//   in: query
//   description: Description of Object
//   required: false
//   type: string[]
//   default: ["Some abandoned object in Grenoble"]
// - name: Attributes
//   in: query
//   description: 'Any other object attributes can be added.
//   They are required depending on the obj type.'
//   required: true
//   type: json
// responses:
//     '201':
//         description: 'Created. A response body will be returned with
//         a meaningful message.'
//     '400':
//         description: 'Bad request. A response body with an error
//         message will be returned.'

var CreateEntity = func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 CreateEntity ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)

	var e string
	var resp map[string]interface{}
	entity := map[string]interface{}{}
	err := json.NewDecoder(r.Body).Decode(&entity)

	//strip the '/api' in URL
	entStr, e1 := mux.Vars(r)["entity"]
	userData := r.Context().Value("user")

	if !e1 || userData == nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message(false, "Error while parsing path params"))
		u.ErrLog("Error while parsing path params", "CREATE "+entStr, "", r)
		return
	}

	domain := userData.(map[string]interface{})["domain"].(string)
	role := userData.(map[string]interface{})["role"].(string)
	uid := userData.(map[string]interface{})["userID"].(uint)

	/*if domainInf == nil || roleInf == nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message(false, "User's Key is not valid please"+
			" check with your administrator"))
		u.ErrLog("User's key does not have domain/role", "CREATE "+entStr, "", r)
		return
	}*/

	println("UserID:", uid)
	println("Role:", role)
	println("Domain:", domain)

	if role == "" {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message(false, "User's Key is not valid please"+
			" check with your administrator"))
		u.ErrLog("User's key does not have assigned role", "CREATE "+entStr, "", r)
		return
	}

	entStr = entStr[:len(entStr)-1] // and the trailing 's'
	entUpper := strings.ToUpper(entStr)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		u.ErrLog("Error while decoding request body", "CREATE "+entStr, "", r)
		return
	}

	//If creating templates, format them
	if idx := strings.Index(entStr, "-"); idx != -1 {
		//entStr[idx] = '_'
		entStr = entStr[:idx] + "_" + entStr[idx+1:]
	}
	i := u.EntityStrToInt(entStr)

	println("ENT: ", entStr)
	println("ENUM VAL: ", i)

	//Prevents Mongo from creating a new unidentified collection
	if i < 0 {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message(false, "Invalid object in URL:"+entStr+" Please provide a valid object"))
		u.ErrLog("Cannot create invalid object", "CREATE "+entStr, "", r)
		return
	}

	if i == u.DOMAIN {
		if !models.EnsureUserIsSuper(role) {
			w.WriteHeader(http.StatusMethodNotAllowed)
			u.Respond(w, u.Message(false, "This "+role+
				" does not have sufficient permissions to create"+
				" this object under this domain. Please refer "+
				" to an administrator or manager for more assistance"))
			u.ErrLog("Unauthorised Domain create made",
				"CREATE "+entStr, "", r)
			return
		}
	} else {
		//Ensure user has permission for this object
		ok, reason := models.EnsureObjectPermission(entity, domain, role)
		if !ok {
			w.WriteHeader(http.StatusMethodNotAllowed)
			if reason == "" {

				u.Respond(w, u.Message(false, "This "+role+
					" does not have sufficient permissions to create"+
					" this object under this domain "))

			} else {
				u.Respond(w, u.Message(false, reason))
			}
			u.ErrLog("Unauthorised Access made", "CREATE "+entStr, "", r)
			return
		}
	}

	//Hard Code the 'category'
	//for all objects that are
	//not obj_template
	if i != u.OBJTMPL {
		entity["category"] = entStr
	}

	//Clean the data of 'id' attribute if present
	if _, ok := entity["id"]; ok {
		delete(entity, "id")
	}

	resp, e = models.CreateEntity(i, entity)

	switch e {
	case "validate", "duplicate":
		w.WriteHeader(http.StatusBadRequest)
		u.ErrLog("Error while creating "+entStr, "CREATE "+entUpper, e, r)
	case "":
		w.WriteHeader(http.StatusCreated)
	default:
		if strings.Split(e, " ")[1] == "duplicate" {
			w.WriteHeader(http.StatusBadRequest)
			u.ErrLog("Error: Duplicate "+entStr+" is forbidden",
				"CREATE "+entUpper, e, r)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			u.ErrLog("Error while creating "+entStr, "CREATE "+entUpper, e, r)
		}
	}

	u.Respond(w, resp)
}

// swagger:operation GET /api/{objs}/{id} objects GetObject
// Gets an Object from the system.
// The ID must be provided in the URL parameter
// The name can be used instead of ID if the obj is site
// ---
// produces:
// - application/json
// parameters:
// - name: objs
//   in: query
//   description: 'Indicates the location. Only values of "sites","domains",
//   "buildings", "rooms", "racks", "devices", "room-templates",
//   "obj-templates", "acs", "panels","cabinets", "groups",
//   "corridors","sensors","stray-devices", "stray-sensors" are acceptable'
//   required: true
//   type: string
//   default: "sites"
// - name: ID
//   in: path
//   description: 'ID of desired object or Name of Site.
//   For templates the slug is the ID. For stray-devices the name is the ID'
//   required: true
//   type: int
//   default: 999
// responses:
//     '200':
//         description: 'Found. A response body will be returned with
//         a meaningful message.'
//     '400':
//         description: Bad request. An error message will be returned.
//     '404':
//         description: Not Found. An error message will be returned.

// swagger:operation OPTIONS /api/{objs}/{id} objects ObjectOptions
// Displays possible operations for the resource in response header.
// ---
// produces:
// - application/json
// parameters:
// - name: objs
//   in: query
//   description: 'Only values of "sites","domains",
//   "buildings", "rooms", "racks", "devices", "room-templates",
//   "obj-templates", "acs", "panels","cabinets", "groups",
//   "corridors","sensors","stray-devices","stray-sensors", are acceptable'
// - name: id
//   in: query
//   description: 'ID of the object or name of Site.
//   For templates the slug is the ID. For stray-devices the name is the ID'
// responses:
//     '200':
//         description: 'Found. A response header will be returned with
//         possible operations.'
//     '400':
//         description: Bad request. An error message will be returned.
//     '404':
//         description: Not Found. An error message will be returned.
var GetEntity = func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 GetEntity ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)

	var data map[string]interface{}
	var id, e1 string
	var x primitive.ObjectID
	var e bool
	var e2 error

	resp := u.Message(true, "success")

	userData := r.Context().Value("user")
	domain := userData.(map[string]interface{})["domain"].(string)
	role := userData.(map[string]interface{})["role"].(string)
	uid := userData.(map[string]interface{})["userID"].(uint)

	println("UserID:", uid)
	println("Domain:", domain)
	println("Role:", role)

	//Get entity type and strip trailing 's'
	s, _ := mux.Vars(r)["entity"]
	s = s[:len(s)-1]

	//If templates, format them
	if idx := strings.Index(s, "-"); idx != -1 {
		s = s[:idx] + "_" + s[idx+1:]
	}

	//GET By ID
	if id, e = mux.Vars(r)["id"]; e == true {

		x, e2 = getObjID(id)
		if e2 != nil {
			u.Respond(w, u.Message(false, "Error while converting ID to ObjectID"))
			u.ErrLog("Error while converting ID to ObjectID", "GET ENTITY", "", r)
			return
		}

		//Prevents API from creating a new unidentified collection
		if i := u.EntityStrToInt(s); i < 0 {
			w.WriteHeader(http.StatusNotFound)
			u.Respond(w, u.Message(false, "Invalid object in URL:"+s+" Please provide a valid object"))
			u.ErrLog("Cannot get invalid object", "GET "+s, "", r)
			return
		}

		//data, e1 = models.GetEntity(bson.M{"_id": x}, s)
		req := bson.M{"_id": x}
		models.RequestGen(req, role, domain)
		data, e1 = models.GetEntity(req, s)

	} else if id, e = mux.Vars(r)["name"]; e == true { //GET By String

		if idx := strings.Contains(s, "_"); idx == true &&
			s != "stray_device" && s != "stray_sensor" { //GET By Slug
			req := bson.M{"slug": id}
			models.RequestGen(req, role, domain)
			data, e1 = models.GetEntity(req, s)
		} else {
			req := bson.M{"name": id}
			models.RequestGen(req, role, domain)
			data, e1 = models.GetEntity(req, s) //GET By Name
		}
	}

	if e == false {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET ENTITY", "", r)
		return
	}

	if data == nil {
		resp = u.Message(false, "Error while getting "+s+": "+e1)
		u.ErrLog("Error while getting "+s, "GET "+strings.ToUpper(s), "", r)

		switch e1 {
		case "record not found":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound) //For now
		}

	} else {

		message := ""
		switch u.EntityStrToInt(s) {
		case u.ROOMTMPL:
			message = "successfully got room_template"
		case u.OBJTMPL:
			message = "successfully got obj_template"
		default:
			message = "successfully got object"
		}
		resp = u.Message(true, message)
	}

	if r.Method == "OPTIONS" && data != nil {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Allow", "GET, DELETE, OPTIONS, PATCH, PUT")
	} else {
		resp["data"] = data
		u.Respond(w, resp)
	}

}

// swagger:operation GET /api/{objs} objects GetAllObjects
// Gets all present objects for specified category in the system.
// Returns JSON body with all specified objects of type and their IDs
// ---
// produces:
// - application/json
// parameters:
// - name: objs
//   in: query
//   description: 'Indicates the location. Only values of "sites","domains",
//   "buildings", "rooms", "racks", "devices", "room-templates",
//   "obj-templates","acs", "panels", "cabinets", "groups",
//   "corridors", "sensors", "stray-devices", "stray-sensors" are acceptable'
//   required: true
//   type: string
//   default: "sites"
// responses:
//     '200':
//         description: 'Found. A response body will be returned with
//         a meaningful message.'
//     '404':
//         description: Nothing Found. An error message will be returned.
var GetAllEntities = func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 GetAllEntities ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)
	var data []map[string]interface{}
	var e, entStr string

	userData := r.Context().Value("user")
	domain := userData.(map[string]interface{})["domain"].(string)
	role := userData.(map[string]interface{})["role"].(string)
	uid := userData.(map[string]interface{})["userID"].(uint)

	println("UserID:", uid)
	println("Domain:", domain)
	println("Role:", role)

	//entStr = arr[2][:len(arr[2])-1]
	entStr, _ = mux.Vars(r)["entity"]
	entStr = entStr[:len(entStr)-1]
	println("ENTSTR: ", entStr)

	//If templates, format them
	if idx := strings.Index(entStr, "-"); idx != -1 {
		entStr = entStr[:idx] + "_" + entStr[idx+1:]
	}

	//Prevents Mongo from creating a new unidentified collection
	if i := u.EntityStrToInt(entStr); i < 0 {
		w.WriteHeader(http.StatusNotFound)
		u.Respond(w, u.Message(false, "Invalid object in URL:"+entStr+" Please provide a valid object"))
		u.ErrLog("Cannot get invalid object", "GET "+entStr, "", r)
		return
	}

	req := bson.M{}
	models.RequestGen(req, role, domain)
	data, e = models.GetManyEntities(entStr, req, nil)

	entUpper := strings.ToUpper(entStr) // and the trailing 's'
	resp := u.Message(true, "success")

	if len(data) == 0 {
		resp = u.Message(false, "Error while getting "+entStr+": "+e)
		u.ErrLog("Error while getting "+entStr+"s", "GET ALL "+entUpper, e, r)

		switch e {
		case "":
			resp = u.Message(false,
				"Error while getting "+entStr+"s: No Records Found")
			w.WriteHeader(http.StatusNotFound)
		default:
		}

	} else {
		message := ""
		switch u.EntityStrToInt(entStr) {
		case u.ROOMTMPL:
			message = "successfully got all room_templates"
		case u.OBJTMPL:
			message = "successfully got all obj_templates"
		default:
			message = "successfully got all objects "
		}
		resp = u.Message(true, message)
	}

	resp["data"] = map[string]interface{}{"objects": data}

	u.Respond(w, resp)
}

// swagger:operation DELETE /api/{objs}/{id} objects DeleteObject
// Deletes an Object in the system.
// ---
// produces:
// - application/json
// parameters:
// - name: objs
//   in: query
//   description: 'Indicates the location. Only values of "sites","domains",
//   "buildings", "rooms", "racks", "devices", "room-templates",
//   "obj-templates","acs", "panels",
//   "cabinets", "groups", "corridors","sensors", "stray-devices"
//   "stray-sensors" are acceptable'
//   required: true
//   type: string
//   default: "sites"
// - name: ID
//   in: path
//   description: 'ID of the object or name of Site.
//   For templates the slug is the ID. For stray-devices the name is the ID'
//   required: true
//   type: int
//   default: 999
// responses:
//     '204':
//        description: 'Successfully deleted object.
//        No response body will be returned'
//     '404':
//        description: Not found. An error message will be returned
var DeleteEntity = func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 DeleteEntity ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)

	var v map[string]interface{}
	id, e := mux.Vars(r)["id"]
	name, e2 := mux.Vars(r)["name"]

	userData := r.Context().Value("user")
	domain := userData.(map[string]interface{})["domain"].(string)
	role := userData.(map[string]interface{})["role"].(string)
	uid := userData.(map[string]interface{})["userID"].(uint)

	println("UserID:", uid)
	println("Domain:", domain)
	println("Role:", role)

	//Get entity from URL and strip trailing 's'
	entity, _ := mux.Vars(r)["entity"]
	entity = entity[:len(entity)-1]

	//If templates, format them
	if idx := strings.Index(entity, "-"); idx != -1 {
		entity = entity[:idx] + "_" + entity[idx+1:]
	}

	//Prevents Mongo from creating a new unidentified collection
	if u.EntityStrToInt(entity) < 0 {
		w.WriteHeader(http.StatusNotFound)
		u.Respond(w, u.Message(false, "Invalid object in URL:"+entity+" Please provide a valid object"))
		u.ErrLog("Cannot delete invalid object", "DELETE "+entity, "", r)
		return
	}

	switch {
	case e2 == true && e == false: // DELETE SLUG or stray-device name
		if entity == "stray_device" || entity == "stray_sensor" {
			req := bson.M{"name": name}
			models.RequestGen(req, role, domain)
			sd, _ := models.GetEntity(req, entity)
			if sd == nil {
				w.WriteHeader(http.StatusNotFound)
				u.Respond(w, u.Message(false, "Error object not found"))
				return
			}
			if _, ok := sd["id"].(primitive.ObjectID); !ok {
				w.WriteHeader(http.StatusInternalServerError)
				u.Respond(w, u.Message(false,
					"Server was not able to process your request"))
				return
			}
			//We have to delete name
			delete(req, "name")

			v, _ = models.DeleteEntity(entity, sd["id"].(primitive.ObjectID), req)
		} else {
			v, _ = models.DeleteEntityManual(entity, bson.M{"slug": name})
		}

	case e == true && e2 == false: // DELETE NORMAL
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			u.Respond(w, u.Message(false, "Error while converting ID to ObjectID"))
			u.ErrLog("Error while converting ID to ObjectID", "DELETE ENTITY", "", r)
			return
		}

		rnd := bson.M{}
		models.RequestGen(rnd, role, domain)

		if entity == "device" {
			v, _ = models.DeleteDeviceF(objID, rnd)
		} else {

			v, _ = models.DeleteEntity(entity, objID, rnd)
		}

	default:
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "DELETE ENTITY", "", r)
		return
	}

	if v["status"] == false {
		w.WriteHeader(http.StatusNotFound)
		u.ErrLog("Error while deleting entity", "DELETE ENTITY", "Not Found", r)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}

	u.Respond(w, v)
}

// swagger:operation PATCH /api/{objs}/{id} objects PartialUpdateObject
// Partially update data in the system.
// This is the preferred method for modifying data in the system.
// If you want to do a full data replace, please use PUT instead.
// If the operation succeeds, the data result will be returned.
// If no new or any information is provided
// an OK will still be returned
// ---
// produces:
// - application/json
// parameters:
// - name: objs
//   in: query
//   description: 'Indicates the location. Only values of "sites","domains",
//   "buildings", "rooms", "racks", "devices", "room-templates",
//   "obj-templates", "rooms", "acs", "panels", "cabinets", "groups",
//   "corridors", "sensors", "stray-devices", "stray-sensors" are acceptable'
//   required: true
//   type: string
//   default: "sites"
// - name: ID
//   in: path
//   description: 'ID of the object or name of Site.
//   For templates the slug is the ID. For stray-devices the name is the ID'
//   required: true
//   type: int
//   default: 999
// - name: Name
//   in: query
//   description: Name of Object
//   required: false
//   type: string
//   default: "INFINITI"
// - name: Category
//   in: query
//   description: Category of Object (ex. Consumer Electronics, Medical)
//   required: false
//   type: string
//   default: "Auto"
// - name: Description
//   in: query
//   description: Description of Object
//   required: false
//   type: string[]
//   default: "High End Worldwide automotive company"
// - name: Domain
//   description: 'Domain of the Object'
//   required: false
//   type: string
//   default: "High End Auto"
// - name: Attributes
//   in: query
//   description: Any other object attributes can be updated
//   required: false
//   type: json
// responses:
//     '200':
//         description: 'Updated. A response body will be returned with
//         a meaningful message.'
//     '400':
//         description: Bad request. An error message will be returned.
//     '404':
//         description: Not Found. An error message will be returned.

// swagger:operation PUT /api/{objs}/{id} objects UpdateObject
// Changes Object data in the system.
// This method will replace the existing data with the JSON
// received, thus fully replacing the data. If you do not
// want to do this, please use PATCH.
// If the operation succeeds, the data result will be returned.
// If no new or any information is provided
// an OK will still be returned
// ---
// produces:
// - application/json
// parameters:
// - name: objs
//   in: query
//   description: 'Indicates the location. Only values of "sites", "domains",
//   "buildings", "rooms", "racks", "devices", "room-templates",
//   "obj-templates", "rooms","acs", "panels", "cabinets", "groups",
//   "corridors","sensors", "stray-devices", "stray-sensors" are acceptable'
//   required: true
//   type: string
//   default: "sites"
// - name: ID
//   in: path
//   description: 'ID of the object or name of Site.
//   For templates the slug is the ID. For stray-devices the name is the ID'
//   required: true
//   type: int
//   default: 999
// - name: Name
//   in: query
//   description: Name of Object
//   required: false
//   type: string
//   default: "INFINITI"
// - name: Category
//   in: query
//   description: Category of Object (ex. Consumer Electronics, Medical)
//   required: false
//   type: string
//   default: "Auto"
// - name: Description
//   in: query
//   description: Description of Object
//   required: false
//   type: string[]
//   default: "High End Worldwide automotive company"
// - name: Domain
//   description: 'Domain of the Object'
//   required: false
//   type: string
//   default: "High End Auto"
// - name: Attributes
//   in: query
//   description: Any other object attributes can be updated
//   required: false
//   type: json
// responses:
//     '200':
//         description: 'Updated. A response body will be returned with
//         a meaningful message.'
//     '400':
//         description: Bad request. An error message will be returned.
//     '404':
//         description: Not Found. An error message will be returned.

var UpdateEntity = func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 UpdateEntity ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)
	var v map[string]interface{}
	var e3 string
	var entity string

	userData := r.Context().Value("user")
	domain := userData.(map[string]interface{})["domain"].(string)
	role := userData.(map[string]interface{})["role"].(string)
	uid := userData.(map[string]interface{})["userID"].(uint)

	println("UserID:", uid)
	println("Domain:", domain)
	println("Role:", role)

	updateData := map[string]interface{}{}
	id, e := mux.Vars(r)["id"]
	name, e2 := mux.Vars(r)["name"]
	isPatch := false
	if r.Method == "PATCH" {
		isPatch = true
	}
	println(r.Method)

	//viewJson(r)

	err := json.NewDecoder(r.Body).Decode(&updateData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		u.ErrLog("Error while decoding request body", "UPDATE ENTITY", "", r)
	}

	//Get entity from URL and strip trailing 's'
	entity, _ = mux.Vars(r)["entity"]
	entity = entity[:len(entity)-1]

	//If templates, format them
	if idx := strings.Index(entity, "-"); idx != -1 {
		entity = entity[:idx] + "_" + entity[idx+1:]
	}

	//Prevents Mongo from creating a new unidentified collection
	if u.EntityStrToInt(entity) < 0 {
		w.WriteHeader(http.StatusNotFound)
		u.Respond(w, u.Message(false, "Invalid object in URL:"+entity+" Please provide a valid object"))
		u.ErrLog("Cannot update invalid object", "UPDATE "+entity, "", r)
		return
	}

	//Flatten updateData if we have
	//a PATCH request
	if isPatch {
		newUpdateData := map[string]interface{}{}
		Flatten("", updateData, newUpdateData)
		updateData = newUpdateData
	}

	switch {
	case e2 == true: // UPDATE SLUG
		var req bson.M
		if entity == "stray_device" || entity == "stray_sensor" {
			req = bson.M{"name": name}
		} else {
			req = bson.M{"slug": name}
		}

		models.RequestGen(req, role, domain)
		v, e3 = models.UpdateEntity(entity, req, &updateData, isPatch)

	case e == true: // UPDATE NORMAL
		println("updating Normale")
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			u.Respond(w, u.Message(false, "Error while converting ID to ObjectID"))
			u.ErrLog("Error while converting ID to ObjectID", "UPDATE ENTITY", "", r)
			return
		}

		println("OBJID:", objID.Hex())
		println("Entity;", entity)

		req := bson.M{"_id": objID}
		models.RequestGen(req, role, domain)
		v, e3 = models.UpdateEntity(entity, req, &updateData, isPatch)

	default:
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message(false, "Error while extracting from path parameters"))
		u.ErrLog("Error while extracting from path parameters", "UPDATE ENTITY", "", r)
		return
	}

	switch e3 {
	case "validate", "Invalid ParentID", "Need ParentID", "invalid":
		w.WriteHeader(http.StatusBadRequest)
		u.ErrLog("Error while updating "+entity, "UPDATE "+strings.ToUpper(entity), e3, r)
	case "internal":
		w.WriteHeader(http.StatusInternalServerError)
		u.ErrLog("Error while updating "+entity, "UPDATE "+strings.ToUpper(entity), e3, r)
	case "mongo: no documents in result", "parent not found":
		w.WriteHeader(http.StatusNotFound)
		u.ErrLog("Error while updating "+entity, "UPDATE "+strings.ToUpper(entity), e3, r)
	default:
	}

	u.Respond(w, v)
}

// swagger:operation GET /api/{objs}? objects GetObject
// Gets an Object using any attribute (with the exception of description)
// via query in the system
// The attributes are in the form {attr}=xyz&{attr1}=abc
// And any combination can be used given that at least 1 is provided.
// ---
// produces:
// - application/json
// parameters:
// - name: objs
//   in: query
//   description: 'Indicates the object. Only values of "domains", "sites",
//   "buildings", "rooms", "racks", "devices", "room-templates",
//   "obj-templates","acs","panels", "groups", "corridors",
//   "sensors", "stray-devices" and "stray-sensors" are acceptable'
//   required: true
//   type: string
//   default: "sites"
// - name: Name
//   in: query
//   description: Name of Site
//   required: false
//   type: string
//   default: "INFINITI"
// - name: Category
//   in: query
//   description: Category of Site (ex. Consumer Electronics, Medical)
//   required: false
//   type: string
//   default: "Auto"
// - name: Domain
//   description: 'Domain of the Site'
//   required: false
//   type: string
//   default: "High End Auto"
// - name: Attributes
//   in: query
//   description: Any other object attributes can be queried
//   required: false
//   type: json
// responses:
//     '204':
//        description: 'Found. A response body will be returned with
//         a meaningful message.'
//     '404':
//        description: Not found. An error message will be returned.
var GetEntityByQuery = func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 GetEntityByQuery ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)
	var data []map[string]interface{}
	var resp map[string]interface{}
	var bsonMap bson.M
	var e, entStr string

	userData := r.Context().Value("user")
	domain := userData.(map[string]interface{})["domain"].(string)
	role := userData.(map[string]interface{})["role"].(string)
	uid := userData.(map[string]interface{})["userID"].(uint)

	println("UserID:", uid)
	println("Domain:", domain)
	println("Role:", role)

	entStr = r.URL.Path[5 : len(r.URL.Path)-1]

	//If templates, format them
	if idx := strings.Index(entStr, "-"); idx != -1 {
		entStr = entStr[:idx] + "_" + entStr[idx+1:]
	}

	query := u.ParamsParse(r.URL, u.EntityStrToInt(entStr))
	js, _ := json.Marshal(query)
	json.Unmarshal(js, &bsonMap)

	//Prevents Mongo from creating a new unidentified collection
	if u.EntityStrToInt(entStr) < 0 {
		w.WriteHeader(http.StatusNotFound)
		u.Respond(w, u.Message(false, "Invalid object in URL:"+entStr+" Please provide a valid object"))
		u.ErrLog("Cannot get invalid object", "GET ENTITYQUERY"+entStr, "", r)
		return
	}

	println("DEBUG entstr", entStr)
	Disp(bsonMap)
	req := bsonMap
	models.RequestGen(req, role, domain)
	data, e = models.GetManyEntities(entStr, bsonMap, nil)

	if len(data) == 0 {
		resp = u.Message(false, "Error: "+e)
		u.ErrLog("Error while getting "+entStr, "GET ENTITYQUERY", e, r)

		switch e {
		case "record not found":
			w.WriteHeader(http.StatusNotFound)
		case "":
			resp = u.Message(false, "Error: No Records Found")
			w.WriteHeader(http.StatusNotFound)
		default:
			resp = u.Message(false, "Error: No Records Found")
			w.WriteHeader(http.StatusNotFound)
		}

	} else {
		message := ""
		switch u.EntityStrToInt(entStr) {
		case u.ROOMTMPL:
			message = "successfully got query for room_template"
		case u.OBJTMPL:
			message = "successfully got query for obj_template"
		default:
			message = "successfully got query for object"
		}
		resp = u.Message(true, message)
	}

	resp["data"] = map[string]interface{}{"objects": data}

	u.Respond(w, resp)
}

// swagger:operation GET /api/{obj}/{id}/{subent} objects GetFromObject
// Obtain all objects 2 levels lower in the system.
// For Example: /api/sites/{id}/buildings
// Will return all buildings of a site
// Returns JSON body with all subobjects under the Object
// ---
// produces:
// - application/json
// parameters:
// - name: obj
//   in: query
//   description: 'Indicates the object. Only values of "sites",
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

// swagger:operation OPTIONS /api/{obj}/{id}/{subent} objects GetFromObjectOptions
// Displays possible operations for the resource in response header.
// ---
// produces:
// - application/json
// parameters:
// - name: objs
//   in: query
//   description: 'Only values of "sites", "domains",
//   "buildings", "rooms", "racks", "devices", and "stray-devices"
//    are acceptable'
// - name: id
//   in: query
//   description: 'ID of the object. For stray-devices and Sites the name
//   can be used as the ID.'
// - name: subent
//   in: query
//   description: 'This refers to the sub object under the objs parameter.
//   Please refer to the OGREE wiki to better understand what objects
//   can be considered as sub objects.'
// responses:
//     '200':
//         description: 'Found. A response body will be returned with
//         a meaningful message.'
//     '404':
//         description: Nothing Found.
var GetEntitiesOfAncestor = func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 GetEntitiesOfAncestor ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)
	var id string
	var e bool
	//Extract string between /api and /{id}
	idx := strings.Index(r.URL.Path[5:], "/") + 4
	entStr := r.URL.Path[5:idx]

	//s, _ := getObjID(id)
	enum := u.EntityStrToInt(entStr)
	//childBase := u.EntityToString(enum + 1)

	userData := r.Context().Value("user")
	domain := userData.(map[string]interface{})["domain"].(string)
	role := userData.(map[string]interface{})["role"].(string)
	uid := userData.(map[string]interface{})["userID"].(uint)

	println("UserID:", uid)
	println("Domain:", domain)
	println("Role:", role)

	//Prevents Mongo from creating a new unidentified collection
	if enum < 0 {
		w.WriteHeader(http.StatusNotFound)
		u.Respond(w, u.Message(false, "Invalid object in URL:"+entStr+" Please provide a valid object"))
		u.ErrLog("Cannot get invalid object", "CREATE CHILDRENOFPARENT"+entStr, "", r)
		return
	}

	resp := u.Message(true, "success")

	if enum == u.SITE {
		id, e = mux.Vars(r)["site_name"]
	} else {
		id, e = mux.Vars(r)["id"]
	}

	if e == false {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET CHILDRENOFPARENT", "", r)
		return
	}

	lastSlashIdx := strings.LastIndex(r.URL.Path, "/")
	indicator := r.URL.Path[lastSlashIdx+1:]
	switch indicator {
	case "acs", "panels", "corridors", "cabinets", "sensors":
		indicator = indicator[:len(indicator)-1]
	default:
		indicator = ""
	}

	req := bson.M{}
	models.RequestGen(req, role, domain)

	data, e1 := models.GetEntitiesOfAncestor(id, req, enum, entStr, indicator)
	if data == nil {
		resp = u.Message(false, "Error while getting "+entStr+"s: "+e1)
		u.ErrLog("Error while getting children of "+entStr,
			"GET CHILDRENOFPARENT", e1, r)

		switch e1 {
		case "record not found":
			w.WriteHeader(http.StatusNotFound)
		default:
		}

	} else {
		resp = u.Message(true,
			"successfully got object")
	}

	if r.Method == "OPTIONS" {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Allow", "GET, OPTIONS")
	} else {
		resp["data"] = map[string]interface{}{"objects": data}
		u.Respond(w, resp)
	}
}

// swagger:operation GET /api/{objs}/{id}/all objects GetFromObject
// Obtain all objects related to specified object in the system.
// Returns JSON body with all subobjects under the Object.
// Note that objects returned will also included relevant objects.
// (ie Room will contain acs, panels etc. Racks and devices will contain sensors)
// ---
// produces:
// - application/json
// parameters:
// - name: objs
//   in: query
//   description: 'Indicates the object. Only values of "sites", "domains"
//   "buildings", "rooms", "racks", "devices", "stray-devices" are acceptable'
//   required: true
//   type: string
//   default: "sites"
// - name: ID
//   in: query
//   description: 'ID of object. For Sites and stray-devices the name
//   can be used as the ID'
//   required: true
//   type: int
//   default: 999
// - name: limit
//   in: query
//   description: 'Limits the level of hierarchy for retrieval. if not
//   specified for devices then the default value is maximum.
//   Example: /api/devices/{id}/all?limit=2'
//   required: false
//   type: string
//   default: 1
// responses:
//     '200':
//         description: 'Found. A response body will be returned with
//         a meaningful message.'
//     '404':
//         description: Nothing Found. An error message will be returned.

// swagger:operation OPTIONS /api/{objs}/{id}/all objects GetFromObjectOptions
// Displays possible operations for the resource in response header.
// ---
// produces:
// - application/json
// parameters:
// - name: objs
//   in: query
//   description: 'Only values of "sites", "domains",
//   "buildings", "rooms", "racks", "devices", and "stray-devices"
//    are acceptable'
// - name: id
//   in: query
//   description: 'ID of the object.For Sites and stray-devices the name
//   can be used as the ID'
// responses:
//     '200':
//         description: 'Found. A response header will be returned with
//         possible operations.'
//     '404':
//         description: Nothing Found.
var GetEntityHierarchy = func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 GetEntityHierarchy ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)
	//Extract string between /api and /{id}
	idx := strings.Index(r.URL.Path[5:], "/") + 4
	entity := r.URL.Path[5:idx]
	resp := u.Message(true, "success")
	var limit int
	var end int
	var lastSlashIdx int
	var data map[string]interface{}
	var e1 string
	var indicator string

	userData := r.Context().Value("user")
	domain := userData.(map[string]interface{})["domain"].(string)
	role := userData.(map[string]interface{})["role"].(string)
	uid := userData.(map[string]interface{})["userID"].(uint)

	println("UserID:", uid)
	println("Domain:", domain)
	println("Role:", role)

	//If template or stray convert '-' -> '_'
	if idx := strings.Index(entity, "-"); idx != -1 {
		entity = entity[:idx] + "_" + entity[idx+1:]
	}

	id, e := mux.Vars(r)["id"]
	if e == false {

		if entity != "site" {
			u.Respond(w, u.Message(false, "Error while parsing path parameters"))
			u.ErrLog("Error while parsing path parameters", "GET ENTITYHIERARCHY", "", r)
			return
		}
		id, e = mux.Vars(r)["site_name"]

		if e == false {
			u.Respond(w, u.Message(false, "Error while parsing site name"))
			u.ErrLog("Error while parsing path parameters", "GET ENTITYHIERARCHY", "", r)
			return
		}
	}

	req := bson.M{}
	models.RequestGen(req, role, domain)

	if entity == "site" {

		//TODO
		newReq := req
		newReq["name"] = id
		_, e := models.GetEntity(newReq, entity)
		if e != "" {
			resp = u.Message(false, "Error while getting :"+entity+","+e)
			u.ErrLog("Error while getting "+entity, "GET "+entity, e, r)
		}

	}

	//Check if the request is a ranged hierarchy
	arr := strings.SplitAfter(r.URL.RawQuery, "limit=")
	if len(arr) == 2 { //limit={number} was provided
		end, _ = strconv.Atoi(arr[1])
		limit = u.EntityStrToInt(entity) + end

		// HACK SECTION FOR FRONT END -- START //

		if end == 0 {

			objID, _ := primitive.ObjectIDFromHex(id)
			newReq := req
			newReq["_id"] = objID
			data, e1 := models.GetEntity(newReq, entity)

			if e1 != "" {
				resp = u.Message(false, "Error while getting :"+entity+","+e1)
				u.ErrLog("Error while getting "+entity, "GET "+entity, e1, r)

				switch e1 {
				case "record not found":
					w.WriteHeader(http.StatusNotFound)
				default:
				}
			} else {
				resp = u.Message(true, "successfully got object")
			}

			resp["data"] = data
			u.Respond(w, resp)
			return
		}

		// HACK SECTION FOR FRONT END -- DONE //

	} else {

		lastSlashIdx = strings.LastIndex(r.URL.Path, "/")
		indicator = r.URL.Path[lastSlashIdx+1:]
		switch indicator {
		case "all":
			//arbitrarily set value to 999 instead of AC
			limit = 999
		case "nonstd":
			//special case
		default:
			//set to int equivalent
			//This strips the trailing s
			limit = u.EntityStrToInt(indicator[:len(indicator)-1])
		}
	}

	println("Indicator: ", indicator)
	println("The limit is: ", limit)

	oID, _ := getObjID(id)

	entNum := u.EntityStrToInt(entity)
	println("EntNum:", entNum)

	//Prevents Mongo from creating a new unidentified collection
	if entNum < 0 {
		w.WriteHeader(http.StatusNotFound)
		u.Respond(w, u.Message(false, "Invalid object in URL:"+entity+" Please provide a valid object"))
		u.ErrLog("Cannot get invalid object", "GET ENTITYHIERARCHY "+entity, "", r)
		return
	}

	println("Entity: ", entity, " & OID: ", oID.Hex())
	if entity == "device" {
		println("RETREIVE")
		end = 999 //Arbitrary value for just obtaining everything
		arr := strings.SplitAfter(r.URL.RawQuery, "limit=")
		if len(arr) == 2 { //limit={number} was provided
			end, _ = strconv.Atoi(arr[1])
			end += 1
		}
		data, e1 = models.GetEntityHierarchy(oID, req, entity, entNum, limit)
	} else {
		data, e1 = models.GetEntityHierarchy(oID, req, entity, entNum, limit)
	}

	if data == nil {
		resp = u.Message(false, "Error while getting :"+entity+","+e1)
		u.ErrLog("Error while getting "+entity, "GET "+entity, e1, r)

		switch e1 {
		case "record not found":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}

	} else {
		//object := ""
		//message := ""
		//if indicator == "" {
		//	object = u.EntityToString(limit)
		//	message = "successfully got " + entity + "'s hierarchy to " + object
		//} else {
		//	if indicator == "all" {
		//		message = "successfully got " + entity + "'s complete hierarchy"
		//	} else {
		//		message = "successfully got " + entity + "'s hierarchy to " + indicator
		//	}

		//}

		resp = u.Message(true, "successfully got object")
	}

	if r.Method == "OPTIONS" {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Allow", "GET, OPTIONS")
	} else {
		resp["data"] = data
		u.Respond(w, resp)
	}
}

// swagger:operation GET /api/{entity}/{name}/all objects GetFromObject
// Obtain all objects related to Site or stray-device in the system using name.
// Returns JSON body with all subobjects
// ---
// produces:
// - application/json
// parameters:
// - name: name
//   in: query
//   description: Name of Site
//   required: true
//   type: int
//   default: 999
// responses:
//     '200':
//         description: 'Found. A response body will be returned with
//         a meaningful message.'
//     '404':
//         description: Nothing Found. An error message will be returned.

// swagger:operation OPTIONS /api/{entity}/{name}/all objects GetFromObjectOptions
// Displays possible operations for the resource in response header.
// ---
// produces:
// - application/json
// parameters:
// - name: name
//   in: query
//   description: 'Name of site.'
// responses:
//     '200':
//         description: 'Found. A response header will be returned with
//         possible operation.'
//     '404':
//         description: Nothing Found.
var GetHierarchyByName = func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 GetHierarchyByName ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)

	entity := "site"
	resp := u.Message(true, "success")
	var indicator string
	var limit int
	var lastSlashIdx int
	var end int

	userData := r.Context().Value("user")
	domain := userData.(map[string]interface{})["domain"].(string)
	role := userData.(map[string]interface{})["role"].(string)
	uid := userData.(map[string]interface{})["userID"].(uint)

	println("UserID:", uid)
	println("Domain:", domain)
	println("Role:", role)

	id, e := mux.Vars(r)["name"]
	if e == false {
		u.Respond(w, u.Message(false, "Error while parsing name"))
		u.ErrLog("Error while parsing path parameters", "GetHierarchyByName", "", r)
		return
	}

	entity, e2 := mux.Vars(r)["entity"]
	if e2 == false {
		u.Respond(w, u.Message(false, "Error while parsing entity"))
		u.ErrLog("Error while parsing path parameters", "GetHierarchyByName", "", r)
		return
	}

	entity = entity[:len(entity)-1]

	//If template or stray convert '-' -> '_'
	if idx := strings.Index(entity, "-"); idx != -1 {
		entity = entity[:idx] + "_" + entity[idx+1:]
	}

	if entity != "site" && entity != "stray_device" && entity != "domain" {
		u.Respond(w, u.Message(false, "Error invalid entity provided"))
		u.ErrLog("Invalid entity found while parsing path parameters", "GetHierarchyByName", "", r)
		return
	}

	//Check if the request is a ranged hierarchy
	arr := strings.SplitAfter(r.URL.RawQuery, "limit=")
	if len(arr) == 2 { //limit={number} was provided
		end, _ = strconv.Atoi(arr[1])
		limit = u.EntityStrToInt(entity) + end

		// HACK SECTION FOR FRONT END -- START //

		if end == 0 {

			//objID, _ := primitive.ObjectIDFromHex(id)
			data, e1 := models.GetEntity(bson.M{"name": id}, "site")

			if e1 != "" {
				resp = u.Message(false, "Error while getting :"+entity+","+e1)
				u.ErrLog("Error while getting "+entity, "GET "+entity, e1, r)

				switch e1 {
				case "record not found":
					w.WriteHeader(http.StatusNotFound)
				default:
				}
			} else {
				resp = u.Message(true, "successfully got object")
			}

			resp["data"] = data
			u.Respond(w, resp)
			return
		}

		// HACK SECTION FOR FRONT END -- DONE //

	} else {

		lastSlashIdx = strings.LastIndex(r.URL.Path, "/")
		indicator = r.URL.Path[lastSlashIdx+1:]
		switch indicator {
		case "all":
			//set to AC1
			if entity == "site" {
				limit = u.AC
			} else { //set limit for stray_device
				limit = 99
			}

		case "nonstd":
			//special case
		default:
			//set to int equivalent
			//This strips the trailing s
			limit = u.EntityStrToInt(indicator[:len(indicator)-1])
		}
	}

	println("Indicator: ", indicator)
	println("The limit is: ", limit)

	entInt := u.EntityStrToInt(entity)

	req := bson.M{}
	models.RequestGen(req, role, domain)
	data, e1 := models.GetHierarchyByName(entity, id, req, entInt, limit)

	if data == nil {
		resp = u.Message(false, "Error while getting :"+entity+","+e1)
		u.ErrLog("Error while getting "+entity, "GET "+entity, e1, r)

		switch e1 {
		case "record not found":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}

	} else {
		//object := ""
		//message := ""
		//if indicator == "" {
		//	object = u.EntityToString(limit)
		//	message = "successfully got " + entity + "'s hierarchy to " + object
		//} else {
		//	if indicator == "all" {
		//		message = "successfully got " + entity + "'s complete hierarchy"
		//	} else {
		//		message = "successfully got " + entity + "'s hierarchy to " + indicator
		//	}

		//}
		resp = u.Message(true, "successfully got object")
	}

	if r.Method == "OPTIONS" {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Allow", "GET, OPTIONS")
	} else {
		resp["data"] = data
		u.Respond(w, resp)
	}
}

// swagger:operation GET /api/{objs}/{id}/* objects GetFromObect
// A category of objects of a Parent Object can be retrieved from the system.
// The path can only contain object type or object names
// ---
// produces:
// - application/json
// parameters:
// - name: objs
//   in: query
//   description: 'Indicates the object. Only values of "sites", "domains",
//   "buildings", "rooms", "racks", "devices", "stray-devices" are acceptable'
//   required: true
//   type: string
//   default: "sites"
// - name: ID
//   in: path
//   description: 'ID of desired object. For Sites and stray-devices the name
//   can be used as the ID'
//   required: true
//   type: string
//   default: "INFINITI"
// - name: '*'
//   in: path
//   description: 'Hierarchal path to desired object(s).
//   For rooms it can additionally have "acs","panels",
//   "corridors", "sensors" and "cabinets".
//   For devices it can have "sensors"
//   For racks it can have "sensors"'
//   required: true
//   type: string
//   default: "/buildings/BuildingB/RoomA"
// responses:
//     '200':
//         description: 'Found. A response body will be returned with
//         a meaningful message.'
//     '404':
//         description: Not Found. An error message will be returned.

// swagger:operation OPTIONS /api/{objs}/{id}/* objects GetFromObjectOptions
// Displays possible operations for the resource in response header.
// ---
// produces:
// - application/json
// parameters:
// - name: objs
//   in: query
//   description: 'Only values of "sites", "domains",
//   "buildings", "rooms", "racks", "devices", and "stray-devices"
//    are acceptable'
// - name: id
//   in: query
//   description: 'ID of the object.For Sites and stray-devices the name
//   can be used as the ID'
// - name: '*'
//   in: path
//   description: 'Hierarchal path to desired object(s).
//   For rooms it can additionally have "acs","panels",
//   "corridors", "sensors" and "cabinets".
//   For devices it can have "sensors"
//   For racks it can have "sensors"'
//   required: true
//   type: string
//   default: "/buildings/BuildingB/RoomA"
// responses:
//     '200':
//         description: 'Found. A response header will be returned with
//         possible operations.'
//     '404':
//         description: Not Found.
var GetEntitiesUsingNamesOfParents = func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 GetEntitiesUsingNamesOfParents ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)
	//Extract string between /api and /{id}
	idx := strings.Index(r.URL.Path[5:], "/") + 4
	entity := r.URL.Path[5:idx]
	resp := u.Message(true, "success")

	//If template or stray convert '-' -> '_'
	if idx := strings.Index(entity, "-"); idx != -1 {
		entity = entity[:idx] + "_" + entity[idx+1:]
	}

	userData := r.Context().Value("user")
	domain := userData.(map[string]interface{})["domain"].(string)
	role := userData.(map[string]interface{})["role"].(string)
	uid := userData.(map[string]interface{})["userID"].(uint)

	println("UserID:", uid)
	println("Domain:", domain)
	println("Role:", role)

	id, e := mux.Vars(r)["id"]
	tname, e1 := mux.Vars(r)["site_name"]
	if e == false && e1 == false {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET ENTITIESUSINGANCESTORNAMES", "", r)
		return
	}

	//Prevents Mongo from creating a new unidentified collection
	if u.EntityStrToInt(entity) < 0 {
		w.WriteHeader(http.StatusNotFound)
		u.Respond(w, u.Message(false, "Invalid object in URL:"+entity+" Please provide a valid object"))
		u.ErrLog("Cannot get invalid object", "GET ENTITIESUSINGANCESTORNAMES "+entity, "", r)
		return
	}

	arr := (strings.Split(r.URL.Path, "/")[4:])
	ancestry := make([]map[string]string, 0)

	for i, k := range arr {

		key := k[:len(k)-1]

		//If templates, format them
		if idx := strings.Index(key, "-"); idx != -1 {
			key = key[:idx] + "_" + key[idx+1:]
		}

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
					u.Respond(w, u.Message(false, "Invalid object in URL:"+key+" Please provide a valid object"))
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
		var e3 string
		req := bson.M{}
		models.RequestGen(req, role, domain)
		if e1 == true {
			println("we are getting entities here")
			data, e3 = models.GetEntitiesUsingSiteAsAncestor(entity, tname, req, ancestry)

		} else {
			data, e3 = models.GetEntitiesUsingAncestorNames(entity, oID, req, ancestry)
		}

		if data == nil || len(data) == 0 {
			resp = u.Message(false, "Error while getting :"+entity+","+e3)
			u.ErrLog("Error while getting "+entity, "GET "+entity, e3, r)

			switch e3 {
			case "record not found":
				w.WriteHeader(http.StatusNotFound)
			default:
				w.WriteHeader(http.StatusNotFound)
			}

		} else {
			if r.Method == "OPTIONS" {
				w.Header().Add("Content-Type", "application/json")
				w.Header().Add("Allow", "GET, OPTIONS")
				return
			}
			resp = u.Message(true, "successfully got object")
		}

		resp["data"] = map[string]interface{}{"objects": data}
		u.Respond(w, resp)
	} else { //We are only retrieving an entity
		var data map[string]interface{}
		var e3 string
		req := bson.M{}
		models.RequestGen(req, role, domain)
		if e1 == true {
			data, e3 = models.GetEntityUsingSiteAsAncestor(entity, tname, req, ancestry)
		} else {
			data, e3 = models.GetEntityUsingAncestorNames(entity, oID, req, ancestry)
		}

		//data, e := models.GetEntityUsingAncestorNames(entity, oID, ancestry)
		if data == nil || len(data) == 0 {
			resp = u.Message(false, "Error while getting :"+entity+","+e3)
			u.ErrLog("Error while getting "+entity, "GET "+entity, e3, r)

			switch e3 {
			case "record not found":
				w.WriteHeader(http.StatusNotFound)
			default:
				w.WriteHeader(http.StatusNotFound)
			}

		} else {
			upperEnt := arr[len(arr)-2]
			upperEnt = upperEnt[:len(upperEnt)-1]
			resp = u.Message(true, "successfully got object")
		}

		if r.Method == "OPTIONS" && data != nil {
			w.Header().Add("Content-Type", "application/json")
			w.Header().Add("Allow", "GET, OPTIONS")
		} else {
			resp["data"] = data
			u.Respond(w, resp)
		}
	}

}

// swagger:operation POST /api/validate/{obj} objects ValidateObject
// Checks the received data and verifies if the object can be created in the system.
// ---
// produces:
// - application/json
// parameters:
// - name: objs
//   in: query
//   description: 'Indicates the Object. Only values of "domains", "sites",
//   "buildings", "rooms", "racks", "devices", "acs", "panels",
//   "cabinets", "groups", "corridors",
//   "room-templates", "obj-templates", "sensors", "stray-devices"
//   "stray-sensors" are acceptable'
//   required: true
//   type: string
//   default: "sites"
// - name: Name
//   in: query
//   description: Name of object
//   required: true
//   type: string
//   default: "Object A"
// - name: Category
//   in: query
//   description: Category of Object (ex. Consumer Electronics, Medical)
//   required: true
//   type: string
//   default: "Research"
// - name: Domain
//   description: 'Domain of Object'
//   required: true
//   type: string
//   default: 999
// - name: ParentID
//   description: 'All objects are linked to a
//   parent with the exception of Tenant since it has no parent'
//   required: true
//   type: int
//   default: 999
// - name: Description
//   in: query
//   description: Description of Object
//   required: false
//   type: string[]
//   default: ["Some abandoned object in Grenoble"]
// - name: Attributes
//   in: query
//   description: 'Any other object attributes can be added.
//   They are required depending on the obj type.'
//   required: true
//   type: json
// responses:
//     '200':
//         description: 'Createable. A response body will be returned with
//         a meaningful message.'
//     '400':
//         description: 'Bad request. A response body with an error
//         message will be returned.'
//     '404':
//         description: Not Found. An error message will be returned.

// swagger:operation OPTIONS /api/validate/{obj} objects ValidateObjectOptions
// Displays possible operations for the resource in response header.
// ---
// produces:
// - application/json
// parameters:
// - name: obj
//   in: query
//   description: 'Only values of "domains", "sites",
//   "buildings", "rooms", "racks", "devices", "room-templates",
//   "obj-templates", "rooms", "acs", "panels",
//   "cabinets", "groups", "corridors","sensors","stray-devices"
//   "stray-sensors" are acceptable'
// responses:
//     '200':
//         description: 'Request is valid.'
//     '404':
//         description: Not Found. An error message will be returned.
var ValidateEntity = func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 ValidateEntity ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)
	var obj map[string]interface{}
	entity, e1 := mux.Vars(r)["entity"]
	entity = entity[:len(entity)-1]

	userData := r.Context().Value("user")
	domain := userData.(map[string]interface{})["domain"].(string)
	role := userData.(map[string]interface{})["role"].(string)
	uid := userData.(map[string]interface{})["userID"].(uint)

	println("User:", uid)
	println("Role:", role)
	println("Domain:", domain)

	//If templates or stray-devices, format them
	if idx := strings.Index(entity, "-"); idx != -1 {
		//entStr[idx] = '_'
		entity = entity[:idx] + "_" + entity[idx+1:]
	}
	entInt := u.EntityStrToInt(entity)

	if e1 == false || entInt == -1 {
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
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		u.ErrLog("Error while decoding request body", "VALIDATE "+entity, "", r)
		return
	}

	if entInt == u.DOMAIN {
		if !models.EnsureUserIsSuper(role) {
			w.WriteHeader(http.StatusMethodNotAllowed)
			u.Respond(w, u.Message(false, "This "+role+
				" does not have sufficient permissions to create"+
				" this object under this domain. Please refer "+
				" to an administrator or manager for more assistance"))
			u.ErrLog("Cannot validate object creation due to limited user privilege",
				"Validate CREATE "+entity, "", r)
			return
		}
	} else {
		//Ensure user has permission for this object
		ok, reason := models.EnsureObjectPermission(obj, domain, role)
		if !ok {
			w.WriteHeader(http.StatusMethodNotAllowed)
			if reason == "" {

				u.Respond(w, u.Message(false, "This "+role+
					" does not have sufficient permissions to create"+
					" this object under this domain "))

			} else {
				u.Respond(w, u.Message(false, reason))
			}
			u.ErrLog("Cannot validate object creation due to limited user privilege",
				"Validate CREATE "+entity, "", r)
			return
		}
	}

	ans, status := models.ValidateEntity(entInt, obj)
	if status == true {
		u.Respond(w, map[string]interface{}{"status": true, "message": "This object can be created"})
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	u.Respond(w, ans)
}

// swagger:operation OPTIONS /api/{objs}/* objects ObjectOptions
// Displays possible operations for the resource in response header.
// ---
// produces:
// - application/json
// parameters:
// - name: objs
//   in: query
//   description: 'Only values of "sites", "domains",
//   "buildings", "rooms", "racks", "devices", "room-templates",
//   "obj-templates", "rooms", "separators", "acs", "panels", "rows",
//   "tiles", "cabinets", "groups", "corridors","sensors","stray-devices",
//    "stray-sensors" are acceptable'
// responses:
//     '200':
//         description: 'Request is valid.'
//     '404':
//         description: Not Found. An error message will be returned.
var BaseOption = func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 BaseOption ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)
	entity, e1 := mux.Vars(r)["entity"]
	entity = entity[:len(entity)-1]
	if e1 == false || u.EntityStrToInt(entity) == -1 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Allow", "GET, DELETE, OPTIONS, PATCH, POST, PUT")

}

var GetEntityHierarchyNonStd = func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 GetEntityHierarchyNonStd ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)
	var e, e1 bool
	var err string
	//Extract string between /api and /{id}
	idx := strings.Index(r.URL.Path[5:], "/") + 4
	entity := r.URL.Path[5:idx]

	id, e := mux.Vars(r)["id"]
	resp := u.Message(true, "success")
	data := map[string]interface{}{}
	//result := map[string][]map[string]interface{}{}

	if e == false {
		if id, e1 = mux.Vars(r)["site_name"]; e1 == false {
			u.Respond(w, u.Message(false, "Error while parsing path parameters"))
			u.ErrLog("Error while parsing path parameters", "GETHIERARCHYNONSTD", "", r)
			return
		}
	}

	entNum := u.EntityStrToInt(entity)

	if entity == "site" {
		println("Getting SITE HEIRARCHY")
		println("With ID: ", id)
		//data, err = models.GetHierarchyByName(entity, id, entNum, u.AC)
		if err != "" {
			println("We have ERR")
		}
	} else {
		//oID, _ := getObjID(id)
		//data, err = models.GetEntityHierarchy(oID, entity, entNum, u.AC)
	}

	if data == nil {
		resp = u.Message(false, "Error while getting NonStandard Hierarchy: "+err)
		u.ErrLog("Error while getting NonStdHierarchy", "GETNONSTDHIERARCHY", err, r)

		switch err {
		case "record not found":
			w.WriteHeader(http.StatusNotFound)
		default:
		}

	} else {
		resp = u.Message(true, "success")
		result := parseDataForNonStdResult(entity, entNum, data)
		resp["data"] = result
		//u.Respond(w, resp)
	}

	//resp["data"] = data
	/*resp["data"] = sites
	resp["buildings"] = bldgs
	resp["rooms"] = rooms
	resp["racks"] = racks
	resp["devices"] = devices*/
	u.Respond(w, resp)
}
