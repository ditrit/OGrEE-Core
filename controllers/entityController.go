package controllers

import (
	"encoding/json"
	"net/http"
	"p3/models"
	u "p3/utils"
	"strings"

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

var CreateEntity = func(w http.ResponseWriter, r *http.Request) {
	//tenant := &models.Tenant{}
	entStr := r.URL.Path[5 : len(r.URL.Path)-1] //strip the '/api' in URL
	entUpper := strings.ToUpper(entStr)         // and the trailing 's'
	entity := map[string]interface{}{}
	err := json.NewDecoder(r.Body).Decode(&entity)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		u.ErrLog("Error while decoding request body", "CREATE "+entStr, "", r)
		return
	}

	i := u.EntityStrToInt(entStr)
	println("ENT: ", entStr)
	println("ENUM VAL: ", i)

	resp, e := models.CreateEntity(i, entity)

	switch e {
	case "validate":
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

var GetEntity = func(w http.ResponseWriter, r *http.Request) {
	id, e := mux.Vars(r)["id"]
	resp := u.Message(true, "success")

	if e == false {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET ENTITY", "", r)
		return
	}

	x, e2 := getObjID(id)
	if e2 != nil {
		u.Respond(w, u.Message(false, "Error while converting ID to ObjectID"))
		u.ErrLog("Error while converting ID to ObjectID", "GET ENTITY", "", r)
		return
	}

	//Get entity type and strip trailing 's'
	s := r.URL.Path[5 : strings.LastIndex(r.URL.Path, "/")-1]

	data, e1 := models.GetEntity(x, s)
	if data == nil {
		resp = u.Message(false, "Error while getting "+s+": "+e1)
		u.ErrLog("Error while getting "+s, "GET "+strings.ToUpper(s), "", r)

		switch e1 {
		case "record not found":
			w.WriteHeader(http.StatusNotFound)
		default:
		}

	} else {
		resp = u.Message(true, "success")
	}

	resp["data"] = data
	u.Respond(w, resp)
}

var GetAllEntities = func(w http.ResponseWriter, r *http.Request) {
	entStr := r.URL.Path[5 : len(r.URL.Path)-1] //strip the '/api' in URL
	entUpper := strings.ToUpper(entStr)         // and the trailing 's'

	resp := u.Message(true, "success")

	//entInt := u.EntityStrToInt(entStr)

	data, e := models.GetAllEntities(entStr)
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
		resp = u.Message(true, "success")
	}

	resp["data"] = map[string]interface{}{"objects": data}

	u.Respond(w, resp)
}

var DeleteEntity = func(w http.ResponseWriter, r *http.Request) {
	id, e := mux.Vars(r)["id"]
	if e == false {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "DELETE ENTITY", "", r)
		return
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while converting ID to ObjectID"))
		u.ErrLog("Error while converting ID to ObjectID", "DELETE ENTITY", "", r)
		return
	}

	//Get entity from URL and strip trailing 's'
	entity := r.URL.Path[5 : strings.LastIndex(r.URL.Path, "/")-1]

	v := models.DeleteEntity(entity, objID)

	if v["status"] == false {
		w.WriteHeader(http.StatusNotFound)
		u.ErrLog("Error while deleting entity", "DELETE ENTITY", "Not Found", r)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}

	u.Respond(w, v)
}

var UpdateEntity = func(w http.ResponseWriter, r *http.Request) {
	updateData := map[string]interface{}{}

	id, e := mux.Vars(r)["id"]
	if e == false {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message(false, "Error while extracting from path parameters"))
		u.ErrLog("Error while extracting from path parameters", "UPDATE ENTITY", "", r)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&updateData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		u.ErrLog("Error while decoding request body", "UPDATE ENTITY", "", r)
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while converting ID to ObjectID"))
		u.ErrLog("Error while converting ID to ObjectID", "DELETE ENTITY", "", r)
		return
	}

	//Get entity from URL and strip trailing 's'
	entity := r.URL.Path[5 : strings.LastIndex(r.URL.Path, "/")-1]

	v, e1 := models.UpdateEntity(entity, objID, &updateData)

	switch e1 {

	case "validate":
		w.WriteHeader(http.StatusBadRequest)
		u.ErrLog("Error while updating "+entity, "UPDATE "+strings.ToUpper(entity), e1, r)
	case "internal":
		w.WriteHeader(http.StatusInternalServerError)
		u.ErrLog("Error while updating "+entity, "UPDATE "+strings.ToUpper(entity), e1, r)
	case "record not found":
		w.WriteHeader(http.StatusNotFound)
		u.ErrLog("Error while updating "+entity, "UPDATE "+strings.ToUpper(entity), e1, r)
	default:
	}

	u.Respond(w, v)
}

var GetEntityByQuery = func(w http.ResponseWriter, r *http.Request) {
	var resp map[string]interface{}
	var bsonMap bson.M
	entStr := r.URL.Path[5 : len(r.URL.Path)-1] //strip the '/api' in URL

	query := u.ParamsParse(r.URL)
	js, _ := json.Marshal(query)
	json.Unmarshal(js, &bsonMap)

	data, e := models.GetEntityByQuery(entStr, bsonMap)

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
			w.WriteHeader(http.StatusNotFound)
		}

	} else {
		resp = u.Message(true, "success")
	}

	resp["data"] = map[string]interface{}{"objects": data}

	u.Respond(w, resp)
}

var GetEntitiesOfParent = func(w http.ResponseWriter, r *http.Request) {
	var id string
	var e bool
	//Extract string between /api and /{id}
	idx := strings.Index(r.URL.Path[5:], "/") + 4
	entStr := r.URL.Path[5:idx]

	//s, _ := getObjID(id)
	enum := u.EntityStrToInt(entStr)
	childBase := u.EntityToString(enum + 1)

	resp := u.Message(true, "success")

	if enum == TENANT {
		id, e = mux.Vars(r)["tenant_name"]
	} else {
		id, e = mux.Vars(r)["id"]
	}

	if e == false {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET CHILDRENOFPARENT", "", r)
		return
	}

	data, e1 := models.GetEntitiesOfParent(childBase, id)
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
		resp = u.Message(true, "success")
	}

	resp["data"] = map[string]interface{}{"objects": data}

	u.Respond(w, resp)
}

var GetEntityHierarchy = func(w http.ResponseWriter, r *http.Request) {
	//Extract string between /api and /{id}
	idx := strings.Index(r.URL.Path[5:], "/") + 4
	entity := r.URL.Path[5:idx]
	resp := u.Message(true, "success")
	var limit int

	id, e := mux.Vars(r)["id"]
	if e == false {

		if entity != "tenant" {
			u.Respond(w, u.Message(false, "Error while parsing path parameters"))
			u.ErrLog("Error while parsing path parameters", "GET ENTITYHIERARCHY", "", r)
			return
		}
		id, e = mux.Vars(r)["tenant_name"]

		if e == false {
			u.Respond(w, u.Message(false, "Error while parsing tenant name"))
			u.ErrLog("Error while parsing path parameters", "GET ENTITYHIERARCHY", "", r)
			return
		}
	}

	if entity == "tenant" {

		_, e := models.GetEntityByName(entity, id)
		if e != "" {
			resp = u.Message(false, "Error while getting :"+entity+","+e)
			u.ErrLog("Error while getting "+entity, "GET "+entity, e, r)
		}

	}

	//Check if the request is a ranged hierarchy
	lastSlashIdx := strings.LastIndex(r.URL.Path, "/")
	indicator := r.URL.Path[lastSlashIdx+1:]
	switch indicator {
	case "all":
		//set to SUBDEV1
		limit = SUBDEV1
	case "nonstd":
		//special case
	default:
		//set to int equivalent
		//This strips the trailing s
		limit = u.EntityStrToInt(indicator[:len(indicator)-1])
	}
	println("Indicator: ", indicator)
	println("The limit is: ", limit)

	oID, _ := getObjID(id)

	entNum := u.EntityStrToInt(entity)

	println("Entity: ", entity, " & OID: ", oID.Hex())
	data, e1 := models.GetEntityHierarchy(entity, oID, entNum, limit+1)

	if data == nil {
		resp = u.Message(false, "Error while getting :"+entity+","+e1)
		u.ErrLog("Error while getting "+entity, "GET "+entity, e1, r)

		switch e1 {
		case "record not found":
			w.WriteHeader(http.StatusNotFound)
		default:
		}

	} else {
		resp = u.Message(true, "success")
	}

	resp["data"] = data
	u.Respond(w, resp)
}

var GetTenantHierarchy = func(w http.ResponseWriter, r *http.Request) {
	entity := "tenant"
	resp := u.Message(true, "success")
	var limit int

	id, e := mux.Vars(r)["tenant_name"]
	if e == false {
		u.Respond(w, u.Message(false, "Error while parsing tenant name"))
		u.ErrLog("Error while parsing path parameters", "GET TENANTHIERARCHY", "", r)
		return
	}

	//Check if the request is a ranged hierarchy
	lastSlashIdx := strings.LastIndex(r.URL.Path, "/")
	indicator := r.URL.Path[lastSlashIdx+1:]
	switch indicator {
	case "all":
		//set to SUBDEV1
		limit = SUBDEV1
	case "nonstd":
		//special case
	default:
		//set to int equivalent
		//This strips the trailing s
		limit = u.EntityStrToInt(indicator[:len(indicator)-1])
	}
	println("Indicator: ", indicator)
	println("The limit is: ", limit)

	data, e1 := models.GetTenantHierarchy(entity, id, TENANT, limit+1)

	if data == nil {
		resp = u.Message(false, "Error while getting :"+entity+","+e1)
		u.ErrLog("Error while getting "+entity, "GET "+entity, e1, r)

		switch e1 {
		case "record not found":
			w.WriteHeader(http.StatusNotFound)
		default:
		}

	} else {
		resp = u.Message(true, "success")
	}

	resp["data"] = data
	u.Respond(w, resp)
}

var GetEntitiesUsingNamesOfParents = func(w http.ResponseWriter, r *http.Request) {
	//Extract string between /api and /{id}
	idx := strings.Index(r.URL.Path[5:], "/") + 4
	entity := r.URL.Path[5:idx]
	ancestry := make(map[string]string, 0)
	resp := u.Message(true, "success")

	id, e := mux.Vars(r)["id"]
	if e == false {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET ENTITIESUSINGANCESTORNAMES", "", r)
		return
	}

	arr := (strings.Split(r.URL.Path, "/")[4:])

	for i, k := range arr {
		if i%2 == 0 { //The keys (entities) are at the even indexes
			if i+1 >= len(arr) {
				ancestry[k[:len(k)-1]] = "all"
			} else {
				ancestry[k[:len(k)-1]] = arr[i+1]
			}
		}
	}

	oID, _ := getObjID(id)

	if len(arr)%2 != 0 { //This means we are getting entities
		data, e := models.GetEntitiesUsingAncestorNames(entity, oID, ancestry)
		if data == nil {
			resp = u.Message(false, "Error while getting :"+entity+","+e)
			u.ErrLog("Error while getting "+entity, "GET "+entity, e, r)

			switch e {
			case "record not found":
				w.WriteHeader(http.StatusNotFound)
			default:
			}

		} else {
			resp = u.Message(true, "success")
		}

		resp["data"] = data
		u.Respond(w, resp)
	} else { //We are only retrieving an entity
		data, e := models.GetEntityUsingAncestorNames(entity, oID, ancestry)
		if data == nil {
			resp = u.Message(false, "Error while getting :"+entity+","+e)
			u.ErrLog("Error while getting "+entity, "GET "+entity, e, r)

			switch e {
			case "record not found":
				w.WriteHeader(http.StatusNotFound)
			default:
			}

		} else {
			resp = u.Message(true, "success")
		}

		resp["data"] = data
		u.Respond(w, resp)
	}

}

var GetEntitiesUsingNameOfTenant = func(w http.ResponseWriter, r *http.Request) {
	//Extract string between /api and /{id}
	idx := strings.Index(r.URL.Path[5:], "/") + 4
	entity := r.URL.Path[5:idx]
	ancestry := make(map[string]string, 0)
	resp := u.Message(true, "success")

	id, e := mux.Vars(r)["tenant_name"]
	if e == false {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET ENTITIESUSINGNAMEOFTENANT", "", r)
		return
	}

	arr := (strings.Split(r.URL.Path, "/")[4:])

	for i, k := range arr {
		if i%2 == 0 { //The keys (entities) are at the even indexes
			if i+1 >= len(arr) {
				ancestry[k[:len(k)-1]] = "all"
			} else {
				ancestry[k[:len(k)-1]] = arr[i+1]
			}
		}
	}

	if len(arr)%2 != 0 { //This means we are getting entities
		data, e := models.GetEntitiesUsingTenantAsAncestor(entity, id, ancestry)
		if data == nil {
			resp = u.Message(false, "Error while getting :"+entity+","+e)
			u.ErrLog("Error while getting "+entity, "GET "+entity, e, r)

			switch e {
			case "record not found":
				w.WriteHeader(http.StatusNotFound)
			default:
			}

		} else {
			resp = u.Message(true, "success")
		}

		resp["data"] = data
		u.Respond(w, resp)
	} else { //We are only retrieving an entity
		data, e := models.GetEntityUsingTenantAsAncestor(entity, id, ancestry)
		if data == nil {
			resp = u.Message(false, "Error while getting :"+entity+","+e)
			u.ErrLog("Error while getting "+entity, "GET "+entity, e, r)

			switch e {
			case "record not found":
				w.WriteHeader(http.StatusNotFound)
			default:
			}

		} else {
			resp = u.Message(true, "success")
		}

		resp["data"] = data
		u.Respond(w, resp)
	}

}

var GetEntityHierarchyNonStd = func(w http.ResponseWriter, r *http.Request) {
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
		if id, e1 = mux.Vars(r)["tenant_name"]; e1 == false {
			u.Respond(w, u.Message(false, "Error while parsing Tpath parameters"))
			u.ErrLog("Error while parsing path parameters", "GETHIERARCHYNONSTD", "", r)
			return
		}
	}

	entNum := u.EntityStrToInt(entity)

	if entity == "tenant" {
		println("Getting TENANT HEIRARCHY")
		println("With ID: ", id)
		data, err = models.GetTenantHierarchy(entity, id, entNum, SUBDEV1)
		if err != "" {
			println("We have ERR")
		}
	} else {
		oID, _ := getObjID(id)
		data, err = models.GetEntityHierarchy(entity, oID, entNum, SUBDEV1)
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
