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

func GetProjects(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 GetProjects ")
	fmt.Println("******************************************************")
	var resp map[string]interface{}

	query, _ := url.ParseQuery(r.URL.RawQuery)

	if len(query["user"]) <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		resp = u.Message("Error: user should be sent as query param")
		u.Respond(w, resp)
		return
	}

	projects, err := models.GetProjectsByUserEmail(query["user"][0])
	if err != nil {
		u.RespondWithError(w, err)
	} else {
		if r.Method == "OPTIONS" {
			w.Header().Add("Content-Type", "application/json")
			w.Header().Add("Allow", "GET, OPTIONS, HEAD")
		} else {
			resp["projects"] = projects
			u.Respond(w, u.RespDataWrapper("successfully got projects", resp))
		}
	}

	u.Respond(w, resp)
}

func CreateOrUpdateProject(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 CreateOrUpdateProject ")
	fmt.Println("******************************************************")

	project := &models.Project{}
	err := json.NewDecoder(r.Body).Decode(project)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message("Invalid request"))
		return
	}

	var modelErr *u.Error
	if r.Method == "POST" {
		// Create project
		modelErr = models.AddProject(*project)
	} else {
		// Update project
		modelErr = models.UpdateProject(*project, mux.Vars(r)["id"])
	}

	if modelErr != nil {
		u.RespondWithError(w, modelErr)
	} else {
		if r.Method == "OPTIONS" {
			w.Header().Add("Content-Type", "application/json")
			w.Header().Add("Allow", "GET, OPTIONS, HEAD")
		} else {
			resp := u.Message("successfully handled project request")
			resp["data"] = project
			u.Respond(w, resp)
		}
	}
}

func DeleteProject(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 DeleteProject ")
	fmt.Println("******************************************************")

	err := models.DeleteProject(mux.Vars(r)["id"])

	if err != nil {
		u.RespondWithError(w, err)
	} else {
		if r.Method == "OPTIONS" {
			w.Header().Add("Content-Type", "application/json")
			w.Header().Add("Allow", "GET, OPTIONS, HEAD")
		} else {
			u.Respond(w, u.Message("successfully removed project"))
		}
	}
}
