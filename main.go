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

var dmatch mux.MatcherFunc = func(request *http.Request, match *mux.RouteMatch) bool {

	//fmt.Println("The URL is: ", request.URL.String())
	//https://benhoyt.com/writings/go-routing/#regex-table
	//https://stackoverflow.com/questions/21664489/
	//golang-mux-routing-wildcard-custom-func-match
	println("Checking MATCH")
	return regexp.MustCompile(`^(\/api\/(tenants|sites|buildings|rooms|rooms\/acs|rooms\/panels|rooms\/walls|racks|devices(room|obj)-templates)\?.*)$`).
		MatchString(request.URL.String())
}

var hmatch mux.MatcherFunc = func(request *http.Request, match *mux.RouteMatch) bool {
	println("CHECKING H-MATCH")
	return regexp.MustCompile(`^(\/api\/(tenants|sites|buildings|rooms|rooms|racks|devices)\/[a-zA-Z0-9]{24}\/all(\/.*)+)$`).
		MatchString(request.URL.String())
}

var pmatch mux.MatcherFunc = func(request *http.Request, match *mux.RouteMatch) bool {
	println("CHECKING P-MATCH")
	return regexp.MustCompile(`^(\/api\/(tenants|sites|buildings|rooms|rooms|racks|devices)\/[a-zA-Z0-9]{24}(\/.*)+)$`).
		MatchString(request.URL.String())
}

func main() {

	router := mux.NewRouter()

	router.HandleFunc("/api",
		controllers.CreateAccount).Methods("POST")

	router.HandleFunc("/api/login",
		controllers.Authenticate).Methods("POST")

	router.HandleFunc("/api/token/valid",
		controllers.Verify).Methods("GET")

	// ------ TENANTS CRUD ------ //

	router.HandleFunc("/api/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}",
		controllers.GetEntitiesUsingNameOfTenant).Methods("GET")

	router.HandleFunc("/api/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices",
		controllers.GetEntitiesUsingNameOfTenant).Methods("GET")

	router.HandleFunc("/api/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}",
		controllers.GetEntitiesUsingNameOfTenant).Methods("GET")

	router.HandleFunc("/api/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms/{room_name}/racks",
		controllers.GetEntitiesUsingNameOfTenant).Methods("GET")

	router.HandleFunc("/api/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms/{room_name}",
		controllers.GetEntitiesUsingNameOfTenant).Methods("GET")

	router.HandleFunc("/api/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms",
		controllers.GetEntitiesUsingNameOfTenant).Methods("GET")

	router.HandleFunc("/api/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}",
		controllers.GetEntitiesUsingNameOfTenant).Methods("GET")

	router.HandleFunc("/api/tenants/{tenant_name}/sites/{site_name}/buildings",
		controllers.GetEntitiesUsingNameOfTenant).Methods("GET")

	router.HandleFunc("/api/tenants/{tenant_name}/sites/{site_name}",
		controllers.GetEntitiesUsingNameOfTenant).Methods("GET")

	router.HandleFunc("/api/tenants/{tenant_name}/all/sites/buildings/rooms/racks/devices",
		controllers.GetTenantHierarchy).Methods("GET")

	router.HandleFunc("/api/tenants/{tenant_name}/all/sites/buildings/rooms/racks",
		controllers.GetTenantHierarchy).Methods("GET")

	router.HandleFunc("/api/tenants/{tenant_name}/all/sites/buildings/rooms",
		controllers.GetTenantHierarchy).Methods("GET")

	/*router.HandleFunc("/api/tenants/{tenant_name}/all/sites/buildings",
	controllers.GetTenantHierarchy).Methods("GET")*/

	// ------ GET ------ //
	//GET ENTITY HIERARCHY
	router.NewRoute().PathPrefix("/api/tenants/{tenant_name}/all").
		MatcherFunc(hmatch).HandlerFunc(controllers.GetTenantHierarchy)

	router.NewRoute().PathPrefix("/api/{entity}/{id:[a-zA-Z0-9]{24}}/all").
		MatcherFunc(hmatch).HandlerFunc(controllers.GetEntityHierarchy)

	router.HandleFunc("/api/{entity}/{id:[a-zA-Z0-9]{24}}/all",
		controllers.GetEntityHierarchy).Methods("GET")

	router.HandleFunc("/api/tenants/{tenant_name}/all",
		controllers.GetTenantHierarchy).Methods("GET")

	//GET BY NAME OF PARENT
	router.NewRoute().PathPrefix("/api/{entity}/{id:[a-zA-Z0-9]{24}}").
		MatcherFunc(pmatch).HandlerFunc(controllers.GetEntitiesUsingNamesOfParents)

	// GET BY QUERY
	router.HandleFunc("/api/{entity:[a-z]+}",
		controllers.GetEntityByQuery).Methods("GET").MatcherFunc(dmatch)

	router.HandleFunc("/api/{entity}/{subent}",
		controllers.GetEntityByQuery).Methods("GET").MatcherFunc(dmatch)

	//GET ENTITY

	router.HandleFunc("/api/{entity}/{id:[a-zA-Z0-9]{24}}/{subent}/{nest}",
		controllers.GetEntity).Methods("GET")

	router.HandleFunc("/api/{entity}/{id:[a-zA-Z0-9]{24}}",
		controllers.GetEntity).Methods("GET")

	router.HandleFunc("/api/{entity}/{name}",
		controllers.GetEntity).Methods("GET")

	// GET IMMEDIATE CHILDREN
	router.HandleFunc("/api/tenants/{tenant_name}/sites",
		controllers.GetEntitiesOfParent).Methods("GET")

	router.HandleFunc("/api/{entity}/{id:[a-zA-Z0-9]{24}}/{child}",
		controllers.GetEntitiesOfParent).Methods("GET")

	//GET ALL NONSTD
	router.HandleFunc("/api/tenants/{tenant_name}/all/nonstd",
		controllers.GetEntityHierarchyNonStd).Methods("GET")

	router.HandleFunc("/api/{entity}/{id:[a-zA-Z0-9]{24}}/all/nonstd",
		controllers.GetEntityHierarchyNonStd).Methods("GET")

	// GET ALL ENTITY
	router.HandleFunc("/api/{entity}",
		controllers.GetAllEntities).Methods("GET")

	router.HandleFunc("/api/{entity}/{id:[a-zA-Z0-9]{24}/{subent}",
		controllers.GetAllEntities).Methods("GET")

	// CREATE ENTITY
	router.HandleFunc("/api/{entity}",
		controllers.CreateEntity).Methods("POST")

	//DELETE ENTITY
	router.HandleFunc("/api/{entity}/{id:[a-zA-Z0-9]{24}/{subent}/{nest}",
		controllers.DeleteEntity).Methods("DELETE")

	router.HandleFunc("/api/{entity}/{id:[a-zA-Z0-9]{24}}",
		controllers.DeleteEntity).Methods("DELETE")

	router.HandleFunc("/api/{entity}/{name}",
		controllers.DeleteEntity).Methods("DELETE")

	// UPDATE ENTITY
	router.HandleFunc("/api/{entity}/{id:[a-zA-Z0-9]{24}}/{subent}/{nest}",
		controllers.UpdateEntity).Methods("PUT")

	router.HandleFunc("/api/{entity}/{id:[a-zA-Z0-9]{24}}",
		controllers.UpdateEntity).Methods("PUT")

	router.HandleFunc("/api/{entity}/{name}",
		controllers.UpdateEntity).Methods("PUT")

	//Attach JWT auth middleware
	//router.Use(app.Log)
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
