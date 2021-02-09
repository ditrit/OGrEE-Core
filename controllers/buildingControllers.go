package controllers

import (
	"encoding/json"
	"net/http"
	"p3/models"
	u "p3/utils"
	"strconv"

	"github.com/gorilla/mux"
)

var CreateBuilding = func(w http.ResponseWriter, r *http.Request) {

	bldg := &models.Building{}
	err := json.NewDecoder(r.Body).Decode(bldg)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		return
	}

	resp := bldg.Create()
	u.Respond(w, resp)
}

//Retrieve bldg using Bldg ID
var GetBuilding = func(w http.ResponseWriter, r *http.Request) {

	id, e := strconv.Atoi(mux.Vars(r)["id"])
	resp := u.Message(true, "success")

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
	}

	data := models.GetBuilding(uint(id))
	if data == nil {
		resp = u.Message(false, "unsuccessful")
	}

	resp["data"] = data
	u.Respond(w, resp)
}

var DeleteBuilding = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
	}

	v := models.DeleteBuilding(uint(id))
	u.Respond(w, v)
}

//Updates work by passing ID in path parameter
var UpdateBuilding = func(w http.ResponseWriter, r *http.Request) {

	bldg := &models.Building{}
	id, e := strconv.Atoi(mux.Vars(r)["id"])

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
	}

	err := json.NewDecoder(r.Body).Decode(bldg)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
	}

	v := models.UpdateBuilding(uint(id), bldg)
	u.Respond(w, v)
}
