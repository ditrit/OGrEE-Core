package controllers

import (
	"encoding/json"
	"net/http"
	"p3/models"
	u "p3/utils"
	"strconv"

	"github.com/gorilla/mux"
)

var CreateTenant = func(w http.ResponseWriter, r *http.Request) {
	tenant := &models.Tenant{}

	err := json.NewDecoder(r.Body).Decode(tenant)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		return
	}
	//tenant.ID = user
	resp := tenant.Create()
	u.Respond(w, resp)
}

var GetTenantFor = func(w http.ResponseWriter, r *http.Request) {
	var resp map[string]interface{}
	id, err := strconv.Atoi(mux.Vars(r)["id"])

	if err != nil {
		u.Respond(w, u.Message(false, "Error while extracting from path parameters"))
	}

	data := models.GetTenant(uint(id))

	if data == nil {
		resp = u.Message(false, "Not found")
	} else {
		resp = u.Message(true, "success")
	}

	resp["data"] = data
	u.Respond(w, resp)
}

var GetAllTenants = func(w http.ResponseWriter, r *http.Request) {

	data := models.GetAllTenants()
	resp := u.Message(true, "success")
	resp["data"] = data
	u.Respond(w, resp)
}

var UpdateTenant = func(w http.ResponseWriter, r *http.Request) {

	id, e := strconv.Atoi(mux.Vars(r)["id"])
	if e != nil {
		u.Respond(w, u.Message(false, "Error while extracting from path parameters"))
	}
	tenant := &models.Tenant{}

	err := json.NewDecoder(r.Body).Decode(tenant)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
	}

	v := models.UpdateTenant(uint(id), tenant)
	u.Respond(w, v)
}

//This delete function is for 1 tenant 1 user
/*var DeleteTenant = func(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value("user").(uint)
	v := models.DeleteTenant(id)
	u.Respond(w, v)
}*/

var DeleteTenant = func(w http.ResponseWriter, r *http.Request) {

	id, e := strconv.Atoi(mux.Vars(r)["id"])
	if e != nil {
		u.Respond(w, u.Message(false, "Error while extracting from path parameters"))
	}

	v := models.DeleteTenant(uint(id))
	u.Respond(w, v)
}
