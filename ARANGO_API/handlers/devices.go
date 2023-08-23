package handlers

import (
	"net/http"

	"arango-api/database"

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

	devices, err := database.GetAll(c, "devices")
	if err != nil {
		c.IndentedJSON(err.StatusCode, gin.H{"message": err.Message})
		return
	}
	if len(devices) == 0 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Devices not found"})
		return
	}
	c.IndentedJSON(http.StatusOK, devices)
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

	key := c.Param("key")
	devices, err := database.GetDevicesConnectedTo(c, key)
	if err != nil {
		c.IndentedJSON(err.StatusCode, gin.H{"message": err.Message})
		return
	}
	if len(devices) == 0 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Devices not found"})
		return
	}
	c.IndentedJSON(http.StatusOK, devices)
}

// swagger:operation POST /devices Devices CreateDevices
// Create new Devices
//
// ---
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
func PostDevices(c *gin.Context) {

	var newDevices map[string]string

	// Call BindJSON to bind the received JSON to
	if err := c.BindJSON(&newDevices); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	//Checking minimal configuration
	if newDevices["_name"] == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Device needs Name"})
		return
	}
	if newDevices["created"] == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Device needs created date"})
		return
	}
	if newDevices["group_name"] == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Device needs roup Name"})
		return
	}
	result, err := database.InsertDevices(c, newDevices)
	if err != nil {
		c.IndentedJSON(err.StatusCode, gin.H{"message": err.Message})
		return
	}
	c.IndentedJSON(http.StatusCreated, result)
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
func DeleteDevice(c *gin.Context) {
	key := c.Param("key")

	devices, err := database.Delete(c, key, "devices")
	if err != nil {
		c.IndentedJSON(err.StatusCode, gin.H{"message": err.Message})
		return
	}
	c.IndentedJSON(http.StatusOK, devices)
}
