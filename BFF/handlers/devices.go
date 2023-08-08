package handlers

import (
	"ogree-bff/controllers"

	"github.com/gin-gonic/gin"
)

// swagger:operation GET /Devices Devices Devices
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

// swagger:operation GET /Devices/ConnecteTo/{device} Devices GetDevicesConnectedTo
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