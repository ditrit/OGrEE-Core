package handlers

import (
	"net/http"

	"go-api/database"
	"github.com/gin-gonic/gin"
)

// swagger:operation GET /Connections Connections Connection
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
// responses:
//   '200':
//     description: successful
//     schema:
//       items:
//         "$ref": "#/definitions/SuccessConResponse"
//   '500':
//     description: Error
//     schema:
//       items:
//         "$ref": "#/definitions/ErrorResponse"
func GetConnection(c *gin.Context) {

	conn, err := database.GetAll(c,"links")
	if err != nil {
		c.IndentedJSON(err.StatusCode, gin.H{"message": err.Message})
		return
	}
	if len(conn) == 0 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "no connection found"})
		return
	}
	c.IndentedJSON(http.StatusOK, conn)
}


// swagger:operation POST /Connections Connections CreateConnection
// Create new Connection
//
// ---
// responses:
//   '200':
//     description: successful
//     schema:
//       items:
//         "$ref": "#/definitions/SuccessConResponse"
//   '500':
//     description: Error
//     schema:
//       items:
//         "$ref": "#/definitions/ErrorResponse"
func PostConnection(c *gin.Context) {
	
	var newConn map[string]string

	// Call BindJSON to bind the received JSON to
	if err := c.BindJSON(&newConn); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	result,err := database.InsertConnection(c, newConn);
	if err != nil {
		c.IndentedJSON(err.StatusCode, gin.H{"message": err.Message})
		return
	}
	c.IndentedJSON(http.StatusCreated, result)
}

// swagger:operation DELETE /Connections/{connection} Connections DeleteConnection
// Delete Connection by key
//
// ---
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
func DeleteConnection(c *gin.Context){
	key := c.Param("key")

	conn, err := database.Delete(c,key,"links")
	if err != nil {
		c.IndentedJSON(err.StatusCode, gin.H{"message": err.Message})
		return
	}
	c.IndentedJSON(http.StatusOK, conn)
}