# Exhaustive Endpoint List
This is an exhaustive list of endpoints for reference 


POST / Create
------------
Perform an HTTP POST operation with the appropriate JSON 
```
/api
/api/login
/api/tenants
/api/sites
/api/buildings
/api/rooms
/api/acs
/api/separators
/api/panels
/api/aisles
/api/tiles
/api/cabinets
/api/groups
/api/corridors
/api/room-sensors
/api/rack-sensors
/api/device-sensors
/api/racks
/api/devices
/api/room-templates
/api/obj-templates
```



DELETE / Delete
------------
Perform an HTTP DELETE operation without JSON body
```
/api/tenants/{id}
/api/sites/{id}
/api/sites
/api/buildings/{id}
/api/rooms/{id}
/api/acs/{id}
/api/panels/{id}
/api/separators/{id}
/api/aisles/{id}
/api/tiles/{id}
/api/cabinets/{id}
/api/groups/{id}
/api/corridors/{id}
/api/room-sensors/{id}
/api/rack-sensors/{id}
/api/device-sensors/{id}
/api/racks/{id}
/api/devices/{id}
/api/room-templates/{template_name}
/api/obj-templates/{template_name}
```


PUT / Update
-------------
Perform an HTTP PUT operation with desired JSON body
```
/api/tenants/{id}
/api/sites/{id}
/api/buildings/{id}
/api/rooms/{id}
/api/acs/{id}
/api/panels/{id}
/api/separators/{id}
/api/aisles/{id}
/api/tiles/{id}
/api/cabinets/{id}
/api/groups/{id}
/api/corridors/{id}
/api/room-sensors/{id}
/api/rack-sensors/{id}
/api/device-sensors/{id}
/api/racks/{id}
/api/devices/{id}
/api/room-templates/{template_name}
/api/obj-templates/{template_name}
```

PATCH / Update
-------------
Perform an HTTP PUT operation with desired JSON body
```
/api/tenants/{id}
/api/sites/{id}
/api/buildings/{id}
/api/rooms/{id}
/api/acs/{id}
/api/panels/{id}
/api/separators/{id}
/api/aisles/{id}
/api/tiles/{id}
/api/cabinets/{id}
/api/groups/{id}
/api/corridors/{id}
/api/room-sensors/{id}
/api/rack-sensors/{id}
/api/device-sensors/{id}
/api/racks/{id}
/api/devices/{id}
/api/room-templates/{template_name}
/api/obj-templates/{template_name}
```

GET / Get
-------------
Perform an HTTP PUT operation without JSON

### Quick Token Check
This URL is for development purposes only
```
/api/token/valid
```

### Get All Objects
```
/api/tenants
/api/sites
/api/buildings
/api/rooms
/api/racks
/api/devices
/api/room-templates
/api/obj-templates
/api/groups
/api/rack-sensors
/api/room-sensors
/api/device-sensors
/api/acs
/api/panels
/api/separators
/api/aisles
/api/tiles
/api/cabinets
/api/groups
/api/corridors
```

### Get by ID (non hierarchal)
ID is a long string    
Template_name is the 'slug'
```
/api/tenants/{id}
/api/sites/{id}
/api/buildings/{id}
/api/rooms/{id}
/api/racks/{id}
/api/devices/{id}
/api/room-templates/{template_name}
/api/obj-templates/{template_name}
/api/groups/{id}
/api/rack-sensors/{id}
/api/room-sensors/{id}
/api/device-sensors/{id}
/api/acs/{id}
/api/panels/{id}
/api/separators/{id}
/api/aisles/{id}
/api/tiles/{id}
/api/cabinets/{id}
/api/groups/{id}
/api/corridors/{id}
```

### Search Objects
Search using a query in the URL
Objects that match the query will be returned
Example: /devices?name=myValue?color=silver
will return silver devices with name 'myValue'
```
/api/acs?
/api/separators?
/api/panels?
/api/tenants?
/api/sites?
/api/buildings?
/api/rooms?
/api/racks?
/api/devices?
/api/aisles?
/api/tiles?
/api/cabinets?
/api/groups?
/api/corridors?
/api/room-templates?
/api/obj-templates?
/api/rack-sensors?
/api/room-sensors?
/api/device-sensors?
```


### Get all Objects 2 levels lower
```
/api/tenants/{tenant_name}/buildings
/api/sites/{id}/rooms
/api/buildings/{id}/racks
/api/rooms/{id}/devices
```

### Get an Object's entire hierarchy
The object and everything related to it will be returned    
in a nested JSON fashion
```
/api/tenants/{tenant_name}/all
/api/sites/{id}/all
/api/buildings/{id}/all
/api/rooms/{id}/all
/api/racks/{id}/all
/api/devices/{id}/all
```

### Get object's ranged hierarchy 
Limits the depth of the hierarchy to retrieve. This is observed by the    
URL given. 
```
/api/tenants/{tenant_name}/all/sites/buildings/rooms/racks/devices
/api/tenants/{tenant_name}/all/sites/buildings/rooms/racks
/api/tenants/{tenant_name}/all/sites/buildings/rooms
/api/tenants/{tenant_name}/all/sites/buildings

/api/sites/{id}/all/buildings/rooms/racks/devices
/api/sites/{id}/all/buildings/rooms/racks
/api/sites/{id}/all/buildings/rooms

/api/buildings/{id}/all/rooms/racks/devices
/api/buildings/{id}/all/rooms/racks

/api/rooms/{id}/all/racks/devices
```

### Get an Object's ranged hierarchy (using limit parameter)
The object and everything related until the number specified by limit will be returned    
in a nested JSON fashion
```
/api/tenants/{tenant_name}/all?limit={#}
/api/sites/{id}/all?limit={#}
/api/buildings/{id}/all?limit={#}
/api/rooms/{id}/all?limit={#}
/api/racks/{id}/all?limit={#}
/api/devices/{id}/all?limit={#}
```

### Get objects through the hierarchy
Returns an object if name given or all the objects immediately under the given URL
```
/api/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}
/api/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices
/api/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}
/api/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms/{room_name}/racks
/api/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms/{room_name}
/api/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms
/api/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}
/api/tenants/{tenant_name}/sites/{site_name}/buildings
/api/tenants/{tenant_name}/sites/{site_name}
/api/tenants/{tenant_name}/sites



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
/api/rooms/{id}/separators/{separator_name}
/api/rooms/{id}/separators
/api/rooms/{id}/aisles/{aisle_name}
/api/rooms/{id}/aisles
/api/rooms/{id}/tiles/{tile_name}
/api/rooms/{id}/tiles
/api/rooms/{id}/cabinets/{cabinet_name}
/api/rooms/{id}/cabinets
/api/rooms/{id}/corridors/{corridor_name}
/api/rooms/{id}/corridors


/api/racks/{id}/rack-sensors/{rack-sensor_name}
/api/racks/{id}/rack-sensors

/api/racks/{id}/devices/{device_name}
/api/racks/{id}/devices

/api/devices/{id}/device-sensors/{device-sensor_name}
/api/devices/{id}/device-sensors
```

### Get object's hierarchy (non standard)
This returns an object's hierarchy in a non standard fashion    
and will be removed in the future
```
/api/tenants/{tenant_name}/all/nonstd
/api/sites/{id}/all/nonstd
/api/buildings/{id}/all/nonstd
/api/rooms/{id}/all/nonstd
/api/racks/{id}/all/nonstd
```