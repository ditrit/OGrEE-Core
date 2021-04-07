package main

import (
	"fmt"
	"p3/app"
	"p3/controllers"

	"net/http"
	"os"
	"regexp"

	"github.com/gorilla/mux"
)

var tmatch mux.MatcherFunc = func(request *http.Request, match *mux.RouteMatch) bool {

	//fmt.Println("The URL is: ", request.URL.String())
	//https://benhoyt.com/writings/go-routing/#regex-table
	//https://stackoverflow.com/questions/21664489/
	//golang-mux-routing-wildcard-custom-func-match

	return regexp.MustCompile(`^(\/api\/user\/tenants\?name=.+)$`).
		MatchString(request.URL.String())
}

func main() {

	router := mux.NewRouter()

	router.HandleFunc("/api/user",
		controllers.CreateAccount).Methods("POST")

	router.HandleFunc("/api/user/login",
		controllers.Authenticate).Methods("POST")

	// ------ TENANTS CRUD ------ //
	router.HandleFunc("/api/user/tenants",
		controllers.GetTenantByName).Methods("GET").MatcherFunc(tmatch)

	router.HandleFunc("/api/user/tenants",
		controllers.GetAllTenants).Methods("GET")

	router.HandleFunc("/api/user/tenants",
		controllers.CreateTenant).Methods("POST")

	router.HandleFunc("/api/user/tenants/{id}",
		controllers.GetTenantFor).Methods("GET")

	router.HandleFunc("/api/user/tenants/{id}",
		controllers.UpdateTenant).Methods("PUT")

	router.HandleFunc("/api/user/tenants/{id}",
		controllers.DeleteTenant).Methods("DELETE")

	// ------ SITES CRUD ------ //
	router.HandleFunc("/api/user/sites",
		controllers.CreateSite).Methods("POST")

	/*router.HandleFunc("/api/user/sites",
		controllers.GetSitesByUserID).Methods("GET")

	router.HandleFunc("/api/user/sites",
		controllers.GetSitesByParentID).Methods("GET")*/

	router.HandleFunc("/api/user/sites",
		controllers.GetAllSites).Methods("GET")

	router.HandleFunc("/api/user/sites/{id}",
		controllers.GetSite).Methods("GET")

	router.HandleFunc("/api/user/sites/{id}",
		controllers.UpdateSite).Methods("PUT")

	router.HandleFunc("/api/user/sites/{id}",
		controllers.DeleteSiteByID).Methods("DELETE")

	router.HandleFunc("/api/user/sites",
		controllers.DeleteSites).Methods("DELETE")

	// ------ BUILDING CRUD ------ //

	router.HandleFunc("/api/user/buildings",
		controllers.CreateBuilding).Methods("POST")

	router.HandleFunc("/api/user/buildings/{id}",
		controllers.UpdateBuilding).Methods("PUT")

	router.HandleFunc("/api/user/buildings/{id}",
		controllers.DeleteBuilding).Methods("DELETE")

	router.HandleFunc("/api/user/buildings/{id}",
		controllers.GetBuilding).Methods("GET")

	router.HandleFunc("/api/user/buildings",
		controllers.GetAllBuildings).Methods("GET")

	// ------ ROOM CRUD ------ //
	router.HandleFunc("/api/user/rooms",
		controllers.CreateRoom).Methods("POST")

	router.HandleFunc("/api/user/rooms/{id}",
		controllers.UpdateRoom).Methods("PUT")

	router.HandleFunc("/api/user/rooms/{id}",
		controllers.DeleteRoom).Methods("DELETE")

	router.HandleFunc("/api/user/rooms/{id}",
		controllers.GetRoom).Methods("GET")

	router.HandleFunc("/api/user/rooms",
		controllers.GetAllRooms).Methods("GET")

	// ------ RACK CRUD ------ //
	router.HandleFunc("/api/user/racks",
		controllers.CreateRack).Methods("POST")

	router.HandleFunc("/api/user/racks",
		controllers.GetAllRacks).Methods("GET")

	router.HandleFunc("/api/user/racks/{id}",
		controllers.UpdateRack).Methods("PUT")

	router.HandleFunc("/api/user/racks/{id}",
		controllers.DeleteRack).Methods("DELETE")

	router.HandleFunc("/api/user/racks/{id}",
		controllers.GetRack).Methods("GET")

	// ------ DEVICE CRUD ------ //
	router.HandleFunc("/api/user/devices",
		controllers.CreateDevice).Methods("POST")

	router.HandleFunc("/api/user/devices/{id}",
		controllers.UpdateDevice).Methods("PUT")

	router.HandleFunc("/api/user/devices/{id}",
		controllers.DeleteDevice).Methods("DELETE")

	router.HandleFunc("/api/user/devices/{id}",
		controllers.GetDevice).Methods("GET")

	router.HandleFunc("/api/user/devices",
		controllers.GetAllDevices).Methods("GET")

	//Attach JWT auth middleware
	router.Use(app.JwtAuthentication)

	//Get port from .env file, no port was specified
	//So this should return an empty string when
	//tested locally
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001" //localhost
	}

	fmt.Println(port)

	//Start app, localhost:8000/api
	err := http.ListenAndServe(":"+port, router)
	if err != nil {
		fmt.Print(err)
	}
}

//https://medium.com/@adigunhammedolalekan/build-and-deploy-a-secure-rest-api-with-go-postgresql-jwt-and-gorm-6fadf3da505b
