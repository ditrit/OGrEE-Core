package controllers

import (
	"encoding/json"
	"net/http"
	"p3/models"
	u "p3/utils"
	"strconv"

	"github.com/gorilla/mux"
)

var CreateSite = func(w http.ResponseWriter, r *http.Request) {

	site := &models.Site{}
	err := json.NewDecoder(r.Body).Decode(site)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		return
	}

	resp := site.Create()
	u.Respond(w, resp)
}

//Retrieve sites using User ID
var GetSitesByUserID = func(w http.ResponseWriter, r *http.Request) {

	id := r.Context().Value("user").(uint)
	resp := u.Message(true, "success")

	data := models.GetSites(uint(id))
	if data == nil {
		resp = u.Message(false, "unsuccessful")
	}

	resp["data"] = data
	u.Respond(w, resp)
}

//Retrieve sites using Tenant ID
var GetSitesByParentID = func(w http.ResponseWriter, r *http.Request) {

	st := &models.Site{}
	err := json.NewDecoder(r.Body).Decode(st)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
	}
	resp := u.Message(true, "success")

	data := models.GetSites(uint(st.Domain))
	if data == nil {
		resp = u.Message(false, "unsuccessful")
	}

	resp["data"] = data
	u.Respond(w, resp)
}

//Retrieve site using Site ID
var GetSite = func(w http.ResponseWriter, r *http.Request) {

	id, e := strconv.Atoi(mux.Vars(r)["id"])
	resp := u.Message(true, "success")

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
	}

	data := models.GetSite(uint(id))
	if data == nil {
		resp = u.Message(false, "unsuccessful")
	}

	resp["data"] = data
	u.Respond(w, resp)
}

/*var DeleteSite = func(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value("user").(uint)
	v := models.DeleteSite(id)
	u.Respond(w, v)
}
*/

var DeleteSiteByID = func(w http.ResponseWriter, r *http.Request) {
	st := &models.Site{}
	err := json.NewDecoder(r.Body).Decode(st)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
	}

	v := models.DeleteSite(st.ID)
	u.Respond(w, v)
}

//Delete all sites of a tenant
var DeleteSites = func(w http.ResponseWriter, r *http.Request) {
	st := &models.Site{}
	err := json.NewDecoder(r.Body).Decode(st)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
	}

	v := models.DeleteSitesOfTenant(uint(st.Domain))
	u.Respond(w, v)
}

//Updates work by passing ID
//and new data in the JSON Body
var UpdateSite = func(w http.ResponseWriter, r *http.Request) {
	site := &models.Site{}
	//id := r.Context().Value("user").(uint)

	err := json.NewDecoder(r.Body).Decode(site)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
	}

	v := models.UpdateSite(site.ID, site)
	u.Respond(w, v)
}
