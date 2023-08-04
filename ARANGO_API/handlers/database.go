package handlers

import (
	"net/http"

	"arango-api/database"
	"arango-api/models"

	"github.com/gin-gonic/gin"
)

// swagger:operation GET /Database Database Database
// Get database name
//
// ---
// responses:
//   '200':
//     description: successful
//     schema:
//       items:
//         "$ref": "#/definitions/ErrorResponse"
//   '500':
//     description: Error
//     schema:
//       items:
//         "$ref": "#/definitions/ErrorResponse"

func GetBDD(c *gin.Context) {
	db, err := database.GetDBConn(c)
	if err != nil {
		c.IndentedJSON(err.StatusCode, gin.H{"message": err.Message})
		return
	}
	addr, ok := c.Value("addr").(*string)
	if !ok {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Fail to get database hostname"})
		return
	}

	info, _ := (*db).Info(nil)

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Connected to " + (*addr) + " on database: " + info.Name})

}

// swagger:operation POST /Database Database ConnectBDD
// Connect to new bdd
//
// ---
// responses:
//
//	'200':
//	  description: successful
//	  schema:
//	    items:
//	      "$ref": "#/definitions/ErrorResponse"
//	'500':
//	  description: Error
//	  schema:
//	    items:
//	      "$ref": "#/definitions/ErrorResponse"
func ConnectBDD(c *gin.Context) {
	db, err := database.GetDBConn(c)
	if err != nil && err.StatusCode != 404 {
		c.IndentedJSON(err.StatusCode, gin.H{"message": err.Message})
		return
	}

	addr, ok := c.Value("addr").(*string)
	if !ok {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Fail to get database hostname"})
		return
	}

	var DBInfo models.DatabaseInfo
	if err := c.BindJSON(&DBInfo); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	newDB, err := database.ConnectToArrengo(DBInfo.Host, DBInfo.Database, DBInfo.User, DBInfo.Password)

	if err != nil {
		c.IndentedJSON(err.StatusCode, gin.H{"message": err.Message})
		return
	}

	(*db) = newDB
	(*addr) = DBInfo.Host
	info, _ := (*db).Info(nil)

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Connected to " + (*addr) + " on database: " + info.Name})

}
