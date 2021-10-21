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
	return regexp.MustCompile(`^(\/api\/(tenants|sites|buildings|rooms|rooms\/acs|rooms\/panels|rooms\/walls|racks|devices|subdevices|subdevices1|(room|obj)-templates)\?.*)$`).
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

	router.HandleFunc("/api/tenants/{tenant_name}/sites",
		controllers.GetEntitiesOfParent).Methods("GET")

	router.HandleFunc("/api/tenants/{tenant_name}/all/sites/buildings/rooms/racks/devices",
		controllers.GetTenantHierarchy).Methods("GET")

	router.HandleFunc("/api/tenants/{tenant_name}/all/sites/buildings/rooms/racks",
		controllers.GetTenantHierarchy).Methods("GET")

	router.HandleFunc("/api/tenants/{tenant_name}/all/sites/buildings/rooms",
		controllers.GetTenantHierarchy).Methods("GET")

	router.HandleFunc("/api/tenants/{tenant_name}/all/sites/buildings",
		controllers.GetTenantHierarchy).Methods("GET")

	router.HandleFunc("/api/tenants/{tenant_name}/all/nonstd",
		controllers.GetEntityHierarchyNonStd).Methods("GET")

	router.HandleFunc("/api/tenants/{tenant_name}/all",
		controllers.GetTenantHierarchy).Methods("GET")

	// ------ SITES CRUD ------ //

	router.HandleFunc("/api/sites/{id}/all/nonstd",
		controllers.GetEntityHierarchyNonStd).Methods("GET")

	router.HandleFunc("/api/sites/{id}/all/buildings/rooms/racks/devices",
		controllers.GetEntityHierarchy).Methods("GET")

	router.HandleFunc("/api/sites/{id}/all/buildings/rooms/racks",
		controllers.GetEntityHierarchy).Methods("GET")

	router.HandleFunc("/api/sites/{id}/all/buildings/rooms",
		controllers.GetEntityHierarchy).Methods("GET")

	router.HandleFunc("/api/sites/{id}/all",
		controllers.GetEntityHierarchy).Methods("GET")

	router.HandleFunc("/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}",
		controllers.GetEntitiesUsingNamesOfParents).Methods("GET")

	router.HandleFunc("/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices",
		controllers.GetEntitiesUsingNamesOfParents).Methods("GET")

	router.HandleFunc("/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}",
		controllers.GetEntitiesUsingNamesOfParents).Methods("GET")

	router.HandleFunc("/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks",
		controllers.GetEntitiesUsingNamesOfParents).Methods("GET")

	router.HandleFunc("/api/sites/{id}/buildings/{building_name}/rooms/{room_name}",
		controllers.GetEntitiesUsingNamesOfParents).Methods("GET")

	router.HandleFunc("/api/sites/{id}/buildings/{building_name}/rooms",
		controllers.GetEntitiesUsingNamesOfParents).Methods("GET")

	router.HandleFunc("/api/sites/{id}/buildings/{building_name}",
		controllers.GetEntitiesUsingNamesOfParents).Methods("GET")

	router.HandleFunc("/api/sites/{id}/buildings",
		controllers.GetEntitiesOfParent).Methods("GET")

	// ------ BUILDING CRUD ------ //

	router.HandleFunc("/api/buildings/{id}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}",
		controllers.GetEntitiesUsingNamesOfParents).Methods("GET")

	router.HandleFunc("/api/buildings/{id}/rooms/{room_name}/racks/{rack_name}/devices",
		controllers.GetEntitiesUsingNamesOfParents).Methods("GET")

	router.HandleFunc("/api/buildings/{id}/rooms/{room_name}/racks/{rack_name}",
		controllers.GetEntitiesUsingNamesOfParents).Methods("GET")

	router.HandleFunc("/api/buildings/{id}/rooms/{room_name}/racks",
		controllers.GetEntitiesUsingNamesOfParents).Methods("GET")

	router.HandleFunc("/api/buildings/{id}/rooms/{room_name}",
		controllers.GetEntitiesUsingNamesOfParents).Methods("GET")

	router.HandleFunc("/api/buildings/{id}/rooms",
		controllers.GetEntitiesOfParent).Methods("GET")

	router.HandleFunc("/api/buildings/{id}/all/nonstd",
		controllers.GetEntityHierarchyNonStd).Methods("GET")

	router.HandleFunc("/api/buildings/{id}/all/rooms/racks/devices",
		controllers.GetEntityHierarchy).Methods("GET")

	router.HandleFunc("/api/buildings/{id}/all/rooms/racks",
		controllers.GetEntityHierarchy).Methods("GET")

	router.HandleFunc("/api/buildings/{id}/all",
		controllers.GetEntityHierarchy).Methods("GET")

	// ------ ROOM CRUD ------ //

	router.HandleFunc("/api/rooms/{id}/racks/{rack_name}/devices/{device_name}",
		controllers.GetEntitiesUsingNamesOfParents).Methods("GET")

	router.HandleFunc("/api/rooms/{id}/racks/{rack_name}/devices",
		controllers.GetEntitiesUsingNamesOfParents).Methods("GET")

	router.HandleFunc("/api/rooms/{id}/racks/{rack_name}",
		controllers.GetEntitiesUsingNamesOfParents).Methods("GET")

	router.HandleFunc("/api/rooms/{id}/racks",
		controllers.GetEntitiesOfParent).Methods("GET")

	router.HandleFunc("/api/rooms/{id}/all/racks/devices",
		controllers.GetEntityHierarchy).Methods("GET")

	router.HandleFunc("/api/rooms/{id}/all",
		controllers.GetEntityHierarchy).Methods("GET")

	router.HandleFunc("/api/rooms/{id}/all/nonstd",
		controllers.GetEntityHierarchyNonStd).Methods("GET")

	// ------ RACK CRUD ------ //

	router.HandleFunc("/api/racks/{id}/devices/{device_name}",
		controllers.GetEntitiesUsingNamesOfParents).Methods("GET")

	router.HandleFunc("/api/racks/{id}/all",
		controllers.GetEntityHierarchy).Methods("GET")

	router.HandleFunc("/api/racks/{id}/all/nonstd",
		controllers.GetEntityHierarchyNonStd).Methods("GET")

	// ------ DEVICE CRUD ------ //

	// ------ TEMPLATE CRUD ------ //

	// ------ GET ------ //
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
