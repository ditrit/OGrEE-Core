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

	return regexp.MustCompile(`^(\/api\/user\/tenants\?.*)$`).
		MatchString(request.URL.String())
}

var smatch mux.MatcherFunc = func(request *http.Request, match *mux.RouteMatch) bool {

	return regexp.MustCompile(`^(\/api\/user\/sites\?.*)$`).
		MatchString(request.URL.String())
}

var bmatch mux.MatcherFunc = func(request *http.Request, match *mux.RouteMatch) bool {

	return regexp.MustCompile(`^(\/api\/user\/buildings\?.*)$`).
		MatchString(request.URL.String())
}

var rmatch mux.MatcherFunc = func(request *http.Request, match *mux.RouteMatch) bool {

	return regexp.MustCompile(`^(\/api\/user\/rooms\?.*)$`).
		MatchString(request.URL.String())
}

var ramatch mux.MatcherFunc = func(request *http.Request, match *mux.RouteMatch) bool {

	return regexp.MustCompile(`^(\/api\/user\/racks\?.*)$`).
		MatchString(request.URL.String())
}

var dmatch mux.MatcherFunc = func(request *http.Request, match *mux.RouteMatch) bool {

	return regexp.MustCompile(`^(\/api\/user\/devices\?.*)$`).
		MatchString(request.URL.String())
}

func main() {

	router := mux.NewRouter()

	router.HandleFunc("/api/user",
		controllers.CreateAccount).Methods("POST")

	router.HandleFunc("/api/user/login",
		controllers.Authenticate).Methods("POST")

	router.HandleFunc("/api/token/valid",
		controllers.Verify).Methods("GET")

	// ------ TENANTS CRUD ------ //
	router.HandleFunc("/api/user/tenants",
		controllers.GetTenantByQuery).Methods("GET").MatcherFunc(tmatch)

	router.HandleFunc("/api/user/tenants",
		controllers.GetAllTenants).Methods("GET")

	router.HandleFunc("/api/user/tenants",
		controllers.CreateTenant).Methods("POST")

	router.HandleFunc("/api/user/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}",
		controllers.GetNamedDeviceOfTenant).Methods("GET")

	router.HandleFunc("/api/user/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices",
		controllers.GetDevicesUsingNamedRackOfTenant).Methods("GET")

	router.HandleFunc("/api/user/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}",
		controllers.GetNamedRackOfTenant).Methods("GET")

	router.HandleFunc("/api/user/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms/{room_name}/racks",
		controllers.GetRacksUsingNamedRoomOfTenant).Methods("GET")

	router.HandleFunc("/api/user/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms/{room_name}",
		controllers.GetNamedRoomOfTenant).Methods("GET")

	router.HandleFunc("/api/user/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms",
		controllers.GetRoomsUsingNamedBuildingOfTenant).Methods("GET")

	router.HandleFunc("/api/user/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}",
		controllers.GetNamedBuildingOfTenant).Methods("GET")

	router.HandleFunc("/api/user/tenants/{tenant_name}/sites/{site_name}/buildings",
		controllers.GetBuildingsUsingNamedSiteOfTenant).Methods("GET")

	router.HandleFunc("/api/user/tenants/{tenant_name}/sites/{site_name}",
		controllers.GetNamedSiteOfTenant).Methods("GET")

	router.HandleFunc("/api/user/tenants/{tenant_name}/sites",
		controllers.GetSitesOfTenant).Methods("GET")

	router.HandleFunc("/api/user/tenants/{tenant_name}/buildings",
		controllers.GetBuildingsOfTenant).Methods("GET")

	router.HandleFunc("/api/user/tenants/{id:[0-9]+}/all/nonstd",
		controllers.GetTenantHierarchyNonStandard).Methods("GET")

	router.HandleFunc("/api/user/tenants/{id:[0-9]+}/all",
		controllers.GetTenantHierarchy).Methods("GET")

	router.HandleFunc("/api/user/tenants/{id}",
		controllers.GetTenantFor).Methods("GET")

	router.HandleFunc("/api/user/tenants/{id}",
		controllers.UpdateTenant).Methods("PUT")

	router.HandleFunc("/api/user/tenants/{id}",
		controllers.DeleteTenant).Methods("DELETE")

	// ------ SITES CRUD ------ //
	router.HandleFunc("/api/user/sites",
		controllers.GetSiteByQuery).Methods("GET").MatcherFunc(smatch)

	router.HandleFunc("/api/user/sites",
		controllers.CreateSite).Methods("POST")

	router.HandleFunc("/api/user/sites/{id:[0-9]+}/all/nonstd",
		controllers.GetSiteHierarchyNonStandard).Methods("GET")

	router.HandleFunc("/api/user/sites/{id:[0-9]+}/all",
		controllers.GetSiteHierarchy).Methods("GET")

	router.HandleFunc("/api/user/sites",
		controllers.GetAllSites).Methods("GET")

	router.HandleFunc("/api/user/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}",
		controllers.GetNamedDeviceOfSite).Methods("GET")

	router.HandleFunc("/api/user/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices",
		controllers.GetDevicesUsingNamedRackOfSite).Methods("GET")

	router.HandleFunc("/api/user/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}",
		controllers.GetNamedRackOfSite).Methods("GET")

	router.HandleFunc("/api/user/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks",
		controllers.GetRacksUsingNamedRoomOfSite).Methods("GET")

	router.HandleFunc("/api/user/sites/{id}/buildings/{building_name}/rooms/{room_name}",
		controllers.GetNamedRoomOfSite).Methods("GET")

	router.HandleFunc("/api/user/sites/{id}/buildings/{building_name}/rooms",
		controllers.GetRoomsUsingNamedBldgOfSite).Methods("GET")

	router.HandleFunc("/api/user/sites/{id}/buildings/{building_name}",
		controllers.GetNamedBuildingOfSite).Methods("GET")

	router.HandleFunc("/api/user/sites/{id}/buildings",
		controllers.GetBuildingsOfSite).Methods("GET")

	router.HandleFunc("/api/user/sites/{id}/rooms",
		controllers.GetRoomsOfSite).Methods("GET")

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
		controllers.GetBuildingByQuery).Methods("GET").MatcherFunc(bmatch)

	router.HandleFunc("/api/user/buildings",
		controllers.CreateBuilding).Methods("POST")

	router.HandleFunc("/api/user/buildings/{id}",
		controllers.UpdateBuilding).Methods("PUT")

	router.HandleFunc("/api/user/buildings/{id}",
		controllers.DeleteBuilding).Methods("DELETE")

	router.HandleFunc("/api/user/buildings/{id:[0-9]+}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}",
		controllers.GetNamedDeviceOfBuilding).Methods("GET")

	router.HandleFunc("/api/user/buildings/{id:[0-9]+}/rooms/{room_name}/racks/{rack_name}/devices",
		controllers.GetDevicesUsingNamedRackOfBuilding).Methods("GET")

	router.HandleFunc("/api/user/buildings/{id:[0-9]+}/rooms/{room_name}/racks/{rack_name}",
		controllers.GetNamedRackOfBuilding).Methods("GET")

	router.HandleFunc("/api/user/buildings/{id:[0-9]+}/rooms/{room_name}/racks",
		controllers.GetRacksUsingNamedRoomOfBuilding).Methods("GET")

	router.HandleFunc("/api/user/buildings/{id:[0-9]+}/rooms/{room_name}",
		controllers.GetNamedRoomOfBuilding).Methods("GET")

	router.HandleFunc("/api/user/buildings/{id:[0-9]+}/rooms",
		controllers.GetRoomsOfBuilding).Methods("GET")

	router.HandleFunc("/api/user/buildings/{id:[0-9]+}/racks",
		controllers.GetRacksOfBuilding).Methods("GET")

	router.HandleFunc("/api/user/buildings/{id:[0-9]+}/all",
		controllers.GetBuildingHierarchy).Methods("GET")

	router.HandleFunc("/api/user/buildings/{id:[0-9]+}/all/nonstd",
		controllers.GetBuildingHierarchyNonStandard).Methods("GET")

	router.HandleFunc("/api/user/buildings/{id}",
		controllers.GetBuilding).Methods("GET")

	router.HandleFunc("/api/user/buildings",
		controllers.GetAllBuildings).Methods("GET")

	// ------ ROOM CRUD ------ //
	router.HandleFunc("/api/user/rooms",
		controllers.GetRoomByQuery).Methods("GET").MatcherFunc(rmatch)

	router.HandleFunc("/api/user/rooms",
		controllers.CreateRoom).Methods("POST")

	router.HandleFunc("/api/user/rooms/{id}",
		controllers.UpdateRoom).Methods("PUT")

	router.HandleFunc("/api/user/rooms/{id}",
		controllers.DeleteRoom).Methods("DELETE")

	router.HandleFunc("/api/user/rooms/{id:[0-9]+}/racks/{rack_name}/devices/{device_name}",
		controllers.GetNamedDeviceOfRoom).Methods("GET")

	router.HandleFunc("/api/user/rooms/{id:[0-9]+}/racks/{rack_name}/devices",
		controllers.GetDevicesUsingNamedRackOfRoom).Methods("GET")

	router.HandleFunc("/api/user/rooms/{id:[0-9]+}/racks/{rack_name}",
		controllers.GetRackOfRoomByName).Methods("GET")

	router.HandleFunc("/api/user/rooms/{id:[0-9]+}/racks",
		controllers.GetRacksOfParent).Methods("GET")

	router.HandleFunc("/api/user/rooms/{id:[0-9]+}/devices",
		controllers.GetDevicesOfRoom).Methods("GET")

	router.HandleFunc("/api/user/rooms/{id:[0-9]+}/all",
		controllers.GetRoomHierarchy).Methods("GET")

	router.HandleFunc("/api/user/rooms/{id:[0-9]+}/all/nonstd",
		controllers.GetRoomHierarchyNonStandard).Methods("GET")

	router.HandleFunc("/api/user/rooms/{id}",
		controllers.GetRoom).Methods("GET")

	router.HandleFunc("/api/user/rooms",
		controllers.GetAllRooms).Methods("GET")

	// ------ RACK CRUD ------ //
	router.HandleFunc("/api/user/racks",
		controllers.GetRackByQuery).Methods("GET").MatcherFunc(ramatch)

	router.HandleFunc("/api/user/racks",
		controllers.CreateRack).Methods("POST")

	router.HandleFunc("/api/user/racks",
		controllers.GetAllRacks).Methods("GET")

	router.HandleFunc("/api/user/racks/{id}",
		controllers.UpdateRack).Methods("PUT")

	router.HandleFunc("/api/user/racks/{id}",
		controllers.DeleteRack).Methods("DELETE")

	router.HandleFunc("/api/user/racks/{id:[0-9]+}/devices/{device_name}",
		controllers.GetRackDeviceByName).Methods("GET")

	router.HandleFunc("/api/user/racks/{id:[0-9]+}/all",
		controllers.GetRackHierarchy).Methods("GET")

	router.HandleFunc("/api/user/racks/{id:[0-9]+}/all/nonstd",
		controllers.GetRackHierarchyNonStandard).Methods("GET")

	router.HandleFunc("/api/user/racks/{id}",
		controllers.GetRack).Methods("GET")

	// ------ DEVICE CRUD ------ //
	router.HandleFunc("/api/user/devices",
		controllers.GetDeviceByQuery).Methods("GET").MatcherFunc(dmatch)

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
