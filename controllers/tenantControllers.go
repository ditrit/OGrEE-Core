package controllers

import (
	"encoding/json"
	"net/http"
	"p3/models"
	u "p3/utils"
)

var CreateTenant = func(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(uint)
	tenant := &models.Tenant{}

	err := json.NewDecoder(r.Body).Decode(tenant)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		return
	}
	tenant.ID = user
	resp := tenant.Create()
	u.Respond(w, resp)
}

var GetTenantFor = func(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value("user").(uint)

	data := models.GetTenant(uint(id))
	resp := u.Message(true, "success")
	resp["data"] = data
	u.Respond(w, resp)
}
