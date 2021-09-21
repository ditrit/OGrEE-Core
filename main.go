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

	return regexp.MustCompile(`^(\/api\/user\/(devices|subdevices|subdevices1)\?.*)$`).
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
		controllers.GetTenantByQuery).Methods("GET").MatcherFunc(tmatch)

	router.HandleFunc("/api/tenants",
		controllers.GetAllEntities).Methods("GET")

	router.HandleFunc("/api/tenants",
		controllers.CreateEntity).Methods("POST")

	router.HandleFunc("/api/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}/subdevices/{subdevice_name}/subdevice1s/{subdevice1_name}",
		controllers.GetNamedSubdevice1OfTenant).Methods("GET")

	router.HandleFunc("/api/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}/subdevices/{subdevice_name}/subdevice1s",
		controllers.GetSubdevice1sUsingNamedSubdeviceOfTenant).Methods("GET")

	router.HandleFunc("/api/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}/subdevices/{subdevice_name}",
		controllers.GetNamedSubdeviceOfTenant).Methods("GET")

	router.HandleFunc("/api/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}/subdevices",
		controllers.GetSubdevicesUsingNamedDeviceOfTenant).Methods("GET")

	router.HandleFunc("/api/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}",
		controllers.GetNamedDeviceOfTenant).Methods("GET")

	router.HandleFunc("/api/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices",
		controllers.GetDevicesUsingNamedRackOfTenant).Methods("GET")

	router.HandleFunc("/api/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}",
		controllers.GetNamedRackOfTenant).Methods("GET")

	router.HandleFunc("/api/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms/{room_name}/racks",
		controllers.GetRacksUsingNamedRoomOfTenant).Methods("GET")

	router.HandleFunc("/api/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms/{room_name}",
		controllers.GetNamedRoomOfTenant).Methods("GET")

	router.HandleFunc("/api/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms",
		controllers.GetRoomsUsingNamedBuildingOfTenant).Methods("GET")

	router.HandleFunc("/api/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}",
		controllers.GetNamedBuildingOfTenant).Methods("GET")

	router.HandleFunc("/api/tenants/{tenant_name}/sites/{site_name}/buildings",
		controllers.GetBuildingsUsingNamedSiteOfTenant).Methods("GET")

	router.HandleFunc("/api/tenants/{tenant_name}/sites/{site_name}",
		controllers.GetNamedSiteOfTenant).Methods("GET")

	router.HandleFunc("/api/tenants/{tenant_name}/sites",
		controllers.GetSitesOfTenant).Methods("GET")

	router.HandleFunc("/api/tenants/{tenant_name}/all/sites/buildings/rooms/racks/devices/subdevices",
		controllers.GetTenantHierarchyToSubdevice).Methods("GET")

	router.HandleFunc("/api/tenants/{tenant_name}/all/sites/buildings/rooms/racks/devices",
		controllers.GetTenantHierarchyToDevice).Methods("GET")

	router.HandleFunc("/api/tenants/{tenant_name}/all/sites/buildings/rooms/racks",
		controllers.GetTenantHierarchyToRack).Methods("GET")

	router.HandleFunc("/api/tenants/{tenant_name}/all/sites/buildings/rooms",
		controllers.GetTenantHierarchyToRoom).Methods("GET")

	router.HandleFunc("/api/tenants/{tenant_name}/all/sites/buildings",
		controllers.GetTenantHierarchyToBuilding).Methods("GET")

	router.HandleFunc("/api/tenants/{id:[0-9]+}/all/nonstd",
		controllers.GetTenantHierarchyNonStandard).Methods("GET")

	router.HandleFunc("/api/tenants/{id:[0-9]+}/all",
		controllers.GetTenantHierarchy).Methods("GET")

	router.HandleFunc("/api/tenants/{id}",
		controllers.GetTenantFor).Methods("GET")

	router.HandleFunc("/api/tenants/{id}",
		controllers.UpdateTenant).Methods("PUT")

	router.HandleFunc("/api/tenants/{id}",
		controllers.DeleteTenant).Methods("DELETE")

	// ------ SITES CRUD ------ //
	router.HandleFunc("/api/sites",
		controllers.GetSiteByQuery).Methods("GET").MatcherFunc(smatch)

	router.HandleFunc("/api/sites",
		controllers.CreateEntity).Methods("POST")

	router.HandleFunc("/api/sites/{id:[0-9]+}/all/nonstd",
		controllers.GetSiteHierarchyNonStandard).Methods("GET")

	router.HandleFunc("/api/sites/{id:[0-9]+}/all/buildings/rooms/racks/devices/subdevices",
		controllers.GetSiteHierarchyToSubdevice).Methods("GET")

	router.HandleFunc("/api/sites/{id:[0-9]+}/all/buildings/rooms/racks/devices",
		controllers.GetSiteHierarchyToDevice).Methods("GET")

	router.HandleFunc("/api/sites/{id:[0-9]+}/all/buildings/rooms/racks",
		controllers.GetSiteHierarchyToRack).Methods("GET")

	router.HandleFunc("/api/sites/{id:[0-9]+}/all/buildings/rooms",
		controllers.GetSiteHierarchyToRoom).Methods("GET")

	router.HandleFunc("/api/sites/{id:[0-9]+}/all",
		controllers.GetSiteHierarchy).Methods("GET")

	router.HandleFunc("/api/sites",
		controllers.GetAllEntities).Methods("GET")

	router.HandleFunc("/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}/subdevices/{subdevice_name}/subdevice1s/{subdevice1_name}",
		controllers.GetNamedSubdevice1OfSite).Methods("GET")

	router.HandleFunc("/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}/subdevices/{subdevice_name}/subdevice1s",
		controllers.GetSubdevice1sUsingNamedSubdeviceOfSite).Methods("GET")

	router.HandleFunc("/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}/subdevices/{subdevice_name}",
		controllers.GetNamedSubdeviceOfSite).Methods("GET")

	router.HandleFunc("/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}/subdevices",
		controllers.GetSubdevicesUsingNamedDeviceOfSite).Methods("GET")

	router.HandleFunc("/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}",
		controllers.GetNamedDeviceOfSite).Methods("GET")

	router.HandleFunc("/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices",
		controllers.GetDevicesUsingNamedRackOfSite).Methods("GET")

	router.HandleFunc("/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}",
		controllers.GetNamedRackOfSite).Methods("GET")

	router.HandleFunc("/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks",
		controllers.GetRacksUsingNamedRoomOfSite).Methods("GET")

	router.HandleFunc("/api/sites/{id}/buildings/{building_name}/rooms/{room_name}",
		controllers.GetNamedRoomOfSite).Methods("GET")

	router.HandleFunc("/api/sites/{id}/buildings/{building_name}/rooms",
		controllers.GetRoomsUsingNamedBldgOfSite).Methods("GET")

	router.HandleFunc("/api/sites/{id}/buildings/{building_name}",
		controllers.GetNamedBuildingOfSite).Methods("GET")

	router.HandleFunc("/api/sites/{id}/buildings",
		controllers.GetBuildingsOfSite).Methods("GET")

	router.HandleFunc("/api/sites/{id}",
		controllers.GetSite).Methods("GET")

	router.HandleFunc("/api/sites/{id}",
		controllers.UpdateSite).Methods("PUT")

	router.HandleFunc("/api/sites/{id}",
		controllers.DeleteSiteByID).Methods("DELETE")

	router.HandleFunc("/api/sites",
		controllers.DeleteSites).Methods("DELETE")

	// ------ BUILDING CRUD ------ //
	router.HandleFunc("/api/buildings",
		controllers.GetBuildingByQuery).Methods("GET").MatcherFunc(bmatch)

	router.HandleFunc("/api/buildings",
		controllers.CreateEntity).Methods("POST")

	router.HandleFunc("/api/buildings/{id}",
		controllers.UpdateBuilding).Methods("PUT")

	router.HandleFunc("/api/buildings/{id}",
		controllers.DeleteBuilding).Methods("DELETE")

	router.HandleFunc("/api/buildings/{id:[0-9]+}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}/subdevices/{subdevice_name}/subdevice1s/{subdevice1_name}",
		controllers.GetNamedSubdevice1OfBuilding).Methods("GET")

	router.HandleFunc("/api/buildings/{id:[0-9]+}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}/subdevices/{subdevice_name}/subdevice1s",
		controllers.GetSubdevice1sUsingNamedSubdeviceOfBuilding).Methods("GET")

	router.HandleFunc("/api/buildings/{id:[0-9]+}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}/subdevices/{subdevice_name}",
		controllers.GetNamedSubdeviceOfBuilding).Methods("GET")

	router.HandleFunc("/api/buildings/{id:[0-9]+}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}/subdevices",
		controllers.GetSubdevicesUsingNamedDeviceOfBuilding).Methods("GET")

	router.HandleFunc("/api/buildings/{id:[0-9]+}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}",
		controllers.GetNamedDeviceOfBuilding).Methods("GET")

	router.HandleFunc("/api/buildings/{id:[0-9]+}/rooms/{room_name}/racks/{rack_name}/devices",
		controllers.GetDevicesUsingNamedRackOfBuilding).Methods("GET")

	router.HandleFunc("/api/buildings/{id:[0-9]+}/rooms/{room_name}/racks/{rack_name}",
		controllers.GetNamedRackOfBuilding).Methods("GET")

	router.HandleFunc("/api/buildings/{id:[0-9]+}/rooms/{room_name}/racks",
		controllers.GetRacksUsingNamedRoomOfBuilding).Methods("GET")

	router.HandleFunc("/api/buildings/{id:[0-9]+}/rooms/{room_name}",
		controllers.GetNamedRoomOfBuilding).Methods("GET")

	router.HandleFunc("/api/buildings/{id:[0-9]+}/rooms",
		controllers.GetRoomsOfBuilding).Methods("GET")

	router.HandleFunc("/api/buildings/{id:[0-9]+}/all/nonstd",
		controllers.GetBuildingHierarchyNonStandard).Methods("GET")

	router.HandleFunc("/api/buildings/{id:[0-9]+}/all/rooms/racks/devices/subdevices",
		controllers.GetBuildingHierarchyToSubdevice).Methods("GET")

	router.HandleFunc("/api/buildings/{id:[0-9]+}/all/rooms/racks/devices",
		controllers.GetBuildingHierarchyToDevice).Methods("GET")

	router.HandleFunc("/api/buildings/{id:[0-9]+}/all/rooms/racks",
		controllers.GetBuildingHierarchyToRack).Methods("GET")

	router.HandleFunc("/api/buildings/{id:[0-9]+}/all",
		controllers.GetBuildingHierarchy).Methods("GET")

	router.HandleFunc("/api/buildings/{id}",
		controllers.GetBuilding).Methods("GET")

	router.HandleFunc("/api/buildings",
		controllers.GetAllEntities).Methods("GET")

	// ------ ROOM CRUD ------ //
	router.HandleFunc("/api/rooms",
		controllers.GetRoomByQuery).Methods("GET").MatcherFunc(rmatch)

	router.HandleFunc("/api/rooms",
		controllers.CreateEntity).Methods("POST")

	router.HandleFunc("/api/rooms/{id}",
		controllers.UpdateRoom).Methods("PUT")

	router.HandleFunc("/api/rooms/{id}",
		controllers.DeleteRoom).Methods("DELETE")

	router.HandleFunc("/api/rooms/{id:[0-9]+}/racks/{rack_name}/devices/{device_name}/subdevices/{subdevice_name}/subdevice1s/{subdevice1_name}",
		controllers.GetNamedSubdevice1OfRoom).Methods("GET")

	router.HandleFunc("/api/rooms/{id:[0-9]+}/racks/{rack_name}/devices/{device_name}/subdevices/{subdevice_name}/subdevice1s",
		controllers.GetSubdevice1sUsingUsingNamedSubdeviceOfRoom).Methods("GET")

	router.HandleFunc("/api/rooms/{id:[0-9]+}/racks/{rack_name}/devices/{device_name}/subdevices/{subdevice_name}",
		controllers.GetNamedSubdeviceOfRoom).Methods("GET")

	router.HandleFunc("/api/rooms/{id:[0-9]+}/racks/{rack_name}/devices/{device_name}/subdevices",
		controllers.GetSubdevicesUsingNamedDeviceOfRoom).Methods("GET")

	router.HandleFunc("/api/rooms/{id:[0-9]+}/racks/{rack_name}/devices/{device_name}",
		controllers.GetNamedDeviceOfRoom).Methods("GET")

	router.HandleFunc("/api/rooms/{id:[0-9]+}/racks/{rack_name}/devices",
		controllers.GetDevicesUsingNamedRackOfRoom).Methods("GET")

	router.HandleFunc("/api/rooms/{id:[0-9]+}/racks/{rack_name}",
		controllers.GetRackOfRoomByName).Methods("GET")

	router.HandleFunc("/api/rooms/{id:[0-9]+}/racks",
		controllers.GetRacksOfParent).Methods("GET")

	router.HandleFunc("/api/rooms/{id:[0-9]+}/all/racks/devices/subdevices",
		controllers.GetRoomHierarchyToSubdevices).Methods("GET")

	router.HandleFunc("/api/rooms/{id:[0-9]+}/all/racks/devices",
		controllers.GetRoomHierarchyToDevices).Methods("GET")

	router.HandleFunc("/api/rooms/{id:[0-9]+}/all",
		controllers.GetRoomHierarchy).Methods("GET")

	router.HandleFunc("/api/rooms/{id:[0-9]+}/all/nonstd",
		controllers.GetRoomHierarchyNonStandard).Methods("GET")

	router.HandleFunc("/api/rooms/{id}",
		controllers.GetRoom).Methods("GET")

	router.HandleFunc("/api/rooms",
		controllers.GetAllEntities).Methods("GET")

	// ------ RACK CRUD ------ //
	router.HandleFunc("/api/racks",
		controllers.GetRackByQuery).Methods("GET").MatcherFunc(ramatch)

	router.HandleFunc("/api/racks",
		controllers.CreateEntity).Methods("POST")

	router.HandleFunc("/api/racks",
		controllers.GetAllEntities).Methods("GET")

	router.HandleFunc("/api/racks/{id}",
		controllers.UpdateRack).Methods("PUT")

	router.HandleFunc("/api/racks/{id}",
		controllers.DeleteRack).Methods("DELETE")

	router.HandleFunc("/api/racks/{id:[0-9]+}/devices/{device_name}/subdevices/{subdevice_name}/subdevice1s/{subdevice1_name}",
		controllers.GetNamedSubdevice1OfRack).Methods("GET")

	router.HandleFunc("/api/racks/{id:[0-9]+}/devices/{device_name}/subdevices/{subdevice_name}/subdevice1s",
		controllers.GetSubdevice1sUsingNamedSubdeviceOfRack).Methods("GET")

	router.HandleFunc("/api/racks/{id:[0-9]+}/devices/{device_name}/subdevices/{subdevice_name}",
		controllers.GetNamedSubdeviceOfRack).Methods("GET")

	router.HandleFunc("/api/racks/{id:[0-9]+}/devices/{device_name}/subdevices",
		controllers.GetSubdevicesUsingNamedDeviceOfRack).Methods("GET")

	router.HandleFunc("/api/racks/{id:[0-9]+}/devices/{device_name}",
		controllers.GetRackDeviceByName).Methods("GET")

	router.HandleFunc("/api/racks/{id:[0-9]+}/devices",
		controllers.GetDevicesOfRack).Methods("GET")

	router.HandleFunc("/api/racks/{id:[0-9]+}/all/devices/subdevices",
		controllers.GetRackHierarchyToSubdevices).Methods("GET")

	router.HandleFunc("/api/racks/{id:[0-9]+}/all",
		controllers.GetRackHierarchy).Methods("GET")

	router.HandleFunc("/api/racks/{id:[0-9]+}/all/nonstd",
		controllers.GetRackHierarchyNonStandard).Methods("GET")

	router.HandleFunc("/api/racks/{id}",
		controllers.GetRack).Methods("GET")

	// ------ DEVICE CRUD ------ //
	router.HandleFunc("/api/devices",
		controllers.GetDeviceByQuery).Methods("GET").MatcherFunc(dmatch)

	router.HandleFunc("/api/devices",
		controllers.CreateEntity).Methods("POST")

	router.HandleFunc("/api/devices/{id}",
		controllers.UpdateDevice).Methods("PUT")

	router.HandleFunc("/api/devices/{id}",
		controllers.DeleteDevice).Methods("DELETE")

	router.HandleFunc("/api/devices/{id:[0-9]+}/subdevices/{subdevice_name}/subdevice1s/{subdevone_name}",
		controllers.GetNamedSubdevice1OfDevice).Methods("GET")

	router.HandleFunc("/api/devices/{id:[0-9]+}/subdevices/{subdevice_name}/subdevice1s",
		controllers.GetSubdevice1sUsingNamedSubdeviceOfDevice).Methods("GET")

	router.HandleFunc("/api/devices/{id:[0-9]+}/subdevices/{subdevice_name}",
		controllers.GetDeviceSubdeviceByName).Methods("GET")

	router.HandleFunc("/api/devices/{id:[0-9]+}/subdevices",
		controllers.GetSubdevicesOfDevice).Methods("GET")

	router.HandleFunc("/api/devices/{id:[0-9]+}/all",
		controllers.GetDeviceHierarchy).Methods("GET")

	router.HandleFunc("/api/devices/{id}",
		controllers.GetDevice).Methods("GET")

	router.HandleFunc("/api/devices",
		controllers.GetAllEntities).Methods("GET")

	// ------ SUBDEVICE CRUD ------ //
	router.HandleFunc("/api/subdevices",
		controllers.GetSubdeviceByQuery).Methods("GET").MatcherFunc(dmatch)

	router.HandleFunc("/api/subdevices",
		controllers.CreateEntity).Methods("POST")

	router.HandleFunc("/api/subdevices/{id}",
		controllers.UpdateSubdevice).Methods("PUT")

	router.HandleFunc("/api/subdevices/{id}",
		controllers.DeleteSubdevice).Methods("DELETE")

	router.HandleFunc("/api/subdevices/{id:[0-9]+}/all",
		controllers.GetSubdeviceHierarchy).Methods("GET")

	router.HandleFunc("/api/subdevices/{id}",
		controllers.GetSubdevice).Methods("GET")

	router.HandleFunc("/api/subdevices",
		controllers.GetAllEntities).Methods("GET")

	// ------ SUBDEVICE1 CRUD ------ //
	router.HandleFunc("/api/subdevice1s",
		controllers.GetSubdevice1ByQuery).Methods("GET").MatcherFunc(dmatch)

	router.HandleFunc("/api/subdevice1s",
		controllers.CreateEntity).Methods("POST")

	router.HandleFunc("/api/subdevice1s/{id}",
		controllers.UpdateSubdevice1).Methods("PUT")

	router.HandleFunc("/api/subdevice1s/{id}",
		controllers.DeleteSubdevice1).Methods("DELETE")

	router.HandleFunc("/api/subdevice1s/{id}",
		controllers.GetSubdevice1).Methods("GET")

	router.HandleFunc("/api/subdevice1s",
		controllers.GetAllEntities).Methods("GET")

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
