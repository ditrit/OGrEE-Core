package main

import (
	"fmt"
	"p3/app"
	"p3/controllers"

	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {

	router := mux.NewRouter()

	router.HandleFunc("/api/user/new",
		controllers.CreateAccount).Methods("POST")

	router.HandleFunc("/api/user/login",
		controllers.Authenticate).Methods("POST")

	router.HandleFunc("/api/user/tenants/all",
		controllers.GetAllTenants).Methods("GET")

	router.HandleFunc("/api/user/tenants/",
		controllers.CreateTenant).Methods("POST")

	router.HandleFunc("/api/user/tenants/{id}",
		controllers.GetTenantFor).Methods("GET")

	router.HandleFunc("/api/user/tenants/{id}",
		controllers.UpdateTenant).Methods("PUT")

	router.HandleFunc("/api/user/tenants/{id}",
		controllers.DeleteTenant).Methods("DELETE")

	router.HandleFunc("/api/user/sites/",
		controllers.CreateSite).Methods("POST")

	router.HandleFunc("/api/user/sites/",
		controllers.GetSitesByUserID).Methods("GET")

	router.HandleFunc("/api/user/sites/all",
		controllers.GetSitesByParentID).Methods("GET")

	router.HandleFunc("/api/user/sites/{id}",
		controllers.GetSite).Methods("GET")

	router.HandleFunc("/api/user/sites/{id}",
		controllers.UpdateSite).Methods("PUT")

	router.HandleFunc("/api/user/sites/{id}",
		controllers.DeleteSiteByID).Methods("DELETE")

	router.HandleFunc("/api/user/sites/",
		controllers.DeleteSites).Methods("DELETE")

	// ------ BUILDING CRUD ------ //

	router.HandleFunc("/api/user/buildings/",
		controllers.CreateBuilding).Methods("POST")

	router.HandleFunc("/api/user/buildings/{id}",
		controllers.UpdateBuilding).Methods("PUT")

	router.HandleFunc("/api/user/buildings/{id}",
		controllers.DeleteBuilding).Methods("DELETE")

	router.HandleFunc("/api/user/buildings/{id}",
		controllers.GetBuilding).Methods("GET")

	/*router.HandleFunc("/api/user/sites/single",
	controllers.GetSite).Methods("GET")*/

	/*router.HandleFunc("/api/me/contacts",
		controllers.GetContactsFor).Methods("GET")

	router.HandleFunc("/api/contacts/new",
		controllers.CreateContact).Methods("POST")*/

	//Attach JWT auth middleware
	router.Use(app.JwtAuthentication)

	//Get port from .env file, no port was specified
	//So this should return an empty string when
	//tested locally
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000" //localhost
	}

	fmt.Println(port)

	//Start app, localhost:8000/api
	err := http.ListenAndServe(":"+port, router)
	if err != nil {
		fmt.Print(err)
	}
}

//https://medium.com/@adigunhammedolalekan/build-and-deploy-a-secure-rest-api-with-go-postgresql-jwt-and-gorm-6fadf3da505b
