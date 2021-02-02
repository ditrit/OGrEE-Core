package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"p3/models"
	u "p3/utils"
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

/*var GetSite = func(w http.ResponseWriter, r *http.Request) {

	siteID := r.Context().Value("siteid").(uint)

	data := models.GetSite(siteID)
	resp := u.Message(true, "success")
	resp["data"] = data
	u.Respond(w, resp)
}*/

var GetSites = func(w http.ResponseWriter, r *http.Request) {

	id := r.Context().Value("user").(uint)
	resp := u.Message(true, "success")

	data := models.GetSites(uint(id))
	if data == nil {
		fmt.Println("Database didn't find anything")
		resp = u.Message(false, "unsuccessful")
	}

	resp["data"] = data
	u.Respond(w, resp)
}
