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

//Obtain by query
var dmatch mux.MatcherFunc = func(request *http.Request, match *mux.RouteMatch) bool {

	//fmt.Println("The URL is: ", request.URL.String())
	//https://benhoyt.com/writings/go-routing/#regex-table
	//https://stackoverflow.com/questions/21664489/
	//golang-mux-routing-wildcard-custom-func-match
	println("Checking MATCH")
	return regexp.MustCompile(`^(\/api\/(tenants|sites|buildings|rooms|rooms\/acs|rooms\/panels|rooms\/walls|rooms\/cabinets|rooms\/aisles|rooms\/tiles|rooms\/groups|rooms\/corridors|racks|devices|racks\/racksensors|devices\/devicesensors|(room|obj)-templates)\?.*)$`).
		MatchString(request.URL.String())
}

//Obtain object hierarchy
var hmatch mux.MatcherFunc = func(request *http.Request, match *mux.RouteMatch) bool {
	println("CHECKING H-MATCH")
	return regexp.MustCompile(`(^(\/api\/(tenants|sites|buildings|rooms|rooms|racks|devices)\/[a-zA-Z0-9]{24}\/all)(\/(tenants|sites|buildings|rooms|rooms|racks|devices))*$)|(^(\/api\/(devices)\/[a-zA-Z0-9]{24}\/all)(\?limit=[0-9]+)*$)`).
		MatchString(request.URL.String())
}

//For Obtaining objects using parent
var pmatch mux.MatcherFunc = func(request *http.Request, match *mux.RouteMatch) bool {
	println("CHECKING P-MATCH")
	return regexp.MustCompile(`^(\/api\/(tenants|sites|buildings|rooms|rooms|racks|devices)\/[a-zA-Z0-9]{24}(\/.*)+)$`).
		MatchString(request.URL.String())
}

//For Obtaining Tenant hierarchy
var tmatch mux.MatcherFunc = func(request *http.Request, match *mux.RouteMatch) bool {
	println("CHECKING T-MATCH")
	return regexp.MustCompile(`^(\/api\/(tenants)(\/[a-zA-Z]+)(\/.*)+)$`).
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

	// ------ GET ------ //
	//GET ENTITY HIERARCHY
	//This matches ranged Tenant Hierarchy
	router.NewRoute().PathPrefix("/api/tenants/{tenant_name}/all").
		MatcherFunc(tmatch).HandlerFunc(controllers.GetTenantHierarchy)

	router.NewRoute().PathPrefix("/api/{entity}/{id:[a-zA-Z0-9]{24}}/all").
		MatcherFunc(hmatch).HandlerFunc(controllers.GetEntityHierarchy)

	//GET EXCEPTIONS
	router.HandleFunc("/api/tenants/{tenant_name}/buildings",
		controllers.GetEntitiesOfAncestor).Methods("GET")

	router.HandleFunc("/api/sites/{id:[a-zA-Z0-9]{24}}/rooms",
		controllers.GetEntitiesOfAncestor).Methods("GET")

	router.HandleFunc("/api/buildings/{id:[a-zA-Z0-9]{24}}/racks",
		controllers.GetEntitiesOfAncestor).Methods("GET")

	router.HandleFunc("/api/rooms/{id:[a-zA-Z0-9]{24}}/devices",
		controllers.GetEntitiesOfAncestor).Methods("GET")

	router.HandleFunc("/api/{entity:racks}/{id:[a-zA-Z0-9]{24}}/{subent:racksensors}",
		controllers.GetAllEntities).Methods("GET")

	router.HandleFunc("/api/{entity:devices}/{id:[a-zA-Z0-9]{24}}/{subent:devicesensors}",
		controllers.GetAllEntities).Methods("GET")

	router.HandleFunc("/api/{entity:rooms}/{id:[a-zA-Z0-9]{24}}/{subent:acs|walls|panels|cabinets|tiles|aisles|groups|corridors}",
		controllers.GetAllEntities).Methods("GET")

	// GET BY QUERY
	router.NewRoute().PathPrefix("/api/{entity:[a-z]+}/{subent:[a-z]+}").MatcherFunc(dmatch).
		HandlerFunc(controllers.GetEntityByQuery).Methods("GET")

	router.NewRoute().PathPrefix("/api/{entity:[a-z]+}").MatcherFunc(dmatch).
		HandlerFunc(controllers.GetEntityByQuery).Methods("GET")

	//GET ENTITY
	router.HandleFunc("/api/{entity}/{id:[a-zA-Z0-9]{24}}/{subent:[a-zA-Z0-9]+}/{nest:[a-zA-Z0-9]{24}}",
		controllers.GetEntity).Methods("GET")

	router.HandleFunc("/api/{entity}/{id:[a-zA-Z0-9]{24}}",
		controllers.GetEntity).Methods("GET")

	router.HandleFunc("/api/{entity}/{name}",
		controllers.GetEntity).Methods("GET")

	//GET BY NAME OF PARENT
	router.NewRoute().PathPrefix("/api/tenants/{tenant_name}").
		MatcherFunc(tmatch).HandlerFunc(controllers.GetEntitiesUsingNamesOfParents).Methods("GET")

	router.NewRoute().PathPrefix("/api/{entity}/{id:[a-zA-Z0-9]{24}}").
		MatcherFunc(pmatch).HandlerFunc(controllers.GetEntitiesUsingNamesOfParents).Methods("GET")

	// GET ALL ENTITY
	router.HandleFunc("/api/{entity}/{id:[a-zA-Z0-9]{24}/{subent:[a-z]+}",
		controllers.GetAllEntities).Methods("GET")

	router.HandleFunc("/api/{entity}",
		controllers.GetAllEntities).Methods("GET")

	//GET ALL NONSTD
	router.HandleFunc("/api/tenants/{tenant_name}/all/nonstd",
		controllers.GetEntityHierarchyNonStd).Methods("GET")

	router.HandleFunc("/api/{entity}/{id:[a-zA-Z0-9]{24}}/all/nonstd",
		controllers.GetEntityHierarchyNonStd).Methods("GET")

	// CREATE ENTITY
	router.HandleFunc("/api/{entity}",
		controllers.CreateEntity).Methods("POST")

	//DELETE ENTITY
	router.HandleFunc("/api/{entity}/{id:[a-zA-Z0-9]{24}}/{subent:[a-z]+}/{nest}",
		controllers.DeleteEntity).Methods("DELETE")

	router.HandleFunc("/api/{entity}/{id:[a-zA-Z0-9]{24}}",
		controllers.DeleteEntity).Methods("DELETE")

	router.HandleFunc("/api/{entity}/{name}",
		controllers.DeleteEntity).Methods("DELETE")

	// UPDATE ENTITY
	router.HandleFunc("/api/{entity}/{id:[a-zA-Z0-9]{24}}/{subent}/{nest}",
		controllers.UpdateEntity).Methods("PUT", "PATCH")

	router.HandleFunc("/api/{entity}/{id:[a-zA-Z0-9]{24}}",
		controllers.UpdateEntity).Methods("PUT", "PATCH")

	router.HandleFunc("/api/{entity}/{name}",
		controllers.UpdateEntity).Methods("PUT", "PATCH")

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
