# Exhaustive Endpoint List
This is an exhaustive list of endpoints for reference 


POST / Create
------------
Perform an HTTP POST operation with the appropriate JSON 
```
/api
/api/login
/api/validate/{obj}

/api/domains
/api/sites
/api/buildings
/api/rooms
/api/racks
/api/devices
/api/acs
/api/cabinets
/api/corridors
/api/panels
/api/sensors
/api/groups
/api/room-templates
/api/obj-templates
/api/bldg-templates
/api/stray-sensors
/api/stray-devices
/api/validate/{obj}
```



DELETE / Delete
------------
Perform an HTTP DELETE operation without JSON body
```
/api/domains/{id}
/api/sites/{id}
/api/buildings/{id}
/api/rooms/{id}
/api/racks/{id}
/api/devices/{id}
/api/acs/{id}
/api/cabinets/{id}
/api/corridors/{id}
/api/panels/{id}
/api/sensors/{id}
/api/groups/{id}
/api/room-templates/{slug}
/api/obj-templates/{slug}
/api/bldg-templates/{slug}
/api/stray-sensors/{id_or_name}
/api/stray-devices/{id_or_name}
```


PATCH OR PUT / Update
-------------
Perform an HTTP PUT or PATCH operation with desired JSON body
```
/api/domains/{id}
/api/sites/{id}
/api/buildings/{id}
/api/rooms/{id}
/api/racks/{id}
/api/devices/{id}
/api/acs/{id}
/api/cabinets/{id}
/api/corridors/{id}
/api/panels/{id}
/api/sensors/{id}
/api/groups/{id}
/api/room-templates/{slug}
/api/obj-templates/{slug}
/api/bldg-templates/{slug}
/api/stray-sensors/{id_or_name}
/api/stray-devices/{id_or_name}
```

OPTIONS
-------------
Perform an HTTP OPTIONS operation without JSON body
```
/api
/api/stats
/api/login
/api/token/valid
/api/version
/api/validate/{obj}
/api/tempunits/{id}

// HIERARCHY - START 
/api/domains/{id_or_name}/all
/api/sites/{id_or_name}/all
/api/buildings/{id}/all
/api/rooms/{id}/all
/api/racks/{id}/all
/api/devices/{id}/all
/api/stray-devices/{id_or_name}/all
// HIERARCHY - END 

// HIERARCHY RANGE - START
/api/domains/{id_or_name}/all?limit={#}

/api/sites/{id_or_name}/all?limit={#}
/api/sites/{id_or_name}/all/buildings
/api/sites/{id_or_name}/all/buildings/rooms
/api/sites/{id_or_name}/all/buildings/rooms/racks
/api/sites/{id_or_name}/all/buildings/rooms/racks/devices

/api/buildings/{id}/all?limit={#}
/api/buildings/{id}/all/rooms/racks
/api/buildings/{id}/all/rooms/racks/devices

/api/rooms/{id}/all?limit={#}
/api/rooms/{id}/all/racks
/api/rooms/{id}/all/racks/devices

/api/racks/{id}/all?limit={#}
/api/racks/{id}/all/devices
// HIERARCHY RANGE - END 


// HIERARCHY JUMP SERIES - START
/api/sites/{site_name}/rooms
/api/buildings/{id}/{acs|corridors|cabinets|panels|sensors|groups|racks}
/api/rooms/{id}/devices
// HIERARCHY JUMP SERIES - END

// HIERARCHY '1 BELOW' SERIES - START
/api/sites/{site_name}/buildings
/api/buildings/{id}/rooms
/api/rooms/{id}/racks
/api/racks/{id}/devices
// HIERARCHY '1 BELOW' SERIES - END

/api/domains/{id}
/api/sites/{id_or_name}
/api/buildings/{id}
/api/rooms/{id}
/api/racks/{id}
/api/devices/{id}
/api/acs/{id}
/api/cabinets/{id}
/api/corridors/{id}
/api/panels/{id}
/api/sensors/{id}
/api/groups/{id}
/api/room-templates/{slug}
/api/obj-templates/{slug}
/api/bldg-templates/{slug}
/api/stray-sensors/{id_or_name}
/api/stray-devices/{id_or_name}


// ENTIT(Y)IES USING NAMES OF PARENTS - START

/api/sites/{id}/buildings/{building_name}
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/acs/{ac_name}
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/cabinets/{cabinet_name}
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/corridors/{corridor_name}
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/panels/{panel_name}
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/sensors/{sensor_name}
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/groups/{group_name}
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}/sensors/{sensor_name}
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/sensors/{sensor_name}



/api/buildings/{id}/rooms/{room_name}
/api/buildings/{id}/rooms/{room_name}/acs/{ac_name}
/api/buildings/{id}/rooms/{room_name}/cabinets/{cabinet_name}
/api/buildings/{id}/rooms/{room_name}/corridors/{corridor_name}
/api/buildings/{id}/rooms/{room_name}/panels/{panel_name}
/api/buildings/{id}/rooms/{room_name}/sensors/{sensor_name}
/api/buildings/{id}/rooms/{room_name}/groups/{group_name}
/api/buildings/{id}/rooms/{room_name}/rack/{rack_name}
/api/buildings/{id}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}
/api/buildings/{id}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}/sensors/{sensor_name}
/api/buildings/{id}/rooms/{room_name}/racks/{rack_name}/sensors/{sensor_name}



/api/rooms/{id}/acs/{ac_name}
/api/rooms/{id}/cabinets/{cabinet_name}
/api/rooms/{id}/corridors/{corridor_name}
/api/rooms/{id}/panels/{panel_name}
/api/rooms/{id}/sensors/{sensor_name}
/api/rooms/{id}/groups/{group_name}
/api/rooms/{id}/racks/{rack_name}/devices/{device_name}
/api/rooms/{id}/racks/{rack_name}/devices/{device_name}/sensors/{sensor_name}
/api/rooms/{id}/racks/{rack_name}/sensors/{sensor_name}

/api/racks/{id}/devices/{device_name}
/api/racks/{id}/devices/{device_name}/sensors/{sensor_name}
/api/racks/{id}/sensors/{sensor_name}

/api/devices/{id}/devices/{device_name}
/api/devices/{id}/sensors/{sensor_name}
// ENTIT(Y)IES USING NAMES OF PARENTS - END


// ENTITIES USING NAMES OF PARENTS - START

/api/sites/{id}/buildings/{building_name}
/api/sites/{id}/buildings/{building_name}/rooms
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/acs
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/cabinets
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/corridors
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/panels
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/sensors
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/groups
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}/sensors
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/sensors



/api/buildings/{id}/rooms
/api/buildings/{id}/rooms/{room_name}/acs
/api/buildings/{id}/rooms/{room_name}/cabinets
/api/buildings/{id}/rooms/{room_name}/corridors
/api/buildings/{id}/rooms/{room_name}/panels
/api/buildings/{id}/rooms/{room_name}/sensors
/api/buildings/{id}/rooms/{room_name}/groups
/api/buildings/{id}/rooms/{room_name}/rack
/api/buildings/{id}/rooms/{room_name}/racks/{rack_name}/devices
/api/buildings/{id}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}/sensors
/api/buildings/{id}/rooms/{room_name}/racks/{rack_name}/sensors



/api/rooms/{id}/acs
/api/rooms/{id}/cabinets
/api/rooms/{id}/corridors
/api/rooms/{id}/panels
/api/rooms/{id}/sensors
/api/rooms/{id}/groups
/api/rooms/{id}/racks/{rack_name}/devices
/api/rooms/{id}/racks/{rack_name}/devices/{device_name}/sensors
/api/rooms/{id}/racks/{rack_name}/sensors

/api/racks/{id}/devices
/api/racks/{id}/devices/{device_name}/sensors
/api/racks/{id}/sensors

/api/devices/{id}/devices
/api/devices/{id}/sensors
// ENTITIES USING NAMES OF PARENTS - END

```

HEAD 
-------------
Perform an HTTP HEAD operation
```
/api/stats
/api/token/valid
/api/version
/api/tempunits/{id}

// ENTIT(Y)IES USING NAMES OF PARENTS - START

/api/sites/{id}/buildings/{building_name}
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/acs/{ac_name}
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/cabinets/{cabinet_name}
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/corridors/{corridor_name}
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/panels/{panel_name}
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/sensors/{sensor_name}
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/groups/{group_name}
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}/sensors/{sensor_name}
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/sensors/{sensor_name}



/api/buildings/{id}/rooms/{room_name}
/api/buildings/{id}/rooms/{room_name}/acs/{ac_name}
/api/buildings/{id}/rooms/{room_name}/cabinets/{cabinet_name}
/api/buildings/{id}/rooms/{room_name}/corridors/{corridor_name}
/api/buildings/{id}/rooms/{room_name}/panels/{panel_name}
/api/buildings/{id}/rooms/{room_name}/sensors/{sensor_name}
/api/buildings/{id}/rooms/{room_name}/groups/{group_name}
/api/buildings/{id}/rooms/{room_name}/rack/{rack_name}
/api/buildings/{id}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}
/api/buildings/{id}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}/sensors/{sensor_name}
/api/buildings/{id}/rooms/{room_name}/racks/{rack_name}/sensors/{sensor_name}



/api/rooms/{id}/acs/{ac_name}
/api/rooms/{id}/cabinets/{cabinet_name}
/api/rooms/{id}/corridors/{corridor_name}
/api/rooms/{id}/panels/{panel_name}
/api/rooms/{id}/sensors/{sensor_name}
/api/rooms/{id}/groups/{group_name}
/api/rooms/{id}/racks/{rack_name}/devices/{device_name}
/api/rooms/{id}/racks/{rack_name}/devices/{device_name}/sensors/{sensor_name}
/api/rooms/{id}/racks/{rack_name}/sensors/{sensor_name}

/api/racks/{id}/devices/{device_name}
/api/racks/{id}/devices/{device_name}/sensors/{sensor_name}
/api/racks/{id}/sensors/{sensor_name}

/api/devices/{id}/devices/{device_name}
/api/devices/{id}/sensors/{sensor_name}
// ENTIT(Y)IES USING NAMES OF PARENTS - END


// ENTITIES USING NAMES OF PARENTS - START

/api/sites/{id}/buildings/{building_name}
/api/sites/{id}/buildings/{building_name}/rooms
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/acs
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/cabinets
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/corridors
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/panels
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/sensors
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/groups
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}/sensors
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/sensors



/api/buildings/{id}/rooms
/api/buildings/{id}/rooms/{room_name}/acs
/api/buildings/{id}/rooms/{room_name}/cabinets
/api/buildings/{id}/rooms/{room_name}/corridors
/api/buildings/{id}/rooms/{room_name}/panels
/api/buildings/{id}/rooms/{room_name}/sensors
/api/buildings/{id}/rooms/{room_name}/groups
/api/buildings/{id}/rooms/{room_name}/rack
/api/buildings/{id}/rooms/{room_name}/racks/{rack_name}/devices
/api/buildings/{id}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}/sensors
/api/buildings/{id}/rooms/{room_name}/racks/{rack_name}/sensors



/api/rooms/{id}/acs
/api/rooms/{id}/cabinets
/api/rooms/{id}/corridors
/api/rooms/{id}/panels
/api/rooms/{id}/sensors
/api/rooms/{id}/groups
/api/rooms/{id}/racks/{rack_name}/devices
/api/rooms/{id}/racks/{rack_name}/devices/{device_name}/sensors
/api/rooms/{id}/racks/{rack_name}/sensors

/api/racks/{id}/devices
/api/racks/{id}/devices/{device_name}/sensors
/api/racks/{id}/sensors

/api/devices/{id}/devices
/api/devices/{id}/sensors
// ENTITIES USING NAMES OF PARENTS - END


//GET EXCEPTIONS
/api/sites/{site_name}/rooms
/api/buildings/{id}/{acs|corridors|cabinets|panels|sensors|groups}
/api/buildings/{id}/racks
/api/rooms/{id}/devices


// SEARCH QUERY TYPE
/api/domains?
/api/sites?
/api/buildings?
/api/rooms?
/api/racks?
/api/devices?
/api/acs?
/api/cabinets?
/api/corridors?
/api/panels?
/api/sensors?
/api/groups?
/api/room-templates?
/api/obj-templates?
/api/bldg-templates?
/api/stray-sensors?
/api/stray-devices?


// ALL ENTITIES 
/api/domains
/api/sites
/api/buildings
/api/rooms
/api/racks
/api/devices
/api/acs
/api/cabinets
/api/corridors
/api/panels
/api/sensors
/api/groups
/api/room-templates
/api/obj-templates
/api/bldg-templates
/api/stray-sensors
/api/stray-devices


// SINGLE ENTITY
/api/domains/{id_or_name}
/api/sites/{id_or_name}
/api/buildings/{id}
/api/rooms/{id}
/api/racks/{id}
/api/devices/{id}
/api/acs/{id}
/api/cabinets/{id}
/api/corridors/{id}
/api/panels/{id}
/api/sensors/{id}
/api/groups/{id}
/api/room-templates/{slug}
/api/obj-templates/{slug}
/api/bldg-templates/{slug}
/api/stray-sensors/{id_or_name}
/api/stray-devices/{id_or_name}
```


GET / Get
-------------

### Quick Token Check
This URL is for development purposes only
```
/api/token/valid
```

### Get DB statistics 
```
/api/stats
```

### Get API Version information
```
/api/version
```

### Get Parent's Temperature Unit Of Object (Site)
```
/api/tempunits/{id}
```

### Get All Objects
```
/api/domains
/api/sites
/api/buildings
/api/rooms
/api/racks
/api/devices
/api/acs
/api/cabinets
/api/corridors
/api/panels
/api/sensors
/api/groups
/api/room-templates
/api/obj-templates
/api/bldg-templates
/api/stray-sensors
/api/stray-devices
```

### Get by ID (non hierarchal)
ID is a string of length 24    
```
/api/domains/{id_or_name}
/api/sites/{id_or_name}
/api/buildings/{id}
/api/rooms/{id}
/api/racks/{id}
/api/devices/{id}
/api/acs/{id}
/api/cabinets/{id}
/api/corridors/{id}
/api/panels/{id}
/api/sensors/{id}
/api/groups/{id}
/api/room-templates/{slug}
/api/obj-templates/{slug}
/api/bldg-templates/{slug}
/api/stray-sensors/{id_or_name}
/api/stray-devices/{id_or_name}
```

### Search Objects
Search using a query in the URL
Objects that match the query will be returned
Example: /devices?name=myValue?color=silver
will return silver devices with name 'myValue'
```
/api/domains?
/api/sites?
/api/buildings?
/api/rooms?
/api/racks?
/api/devices?
/api/acs?
/api/cabinets?
/api/corridors?
/api/panels?
/api/sensors?
/api/groups?
/api/room-templates?
/api/obj-templates?
/api/bldg-templates?
/api/stray-sensors?
/api/stray-devices?
```


### Get all Objects 2 levels lower
```
/api/sites/{site_name}/rooms
/api/buildings/{id}/racks
/api/rooms/{id}/devices
```

### Get an Object's entire hierarchy
The object and everything related to it will be returned    
in a nested JSON fashion
```
/api/domains/{id_or_name}/all
/api/sites/{site_name}/all
/api/buildings/{id}/all
/api/rooms/{id}/all
/api/racks/{id}/all
/api/devices/{id}/all
/api/stray-devices/{id}/all
```

### Get object's ranged hierarchy 
Limits the depth of the hierarchy to retrieve. This is observed by the    
URL given. 
```
/api/sites/{id}/all/buildings/rooms/racks/devices
/api/sites/{id}/all/buildings/rooms/racks
/api/sites/{id}/all/buildings/rooms
/api/sites/{id}/all/buildings

/api/buildings/{id}/all/rooms/racks/devices
/api/buildings/{id}/all/rooms/racks
/api/buildings/{id}/all/rooms

/api/rooms/{id}/all/racks/devices
/api/rooms/{id}/all/racks

/api/stray-devices/{id}/all/devices
```

### Get an Object's ranged hierarchy (using limit parameter)
The object and everything related until the number specified by limit will be returned    
in a nested JSON fashion
```
/api/domains/{id_or_name}/all?limit={#}
/api/sites/{site_name}/all?limit={#}
/api/buildings/{id}/all?limit={#}
/api/rooms/{id}/all?limit={#}
/api/racks/{id}/all?limit={#}
/api/devices/{id}/all?limit={#}
/api/stray-devices/{id}/all?limit={#}
```

### Get objects through the hierarchy
Returns an object if name given or all the objects immediately under the given URL
```

/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks
/api/sites/{id}/buildings/{building_name}/rooms/{room_name}
/api/sites/{id}/buildings/{building_name}/rooms
/api/sites/{id}/buildings/{building_name}
/api/sites/{id}/buildings


/api/buildings/{id}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}
/api/buildings/{id}/rooms/{room_name}/racks/{rack_name}/devices
/api/buildings/{id}/rooms/{room_name}/racks/{rack_name}
/api/buildings/{id}/rooms/{room_name}/racks
/api/buildings/{id}/rooms/{room_name}
/api/buildings/{id}/rooms


/api/rooms/{id}/racks/{rack_name}/devices/{device_name}
/api/rooms/{id}/racks/{rack_name}/devices
/api/rooms/{id}/racks/{rack_name}
/api/rooms/{id}/racks


/api/rooms/{id}/acs/{ac_name}
/api/rooms/{id}/acs
/api/rooms/{id}/panels/{panel_name}
/api/rooms/{id}/panels
/api/rooms/{id}/cabinets/{cabinet_name}
/api/rooms/{id}/cabinets
/api/rooms/{id}/corridors/{corridor_name}
/api/rooms/{id}/corridors

/api/rooms/{id}/sensors/{room-sensor_name}
/api/rooms/{id}/sensors

/api/racks/{id}/sensors/{rack-sensor_name}
/api/racks/{id}/sensors

/api/racks/{id}/devices/{device_name}
/api/racks/{id}/devices

/api/devices/{id}/sensors/{sensor_name}
/api/devices/{id}/sensors
```

### Get object's hierarchy (non standard)
This returns an object's hierarchy in a non standard fashion    
and will be removed in the future
```
/api/sites/{site_name}/all/nonstd
/api/buildings/{id}/all/nonstd
/api/rooms/{id}/all/nonstd
/api/racks/{id}/all/nonstd
```