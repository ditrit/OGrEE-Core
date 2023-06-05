package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"p3/models"
	u "p3/utils"

	"github.com/gorilla/mux"
)

var GetProjects = func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 GetProjects ")
	fmt.Println("******************************************************")
	var resp map[string]interface{}

	query, _ := url.ParseQuery(r.URL.RawQuery)

	if len(query["user"]) <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		resp = u.Message(false, "Error: user should be sent as query param")
		u.Respond(w, resp)
		return
	}

	data, err := models.GetProjectsByUserEmail(query["user"][0])
	if err != "" {
		w.WriteHeader(http.StatusNotFound)
		resp = u.Message(false, "Error: "+err)
	} else {
		if r.Method == "OPTIONS" {
			w.Header().Add("Content-Type", "application/json")
			w.Header().Add("Allow", "GET, OPTIONS, HEAD")
		} else {
			resp = u.Message(true, "successfully got projects")
			resp["data"] = data
		}
	}

	u.Respond(w, resp)
}

var CreateOrUpdateProject = func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 CreateOrUpdateProject ")
	fmt.Println("******************************************************")
	var resp map[string]interface{}

	project := &models.Project{}
	err := json.NewDecoder(r.Body).Decode(project)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message(false, "Invalid request"))
		return
	}

	var errStr string
	if r.Method == "POST" {
		// Create project
		errStr = models.AddProject(*project)
	} else {
		// Update project
		errStr = models.UpdateProject(*project, mux.Vars(r)["id"])
	}

	if errStr != "" {
		w.WriteHeader(http.StatusNotFound)
		resp = u.Message(false, "Error: "+errStr)
	} else {
		if r.Method == "OPTIONS" {
			w.Header().Add("Content-Type", "application/json")
			w.Header().Add("Allow", "GET, OPTIONS, HEAD")
		} else {
			resp = u.Message(true, "successfully handled project request")
			resp["data"] = project
		}
	}

	u.Respond(w, resp)
}

var DeleteProject = func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 DeleteProject ")
	fmt.Println("******************************************************")
	var resp map[string]interface{}

	errStr := models.DeleteProject(mux.Vars(r)["id"])

	if errStr != "" {
		w.WriteHeader(http.StatusNotFound)
		resp = u.Message(false, "Error: "+errStr)
	} else {
		if r.Method == "OPTIONS" {
			w.Header().Add("Content-Type", "application/json")
			w.Header().Add("Allow", "GET, OPTIONS, HEAD")
		} else {
			resp = u.Message(true, "successfully removed project request")
		}
	}

	u.Respond(w, resp)
}
