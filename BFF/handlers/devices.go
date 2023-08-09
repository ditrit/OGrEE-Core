package handlers

import (
	"ogree-bff/controllers"

	"github.com/gin-gonic/gin"
)

// swagger:operation GET /devices Devices Devices
// Get Devices list
//
// ---
// parameters:
//   - name: _key
//     in: query
//     description: Key of device
//     required: false
//     type: string
//   - name: _name
//     in: query
//     description: Name of device
//     required: false
//     type: string
//   - name: group_name
//     in: query
//     description: Group_name of device
//     required: false
//     type: string
//   - name: serial
//     in: query
//     description: Serial number of device
//     required: false
//     type: string
// security:
//   - Bearer: []
// responses:
//   '200':
//     description: successful
//     schema:
//       items:
//         "$ref": "#/definitions/SuccessResponse"
//   '500':
//     description: Error
//     schema:
//       items:
//         "$ref": "#/definitions/ErrorResponse"

func GetDevices(c *gin.Context) {
	controllers.Get(c,"arango")
}

// swagger:operation GET /devices/ConnecteTo/{device} Devices GetDevicesConnectedTo
// Get Devices connected to a device
//
// ---
// parameters:
//   - name: device
//     in: path
//     description: Key of device
//     required: true
//     type: string
//   - name: _key
//     in: query
//     description: Filter devices by key
//     required: false
//     type: string
//   - name: _name
//     in: query
//     description: Name of device
//     required: false
//     type: string
//   - name: group_name
//     in: query
//     description: Group_name of device
//     required: false
//     type: string
//   - name: serial
//     in: query
//     description: Serial number of device
//     required: false
//     type: string
// security:
//   - Bearer: []
// responses:
//   '200':
//     description: successful
//     schema:
//       items:
//         "$ref": "#/definitions/SuccessResponse"
//   '500':
//     description: Error
//     schema:
//       items:
//         "$ref": "#/definitions/ErrorResponse"
func GetDevicesConnectedTo(c *gin.Context) {
	controllers.Get(c,"arango")
}
// swagger:operation POST /devices Devices CreateDevices
// Create new Devices
//
// ---
// security:
//   - Bearer: []
// parameters:
//   - name: body
//     in: body
//     description: 'Mandatory: _name, group_name,created.'
//     required: true
//     format: object
//     example: '{"_name": "server", "group_name": "exwipen22","created": "2022-07-18"}'
// responses:
//   '200':
//     description: successful
//     schema:
//       items:
//         "$ref": "#/definitions/SuccessResponse"
//   '500':
//     description: Error
//     schema:
//       items:
//         "$ref": "#/definitions/ErrorResponse"
func CreateDevices(c *gin.Context) {
	controllers.Post(c,"arango")
}

// swagger:operation DELETE /devices/{device} Devices DeleteDevices
// Delete Devices by key
//
// ---
// parameters:
//   - name: device
//     in: path
//     description: device looking for
//     required: true
//     type: string
// security:
//   - Bearer: [] 
// responses:
//   '200':
//     description: successful
//     schema:
//       items:
//         "$ref": "#/definitions/SuccessResponse"
//   '500':
//     description: Error
//     schema:
//       items:
//         "$ref": "#/definitions/ErrorResponse"
func DeleteDevice(c *gin.Context){
	controllers.Delete(c,"arango")
}

// swagger:operation GET /Connections Devices GetConnections
// Get Connection list
//
// ---
// parameters:
//   - name: _key
//     in: query
//     description: Key of connection
//     required: false
//     type: string
//   - name: _from
//     in: query
//     description: From witch device
//     required: false
//     type: string
//   - name: _to
//     in: query
//     description: To witch device
//     required: false
//     type: string
// security:
//   - Bearer: []
// responses:
//   '200':
//     description: successful
//     schema:
//       items:
//         "$ref": "#/definitions/SuccessResponse"
//   '500':
//     description: Error
//     schema:
//       items:
//         "$ref": "#/definitions/ErrorResponse"
func GetConnections(c *gin.Context){
	controllers.Get(c,"arango")
}
// swagger:operation POST /Connections Devices CreateConnection
// Create new Connection
//
// ---
// security:
//   - Bearer: []
// parameters:
//   - name: body
//     in: body
//     description: 'Mandatory: _from, _to.'
//     required: true
//     format: object
//     example: '{"_from": "devices/123", "_to": "devices/111"}'
// responses:
//   '200':
//     description: successful
//     schema:
//       items:
//         "$ref": "#/definitions/SuccessResponse"
//   '500':
//     description: Error
//     schema:
//       items:
//         "$ref": "#/definitions/ErrorResponse"
func CreateConnections(c *gin.Context){
	controllers.Post(c,"arango")
}

// swagger:operation DELETE /Connections/{connection} Devices DeleteConnection
// Delete Connection by key
//
// ---
// security:
//   - Bearer: []
// parameters:
//   - name: connection
//     in: path
//     description: connection looking for
//     required: true
//     type: string
//
// responses:
//   '200':
//     description: successful
//     schema:
//       items:
//         "$ref": "#/definitions/SuccessResponse"
//   '500':
//     description: Error
//     schema:
//       items:
//         "$ref": "#/definitions/ErrorResponse"
func DeleteConnections(c *gin.Context){
	controllers.Delete(c,"arango")
}



// swagger:operation GET /devices/{obj}/objAttr/{objAttr}/deviceAttr/{deviceAttr} Devices GetDeviceBindingObject
// Get Devices list
//
// ---
// parameters:
//   - name: obj
//     in: path
//     description: object for binding
//     required: true
//     type: string
//   - name: objAttr
//     in: path
//     description: object attribute for binding
//     required: true
//     type: string
//   - name: deviceAttr
//     in: path
//     description: devices attribute for binding
//     required: true
//     type: string
//   - name: _key
//     in: query
//     description: Key of device
//     required: false
//     type: string
//   - name: _name
//     in: query
//     description: Name of device
//     required: false
//     type: string
//   - name: group_name
//     in: query
//     description: Group_name of device
//     required: false
//     type: string
//   - name: serial
//     in: query
//     description: Serial number of device
//     required: false
//     type: string
// security:
//   - Bearer: []
// responses:
//   '200':
//     description: successful
//     schema:
//       items:
//         "$ref": "#/definitions/SuccessResponse"
//   '500':
//     description: Error
//     schema:
//       items:
//         "$ref": "#/definitions/ErrorResponse"
func GetDeviceBindingObject(c*gin.Context){
	controllers.DeviceBindingObject(c,"arango")
}