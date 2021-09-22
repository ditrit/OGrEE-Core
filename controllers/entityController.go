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
