package handlers

import (
	"ogree-bff/controllers"
	
	"github.com/gin-gonic/gin"
)

// swagger:operation GET /stats About GetStats
// Displays DB statistics.
// ---
// security:
//   - Bearer: []
// produces:
// 	 - application/json
// responses:
//		'200':
//			description: 'Request is valid.'
//		'504':
//			description: Server error.
func GetStats(c *gin.Context){
	controllers.Get(c,"mongo")
}

// swagger:operation GET /version About GetAPIVersion
// Gets the API version.
// ---
// security:
//   - Bearer: []
// produces:
// 	 - application/json
// responses:
//     '200':
//         description: 'OK. A response body will be returned with
//         version details.'
func GetAPIVersion(c *gin.Context){
	controllers.Get(c,"mongo")
}