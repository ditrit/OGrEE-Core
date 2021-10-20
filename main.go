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
	return regexp.MustCompile(`^(\/api\/(tenants|sites|buildings|rooms|rooms\/acs|rooms\/panels|rooms\/walls|racks|devices|subdevices|subdevices1|(room|rack|device)-templates)\?.*)$`).
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
	router.HandleFunc("/api/tenants",
		controllers.GetEntityByQuery).Methods("GET").MatcherFunc(dmatch)

	router.HandleFunc("/api/tenants",
		controllers.GetAllEntities).Methods("GET")

	router.HandleFunc("/api/tenants",
		controllers.CreateEntity).Methods("POST")

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

	router.HandleFunc("/api/tenants/{id}",
		controllers.GetEntity).Methods("GET")

	router.HandleFunc("/api/tenants/{id}",
		controllers.UpdateEntity).Methods("PUT")

	router.HandleFunc("/api/tenants/{id}",
		controllers.DeleteEntity).Methods("DELETE")

	// ------ SITES CRUD ------ //
	router.HandleFunc("/api/sites",
		controllers.GetEntityByQuery).Methods("GET").MatcherFunc(dmatch)

	router.HandleFunc("/api/sites",
		controllers.CreateEntity).Methods("POST")

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

	router.HandleFunc("/api/sites",
		controllers.GetAllEntities).Methods("GET")

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

	router.HandleFunc("/api/sites/{id}",
		controllers.GetEntity).Methods("GET")

	router.HandleFunc("/api/sites/{id}",
		controllers.UpdateEntity).Methods("PUT")

	router.HandleFunc("/api/sites/{id}",
		controllers.DeleteEntity).Methods("DELETE")

	router.HandleFunc("/api/sites",
		controllers.DeleteEntity).Methods("DELETE")

	// ------ BUILDING CRUD ------ //
	router.HandleFunc("/api/buildings",
		controllers.GetEntityByQuery).Methods("GET").MatcherFunc(dmatch)

	router.HandleFunc("/api/buildings",
		controllers.CreateEntity).Methods("POST")

	router.HandleFunc("/api/buildings/{id}",
		controllers.UpdateEntity).Methods("PUT")

	router.HandleFunc("/api/buildings/{id}",
		controllers.DeleteEntity).Methods("DELETE")

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

	router.HandleFunc("/api/buildings/{id}",
		controllers.GetEntity).Methods("GET")

	router.HandleFunc("/api/buildings",
		controllers.GetAllEntities).Methods("GET")

	// ------ ROOM CRUD ------ //
	router.HandleFunc("/api/rooms/acs",
		controllers.GetEntityByQuery).Methods("GET").MatcherFunc(dmatch)

	router.HandleFunc("/api/rooms/walls",
		controllers.GetEntityByQuery).Methods("GET").MatcherFunc(dmatch)

	router.HandleFunc("/api/rooms/panels",
		controllers.GetEntityByQuery).Methods("GET").MatcherFunc(dmatch)

	router.HandleFunc("/api/rooms",
		controllers.GetEntityByQuery).Methods("GET").MatcherFunc(dmatch)

	router.HandleFunc("/api/rooms",
		controllers.CreateEntity).Methods("POST")

	router.HandleFunc("/api/rooms/{id}/acs/{nest}",
		controllers.UpdateEntity).Methods("PUT")

	router.HandleFunc("/api/rooms/{id}/panels/{nest}",
		controllers.UpdateEntity).Methods("PUT")

	router.HandleFunc("/api/rooms/{id}/walls/{nest}",
		controllers.UpdateEntity).Methods("PUT")

	router.HandleFunc("/api/rooms/{id}",
		controllers.UpdateEntity).Methods("PUT")

	router.HandleFunc("/api/rooms/{id}/walls/{nest}",
		controllers.DeleteEntity).Methods("DELETE")

	router.HandleFunc("/api/rooms/{id}/acs/{nest}",
		controllers.DeleteEntity).Methods("DELETE")

	router.HandleFunc("/api/rooms/{id}/panels/{nest}",
		controllers.DeleteEntity).Methods("DELETE")

	router.HandleFunc("/api/rooms/{id}",
		controllers.DeleteEntity).Methods("DELETE")

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

	router.HandleFunc("/api/rooms/{id}/walls",
		controllers.GetAllEntities).Methods("GET")

	router.HandleFunc("/api/rooms/{id}/acs",
		controllers.GetAllEntities).Methods("GET")

	router.HandleFunc("/api/rooms/{id}/panels",
		controllers.GetAllEntities).Methods("GET")

	router.HandleFunc("/api/rooms/{id}/acs/{nest}",
		controllers.GetEntity).Methods("GET")

	router.HandleFunc("/api/rooms/{id}/walls/{nest}",
		controllers.GetEntity).Methods("GET")

	router.HandleFunc("/api/rooms/{id}/panels/{nest}",
		controllers.GetEntity).Methods("GET")

	router.HandleFunc("/api/rooms/{id}",
		controllers.GetEntity).Methods("GET")

	router.HandleFunc("/api/rooms",
		controllers.GetAllEntities).Methods("GET")

	// ------ RACK CRUD ------ //
	router.HandleFunc("/api/racks",
		controllers.GetEntityByQuery).Methods("GET").MatcherFunc(dmatch)

	router.HandleFunc("/api/racks",
		controllers.CreateEntity).Methods("POST")

	router.HandleFunc("/api/racks",
		controllers.GetAllEntities).Methods("GET")

	router.HandleFunc("/api/racks/{id}",
		controllers.UpdateEntity).Methods("PUT")

	router.HandleFunc("/api/racks/{id}",
		controllers.DeleteEntity).Methods("DELETE")

	router.HandleFunc("/api/racks/{id}/devices/{device_name}",
		controllers.GetEntitiesUsingNamesOfParents).Methods("GET")

	router.HandleFunc("/api/racks/{id}/all",
		controllers.GetEntityHierarchy).Methods("GET")

	router.HandleFunc("/api/racks/{id}/all/nonstd",
		controllers.GetEntityHierarchyNonStd).Methods("GET")

	router.HandleFunc("/api/racks/{id}",
		controllers.GetEntity).Methods("GET")

	// ------ DEVICE CRUD ------ //
	router.HandleFunc("/api/devices",
		controllers.GetEntityByQuery).Methods("GET").MatcherFunc(dmatch)

	router.HandleFunc("/api/devices",
		controllers.CreateEntity).Methods("POST")

	router.HandleFunc("/api/devices/{id}",
		controllers.UpdateEntity).Methods("PUT")

	router.HandleFunc("/api/devices/{id}",
		controllers.DeleteEntity).Methods("DELETE")

	router.HandleFunc("/api/devices/{id}",
		controllers.GetEntity).Methods("GET")

	router.HandleFunc("/api/devices",
		controllers.GetAllEntities).Methods("GET")

	// ------ TEMPLATE CRUD ------ //
	router.HandleFunc("/api/room-templates",
		controllers.GetEntityByQuery).Methods("GET").MatcherFunc(dmatch)

	router.HandleFunc("/api/obj-templates",
		controllers.GetEntityByQuery).Methods("GET").MatcherFunc(dmatch)

	router.HandleFunc("/api/room-templates",
		controllers.CreateEntity).Methods("POST")

	router.HandleFunc("/api/obj-templates",
		controllers.CreateEntity).Methods("POST")

	router.HandleFunc("/api/room-templates/{name}",
		controllers.GetEntity).Methods("GET")

	router.HandleFunc("/api/obj-templates/{name}",
		controllers.GetEntity).Methods("GET")

	router.HandleFunc("/api/room-templates",
		controllers.GetAllEntities).Methods("GET")

	router.HandleFunc("/api/obj-templates",
		controllers.GetAllEntities).Methods("GET")

	router.HandleFunc("/api/room-templates/{name}",
		controllers.DeleteEntity).Methods("DELETE")

	router.HandleFunc("/api/obj-templates/{name}",
		controllers.DeleteEntity).Methods("DELETE")

	router.HandleFunc("/api/room-templates/{name}",
		controllers.UpdateEntity).Methods("PUT")

	router.HandleFunc("/api/obj-templates/{name}",
		controllers.UpdateEntity).Methods("PUT")

	// ------ AC/PWR/WALL CRUD ------ //
	router.HandleFunc("/api/acs",
		controllers.CreateEntity).Methods("POST")

	router.HandleFunc("/api/panels",
		controllers.CreateEntity).Methods("POST")

	router.HandleFunc("/api/walls",
		controllers.CreateEntity).Methods("POST")

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
