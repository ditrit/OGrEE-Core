package handlers

import (
	"ogree-bff/controllers"

	"github.com/gin-gonic/gin"
)

// swagger:operation POST /{entity} Objects CreateObject
// Creates an object of the given entity in the system.
// ---
// security:
//   - Bearer: []
//
// produces:
//   - application/json
//
// parameters:
//   - name: entity
//     in: path
//     description: 'Entity (same as category) of the object. Accepted values: sites, domains,
//     buildings, rooms, racks, devices, acs, panels,
//     cabinets, groups, corridors,
//     room-templates, obj-templates, bldg-templates, stray-devices.'
//     required: true
//     type: string
//     default: "sites"
//   - name: body
//     in: body
//     required: true
//     default: {}
//
// responses:
//
//	'201':
//	    description: 'Created. A response body will be returned with
//	    a meaningful message.'
//	'400':
//	    description: 'Bad request. A response body with an error
//	    message will be returned.'
func CreateObject(c *gin.Context) {
	controllers.Post(c, "objects")
}

// swagger:operation GET /objects/{hierarchyName} Objects GetGenericObject
// Get an object from any entity.
// Gets an object from any of the physical entities with no need to specify it.
// The hierarchyName must be provided in the URL as a parameter.
// ---
// security:
//   - Bearer: []
//
// produces:
//   - application/json
//
// parameters:
//   - name: hierarchyName
//     in: path
//     description: hierarchyName of the object
//     required: true
//   - name: fieldOnly
//     in: query
//     description: 'specify which object field to show in response.
//     Multiple fieldOnly can be added. An invalid field is simply ignored.'
//   - name: startDate
//     in: query
//     description: 'filter objects by lastUpdated >= startDate.
//     Format: yyyy-mm-dd'
//   - name: endDate
//     in: query
//     description: 'filter objects by lastUpdated <= endDate.
//     Format: yyyy-mm-dd'
//
// responses:
//
//		'200':
//		    description: 'Found. A response body will be returned with
//	        a meaningful message.'
//		'404':
//		    description: Not Found. An error message will be returned.
func GetGenericObject(c *gin.Context) {
	controllers.Get(c, "objects")
}

// swagger:operation GET /{entity}/{IdOrHierarchyName} Objects GetEntity
// Gets an Object of the given entity.
// The ID or hierarchy name must be provided in the URL parameter.
// ---
// security:
//   - Bearer: []
//
// produces:
//   - application/json
//
// parameters:
//   - name: entity
//     in: path
//     description: 'Entity (same as category) of the object. Accepted values: sites, domains,
//     buildings, rooms, racks, devices, acs, panels,
//     cabinets, groups, corridors,
//     room-templates, obj-templates, bldg-templates, stray-devices.'
//     required: true
//     type: string
//     default: "sites"
//   - name: IdOrHierarchyName
//     in: path
//     description: 'ID or hierarchy name of desired object.
//     For templates the slug is the ID.'
//     required: true
//     type: string
//     default: "siteA"
//   - name: fieldOnly
//     in: query
//     description: 'specify which object field to show in response.
//     Multiple fieldOnly can be added. An invalid field is simply ignored.'
//   - name: startDate
//     in: query
//     description: 'filter objects by lastUpdated >= startDate.
//     Format: yyyy-mm-dd'
//   - name: endDate
//     in: query
//     description: 'filter objects by lastUpdated <= endDate.
//     Format: yyyy-mm-dd'
//
// responses:
//
//	'200':
//	  description: 'Found. A response body will be returned with
//	  a meaningful message.'
//	'400':
//	  description: Bad request. An error message will be returned.
//	'404':
//	  description: Not Found. An error message will be returned.
func GetEntity(c *gin.Context) {
	controllers.Get(c, "objects")
}

// swagger:operation GET /{entity} Objects GetAllEntities
// Gets all present objects for specified entity (category).
// Returns JSON body with all specified objects of type.
// ---
// security:
//   - Bearer: []
//
// produces:
//   - application/json
//
// parameters:
//   - name: entity
//     in: path
//     description: 'Entity (same as category) of the object. Accepted values: sites, domains,
//     buildings, rooms, racks, devices, acs, panels,
//     cabinets, groups, corridors,
//     room-templates, obj-templates, bldg-templates, stray-devices'
//     required: true
//     type: string
//     default: "sites"
//   - name: fieldOnly
//     in: query
//     description: 'specify which object field to show in response.
//     Multiple fieldOnly can be added. An invalid field is simply ignored.'
//   - name: startDate
//     in: query
//     description: 'filter objects by lastUpdated >= startDate.
//     Format: yyyy-mm-dd'
//   - name: endDate
//     in: query
//     description: 'filter objects by lastUpdated <= endDate.
//     Format: yyyy-mm-dd'
//
// responses:
//
//	'200':
//		description: 'Found. A response body will be returned with
//		a meaningful message.'
//	'404':
//		description: Nothing Found. An error message will be returned.
func GetAllEntities(c *gin.Context) {
	controllers.Get(c, "objects")
}

// swagger:operation DELETE /{entity}/{IdOrHierarchyName} Objects DeleteObject
// Deletes an Object in the system.
// ---
// security:
//   - Bearer: []
//
// produces:
//   - application/json
//
// parameters:
//   - name: entity
//     in: path
//     description: 'Entity (same as category) of the object. Accepted values: sites, domains,
//     buildings, rooms, racks, devices, acs, panels,
//     cabinets, groups, corridors,
//     room-templates, obj-templates, bldg-templates, stray-devices.'
//     required: true
//     type: string
//     default: "sites"
//   - name: IdOrHierarchyName
//     in: path
//     description: 'ID or hierarchy name of desired object.
//     For templates the slug is the ID.'
//     required: true
//     type: string
//     default: "siteA"
//
// responses:
//
//	'204':
//		description: 'Successfully deleted object.
//		No response body will be returned'
//	'404':
//		description: Not found. An error message will be returned
func DeleteObject(c *gin.Context) {
	controllers.Delete(c, "objects")
}

// swagger:operation PATCH /{entity}/{IdOrHierarchyName} Objects PartialUpdateObject
// Partially update object.
// This is the preferred method for modifying data in the system.
// If you want to do a full data replace, please use PUT instead.
// If no data is effectively changed, an OK will still be returned.
// ---
// security:
//   - Bearer: []
//
// produces:
//   - application/json
//
// parameters:
//   - name: entity
//     in: path
//     description: 'Entity (same as category) of the object. Accepted values: sites, domains,
//     buildings, rooms, racks, devices, acs, panels,
//     cabinets, groups, corridors,
//     room-templates, obj-templates, bldg-templates, stray-devices.'
//     required: true
//     type: string
//     default: "sites"
//   - name: IdOrHierarchyName
//     in: path
//     description: 'ID or hierarchy name of desired object.
//     For templates the slug is the ID.'
//     required: true
//     type: string
//     default: "siteA"
//
// responses:
//
//	'200':
//	    description: 'Updated. A response body will be returned with
//	    a meaningful message.'
//	'400':
//	    description: Bad request. An error message will be returned.
//	'404':
//	    description: Not Found. An error message will be returned.
func PartialUpdateObject(c *gin.Context) {
	controllers.Patch(c, "objects")
}

// swagger:operation PUT /{objs}/{IdOrHierarchyName} Objects UpdateObject
// Completely update object.
// This method will replace the existing data with the JSON
// received, thus fully replacing the data. If you do not
// want to do this, please use PATCH.
// If no data is effectively changed, an OK will still be returned.
// ---
// security:
//   - Bearer: []
//
// produces:
//   - application/json
//
// parameters:
//   - name: entity
//     in: path
//     description: 'Entity (same as category) of the object. Accepted values: sites, domains,
//     buildings, rooms, racks, devices, acs, panels,
//     cabinets, groups, corridors,
//     room-templates, obj-templates, bldg-templates, stray-devices.'
//     required: true
//     type: string
//     default: "sites"
//   - name: IdOrHierarchyName
//     in: path
//     description: 'ID or hierarchy name of desired object.
//     For templates the slug is the ID.'
//     required: true
//     type: string
//     default: "siteA"
//
// responses:
//
//	'200':
//	    description: 'Updated. A response body will be returned with
//	    a meaningful message.'
//	'400':
//	    description: Bad request. An error message will be returned.
//	'404':
//	    description: Not Found. An error message will be returned.
func UpdateObject(c *gin.Context) {
	controllers.Put(c, "objects")
}

// swagger:operation GET /{entity}? Objects GetEntityByQuery
// Gets an object filtering by attribute.
// Gets an Object using any attribute (with the exception of description)
// via query in the system
// The attributes are in the form {attr}=xyz&{attr1}=abc
// And any combination can be used given that at least 1 is provided.
// ---
// security:
//   - Bearer: []
//
// produces:
//   - application/json
//
// parameters:
//   - name: entity
//     in: path
//     description: 'Entity (same as category) of the object. Accepted values: sites, domains,
//     buildings, rooms, racks, devices, acs, panels,
//     cabinets, groups, corridors,
//     room-templates, obj-templates, bldg-templates, stray-devices.'
//     required: true
//     type: string
//     default: "sites"
//   - name: fieldOnly
//     in: query
//     description: 'specify which object field to show in response.
//     Multiple fieldOnly can be added. An invalid field is simply ignored.'
//   - name: startDate
//     in: query
//     description: 'filter objects by lastUpdated >= startDate.
//     Format: yyyy-mm-dd'
//   - name: endDate
//     in: query
//     description: 'filter objects by lastUpdated <= endDate.
//     Format: yyyy-mm-dd'
//   - name: attributes
//     in: query
//     description: 'Any other object attributes can be queried.
//     Replace attributes here by the name of the attribute followed by its value.'
//     required: false
//     type: string
//     default: domain=DemoDomain
//     example: vendor=ibm ; name=siteA ; orientation=front
//
// responses:
//
//	'204':
//		description: 'Found. A response body will be returned with
//		a meaningful message.'
//	'404':
//		description: Not found. An error message will be returned.
func GetEntityByQuery(c *gin.Context) {
	controllers.Get(c, "objects")
}

// swagger:operation GET /tempunits/{IdOrHierarchyName} Objects GetTempUnit
// Gets the temperatureUnit attribute of the parent site of given object.
// ---
// security:
//   - Bearer: []
//
// produces:
//   - application/json
//
// parameters:
//   - name: IdOrHierarchyName
//     in: path
//     description: 'ID or hierarchy name of desired object.
//     For templates the slug is the ID.'
//     required: true
//     type: string
//     default: "siteA"
//
// responses:
//
//	'200':
//	   description: 'Found. A response body will be returned with
//	   a meaningful message.'
//	'404':
//	   description: 'Nothing Found. An error message will be returned.'
func GetTempUnit(c *gin.Context) {
	controllers.Get(c, "objects")
}

// swagger:operation GET /{entity}/{id}/{subent} Objects GetEntitiesOfAncestor
// Obtain all objects 2 levels lower in the system.
// For Example: /api/sites/{id}/buildings
// Will return all buildings of a site
// Returns JSON body with all subobjects under the Object
// ---
// security:
//   - Bearer: []
//
// produces:
//   - application/json
//
// parameters:
//   - name: entity
//     in: query
//     description: 'Indicates the entity. Only values of "sites",
//     "buildings", "rooms" are acceptable'
//     required: true
//     type: string
//     default: "sites"
//   - name: ID
//     in: query
//     description: ID of object
//     required: true
//     type: int
//     default: 999
//   - name: subent
//     in: query
//     description: Objects which 2 are levels lower in the hierarchy.
//     required: true
//     type: string
//     default: buildings
//
// responses:
//
//	'200':
//	    description: 'Found. A response body will be returned with
//	    a meaningful message.'
//	'404':
//	    description: Nothing Found. An error message will be returned.
func GetEntitiesOfAncestor(c *gin.Context) {
	controllers.Get(c, "objects")
}

// swagger:operation GET /{entity}/{IdOrHierarchyName}/all Objects GetEntityHierarchy
// Get object and all its children.
// Returns JSON body with all subobjects under the Object.
// ---
// security:
//   - Bearer: []
//
// produces:
//   - application/json
//
// parameters:
//   - name: entity
//     in: path
//     description: 'Entity (same as category) of the object. Accepted values: sites, domains,
//     buildings, rooms, racks, devices, acs, panels,
//     cabinets, groups, corridors,
//     room-templates, obj-templates, bldg-templates, stray-devices.'
//     required: true
//     type: string
//     default: "sites"
//   - name: IdOrHierarchyName
//     in: path
//     description: 'ID or hierarchy name of desired object.
//     For templates the slug is the ID.'
//     required: true
//     type: string
//     default: "siteA"
//   - name: limit
//     in: query
//     description: 'Limits the level of hierarchy for retrieval. if not
//     specified for devices then the default value is maximum.
//     Example: /api/devices/{id}/all?limit=2'
//     required: false
//     type: string
//     default: 1
//   - name: fieldOnly
//     in: query
//     description: 'specify which object field to show in response.
//     Multiple fieldOnly can be added. An invalid field is simply ignored.'
//   - name: startDate
//     in: query
//     description: 'filter objects by lastUpdated >= startDate.
//     Format: yyyy-mm-dd'
//   - name: endDate
//     in: query
//     description: 'filter objects by lastUpdated <= endDate.
//     Format: yyyy-mm-dd'
//
// responses:
//
//	'200':
//	    description: 'Found. A response body will be returned with
//	    a meaningful message.'
//	'404':
//	    description: Nothing Found. An error message will be returned.
func GetEntityHierarchy(c *gin.Context) {
	controllers.Get(c, "objects")
}

// swagger:operation GET /hierarchy Objects GetCompleteHierarchy
// Returns system complete hierarchy.
// Return is arranged by relationship (father:[children])
// and category (category:[objects]), starting with "Root":[sites].
// User permissions apply.
// ---
// security:
//   - Bearer: []
// produces:
// 	 - application/json
// responses:
//		'200':
//			description: 'Request is valid.'
//		'500':
//			description: Server error.

func GetCompleteHierarchy(c *gin.Context) {
	controllers.Get(c, "objects")
}

// swagger:operation GET /hierarchy/attributes Objects GetCompleteHierarchyAttrs
// Returns attributes of all objects.
// Return is arranged by hierarchyName (objHierarchyName:{attributes}).
// User permissions apply.
// ---
// security:
//   - Bearer: []
//
// produces:
//   - application/json
//
// responses:
//
//	'200':
//		description: 'Request is valid.'
//	'500':
//		description: Server error.
func GetCompleteHierarchyAttrs(c *gin.Context) {
	controllers.Get(c, "objects")
}

// swagger:operation GET /{firstEntity}/{id}/{HierarchalPath} Objects GetEntitiesUsingNamesOfParents
// Get an object with its full hierarchal path.
// The path should begin with an entity name (firstEntity) and the ID of an object of this entity
// followed by a hierarchal path until the desired objet, that is,
// a sequence of entity names (category) and object names.
// ---
// security:
//   - Bearer: []
// produces:
// 	 - application/json
// parameters:
// - name: firstEntity
//   in: query
//   description: 'Root entity followed by an id. Can be: sites, buildings, rooms, racks or devices'
//   required: true
//   type: string
//   default: "sites"
// - name: id
//   in: path
//   description: 'id of object of firstEntity'
//   required: true
//   type: string
//   default: "123"
// - name: HierarchalPath
//   in: path
//   description: 'Hierarchal path to desired object.'
//   required: true
//   type: string
//   example: '/api/sites/{id}/buildings/{building_name} ;
//   /api/sites/{id}/buildings/{building_name}/rooms/{room_name} ;
//   /api/sites/{id}/buildings/{building_name}/rooms/{room_name}/acs/{ac_name} ;
//   /api/sites/{id}/buildings/{building_name}/rooms/{room_name}/corridors/{corridor_name} ;
//   /api/sites/{id}/buildings/{building_name}/rooms/{room_name}/panels/{panel_name} ;
//   /api/sites/{id}/buildings/{building_name}/rooms/{room_name}/groups/{group_name} ;
//   /api/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices/{device_name} ;
//   /api/sites/{id}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name} ;
//   /api/buildings/{id}/rooms/{room_name} ;
//   /api/buildings/{id}/rooms/{room_name}/acs/{ac_name} ;
//   /api/buildings/{id}/rooms/{room_name}/corridors/{corridor_name} ;
//   /api/buildings/{id}/rooms/{room_name}/panels/{panel_name} ;
//   /api/buildings/{id}/rooms/{room_name}/groups/{group_name} ;
//   /api/buildings/{id}/rooms/{room_name}/rack/{rack_name} ;
//   /api/buildings/{id}/rooms/{room_name}/racks/{rack_name}/devices/{device_name} ;
//   /api/rooms/{id}/acs/{ac_name} ;
//   /api/rooms/{id}/corridors/{corridor_name} ;
//   /api/rooms/{id}/panels/{panel_name} ;
//   /api/rooms/{id}/groups/{group_name} ;
//   /api/rooms/{id}/racks/{rack_name}/devices/{device_name} ;
//   /api/racks/{id}/devices/{device_name} ;
//   /api/devices/{id}/devices/{device_name} ;'
//   default: "/buildings/BuildingB/rooms/RoomA"
// responses:
//     '200':
//         description: 'Found. A response body will be returned with
//         a meaningful message.'
//     '404':
//         description: Not Found. An error message will be returned.

func GetEntitiesUsingNamesOfParents(c *gin.Context) {
	controllers.Get(c, "objects")
}

// swagger:operation POST /validate/{entity} Objects ValidateObject
// Checks the received data and verifies if the object can be created in the system.
// ---
// security:
//   - Bearer: []
//
// produces:
//   - application/json
//
// parameters:
//   - name: entity
//     in: path
//     description: 'Entity (same as category) of the object. Accepted values: sites, domains,
//     buildings, rooms, racks, devices, acs, panels,
//     cabinets, groups, corridors,
//     room-templates, obj-templates, bldg-templates, stray-devices.'
//     required: true
//     type: string
//     default: "sites"
//   - name: body
//     in: body
//     required: true
//     default: {}
//
// responses:
//
//	'200':
//	    description: 'Createable. A response body will be returned with
//	    a meaningful message.'
//	'400':
//	    description: 'Bad request. A response body with an error
//	    message will be returned.'
//	'404':
//	    description: Not Found. An error message will be returned.
func ValidateObject(c *gin.Context) {
	controllers.Post(c, "objects")
}
